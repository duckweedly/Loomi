package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM22BoundedAgentLoopWorkspaceSmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved workspace read tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWorkspaceGlob, productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceRead},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m22WorkspaceLoopProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m22_workspace_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M22 workspace loop","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Inspect workspace","client_message_id":"m22-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("initial ProcessOne ok=%v err=%v", ok, err)
	}
	drainM22Worker(t, worker, svc, ident, runID)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_glob_1", productdata.ToolNameWorkspaceGlob)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_read_2", productdata.ToolNameWorkspaceRead)
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	messages, err := svc.ListMessages(context.Background(), ident, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "M22 workspace loop complete." {
		t.Fatalf("messages = %+v", messages)
	}
	readCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_read_2")
	if err != nil {
		t.Fatal(err)
	}
	if readCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || stringValue(readCall.ResultSummary, "content") != "needle" {
		t.Fatalf("read call = %+v", readCall)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{productdata.EventToolCallSucceeded, `"tool_call_id":"tc_read_2"`, `"loop_index":2`, `"loop_max":6`, `"model_phase":"continuation"`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	if strings.Contains(eventsBody, productdata.EventToolCallApprovalRequired) {
		t.Fatalf("read-only workspace loop should not require approval: %s", eventsBody)
	}
	assertBodyExcludes(t, eventsBody, "m22 bounded loop events", root, "fixture-secret", ".env", "secrets/token")
}

func TestM22CodeAgentReadEditExecReadLoopSmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved code-agent tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspaceEdit, productdata.ToolNameSandboxExecCommand},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m22CodeAgentLoopProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m22_code_agent_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M22 code-agent loop","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Patch and verify workspace file","client_message_id":"m22-code-agent-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("read-before ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_read_before_1", productdata.ToolNameWorkspaceRead)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_edit_2", productdata.ToolNameWorkspaceEdit)
	approveEdit := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/tc_edit_2/approve", "")
	assertStatus(t, approveEdit.Code, http.StatusOK, approveEdit.Body.String())

	for _, step := range []struct {
		id   string
		name string
	}{
		{"tc_exec_3", productdata.ToolNameSandboxExecCommand},
	} {
		if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
			run, _ := svc.GetRun(context.Background(), ident, runID)
			errorCode := ""
			errorMessage := ""
			if run.ErrorCode != nil {
				errorCode = *run.ErrorCode
			}
			if run.ErrorMessage != nil {
				errorMessage = *run.ErrorMessage
			}
			t.Fatalf("%s ProcessOne ok=%v err=%v status=%s error=%s message=%s", step.id, ok, err, run.Status, errorCode, errorMessage)
		}
		assertM22ToolBlocked(t, svc, threadID, runID, step.id, step.name)
		approve := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+step.id+"/approve", "")
		assertStatus(t, approve.Code, http.StatusOK, approve.Body.String())
	}

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("read-after ProcessOne ok=%v err=%v", ok, err)
	}
	drainM22Worker(t, worker, svc, ident, runID)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_read_after_4", productdata.ToolNameWorkspaceRead)

	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	messages, err := svc.ListMessages(context.Background(), ident, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "Code-agent loop complete." {
		t.Fatalf("messages = %+v", messages)
	}
	readAfter, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_read_after_4")
	if err != nil {
		t.Fatal(err)
	}
	if readAfter.ExecutionStatus != productdata.ToolCallExecutionSucceeded || !strings.Contains(stringValue(readAfter.ResultSummary, "content"), "thread") {
		t.Fatalf("read after call = %+v", readAfter)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{`"tool_call_id":"tc_read_after_4"`, `"loop_index":4`, `"loop_max":6`, `"tool_name":"sandbox.exec_command"`, productdata.EventWorkTodoUpdated, "Run validation command", productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m22 code-agent loop events", root, "fixture-secret", ".env", "secrets/token")
}

func assertM22ToolBlocked(t *testing.T, svc productdata.Service, threadID string, runID string, toolCallID string, toolName string) {
	t.Helper()
	call, err := svc.GetToolCall(context.Background(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != toolName || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("call = %+v", call)
	}
}

func assertM22ToolSucceeded(t *testing.T, svc productdata.Service, threadID string, runID string, toolCallID string, toolName string) {
	t.Helper()
	call, err := svc.GetToolCall(context.Background(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != toolName || call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("call = %+v", call)
	}
}

func drainM22Worker(t *testing.T, worker *productruntime.Worker, svc productdata.Service, ident identity.LocalIdentity, runID string) {
	t.Helper()
	for i := 0; i < 6; i++ {
		run, err := svc.GetRun(context.Background(), ident, runID)
		if err != nil {
			t.Fatal(err)
		}
		if productdata.IsRunTerminal(run.Status) || run.Status == productdata.RunStatusBlockedOnToolApproval {
			return
		}
		ok, err := worker.ProcessOne(context.Background())
		if err != nil {
			t.Fatalf("drain ProcessOne ok=%v err=%v", ok, err)
		}
		if !ok {
			return
		}
	}
}

type m22WorkspaceLoopProvider struct {
	calls int
}

func (p *m22WorkspaceLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m22WorkspaceLoopProvider) Stream(_ context.Context, _ productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceGlob, Metadata: map[string]any{"tool_call_id": "tc_glob_1", "arguments_summary": map[string]any{"pattern": "src/*.txt", "limit": 10}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_2", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 6}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M22 workspace "}, {Type: productruntime.ProviderEventCompleted, Text: "M22 workspace loop complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

type m22CodeAgentLoopProvider struct {
	calls int
}

func (p *m22CodeAgentLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m22CodeAgentLoopProvider) Stream(_ context.Context, _ productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_before_1", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 32}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceEdit, Metadata: map[string]any{"tool_call_id": "tc_edit_2", "arguments_summary": map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "thread\n"}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameSandboxExecCommand, Metadata: map[string]any{"tool_call_id": "tc_exec_3", "arguments_summary": map[string]any{"argv": []any{"cat", "src/notes.txt"}, "timeout_ms": 1000, "max_output_bytes": 4096}}}}
	case 4:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_after_4", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 64}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "Code-agent "}, {Type: productruntime.ProviderEventCompleted, Text: "Code-agent loop complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}
