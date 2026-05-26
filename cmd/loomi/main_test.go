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
