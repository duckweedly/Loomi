package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunDoctorReportsCompletionFailed503(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/readyz":
			w.WriteHeader(http.StatusOK)
		case "/v1/model-providers":
			writeCLITestJSON(w, `{"providers":[{"id":"custom","family":"openai_compatible","model":"gpt-5.5","status":"configured","execution_state":"supported"}]}`)
		case "/v1/model-providers/check":
			writeCLITestJSON(w, `{"provider":{"id":"custom","family":"openai_compatible","model":"gpt-5.5","status":"completion-failed","check_code":"completion-failed-503","http_status":503,"message":"Provider completion check failed with HTTP 503.","execution_state":"supported"}}`)
		case "/v1/tools/catalog":
			writeCLITestJSON(w, `{"tools":[{"name":"runtime.get_current_time","group":"runtime","enabled":true}]}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	report := RunDoctor(context.Background(), NewClient(server.URL), Config{Provider: "custom"})

	var provider DoctorCheck
	for _, check := range report.Checks {
		if check.Name == "providers" {
			provider = check
		}
	}
	if provider.Status != "warn" || !strings.Contains(provider.Detail, "completion-failed-503") || !strings.Contains(provider.Detail, "HTTP 503") {
		t.Fatalf("provider check = %+v", provider)
	}
}

func writeCLITestJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}
