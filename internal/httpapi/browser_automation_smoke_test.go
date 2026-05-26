package httpapi

import (
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

func TestM27BrowserOpenApproveExecuteFinalSmoke(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><title>M27 Browser</title><body>Open smoke</body></html>"))
	}))
	defer target.Close()
	const toolCallID = "tc_m27_browser_open"

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := syncM27BrowserPersona(svc, ident); err != nil {
		t.Fatal(err)
	}
	provider := &m21WorkspaceProvider{toolName: productdata.ToolNameBrowserOpen, toolCallID: toolCallID, args: map[string]any{"url": target.URL, "max_bytes": 4096}}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway, BrowserExecutor: productruntime.BrowserToolExecutor{Store: productruntime.NewBrowserSessionStore(), AllowPrivateHosts: true}})
	worker.WorkerID = "worker_m27_browser_open"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadID, runID := startM27BrowserRun(t, srv)
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolBlocked(t, svc, threadID, runID, toolCallID, productdata.ToolNameBrowserOpen)
	approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
	assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("second ProcessOne ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "open" || call.ResultSummary["scope"] != "browser" || call.ResultSummary["title"] != "M27 Browser" {
		t.Fatalf("call = %+v", call)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{productdata.EventToolCallApprovalRequired, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, `"tool_name":"browser.open"`, `"scope":"browser"`} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m27 browser open events", "<html>", "Set-Cookie", "Authorization", "/Users/")
}

func TestM27BrowserOpenClickSnapshotLoopSmoke(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><title>M27 Home</title><body>Home <a href="/docs">Docs</a></body></html>`))
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><title>M27 Docs</title><body>Docs page</body></html>`))
	})
	target := httptest.NewServer(mux)
	defer target.Close()

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := syncM27BrowserPersona(svc, ident); err != nil {
		t.Fatal(err)
	}
	provider := &m27BrowserLoopProvider{url: target.URL}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway, BrowserExecutor: productruntime.BrowserToolExecutor{Store: productruntime.NewBrowserSessionStore(), AllowPrivateHosts: true}})
	worker.WorkerID = "worker_m27_browser_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadID, runID := startM27BrowserRun(t, srv)
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("open request ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_browser_open_1", productdata.ToolNameBrowserOpen)
	approveOpen := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/tc_browser_open_1/approve", "")
	assertStatus(t, approveOpen.Code, http.StatusOK, approveOpen.Body.String())

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("open execution ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_browser_click_2", productdata.ToolNameBrowserClickLink)
	approveClick := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/tc_browser_click_2/approve", "")
	assertStatus(t, approveClick.Code, http.StatusOK, approveClick.Body.String())

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("click execution ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_browser_snapshot_3", productdata.ToolNameBrowserSnapshot)
	approveSnapshot := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/tc_browser_snapshot_3/approve", "")
	assertStatus(t, approveSnapshot.Code, http.StatusOK, approveSnapshot.Body.String())

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("snapshot execution ProcessOne ok=%v err=%v", ok, err)
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted || provider.calls != 4 {
		t.Fatalf("run=%+v provider calls=%d", run, provider.calls)
	}
	snapshotCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_browser_snapshot_3")
	if err != nil {
		t.Fatal(err)
	}
	if snapshotCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || snapshotCall.ResultSummary["operation"] != "snapshot" || snapshotCall.ResultSummary["title"] != "M27 Docs" {
		t.Fatalf("snapshot call = %+v", snapshotCall)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{`"tool_name":"browser.open"`, `"tool_name":"browser.click_link"`, `"tool_name":"browser.snapshot"`, `"loop_index":3`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m27 browser loop events", "<html>", "Set-Cookie", "Authorization", "/Users/")
}

func syncM27BrowserPersona(svc *productdata.MemoryService, ident identity.LocalIdentity) (productdata.PersonaSyncResult, error) {
	return svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved browser tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameBrowserOpen, productdata.ToolNameBrowserClickLink, productdata.ToolNameBrowserSnapshot},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}})
}

func startM27BrowserRun(t *testing.T, srv http.Handler) (string, string) {
	t.Helper()
	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M27 browser smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Browse docs","client_message_id":"m27-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	return threadID, decodeStringField(t, runRes.Body.Bytes(), "run", "id")
}

type m27BrowserLoopProvider struct {
	url       string
	calls     int
	sessionID string
}

func (p *m27BrowserLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m27BrowserLoopProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{}
	if p.calls > 1 && p.sessionID == "" {
		p.sessionID = sessionIDFromProviderRequest(request)
	}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameBrowserOpen, Metadata: map[string]any{"tool_call_id": "tc_browser_open_1", "arguments_summary": map[string]any{"url": p.url, "max_bytes": 4096}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameBrowserClickLink, Metadata: map[string]any{"tool_call_id": "tc_browser_click_2", "arguments_summary": map[string]any{"session_id": p.sessionID, "link_index": 0}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameBrowserSnapshot, Metadata: map[string]any{"tool_call_id": "tc_browser_snapshot_3", "arguments_summary": map[string]any{"session_id": p.sessionID}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M27 browser "}, {Type: productruntime.ProviderEventCompleted, Text: "M27 browser complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func sessionIDFromProviderRequest(request productruntime.ProviderRequest) string {
	for _, message := range request.Messages {
		if message.Role != productruntime.ProviderMessageRoleToolResult {
			continue
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(message.Content), &result); err != nil {
			continue
		}
		if id, _ := result["session_id"].(string); id != "" {
			return id
		}
	}
	return ""
}
