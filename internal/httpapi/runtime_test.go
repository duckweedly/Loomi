package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestStartRunHandlerCreatesLocalSimulatedRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"script_name":"m4_smoke"}`)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Run productdata.Run `json:"run"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Run.ThreadID != thread.ID || body.Run.Source != productdata.RunSourceLocalSimulated || body.Run.Status != productdata.RunStatusRunning {
		t.Fatalf("run = %+v", body.Run)
	}
}

func TestRunEventHistoryHandlerReturnsOrderedEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryProgress, "context_loaded")
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events?after_sequence=1", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Events []productdata.RunEvent `json:"events"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Events) != 2 || body.Events[0].Sequence != 2 || body.Events[1].Sequence != 3 {
		t.Fatalf("events = %+v", body.Events)
	}
}

func TestRunEventStreamDeliversHistoryBeforeCloseMarker(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryProgress, "context_loaded")
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events/stream", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	if !strings.Contains(res.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("content-type = %q", res.Header().Get("Content-Type"))
	}
	if strings.Index(body, "run_created") > strings.Index(body, "context_loaded") {
		t.Fatalf("history order body=%s", body)
	}
	if !strings.Contains(body, "event: stream_closed") {
		t.Fatalf("missing close marker body=%s", body)
	}
}

func TestRunEventStreamReconnectUsesAfterSequence(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryProgress, "context_loaded")
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events/stream?after_sequence=1", "")

	body := res.Body.String()
	if strings.Contains(body, "run_created") || !strings.Contains(body, "context_loaded") {
		t.Fatalf("reconnect body=%s", body)
	}
}

func TestRunEventStreamSubscribesBeforeHistoryRead(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	broadcaster := productruntime.NewBroadcaster()
	srv := NewServerWithRuntime(config.Config{AppEnv: "local"}, fakeChecker{}, publishDuringListService{Service: svc, broadcaster: broadcaster}, broadcaster, nil)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events/stream", "")

	body := res.Body.String()
	if !strings.Contains(body, "context_loaded") || !strings.Contains(body, "stream_closed") {
		t.Fatalf("body = %s", body)
	}
}

func TestRunEventStreamFlushesHistoryAndCloseMarker(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	res := newFlushRecorder()
	srv.ServeHTTP(res, httptestRequest(http.MethodGet, "/v1/runs/"+run.ID+"/events/stream", ""))

	if res.flushes == 0 {
		t.Fatal("flushes = 0, want at least one flush for history/close marker")
	}
}

func TestStopRunHandlerPublishesStopEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	broadcaster := productruntime.NewBroadcaster()
	srv := NewServerWithRuntime(config.Config{AppEnv: "local"}, fakeChecker{}, svc, broadcaster, nil)
	events := broadcaster.Subscribe(context.Background(), run.ID)

	res := requestJSON(t, srv, http.MethodPost, "/v1/runs/"+run.ID+"/stop", "")
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	first := <-events
	second := <-events
	if first.Type != "run_stopped" || second.Category != productdata.RunEventCategoryFinal {
		t.Fatalf("published events = %+v %+v", first, second)
	}
}

func TestStopRunHandlerReturnsAlreadyTerminalForCompletedRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/runs/"+run.ID+"/stop", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "already_terminal") {
		t.Fatalf("body = %s", res.Body.String())
	}
}

type publishDuringListService struct {
	productdata.Service
	broadcaster *productruntime.Broadcaster
	published   bool
}

func (s publishDuringListService) ListRunEvents(ctx context.Context, ident identity.LocalIdentity, runID string, afterSequence int) ([]productdata.RunEvent, error) {
	events, err := s.Service.ListRunEvents(ctx, ident, runID, afterSequence)
	if err != nil || s.published {
		return events, err
	}
	event, err := s.Service.AppendRunEvent(ctx, ident, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: "context_loaded", Summary: "Context loaded"})
	if err != nil {
		return events, err
	}
	final, err := s.Service.AppendRunEvent(ctx, ident, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"})
	if err != nil {
		return events, err
	}
	s.broadcaster.Publish(event)
	s.broadcaster.Publish(final)
	return events, nil
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushes int
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func (r *flushRecorder) Flush() {
	r.flushes++
}

func httptestRequest(method string, path string, body string) *http.Request {
	if body == "" {
		return httptest.NewRequest(method, path, nil)
	}
	return httptest.NewRequest(method, path, bytes.NewBufferString(body))
}

func createRuntimeTestThread(t *testing.T, svc *productdata.MemoryService) productdata.Thread {
	t.Helper()
	thread, err := svc.CreateThread(context.Background(), identity.LocalDevIdentity(), productdata.CreateThreadInput{Title: "Runtime", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	return thread
}

func createRuntimeTestRun(t *testing.T, svc *productdata.MemoryService) productdata.Run {
	t.Helper()
	thread := createRuntimeTestThread(t, svc)
	run, err := svc.StartRun(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.StartRunInput{ScriptName: "m4_smoke"})
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func appendRuntimeTestEvent(t *testing.T, svc *productdata.MemoryService, runID string, category productdata.RunEventCategory, eventType string) productdata.RunEvent {
	t.Helper()
	event, err := svc.AppendRunEvent(context.Background(), identity.LocalDevIdentity(), runID, productdata.AppendRunEventInput{Category: category, Type: eventType, Summary: eventType})
	if err != nil {
		t.Fatal(err)
	}
	return event
}
