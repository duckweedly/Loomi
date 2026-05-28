package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunToolsListCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/tools/catalog" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tools":[{"name":"workspace.read","group":"workspace","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true},{"name":"runtime.get_current_time","group":"runtime","risk_level":"low","approval_policy":"read_only","execution_state":"executable","enabled":true},{"name":"browser.open","group":"browser","risk_level":"medium","approval_policy":"always_required","execution_state":"disabled","enabled":false}],"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"tools", "list", "--host", server.URL, "--group", "workspace", "--enabled-only"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"[workspace]", "workspace.read", "always_required", "low"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if strings.Contains(output, "browser.open") || strings.Contains(output, "runtime.get_current_time") {
		t.Fatalf("stdout contains filtered tool: %s", output)
	}
}

func TestRunMCPServersCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/mcp/servers" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"servers":[{"server_safe_id":"mcp:local-smoke","server_slug":"local-smoke","display_name":"Local Smoke","transport":"stdio","enabled":true,"config_source":"local","discovery_status":"succeeded","candidate_count":1,"candidate_names":["mcp.local-smoke.echo"],"execution_mode":"approval_gated"}],"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"mcp", "servers", "--host", server.URL}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"local-smoke", "Local Smoke", "discovery=succeeded", "candidates=1", "execution=approval_gated", "mcp.local-smoke.echo"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestRunLSPToolsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/tools/catalog" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tools":[{"name":"workspace.read","group":"workspace","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true},{"name":"lsp.symbols","group":"lsp","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true},{"name":"lsp.hover","group":"lsp","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true}],"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"lsp", "tools", "--host", server.URL}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"[lsp]", "lsp.symbols", "lsp.hover", "always_required"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if strings.Contains(output, "workspace.read") {
		t.Fatalf("stdout contains non-lsp tool: %s", output)
	}
}

func TestArtifactsListCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/threads/thr/artifacts" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Fatalf("limit = %q", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"artifacts":[{"id":"art_1","thread_id":"thr","run_id":"run","title":"Notes","artifact_type":"text","content_bytes":42,"text_excerpt":"hello artifact","truncated":false,"created_at":"2026-05-26T00:00:00Z","updated_at":"2026-05-26T00:00:00Z"}],"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"artifacts", "list", "--host", server.URL, "--limit", "5", "thr"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"art_1", "text", "Notes", "bytes=42", "hello artifact"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestArtifactsReadCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/threads/thr/artifacts/art_1" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("max_bytes") != "7" {
			t.Fatalf("max_bytes = %q", r.URL.Query().Get("max_bytes"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"artifact":{"id":"art_1","thread_id":"thr","run_id":"run","title":"Notes","artifact_type":"text","content_bytes":42,"text_excerpt":"hello a","truncated":true,"created_at":"2026-05-26T00:00:00Z","updated_at":"2026-05-26T00:00:00Z"},"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"artifacts", "read", "--host", server.URL, "--max-bytes", "7", "thr", "art_1"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"artifact art_1 text", "thread thr", "run run", "title Notes", "truncated true", "hello a"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestMemoryListCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/memory" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("scope_type") != "thread" || r.URL.Query().Get("scope_id") != "thr" || r.URL.Query().Get("limit") != "3" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[{"id":"mem_1","title":"Preference","summary":"likes concise output","scope_type":"thread","scope_id":"thr","status":"approved","safety_state":"safe","source_type":"thread","redaction_applied":false}],"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"memory", "list", "--host", server.URL, "--scope-type", "thread", "--scope-id", "thr", "--limit", "3"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"mem_1", "approved", "safe", "thread/thr", "likes concise output"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestMemorySearchCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/memory/search" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		body := string(raw)
		for _, expected := range []string{`"query":"concise memory"`, `"scope_type":"thread"`, `"scope_id":"thr"`} {
			if !strings.Contains(body, expected) {
				t.Fatalf("body missing %q: %s", expected, body)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[{"id":"mem_2","title":"Style","summary":"concise memory match","scope_type":"thread","scope_id":"thr","status":"approved","safety_state":"safe","source_type":"run","source_run_id":"run","redaction_applied":true}],"excluded_count":0,"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"memory", "search", "--host", server.URL, "--scope-type", "thread", "--scope-id", "thr", "concise", "memory"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if output := stdout.String(); !strings.Contains(output, "mem_2") || !strings.Contains(output, "redacted=true") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestMemoryShowCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/memory/entries/mem_1" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("scope_type") != "thread" || r.URL.Query().Get("scope_id") != "thr" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"entry":{"id":"mem_1","title":"Preference","summary":"likes concise output","scope_type":"thread","scope_id":"thr","status":"approved","safety_state":"safe","source_type":"thread","redaction_applied":false},"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"memory", "show", "--host", server.URL, "--scope-type", "thread", "--scope-id", "thr", "mem_1"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"memory mem_1", "title Preference", "summary likes concise output"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestMemoryAuditCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/memory/audit" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("source_thread_id") != "thr" || r.URL.Query().Get("event_type") != "memory_write_approved" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[{"id":"evt_1","event_type":"memory_write_approved","summary":"Memory write approved","thread_id":"thr","run_id":"run","memory_entry_id":"mem_1","status":"approved","scope_type":"thread","source_type":"run","redaction_applied":true,"occurred_at":"2026-05-26T00:00:00Z"}],"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"memory", "audit", "--host", server.URL, "--thread-id", "thr", "--event-type", "memory_write_approved"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if output := stdout.String(); !strings.Contains(output, "evt_1") || !strings.Contains(output, "memory_write_approved") || !strings.Contains(output, "mem_1") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestAgentTasksCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/threads/thr/agent-tasks" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "4" {
			t.Fatalf("limit = %q", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tasks":[{"id":"agt_1","thread_id":"thr","run_id":"run","role":"reviewer","goal":"Review implementation","status":"spawned","created_at":"2026-05-26T00:00:00Z","updated_at":"2026-05-26T00:00:00Z"}],"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"agent", "tasks", "--host", server.URL, "--limit", "4", "thr"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"agt_1", "spawned", "reviewer", "run", "Review implementation"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestBrowserToolsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/tools/catalog" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tools":[{"name":"browser.open","group":"browser","risk_level":"medium","approval_policy":"always_required","execution_state":"executable","enabled":true},{"name":"agent.spawn","group":"agent","risk_level":"medium","approval_policy":"always_required","execution_state":"executable","enabled":true}],"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"browser", "tools", "--host", server.URL}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if !strings.Contains(output, "browser.open") || strings.Contains(output, "agent.spawn") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestBrowserEventsCommandFiltersBrowserToolEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/runs/run/events/stream" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run\",\"thread_id\":\"thr\",\"sequence\":1,\"type\":\"tool_call_succeeded\",\"metadata\":{\"tool_call_id\":\"tc_browser\",\"tool_name\":\"browser.open\",\"result_summary\":{\"session_id\":\"br_1\",\"title\":\"Docs\",\"url\":\"https://example.test\"}}}\n\n")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run\",\"thread_id\":\"thr\",\"sequence\":2,\"type\":\"tool_call_succeeded\",\"metadata\":{\"tool_call_id\":\"tc_agent\",\"tool_name\":\"agent.spawn\"}}\n\n")
		fmt.Fprint(w, "event: close\n")
		fmt.Fprint(w, "data: {\"run_id\":\"run\"}\n\n")
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"browser", "events", "--host", server.URL, "--compact", "run"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if !strings.Contains(output, "browser.open") || strings.Contains(output, "agent.spawn") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestVersionCommand(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"version"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"loomi dev", "commit unknown", "date unknown"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestVersionCommandJSON(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"version", "--output", "json"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{`"version": "dev"`, `"commit": "unknown"`, `"date": "unknown"`} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestCompletionCommandBash(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"completion", "bash"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"_loomi_completion()", "complete -F _loomi_completion loomi", "doctor", "approvals"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestCompletionCommandZsh(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"completion", "zsh"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"#compdef loomi", "_loomi()", "'completion:print shell completion script'"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestCompletionCommandFish(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"completion", "fish"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"complete -c loomi -f", "completion", "bash zsh fish"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestCompletionCommandRejectsUnknownShell(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"completion", "powershell"}, &stdout, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unsupported shell powershell") {
		t.Fatalf("err = %#v", err)
	}
}

func TestDoctorCommandReportsHealthyChecks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case "/v1/model-providers":
			fmt.Fprint(w, `{"providers":[{"id":"local_codex","status":"available","execution_state":"supported","model":"gpt-5-codex"}],"request_id":"req_providers"}`)
		case "/v1/tools/catalog":
			fmt.Fprint(w, `{"tools":[{"name":"workspace.read","group":"workspace","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true}],"request_id":"req_tools"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"doctor", "--host", server.URL}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"doctor ok",
		"ok\tapi\t" + server.URL,
		"ok\tproviders\tlocal_codex status=available execution=supported model=gpt-5-codex",
		"ok\ttools\t1 tools, 1 enabled, 1 groups",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestDoctorCommandReturnsExitErrorWhenAPIUnavailable(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"doctor", "--host", "http://127.0.0.1:1"}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if exit, ok := err.(exitError); !ok || exit.code != 1 {
		t.Fatalf("err = %#v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "doctor fail") || !strings.Contains(output, "fail\tapi\t") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestDoctorCommandExplainsMissingDefaultLocalCodexProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case "/v1/model-providers":
			fmt.Fprint(w, `{"providers":[{"id":"custom","status":"available","execution_state":"supported","model":"test-model"}],"request_id":"req_providers"}`)
		case "/v1/local-provider-detections":
			fmt.Fprint(w, `{"providers":[{"provider_id":"local_claude_code","status":"unavailable"}],"request_id":"req_local"}`)
		case "/v1/tools/catalog":
			fmt.Fprint(w, `{"tools":[{"name":"workspace.read","group":"workspace","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true}],"request_id":"req_tools"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"doctor", "--host", server.URL}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"warn\tproviders\tdefault provider local_codex is not registered",
		"fix\tproviders\tset LOOMI_PROVIDER or run loomi config set provider <id>; use loomi models list to see registered providers",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestDoctorCommandExplainsDetectedButDisabledLocalCodexProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case "/v1/model-providers":
			fmt.Fprint(w, `{"providers":[],"request_id":"req_providers"}`)
		case "/v1/local-provider-detections":
			fmt.Fprint(w, `{"providers":[{"provider_id":"local_codex","auth_mode":"oauth","status":"available","message":"Detected OAuth token presence but not enabled. Explicit opt-in is required before use."}],"request_id":"req_local"}`)
		case "/v1/tools/catalog":
			fmt.Fprint(w, `{"tools":[{"name":"workspace.read","group":"workspace","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true}],"request_id":"req_tools"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"doctor", "--host", server.URL, "--provider", "local_codex"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"warn\tproviders\tlocal_codex blocked status=available auth=oauth",
		"fix\tproviders\tenable Local Codex in Settings > Providers or POST /v1/local-provider-detections/local_codex/enable",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestDoctorCommandExplainsProviderAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case "/v1/model-providers":
			fmt.Fprint(w, `{"providers":[{"id":"custom","status":"configured","execution_state":"supported","model":"bad-model"}],"request_id":"req_providers"}`)
		case "/v1/model-providers/check":
			fmt.Fprint(w, `{"provider":{"id":"custom","status":"completion-failed","execution_state":"supported","model":"bad-model","check_stage":"completion","check_code":"completion-failed-auth","http_status":401,"message":"Provider token was rejected by the upstream completion endpoint."},"request_id":"req_check"}`)
		case "/v1/tools/catalog":
			fmt.Fprint(w, `{"tools":[{"name":"workspace.read","group":"workspace","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true}],"request_id":"req_tools"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"doctor", "--host", server.URL, "--provider", "custom"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"warn\tproviders\tcustom status=completion-failed execution=supported model=bad-model check_stage=completion check=completion-failed-auth http=401",
		"fix\tproviders\tRefresh the provider API token",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestDoctorCommandUsesConfiguredAPITokenWithoutPrintingIt(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{"api_token":"doctor-secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOOMI_CONFIG", configPath)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" && r.Header.Get("Authorization") != "Bearer doctor-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"missing bearer token"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case "/v1/model-providers":
			fmt.Fprint(w, `{"providers":[{"id":"local_codex","status":"available","execution_state":"supported","model":"gpt-5-codex"}],"request_id":"req_providers"}`)
		case "/v1/tools/catalog":
			fmt.Fprint(w, `{"tools":[{"name":"workspace.read","group":"workspace","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true}],"request_id":"req_tools"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"doctor", "--host", server.URL}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if !strings.Contains(output, "doctor ok") || !strings.Contains(output, "api_token_set=true") || strings.Contains(output, "doctor-secret") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestDoctorCommandDoesNotPrintProviderErrorBodyOrConfigPath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{"api_token":"doctor-secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOOMI_CONFIG", configPath)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			fmt.Fprint(w, `{"status":"ready","checks":[{"name":"database","status":"ok"},{"name":"schema","status":"ok"}]}`)
		case "/v1/model-providers":
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, `{"error":{"code":"provider_failed","message":"upstream body includes doctor-secret and /Users/someone/.codex/auth.json"}}`)
		case "/v1/tools/catalog":
			fmt.Fprint(w, `{"tools":[{"name":"workspace.read","group":"workspace","risk_level":"low","approval_policy":"always_required","execution_state":"executable","enabled":true}],"request_id":"req_tools"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"doctor", "--host", server.URL}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected provider failure")
	}
	output := stdout.String()
	for _, leaked := range []string{"doctor-secret", configPath, ".codex/auth.json", "/Users/someone"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("stdout leaked %q: %s", leaked, output)
		}
	}
	if !strings.Contains(output, "GET /v1/model-providers returned 502") || !strings.Contains(output, "code=provider_failed") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestDoctorCommandExplainsMissingAPIToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case "/v1/model-providers", "/v1/tools/catalog":
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"missing bearer token"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"doctor", "--host", server.URL}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	output := stdout.String()
	if !strings.Contains(output, "doctor fail") ||
		!strings.Contains(output, "fix\tproviders\tset LOOMI_API_TOKEN or run loomi config set api_token <token>") ||
		!strings.Contains(output, "fix\ttools\tset LOOMI_API_TOKEN or run loomi config set api_token <token>") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestSmokeAgentCommandReportsProviderBoundaryBlocker(t *testing.T) {
	failureLog := filepath.Join(t.TempDir(), "smoke-failure.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case "/v1/model-providers/check":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"provider":{"id":"custom","family":"openai_compatible","model":"bad-model","status":"completion-failed","execution_state":"supported","check_stage":"completion","check_code":"completion-failed-auth","http_status":401,"message":"Provider token was rejected by the upstream completion endpoint."},"request_id":"req_check"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"smoke", "agent", "--host", server.URL, "--provider", "custom", "--prompt", "hello", "--failure-log", failureLog}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected provider-boundary blocker exit")
	}
	output := stdout.String()
	for _, expected := range []string{
		"smoke blocked",
		"stage\tprovider_check",
		"provider\tcustom status=completion-failed check_stage=completion check=completion-failed-auth http=401",
		"blocked_reason\tprovider_auth",
		"fix\tRefresh the provider API token",
		"failure_log\t" + failureLog,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	content, err := os.ReadFile(failureLog)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `"blocked_reason": "provider_auth"`) || strings.Contains(string(content), "doctor-secret") {
		t.Fatalf("failure log content = %s", string(content))
	}
}

func TestSmokeAgentCommandRunsWithAutoApprovalAndPrintsSummary(t *testing.T) {
	var calls []string
	workspaceRoot := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workspace/root":
			fmt.Fprint(w, `{"config":{"configured":true,"display_name":"smoke-workspace"},"request_id":"req_workspace"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/model-providers/check":
			fmt.Fprint(w, `{"provider":{"id":"custom","family":"openai_compatible","model":"model","status":"completion-ok","execution_state":"supported","check_stage":"completion","check_code":"completion-ok","http_status":200,"message":"Provider completion check succeeded."},"request_id":"req_check"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_smoke","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_smoke"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/runs":
			fmt.Fprint(w, `{"run":{"id":"run_smoke","thread_id":"thr_smoke","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/runs/run_smoke/tool-calls/tc_read/approve":
			fmt.Fprint(w, `{"tool_call":{"tool_call_id":"tc_read","approval_status":"approved","execution_status":"not_started"},"request_id":"req_approve"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_smoke/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			if r.URL.Query().Get("after_sequence") == "" {
				fmt.Fprint(w, "event: run_event\n")
				fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":1,\"type\":\"model_request_started\",\"metadata\":{\"model_phase\":\"initial\",\"provider_id\":\"custom\"}}\n\n")
				fmt.Fprint(w, "event: run_event\n")
				fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":2,\"type\":\"tool_call_approval_required\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\",\"arguments_summary\":{\"path\":\"AGENTS.md\"}}}\n\n")
				return
			}
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":3,\"type\":\"tool_call_approved\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\"}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_4\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":4,\"type\":\"tool_call_succeeded\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\",\"result_summary\":{\"path\":\"AGENTS.md\",\"truncated\":false}}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_5\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":5,\"type\":\"model_output_delta\",\"content\":\"ok\"}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_6\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":6,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_smoke\"}\n\n")
		case r.Method == http.MethodGet && r.URL.Path == "/v1/threads/thr_smoke/messages":
			fmt.Fprint(w, `{"messages":[{"id":"msg_smoke","thread_id":"thr_smoke","role":"user","content":"hello"},{"id":"msg_final","thread_id":"thr_smoke","role":"assistant","content":"ok final summary","run_id":"run_smoke"}],"request_id":"req_messages"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_smoke/events":
			fmt.Fprint(w, `{"events":[{"id":"evt_1","run_id":"run_smoke","thread_id":"thr_smoke","sequence":1,"type":"model_request_started","metadata":{"model_phase":"initial","provider_id":"custom"}},{"id":"evt_2","run_id":"run_smoke","thread_id":"thr_smoke","sequence":2,"type":"tool_call_approval_required","metadata":{"tool_call_id":"tc_read","tool_name":"workspace.read","arguments_summary":{"path":"AGENTS.md"}}},{"id":"evt_3","run_id":"run_smoke","thread_id":"thr_smoke","sequence":3,"type":"tool_call_approved","metadata":{"tool_call_id":"tc_read","tool_name":"workspace.read"}},{"id":"evt_4","run_id":"run_smoke","thread_id":"thr_smoke","sequence":4,"type":"tool_call_succeeded","metadata":{"tool_call_id":"tc_read","tool_name":"workspace.read","result_summary":{"path":"AGENTS.md","truncated":false}}},{"id":"evt_5","run_id":"run_smoke","thread_id":"thr_smoke","sequence":5,"type":"model_output_delta","content":"ok"},{"id":"evt_6","run_id":"run_smoke","thread_id":"thr_smoke","sequence":6,"type":"run_completed"}],"request_id":"req_replay"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"smoke", "agent", "--host", server.URL, "--provider", "custom", "--workspace", workspaceRoot, "--auto-approve", "--prompt", "hello"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"smoke ok",
		"thread_id\tthr_smoke",
		"run_id\trun_smoke",
		"final_stage\trun_completed",
		"provider\tcustom status=completion-ok check_stage=completion check=completion-ok http=200",
		"workspace\tsmoke-workspace",
		"events\t6 total, 3 tool, 1 approvals",
		"pending_approvals\t0",
		"tool_chain\tworkspace.read",
		"tool_order\tworkspace.read",
		"final_persisted\tok",
		"replay\tok events=6 terminal=run_completed",
		"final_message\tok final summary",
		"last_events\t0004 tool_call_succeeded workspace.read tc_read; 0005 model_output_delta; 0006 run_completed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if !containsMainString(calls, "POST /v1/threads/thr_smoke/runs/run_smoke/tool-calls/tc_read/approve") {
		t.Fatalf("calls = %v", calls)
	}
	if !containsMainString(calls, "POST /v1/workspace/root") || !containsMainString(calls, "GET /v1/threads/thr_smoke/messages") {
		t.Fatalf("calls = %v", calls)
	}
	if !containsMainString(calls, "GET /v1/runs/run_smoke/events") {
		t.Fatalf("calls = %v", calls)
	}
}

func TestSmokeAgentCommandBlocksRedactedFinalMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/model-providers/check":
			fmt.Fprint(w, `{"provider":{"id":"custom","family":"openai_compatible","model":"model","status":"completion-ok","execution_state":"supported"},"request_id":"req_check"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_smoke","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_smoke"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/runs":
			fmt.Fprint(w, `{"run":{"id":"run_smoke","thread_id":"thr_smoke","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_smoke/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":1,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_smoke\"}\n\n")
		case r.Method == http.MethodGet && r.URL.Path == "/v1/threads/thr_smoke/messages":
			fmt.Fprint(w, `{"messages":[{"id":"msg_final","thread_id":"thr_smoke","role":"assistant","content":"[redacted]","run_id":"run_smoke"}],"request_id":"req_messages"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"smoke", "agent", "--host", server.URL, "--provider", "custom", "--prompt", "hello"}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected redacted final message blocker")
	}
	output := stdout.String()
	for _, expected := range []string{
		"smoke blocked",
		"stage\tfinal_message",
		"final_message\t[redacted]",
		"blocked_reason\tfinal_message_redacted",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestSmokeAgentCommandWithThreadRejectsOldAssistantMessageWithoutCurrentRunID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/model-providers/check":
			fmt.Fprint(w, `{"provider":{"id":"custom","family":"openai_compatible","model":"model","status":"completion-ok","execution_state":"supported"},"request_id":"req_check"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_existing/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_smoke"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_existing/runs":
			fmt.Fprint(w, `{"run":{"id":"run_current","thread_id":"thr_existing","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_current/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_current\",\"thread_id\":\"thr_existing\",\"sequence\":1,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_current\"}\n\n")
		case r.Method == http.MethodGet && r.URL.Path == "/v1/threads/thr_existing/messages":
			fmt.Fprint(w, `{"messages":[{"id":"old_user","thread_id":"thr_existing","role":"user","content":"old prompt"},{"id":"old_assistant","thread_id":"thr_existing","role":"assistant","content":"old assistant final"}],"request_id":"req_messages"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"smoke", "agent", "--host", server.URL, "--provider", "custom", "--thread", "thr_existing", "--prompt", "hello"}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected missing final message blocker")
	}
	output := stdout.String()
	for _, expected := range []string{
		"smoke blocked",
		"stage\tfinal_message",
		"blocked_reason\tfinal_message_missing",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if strings.Contains(output, "old assistant final") {
		t.Fatalf("stdout used stale assistant message: %s", output)
	}
}

func TestSmokeAgentCommandBlocksCompletedRunWithPendingApproval(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/model-providers/check":
			fmt.Fprint(w, `{"provider":{"id":"custom","family":"openai_compatible","model":"model","status":"completion-ok","execution_state":"supported"},"request_id":"req_check"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_smoke","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_smoke"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/runs":
			fmt.Fprint(w, `{"run":{"id":"run_smoke","thread_id":"thr_smoke","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/runs/run_smoke/tool-calls/tc_edit/approve":
			fmt.Fprint(w, `{"tool_call":{"tool_call_id":"tc_edit","approval_status":"approved","execution_status":"not_started"},"request_id":"req_approve"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_smoke/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			if r.URL.Query().Get("after_sequence") == "" {
				fmt.Fprint(w, "event: run_event\n")
				fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":1,\"type\":\"tool_call_approval_required\",\"metadata\":{\"tool_call_id\":\"tc_edit\",\"tool_name\":\"workspace.edit\"}}\n\n")
				return
			}
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":2,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_smoke\"}\n\n")
		case r.Method == http.MethodGet && r.URL.Path == "/v1/threads/thr_smoke/messages":
			fmt.Fprint(w, `{"messages":[{"id":"msg_final","thread_id":"thr_smoke","role":"assistant","content":"done","run_id":"run_smoke"}],"request_id":"req_messages"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"smoke", "agent", "--host", server.URL, "--provider", "custom", "--auto-approve", "--prompt", "hello"}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected pending approval blocker")
	}
	output := stdout.String()
	for _, expected := range []string{
		"smoke blocked",
		"stage\tpending_approval",
		"blocked_reason\tpending_approval",
		"tool_chain\tworkspace.edit",
		"final_message\tdone",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestSmokeAgentCommandBlocksGeneratedFailurePlaceholderFinalMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/readyz":
			fmt.Fprint(w, `{"ok":true}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/model-providers/check":
			fmt.Fprint(w, `{"provider":{"id":"custom","family":"openai_compatible","model":"model","status":"completion-ok","execution_state":"supported"},"request_id":"req_check"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_smoke","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_smoke"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_smoke/runs":
			fmt.Fprint(w, `{"run":{"id":"run_smoke","thread_id":"thr_smoke","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_smoke/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_smoke\",\"thread_id\":\"thr_smoke\",\"sequence\":1,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_smoke\"}\n\n")
		case r.Method == http.MethodGet && r.URL.Path == "/v1/threads/thr_smoke/messages":
			fmt.Fprint(w, `{"messages":[{"id":"msg_final","thread_id":"thr_smoke","role":"assistant","content":"未生成成功回复","run_id":"run_smoke"}],"request_id":"req_messages"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"smoke", "agent", "--host", server.URL, "--provider", "custom", "--prompt", "hello"}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected placeholder final message blocker")
	}
	output := stdout.String()
	for _, expected := range []string{
		"smoke blocked",
		"stage\tfinal_message",
		"blocked_reason\tfinal_message_placeholder",
		"final_message\t未生成成功回复",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestRunHelpCommandShowsTopics(t *testing.T) {
	var stdout bytes.Buffer
	if err := run([]string{"help", "tools"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"usage: loomi tools list", "--group <name>", "--enabled-only", "--flat"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("stdout missing %q: %s", expected, stdout.String())
		}
	}
}

func TestRunConfigSetUnsetCommands(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("LOOMI_CONFIG", configPath)

	var stdout bytes.Buffer
	if err := run([]string{"config", "set", "host", "http://127.0.0.1:9999"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "set host") {
		t.Fatalf("stdout = %s", stdout.String())
	}
	stdout.Reset()
	if err := run([]string{"config", "show"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "host\thttp://127.0.0.1:9999") {
		t.Fatalf("stdout = %s", stdout.String())
	}
	stdout.Reset()
	if err := run([]string{"config", "unset", "host"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "unset host") {
		t.Fatalf("stdout = %s", stdout.String())
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "127.0.0.1:9999") {
		t.Fatalf("config file = %s", string(raw))
	}
}

func TestRunApprovalDecisionCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/threads/thr/runs/run/tool-calls/tc/approve" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tool_call":{"tool_call_id":"tc","approval_status":"approved","execution_status":"not_started"},"request_id":"req"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"approvals", "approve", "--host", server.URL, "thr", "run", "tc"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(stdout.String()) != "tc approved not_started" {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestRunStopCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/runs/run_cli/stop" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"stopped"},"result":"stopped","request_id":"req_stop"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"runs", "stop", "--host", server.URL, "run_cli"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(stdout.String()) != "run run_cli stopped stopped" {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestRunStatusCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/runs/run_cli" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"runs", "status", "--host", server.URL, "run_cli"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "run run_cli running") || !strings.Contains(stdout.String(), "thread thr_cli") {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestRunsAttachReplaysThenStreamsFromLastSequence(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli" && r.URL.RawQuery == "":
			fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events" && r.URL.Query().Get("after_sequence") == "":
			fmt.Fprint(w, `{"events":[{"id":"evt_1","run_id":"run_cli","thread_id":"thr_cli","sequence":1,"type":"model_output_delta","content":"hello"},{"id":"evt_2","run_id":"run_cli","thread_id":"thr_cli","sequence":2,"type":"tool_call_approval_required","metadata":{"tool_call_id":"tc","tool_name":"workspace.read"}}],"request_id":"req_events"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events/stream":
			if r.URL.Query().Get("after_sequence") != "2" {
				t.Fatalf("after_sequence = %q", r.URL.Query().Get("after_sequence"))
			}
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"runs", "attach", "--host", server.URL, "--compact", "run_cli"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"run run_cli running",
		"0001 hello",
		"0002 approval_required workspace.read tc",
		"0003 run_completed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if len(calls) != 3 {
		t.Fatalf("calls = %#v", calls)
	}
}

func TestRunsAttachTerminalRunReplaysHistoryWithoutOpeningStream(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli":
			fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"completed"},"request_id":"req_run"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events":
			fmt.Fprint(w, `{"events":[{"id":"evt_1","run_id":"run_cli","thread_id":"thr_cli","sequence":1,"type":"model_output_delta","content":"done"},{"id":"evt_2","run_id":"run_cli","thread_id":"thr_cli","sequence":2,"type":"run_completed"}],"request_id":"req_events"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"runs", "attach", "--host", server.URL, "--compact", "run_cli"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if !strings.Contains(output, "run run_cli completed") || !strings.Contains(output, "0001 done") || !strings.Contains(output, "0002 run_completed") {
		t.Fatalf("stdout = %s", output)
	}
	if len(calls) != 2 {
		t.Fatalf("calls = %#v", calls)
	}
}

func TestRunsFollowDefaultsToFutureEventsOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events":
			fmt.Fprint(w, `{"events":[{"id":"evt_1","run_id":"run_cli","thread_id":"thr_cli","sequence":1,"type":"model_output_delta","content":"old"}],"request_id":"req_events"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events/stream":
			if r.URL.Query().Get("after_sequence") != "1" {
				t.Fatalf("after_sequence = %q", r.URL.Query().Get("after_sequence"))
			}
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":2,\"type\":\"model_output_delta\",\"content\":\"new\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"runs", "follow", "--host", server.URL, "run_cli"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if strings.Contains(output, "old") || !strings.Contains(output, "0002 model_output_delta new") {
		t.Fatalf("stdout = %s", output)
	}
}

func TestRunApprovalDecisionCommandCanFollowAfterCurrentSequence(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run/events":
			fmt.Fprint(w, `{"events":[{"id":"evt_1","sequence":1,"type":"tool_call_approval_required","metadata":{"tool_call_id":"tc","tool_name":"workspace.read"}}],"request_id":"req_events"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr/runs/run/tool-calls/tc/approve":
			fmt.Fprint(w, `{"tool_call":{"tool_call_id":"tc","approval_status":"approved","execution_status":"not_started"},"request_id":"req"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run/events/stream":
			if r.URL.Query().Get("after_sequence") != "1" {
				t.Fatalf("after_sequence = %q", r.URL.Query().Get("after_sequence"))
			}
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run\",\"thread_id\":\"thr\",\"sequence\":2,\"type\":\"tool_call_succeeded\",\"metadata\":{\"tool_call_id\":\"tc\",\"tool_name\":\"workspace.read\",\"result_summary\":{\"path\":\"README.md\",\"truncated\":false}}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run\",\"thread_id\":\"thr\",\"sequence\":3,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run\"}\n\n")
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"approvals", "approve", "--host", server.URL, "--follow", "thr", "run", "tc"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"tc approved not_started",
		`0002 tool_call_succeeded workspace.read tc result=path=README.md truncated=false`,
		"0003 run_completed run_completed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if strings.Join(calls, ",") != "GET /v1/runs/run/events,POST /v1/threads/thr/runs/run/tool-calls/tc/approve,GET /v1/runs/run/events/stream?after_sequence=1" {
		t.Fatalf("calls = %v", calls)
	}
}

func TestRunSessionsModelsAndPersonasListCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/threads":
			fmt.Fprint(w, `{"threads":[{"id":"thr_cli","title":"Thread CLI","mode":"work","lifecycle_status":"active","updated_at":"2026-05-26T12:00:00Z"}],"request_id":"req_threads"}`)
		case "/v1/model-providers":
			fmt.Fprint(w, `{"providers":[{"id":"local_codex","family":"openai","model":"gpt-5-codex","status":"ready","execution_state":"executable"}],"request_id":"req_models"}`)
		case "/v1/personas":
			fmt.Fprint(w, `{"personas":[{"id":"persona_dev","slug":"developer","name":"Developer","source":"builtin","is_default":true,"active_version":1}],"request_id":"req_personas"}`)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	for _, tc := range []struct {
		name     string
		args     []string
		expected string
	}{
		{name: "sessions", args: []string{"sessions", "list", "--host", server.URL}, expected: "Thread CLI"},
		{name: "models", args: []string{"models", "list", "--host", server.URL}, expected: "gpt-5-codex"},
		{name: "personas", args: []string{"personas", "list", "--host", server.URL}, expected: "developer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			err := run(tc.args, &stdout, &bytes.Buffer{})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(stdout.String(), tc.expected) {
				t.Fatalf("stdout = %s", stdout.String())
			}
		})
	}
}

func TestRunCommandReadsPromptFileFromStdinAndPrintsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case "/v1/threads/thr_cli/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case "/v1/threads/thr_cli/runs":
			fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case "/v1/runs/run_cli/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"sequence\":1,\"type\":\"model_output_delta\",\"content\":\"Hi\"}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"sequence\":2,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := runWithIO([]string{"run", "--host", server.URL, "--prompt-file", "-", "--output", "json"}, strings.NewReader("hello from stdin"), &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if !strings.Contains(output, `"RunID": "run_cli"`) || !strings.Contains(output, `"Output": "Hi"`) {
		t.Fatalf("stdout = %s", output)
	}
}

func TestRunCommandInteractiveApprovalsApprovesAndContinues(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs":
			fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs/run_cli/tool-calls/tc_read/approve":
			fmt.Fprint(w, `{"tool_call":{"tool_call_id":"tc_read","approval_status":"approved","execution_status":"not_started"},"request_id":"req_approve"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			if r.URL.Query().Get("after_sequence") == "" {
				fmt.Fprint(w, "event: run_event\n")
				fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":1,\"type\":\"tool_call_approval_required\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\",\"arguments_summary\":{\"path\":\"README.md\"}}}\n\n")
				return
			}
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":2,\"type\":\"tool_call_approved\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\"}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := runWithIO([]string{"run", "--host", server.URL, "--interactive-approvals", "hello"}, strings.NewReader("a\n"), &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		`0001 tool_call_approval_required workspace.read tc_read args=`,
		"approve workspace.read tc_read?",
		"0002 tool_call_approved workspace.read tc_read",
		"run run_cli completed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if !containsMainString(calls, "POST /v1/threads/thr_cli/runs/run_cli/tool-calls/tc_read/approve") {
		t.Fatalf("calls = %v", calls)
	}
}

func TestRunCommandCompactPrintsShortTranscript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case "/v1/threads/thr_cli/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case "/v1/threads/thr_cli/runs":
			fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case "/v1/runs/run_cli/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":1,\"type\":\"model_output_delta\",\"content\":\"thinking\"}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":2,\"type\":\"tool_call_succeeded\",\"metadata\":{\"tool_call_id\":\"tc_exec\",\"tool_name\":\"sandbox.exec_command\",\"result_summary\":{\"exit_code\":0,\"timed_out\":false,\"stdout\":\"ok\\nnext\"}}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := runWithIO([]string{"run", "--host", server.URL, "--compact", "hello"}, strings.NewReader(""), &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"0001 thinking",
		`0002 succeeded sandbox.exec_command tc_exec result=exit=0 timeout=false stdout="ok"`,
		"run run_cli completed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if strings.Contains(output, "tool_call_succeeded") {
		t.Fatalf("stdout was not compact: %s", output)
	}
}

func TestRunCommandStopsWhenTerminalEventArrivesWithoutSSEClose(t *testing.T) {
	streamClosed := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case "/v1/threads/thr_cli/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case "/v1/threads/thr_cli/runs":
			fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case "/v1/runs/run_cli/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":1,\"type\":\"run_failed\"}\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			<-r.Context().Done()
			close(streamClosed)
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := runWithIO([]string{"run", "--host", server.URL, "hello"}, strings.NewReader(""), &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-streamClosed:
	case <-time.After(time.Second):
		t.Fatalf("stream was not closed after terminal event")
	}
	if !strings.Contains(stdout.String(), "run run_cli failed") {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestRunCommandUsesConfigDefaults(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{"host":"http://unused","mode":"work","provider":"configured_provider","model":"configured_model","persona":"configured_persona"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOOMI_CONFIG", configPath)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/threads":
			fmt.Fprint(w, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case "/v1/threads/thr_cli/messages":
			fmt.Fprint(w, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case "/v1/threads/thr_cli/runs":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			payload := string(body)
			for _, expected := range []string{`"provider_id":"configured_provider"`, `"model":"configured_model"`, `"persona_id":"configured_persona"`} {
				if !strings.Contains(payload, expected) {
					t.Fatalf("run payload missing %s: %s", expected, payload)
				}
			}
			fmt.Fprint(w, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case "/v1/runs/run_cli/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"sequence\":1,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := runWithIO([]string{"run", "--host", server.URL, "hello"}, strings.NewReader(""), &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunApprovalsFollowCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/runs/run_cli/events/stream" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":1,\"type\":\"model_output_delta\",\"content\":\"ignore\"}\n\n")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":2,\"type\":\"tool_call_approval_required\",\"metadata\":{\"tool_call_id\":\"tc_exec\",\"tool_name\":\"sandbox.exec_command\"}}\n\n")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"tool_call_approved\",\"metadata\":{\"tool_call_id\":\"tc_exec\",\"tool_name\":\"sandbox.exec_command\"}}\n\n")
		fmt.Fprint(w, "event: close\n")
		fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"approvals", "follow", "--host", server.URL, "run_cli"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{"approval required: sandbox.exec_command tc_exec", "loomi approvals approve thr_cli run_cli tc_exec", "tool_call_approved: sandbox.exec_command tc_exec"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if strings.Contains(output, "ignore") {
		t.Fatalf("stdout includes non-approval event: %s", output)
	}
}

func TestRunEventsTailToolsOnlyCompact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/runs/run_cli/events/stream" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":1,\"type\":\"model_output_delta\",\"content\":\"ignore\"}\n\n")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":2,\"type\":\"tool_call_approval_required\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\",\"arguments_summary\":{\"path\":\"README.md\"}}}\n\n")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"tool_call_succeeded\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\",\"result_summary\":{\"path\":\"README.md\",\"truncated\":false}}}\n\n")
		fmt.Fprint(w, "event: run_event\n")
		fmt.Fprint(w, "data: {\"id\":\"evt_4\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":4,\"type\":\"run_completed\"}\n\n")
		fmt.Fprint(w, "event: close\n")
		fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"events", "tail", "--host", server.URL, "--tools-only", "--compact", "run_cli"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		`0002 approval_required workspace.read tc_read args=path=README.md`,
		`0003 succeeded workspace.read tc_read result=path=README.md truncated=false`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if strings.Contains(output, "ignore") || strings.Contains(output, "run_completed") {
		t.Fatalf("stdout includes filtered event: %s", output)
	}
}

func containsMainString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
