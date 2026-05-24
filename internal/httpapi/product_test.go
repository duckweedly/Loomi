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
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestGetMeReturnsLocalUser(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/v1/me", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	user := body["user"].(map[string]any)
	if user["id"] != "user_local_dev" || body["request_id"] == "" {
		t.Fatalf("body = %+v", body)
	}
}

func TestThreadHandlers(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	create := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"Thread","mode":"chat"}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", create.Code, create.Body.String())
	}
	threadID := decodeThreadID(t, create.Body.Bytes())

	list := requestJSON(t, srv, http.MethodGet, "/v1/threads", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), threadID) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}

	patch := requestJSON(t, srv, http.MethodPatch, "/v1/threads/"+threadID, `{"title":"Renamed","mode":"work"}`)
	if patch.Code != http.StatusOK || !strings.Contains(patch.Body.String(), "Renamed") {
		t.Fatalf("patch status=%d body=%s", patch.Code, patch.Body.String())
	}

	archive := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/archive", "")
	if archive.Code != http.StatusOK || !strings.Contains(archive.Body.String(), "archived") {
		t.Fatalf("archive status=%d body=%s", archive.Code, archive.Body.String())
	}
}

func TestMessageHandlers(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	create := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"Thread","mode":"chat"}`)
	threadID := decodeThreadID(t, create.Body.Bytes())

	msg := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Hello","client_message_id":"client-1"}`)
	if msg.Code != http.StatusCreated || strings.Contains(msg.Body.String(), "assistant") {
		t.Fatalf("message status=%d body=%s", msg.Code, msg.Body.String())
	}
	messageID := decodeMessageID(t, msg.Body.Bytes())
	dup := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Hello again","client_message_id":"client-1"}`)
	if dup.Code != http.StatusOK || !strings.Contains(dup.Body.String(), messageID) {
		t.Fatalf("dup status=%d body=%s", dup.Code, dup.Body.String())
	}
	list := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+threadID+"/messages", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), messageID) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
}

func TestAPIPreflightAllowsBrowserWrites(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	for _, origin := range []string{"http://127.0.0.1:5173", "http://localhost:5173"} {
		t.Run(origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/v1/threads", nil)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			req.Header.Set("Access-Control-Request-Headers", "content-type")
			res := httptest.NewRecorder()

			srv.ServeHTTP(res, req)

			if res.Code != http.StatusNoContent {
				t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
			}
			if res.Header().Get("Access-Control-Allow-Origin") != origin {
				t.Fatalf("allow origin = %q", res.Header().Get("Access-Control-Allow-Origin"))
			}
			if res.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PATCH, OPTIONS" {
				t.Fatalf("allow methods = %q", res.Header().Get("Access-Control-Allow-Methods"))
			}
			if res.Header().Get("Access-Control-Allow-Headers") != "Content-Type" {
				t.Fatalf("allow headers = %q", res.Header().Get("Access-Control-Allow-Headers"))
			}
			if res.Header().Get("Access-Control-Allow-Credentials") != "" {
				t.Fatalf("allow credentials = %q", res.Header().Get("Access-Control-Allow-Credentials"))
			}
		})
	}
}

func TestAPIPreflightRejectsNonLocalOrigin(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	for _, origin := range []string{"https://example.com", "http://127.0.0.1:5174", "http://localhost:3000"} {
		t.Run(origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/v1/threads", nil)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			res := httptest.NewRecorder()

			srv.ServeHTTP(res, req)

			if res.Code != http.StatusNoContent {
				t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
			}
			if res.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Fatalf("allow origin = %q", res.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestAPICORSHeadersOnNormalRequest(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	req := httptest.NewRequest(http.MethodGet, "/v1/threads", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if res.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("allow origin = %q", res.Header().Get("Access-Control-Allow-Origin"))
	}
	if res.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("allow credentials = %q", res.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestAPICORSDisabledOutsideLocalDev(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "test"}, fakeChecker{}, productdata.NewMemoryService())
	req := httptest.NewRequest(http.MethodOptions, "/v1/threads", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	res := httptest.NewRecorder()

	srv.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if res.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("allow origin = %q", res.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestUnsupportedThreadMethodReturnsMethodNotAllowed(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	res := requestJSON(t, srv, http.MethodDelete, "/v1/threads", "")
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if res.Header().Get("Allow") != "GET, POST" {
		t.Fatalf("allow = %q", res.Header().Get("Allow"))
	}
	if !strings.Contains(res.Body.String(), "method_not_allowed") {
		t.Fatalf("body = %s", res.Body.String())
	}
}

func TestCreateThreadRejectsUnknownJSONFields(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	res := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"Thread","mode":"chat","extra":true}`)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "invalid_request") {
		t.Fatalf("body = %s", res.Body.String())
	}
}

func TestEmptyThreadPatchIsRejected(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	create := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"Thread","mode":"chat"}`)
	threadID := decodeThreadID(t, create.Body.Bytes())

	patch := requestJSON(t, srv, http.MethodPatch, "/v1/threads/"+threadID, `{}`)

	if patch.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", patch.Code, patch.Body.String())
	}
	if !strings.Contains(patch.Body.String(), "invalid_request") {
		t.Fatalf("body = %s", patch.Body.String())
	}
}

func TestProductRoutesReturnStructuredErrorWhenServiceUnavailable(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, nil)
	res := requestJSON(t, srv, http.MethodGet, "/v1/me", "")
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "internal_error") || !strings.Contains(res.Body.String(), "request_id") {
		t.Fatalf("body = %s", res.Body.String())
	}
}

func TestStructuredErrors(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	res := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":" ","mode":"chat"}`)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "invalid_request") || !strings.Contains(res.Body.String(), "request_id") {
		t.Fatalf("body = %s", res.Body.String())
	}
	if strings.Contains(res.Body.String(), "postgres://") || strings.Contains(res.Body.String(), "secret") {
		t.Fatalf("body leaked secret: %s", res.Body.String())
	}
}

func requestJSON(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		r.Header.Set("Content-Type", "application/json")
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, r)
	return res
}

func decodeThreadID(t *testing.T, raw []byte) string {
	t.Helper()
	var body struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body.Thread.ID
}

func decodeMessageID(t *testing.T, raw []byte) string {
	t.Helper()
	var body struct {
		Message struct {
			ID string `json:"id"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body.Message.ID
}

var _ = context.Background
