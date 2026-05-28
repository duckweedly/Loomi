package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type fakeChecker struct {
	pingErr   error
	schemaErr error
}

func (f fakeChecker) Ping(context.Context) error        { return f.pingErr }
func (f fakeChecker) SchemaReady(context.Context) error { return f.schemaErr }

type deadlineChecker struct {
	deadlineSeen bool
}

func (c *deadlineChecker) Ping(ctx context.Context) error {
	_, ok := ctx.Deadline()
	c.deadlineSeen = ok
	return nil
}

func (c *deadlineChecker) SchemaReady(context.Context) error { return nil }

func TestHealthz(t *testing.T) {
	srv := NewServer(config.Config{AppEnv: "local"}, fakeChecker{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "alive" || body["service"] != "loomi-api" || body["request_id"] == "" {
		t.Fatalf("body = %+v", body)
	}
}

func TestReadyzReady(t *testing.T) {
	srv := NewServer(config.Config{AppEnv: "local"}, fakeChecker{})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
}

func TestReadyzIncludesLocalDevCORSHeaders(t *testing.T) {
	srv := NewServer(config.Config{AppEnv: "local"}, fakeChecker{})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5180")
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	if res.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:5180" {
		t.Fatalf("allow origin = %q", res.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestReadyzNotReady(t *testing.T) {
	srv := NewServer(config.Config{AppEnv: "local", ReadinessTimeoutSeconds: 5}, fakeChecker{pingErr: errors.New("secret dial error")})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", res.Code)
	}
	if contains(res.Body.String(), "secret") {
		t.Fatalf("body leaked secret: %s", res.Body.String())
	}
}

func TestWorkerQueueDiagnosticsReadyPausedStoppedAndDegraded(t *testing.T) {
	readyService := workerQueueDiagnosticsService{status: productdata.WorkerQueueDiagnostics{QueueStatus: productdata.WorkerQueueStatusReady, WorkerStatus: productdata.WorkerStatusReady, QueuedCount: 2, UpdatedAt: time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)}}
	ready := requestWorkerQueueDiagnostics(t, NewServerWithProduct(config.Config{AppEnv: "local", WorkerQueueEnabled: true}, fakeChecker{}, readyService))
	if ready["queue_status"] != string(productdata.WorkerQueueStatusReady) || ready["worker_status"] != string(productdata.WorkerStatusReady) || ready["queued_count"].(float64) != 2 {
		t.Fatalf("ready diagnostics = %+v", ready)
	}
	paused := requestWorkerQueueDiagnostics(t, NewServerWithProduct(config.Config{AppEnv: "local", WorkerQueueEnabled: true, WorkerQueuePaused: true}, fakeChecker{}, readyService))
	if paused["queue_status"] != string(productdata.WorkerQueueStatusPaused) || paused["worker_status"] != string(productdata.WorkerStatusPaused) {
		t.Fatalf("paused diagnostics = %+v", paused)
	}
	stopped := requestWorkerQueueDiagnostics(t, NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, readyService))
	if stopped["worker_status"] != string(productdata.WorkerStatusStopped) {
		t.Fatalf("stopped diagnostics = %+v", stopped)
	}
	degradedService := workerQueueDiagnosticsService{status: productdata.WorkerQueueDiagnostics{QueueStatus: productdata.WorkerQueueStatusDegraded, WorkerStatus: productdata.WorkerStatusDegraded, StaleCount: 1, RetryingCount: 1, DeadCount: 1, UpdatedAt: time.Date(2026, 5, 24, 10, 1, 0, 0, time.UTC)}}
	degraded := requestWorkerQueueDiagnostics(t, NewServerWithProduct(config.Config{AppEnv: "local", WorkerQueueEnabled: true}, fakeChecker{}, degradedService))
	if degraded["queue_status"] != string(productdata.WorkerQueueStatusDegraded) || degraded["stale_count"].(float64) != 1 || degraded["dead_count"].(float64) != 1 {
		t.Fatalf("degraded diagnostics = %+v", degraded)
	}
}

type workerQueueDiagnosticsService struct {
	productdata.Service
	status productdata.WorkerQueueDiagnostics
}

func (s workerQueueDiagnosticsService) WorkerQueueDiagnostics(context.Context, identity.LocalIdentity) (productdata.WorkerQueueDiagnostics, error) {
	return s.status, nil
}

func requestWorkerQueueDiagnostics(t *testing.T, srv *Server) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/v1/diagnostics/worker-queue", nil)
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	diagnostics, ok := body["diagnostics"].(map[string]any)
	if !ok {
		t.Fatalf("body = %+v", body)
	}
	return diagnostics
}

func TestReadyzUsesConfiguredTimeout(t *testing.T) {
	checker := &deadlineChecker{}
	srv := NewServer(config.Config{AppEnv: "local", ReadinessTimeoutSeconds: 1}, checker)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	if !checker.deadlineSeen {
		t.Fatal("readyz checker did not receive a context deadline")
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
