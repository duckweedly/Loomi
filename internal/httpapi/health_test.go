package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
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
