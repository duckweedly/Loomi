package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRunDoctorReportsDatabaseAndSchemaReadinessBlockers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/readyz":
			w.WriteHeader(http.StatusServiceUnavailable)
			writeCLITestJSON(w, `{"status":"not_ready","checks":[{"name":"database","status":"failed","reason":"database ping failed"},{"name":"schema","status":"failed","reason":"schema version unavailable"}]}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	report := RunDoctor(context.Background(), NewClient(server.URL), Config{Provider: "custom"})

	if report.OK {
		t.Fatal("expected doctor report to fail")
	}
	var database, schema DoctorCheck
	for _, check := range report.Checks {
		if check.Name == "database" {
			database = check
		}
		if check.Name == "schema" {
			schema = check
		}
	}
	if database.Status != "fail" || !strings.Contains(database.Detail, "database ping failed") {
		t.Fatalf("database check = %+v", database)
	}
	if schema.Status != "fail" || !strings.Contains(schema.Detail, "schema version unavailable") || !strings.Contains(schema.Remedy, "migrate") {
		t.Fatalf("schema check = %+v", schema)
	}
}

func TestRunDoctorReportsWebBaseURLMismatch(t *testing.T) {
	t.Setenv("VITE_LOOMI_API_BASE_URL", "http://127.0.0.1:5173")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/readyz":
			writeCLITestJSON(w, `{"status":"ready","checks":[{"name":"database","status":"ok"},{"name":"schema","status":"ok"}]}`)
		case "/v1/model-providers":
			writeCLITestJSON(w, `{"providers":[{"id":"custom","family":"openai_compatible","model":"gpt-5.5","status":"available","execution_state":"supported"}]}`)
		case "/v1/tools/catalog":
			writeCLITestJSON(w, `{"tools":[{"name":"workspace.read","group":"workspace","enabled":true,"execution_state":"executable"}]}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	report := RunDoctor(context.Background(), NewClient(server.URL), Config{Host: server.URL, Provider: "custom"})

	var web DoctorCheck
	for _, check := range report.Checks {
		if check.Name == "web" {
			web = check
		}
	}
	if web.Status != "warn" || !strings.Contains(web.Detail, "VITE_LOOMI_API_BASE_URL") || strings.Contains(web.Detail, os.Getenv("HOME")) {
		t.Fatalf("web check = %+v", web)
	}
}

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
	if provider.Status != "warn" || !strings.Contains(provider.Detail, "completion-failed-503") || !strings.Contains(provider.Detail, "http=503") {
		t.Fatalf("provider check = %+v", provider)
	}
}

func TestRunDesktopDoctorReportsWorkspaceMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/readyz":
			writeCLITestJSON(w, `{"status":"ready","checks":[{"name":"schema","status":"ok"}]}`)
		case "/v1/model-providers":
			writeCLITestJSON(w, `{"providers":[{"id":"custom","family":"openai_compatible","model":"gpt-5.5","status":"available","execution_state":"supported"}]}`)
		case "/v1/tools/catalog":
			writeCLITestJSON(w, `{"tools":[{"name":"workspace.read","group":"workspace","enabled":true,"execution_state":"executable"}]}`)
		case "/v1/workspace/root":
			writeCLITestJSON(w, `{"config":{"configured":false,"display_name":"No folder selected"}}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	report := RunDesktopDoctor(context.Background(), NewClient(server.URL), Config{Provider: "custom"})

	var workspace DoctorCheck
	for _, check := range report.Checks {
		if check.Name == "workspace" {
			workspace = check
		}
	}
	if workspace.Status != "warn" || !strings.Contains(workspace.Detail, "not selected") || !strings.Contains(workspace.Remedy, "choose a workspace") {
		t.Fatalf("workspace check = %+v", workspace)
	}
}

func writeCLITestJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}
