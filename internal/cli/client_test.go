package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerExecuteCreatesThreadMessageRunAndConsumesSSE(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.String())
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			writeTestJSON(w, http.StatusCreated, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/messages":
			writeTestJSON(w, http.StatusCreated, `{"message":{"id":"msg_cli","thread_id":"thr_cli","role":"user","content":"hello"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs":
			writeTestJSON(w, http.StatusAccepted, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"sequence\":1,\"type\":\"model_output_delta\",\"content\":\"Hi\"}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":2,\"type\":\"tool_call_approval_required\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\"}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"tool_call_approved\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\"}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_4\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":4,\"type\":\"tool_call_approval_required\",\"metadata\":{\"tool_call_id\":\"tc_exec\",\"tool_name\":\"sandbox.exec_command\"}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_5\",\"sequence\":5,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	var rendered []RunEvent
	result, err := Runner{Client: client}.Execute(context.Background(), RunOptions{
		Prompt:   "hello",
		Mode:     "work",
		Provider: "local_codex",
		Model:    "gpt-local-fixture",
		OnEvent: func(event RunEvent) {
			rendered = append(rendered, event)
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.ThreadID != "thr_cli" || result.RunID != "run_cli" || result.Output != "Hi" || result.Status != "completed" {
		t.Fatalf("result = %+v", result)
	}
	if len(rendered) != 5 || rendered[0].Type != "model_output_delta" || rendered[4].Type != "run_completed" {
		t.Fatalf("rendered = %+v", rendered)
	}
	if len(result.PendingApprovals) != 1 || result.PendingApprovals[0].ToolCallID != "tc_exec" || result.PendingApprovals[0].ToolName != "sandbox.exec_command" {
		t.Fatalf("pending approvals = %+v", result.PendingApprovals)
	}
	for _, expected := range []string{
		"POST /v1/threads",
		"POST /v1/threads/thr_cli/messages",
		"POST /v1/threads/thr_cli/runs",
		"GET /v1/runs/run_cli/events/stream",
	} {
		if !containsString(calls, expected) {
			t.Fatalf("calls missing %q: %v", expected, calls)
		}
	}
}

func TestRunnerExecuteCreatesThreadWithPromptTitle(t *testing.T) {
	var threadPayload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			if err := json.NewDecoder(r.Body).Decode(&threadPayload); err != nil {
				t.Fatal(err)
			}
			writeTestJSON(w, http.StatusCreated, `{"thread":{"id":"thr_cli","mode":"chat","title":"请只回复 pong"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/messages":
			writeTestJSON(w, http.StatusCreated, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs":
			writeTestJSON(w, http.StatusAccepted, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":1,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	_, err := Runner{Client: NewClient(server.URL)}.Execute(context.Background(), RunOptions{
		Prompt:   "请只回复 pong\n\n这部分很长，不能全部塞进标题里。" + strings.Repeat("x", 120),
		Mode:     "chat",
		Provider: "local_codex",
	})
	if err != nil {
		t.Fatal(err)
	}

	if threadPayload["mode"] != "chat" {
		t.Fatalf("mode = %q", threadPayload["mode"])
	}
	title := threadPayload["title"]
	if title == "" {
		t.Fatalf("thread payload missing title: %+v", threadPayload)
	}
	if len([]rune(title)) > 80 {
		t.Fatalf("title too long: %q", title)
	}
	if strings.Contains(title, "这部分很长") || strings.Contains(title, strings.Repeat("x", 20)) {
		t.Fatalf("title leaked too much prompt: %q", title)
	}
}

func TestRunnerReconnectsEventStreamWithAfterSequence(t *testing.T) {
	streamCalls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			writeTestJSON(w, http.StatusCreated, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/messages":
			writeTestJSON(w, http.StatusCreated, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs":
			writeTestJSON(w, http.StatusAccepted, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events/stream":
			streamCalls = append(streamCalls, r.URL.RawQuery)
			w.Header().Set("Content-Type", "text/event-stream")
			if r.URL.Query().Get("after_sequence") == "" {
				fmt.Fprint(w, "event: run_event\n")
				fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":1,\"type\":\"model_output_delta\",\"content\":\"Hi\"}\n\n")
				return
			}
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":2,\"type\":\"model_output_delta\",\"content\":\" there\"}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	result, err := Runner{Client: NewClient(server.URL)}.Execute(context.Background(), RunOptions{
		Prompt:   "hello",
		Provider: "local_codex",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Status != "completed" || result.Output != "Hi there" {
		t.Fatalf("result = %+v", result)
	}
	if strings.Join(streamCalls, ",") != ",after_sequence=1" {
		t.Fatalf("streamCalls = %v", streamCalls)
	}
}

func TestRunnerInteractiveApprovalDecidesAndContinues(t *testing.T) {
	var calls []string
	streamCalls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.String())
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			writeTestJSON(w, http.StatusCreated, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/messages":
			writeTestJSON(w, http.StatusCreated, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs":
			writeTestJSON(w, http.StatusAccepted, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs/run_cli/tool-calls/tc_read/approve":
			writeTestJSON(w, http.StatusOK, `{"tool_call":{"tool_call_id":"tc_read","approval_status":"approved","execution_status":"not_started"},"request_id":"req_approve"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli/events/stream":
			streamCalls = append(streamCalls, r.URL.RawQuery)
			w.Header().Set("Content-Type", "text/event-stream")
			if r.URL.Query().Get("after_sequence") == "" {
				fmt.Fprint(w, "event: run_event\n")
				fmt.Fprint(w, "data: {\"id\":\"evt_1\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":1,\"type\":\"tool_call_approval_required\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\"}}\n\n")
				return
			}
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_2\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":2,\"type\":\"tool_call_approved\",\"metadata\":{\"tool_call_id\":\"tc_read\",\"tool_name\":\"workspace.read\"}}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"model_output_delta\",\"content\":\"done\"}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_4\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":4,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	approvals := []PendingApproval{}
	result, err := Runner{Client: NewClient(server.URL)}.Execute(context.Background(), RunOptions{
		Prompt:   "hello",
		Provider: "local_codex",
		OnApproval: func(approval PendingApproval, event RunEvent) (string, error) {
			approvals = append(approvals, approval)
			return "approve", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Status != "completed" || result.Output != "done" || len(result.PendingApprovals) != 0 {
		t.Fatalf("result = %+v", result)
	}
	if len(approvals) != 1 || approvals[0].ToolCallID != "tc_read" || approvals[0].ThreadID != "thr_cli" || approvals[0].RunID != "run_cli" {
		t.Fatalf("approvals = %+v", approvals)
	}
	if strings.Join(streamCalls, ",") != ",after_sequence=1" {
		t.Fatalf("streamCalls = %v", streamCalls)
	}
	if !containsString(calls, "POST /v1/threads/thr_cli/runs/run_cli/tool-calls/tc_read/approve") {
		t.Fatalf("calls = %v", calls)
	}
}

func TestREPLInteractiveApprovalDecidesAndContinues(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.String())
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads":
			writeTestJSON(w, http.StatusCreated, `{"thread":{"id":"thr_cli","mode":"work"},"request_id":"req_thread"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/messages":
			writeTestJSON(w, http.StatusCreated, `{"message":{"id":"msg_cli"},"request_id":"req_msg"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs":
			writeTestJSON(w, http.StatusAccepted, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/threads/thr_cli/runs/run_cli/tool-calls/tc_read/approve":
			writeTestJSON(w, http.StatusOK, `{"tool_call":{"tool_call_id":"tc_read","approval_status":"approved","execution_status":"not_started"},"request_id":"req_approve"}`)
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
			fmt.Fprint(w, "data: {\"id\":\"evt_3\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":3,\"type\":\"model_output_delta\",\"content\":\"done\"}\n\n")
			fmt.Fprint(w, "event: run_event\n")
			fmt.Fprint(w, "data: {\"id\":\"evt_4\",\"run_id\":\"run_cli\",\"thread_id\":\"thr_cli\",\"sequence\":4,\"type\":\"run_completed\"}\n\n")
			fmt.Fprint(w, "event: close\n")
			fmt.Fprint(w, "data: {\"run_id\":\"run_cli\"}\n\n")
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	repl := REPL{
		Client:   NewClient(server.URL),
		In:       strings.NewReader("hello\na\n/quit\n"),
		Out:      &stdout,
		Err:      &stderr,
		Provider: "local_codex",
		Mode:     "work",
	}
	if err := repl.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"approve workspace.read tc_read?",
		"0002 tool_call_approved workspace.read tc_read",
		"run run_cli completed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %s", stderr.String())
	}
	if !containsString(calls, "POST /v1/threads/thr_cli/runs/run_cli/tool-calls/tc_read/approve") {
		t.Fatalf("calls = %v", calls)
	}
}

func TestREPLSlashCommandsListToolsApprovalsAndEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/tools/catalog":
			writeTestJSON(w, http.StatusOK, `{"tools":[{"name":"workspace.read","group":"workspace","execution_state":"enabled","approval_policy":"always_required","risk_level":"low","enabled":true},{"name":"sandbox.exec_command","group":"sandbox","execution_state":"enabled","approval_policy":"always_required","risk_level":"high","enabled":true}],"request_id":"req_tools"}`)
		case "/v1/runs/run_cli/events":
			writeTestJSON(w, http.StatusOK, `{"events":[{"id":"evt_1","run_id":"run_cli","thread_id":"thr_cli","sequence":1,"type":"tool_call_approval_required","metadata":{"tool_call_id":"tc_read","tool_name":"workspace.read","arguments_summary":{"path":"README.md"}}},{"id":"evt_2","run_id":"run_cli","thread_id":"thr_cli","sequence":2,"type":"model_output_delta","content":"working"}],"request_id":"req_events"}`)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	repl := REPL{
		Client:    NewClient(server.URL),
		In:        strings.NewReader("/tools workspace\n/approvals\n/events compact\n/quit\n"),
		Out:       &stdout,
		Err:       &stderr,
		LastRunID: "run_cli",
	}
	if err := repl.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"[workspace]",
		"workspace.read",
		"thr_cli\trun_cli\ttc_read\tworkspace.read",
		"0001 approval_required workspace.read tc_read args=path=README.md",
		"0002 working",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
	if strings.Contains(output, "[sandbox]") {
		t.Fatalf("stdout includes unrequested group: %s", output)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestREPLStopCommandStopsLastRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/runs/run_cli":
			writeTestJSON(w, http.StatusOK, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"running"},"request_id":"req_run"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/runs/run_cli/stop":
			writeTestJSON(w, http.StatusOK, `{"run":{"id":"run_cli","thread_id":"thr_cli","status":"stopped"},"result":"stopped","request_id":"req_stop"}`)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	repl := REPL{
		Client:    NewClient(server.URL),
		In:        strings.NewReader("/run\n/stop\n/quit\n"),
		Out:       &stdout,
		Err:       &stderr,
		LastRunID: "run_cli",
	}
	if err := repl.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "run run_cli running") || !strings.Contains(stdout.String(), "run run_cli stopped stopped") {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestPendingApprovalEventsOnlyReturnsUnresolvedApprovals(t *testing.T) {
	events := []RunEvent{
		{Type: "tool_call_approval_required", RunID: "run", ThreadID: "thr", Metadata: map[string]any{"tool_call_id": "tc_read", "tool_name": "workspace.read"}},
		{Type: "tool_call_approved", RunID: "run", ThreadID: "thr", Metadata: map[string]any{"tool_call_id": "tc_read", "tool_name": "workspace.read"}},
		{Type: "tool_call_approval_required", RunID: "run", ThreadID: "thr", Metadata: map[string]any{"tool_call_id": "tc_exec", "tool_name": "sandbox.exec_command"}},
	}

	pending := PendingApprovals(events)

	if len(pending) != 1 || pending[0].ToolCallID != "tc_exec" || pending[0].ThreadID != "thr" || pending[0].RunID != "run" {
		t.Fatalf("pending = %+v", pending)
	}
}

func TestRendererPrintRunResultIncludesApprovalCommands(t *testing.T) {
	var stdout bytes.Buffer
	err := Renderer{Out: &stdout}.PrintRunResult(RunResult{
		ThreadID: "thr",
		RunID:    "run",
		Status:   "blocked_on_tool_approval",
		PendingApprovals: []PendingApproval{{
			ThreadID:   "thr",
			RunID:      "run",
			ToolCallID: "tc_exec",
			ToolName:   "sandbox.exec_command",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, expected := range []string{
		"run run blocked_on_tool_approval",
		"pending approvals:",
		"sandbox.exec_command tc_exec",
		"loomi approvals approve thr run tc_exec",
		"loomi approvals deny thr run tc_exec",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestRendererPrintEventIncludesToolArgumentsAndResult(t *testing.T) {
	var stdout bytes.Buffer
	renderer := Renderer{Out: &stdout}
	events := []RunEvent{
		{Sequence: 1, Type: "tool_call_approval_required", Metadata: map[string]any{"tool_call_id": "tc_read", "tool_name": "workspace.read", "arguments_summary": map[string]any{"path": "internal/cli/render.go", "limit": 128}}},
		{Sequence: 2, Type: "tool_call_succeeded", Metadata: map[string]any{"tool_call_id": "tc_read", "tool_name": "workspace.read", "result_summary": map[string]any{"path": "internal/cli/render.go", "truncated": false}}},
	}
	for _, event := range events {
		if err := renderer.PrintEvent(event); err != nil {
			t.Fatal(err)
		}
	}

	output := stdout.String()
	for _, expected := range []string{
		`0001 tool_call_approval_required workspace.read tc_read args=path=internal/cli/render.go limit=128`,
		`0002 tool_call_succeeded workspace.read tc_read result=path=internal/cli/render.go truncated=false`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestRendererPrintEventFormatsCommonToolResults(t *testing.T) {
	var stdout bytes.Buffer
	renderer := Renderer{Out: &stdout}
	events := []RunEvent{
		{Sequence: 1, Type: "tool_call_succeeded", Metadata: map[string]any{"tool_call_id": "tc_exec", "tool_name": "sandbox.exec_command", "result_summary": map[string]any{"exit_code": 0, "timed_out": false, "stdout": "ok\nsecond line"}}},
		{Sequence: 2, Type: "tool_call_succeeded", Metadata: map[string]any{"tool_call_id": "tc_browser", "tool_name": "browser.snapshot", "result_summary": map[string]any{"session_id": "sess_1", "title": "Example", "url": "https://example.com", "links": []any{"a", "b"}, "inputs": []any{"q"}}}},
	}
	for _, event := range events {
		if err := renderer.PrintEvent(event); err != nil {
			t.Fatal(err)
		}
	}

	output := stdout.String()
	for _, expected := range []string{
		`0001 tool_call_succeeded sandbox.exec_command tc_exec result=exit=0 timeout=false stdout="ok"`,
		`0002 tool_call_succeeded browser.snapshot tc_browser result=session=sess_1 title="Example" url=https://example.com links=2 inputs=1`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("stdout missing %q: %s", expected, output)
		}
	}
}

func TestClientApproveAndDenyToolCalls(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		switch r.URL.Path {
		case "/v1/threads/thr/runs/run/tool-calls/tc/approve":
			writeTestJSON(w, http.StatusOK, `{"tool_call":{"tool_call_id":"tc","approval_status":"approved"},"request_id":"req_approve"}`)
		case "/v1/threads/thr/runs/run/tool-calls/tc/deny":
			writeTestJSON(w, http.StatusOK, `{"tool_call":{"tool_call_id":"tc","approval_status":"denied"},"request_id":"req_deny"}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	approved, err := client.DecideToolCall(context.Background(), "thr", "run", "tc", "approve")
	if err != nil {
		t.Fatal(err)
	}
	denied, err := client.DecideToolCall(context.Background(), "thr", "run", "tc", "deny")
	if err != nil {
		t.Fatal(err)
	}

	if approved.ApprovalStatus != "approved" || denied.ApprovalStatus != "denied" {
		t.Fatalf("approved=%+v denied=%+v", approved, denied)
	}
	if strings.Join(calls, ",") != "POST /v1/threads/thr/runs/run/tool-calls/tc/approve,POST /v1/threads/thr/runs/run/tool-calls/tc/deny" {
		t.Fatalf("calls = %v", calls)
	}
}

func TestClientListsThreadsPersonasAndModelProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/threads":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s", r.Method)
			}
			writeTestJSON(w, http.StatusOK, `{"threads":[{"id":"thr_1","title":"New thread","mode":"work","lifecycle_status":"active","updated_at":"2026-05-26T12:00:00Z"}],"request_id":"req_threads"}`)
		case "/v1/personas":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s", r.Method)
			}
			writeTestJSON(w, http.StatusOK, `{"personas":[{"id":"persona_dev","slug":"developer","name":"Developer","source":"builtin","is_default":true,"active_version":2}],"request_id":"req_personas"}`)
		case "/v1/model-providers":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s", r.Method)
			}
			writeTestJSON(w, http.StatusOK, `{"providers":[{"id":"local_codex","family":"openai","model":"gpt-5-codex","status":"ready","execution_state":"executable"}],"request_id":"req_models"}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	threads, err := client.ListThreads(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	personas, err := client.ListPersonas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	providers, err := client.ListModelProviders(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(threads) != 1 || threads[0].Title != "New thread" {
		t.Fatalf("threads = %+v", threads)
	}
	if len(personas) != 1 || personas[0].Slug != "developer" || !personas[0].IsDefault {
		t.Fatalf("personas = %+v", personas)
	}
	if len(providers) != 1 || providers[0].ID != "local_codex" || providers[0].Model != "gpt-5-codex" {
		t.Fatalf("providers = %+v", providers)
	}
}

func TestLoadConfigFromEnvReadsConfigAndEnvOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"host":"http://config.local","mode":"chat","provider":"from_config","model":"model_config","persona":"persona_config"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOOMI_CONFIG", path)
	t.Setenv("LOOMI_PROVIDER", "from_env")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.Found || cfg.Path != path || cfg.Host != "http://config.local" || cfg.Mode != "chat" || cfg.Provider != "from_env" || cfg.Model != "model_config" || cfg.Persona != "persona_config" {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestConfigFileSetUnsetAndSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("LOOMI_CONFIG", path)

	cfg, err := LoadConfigFileFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Found || cfg.Path != path || cfg.Host != "" || cfg.Provider != "" {
		t.Fatalf("initial cfg = %+v", cfg)
	}
	if err := SetConfigValue(&cfg, "host", " http://loomi.local "); err != nil {
		t.Fatal(err)
	}
	if err := SetConfigValue(&cfg, "provider", "local_codex"); err != nil {
		t.Fatal(err)
	}
	if err := SaveConfigFile(cfg); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config mode = %v", info.Mode().Perm())
	}

	loaded, err := LoadConfigFileFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Found || loaded.Host != "http://loomi.local" || loaded.Provider != "local_codex" {
		t.Fatalf("loaded cfg = %+v", loaded)
	}
	if err := UnsetConfigValue(&loaded, "provider"); err != nil {
		t.Fatal(err)
	}
	if err := SaveConfigFile(loaded); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "provider") || !strings.Contains(string(raw), "http://loomi.local") {
		t.Fatalf("config file = %s", string(raw))
	}
}

func writeTestJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
