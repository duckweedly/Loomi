package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM24SandboxExecCommandSmoke(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	const toolCallID = "tc_m24_exec"

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved sandbox exec only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameSandboxExecCommand},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m21WorkspaceProvider{toolName: productdata.ToolNameSandboxExecCommand, toolCallID: toolCallID, args: map[string]any{"argv": []any{"ls", "."}, "cwd": "."}}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m24_sandbox_exec"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M24 sandbox exec smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Run sandbox command","client_message_id":"m24-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("pre-approval call = %+v", call)
	}

	approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
	assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("second ProcessOne ok=%v err=%v", ok, err)
	}
	finalCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if finalCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || finalCall.ResultSummary["operation"] != "exec_command" || finalCall.ResultSummary["scope"] != "bounded_command" {
		t.Fatalf("final call = %+v", finalCall)
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{productdata.EventToolCallApprovalRequired, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, `"tool_name":"sandbox.exec_command"`, `"scope":"bounded_command"`} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m24 sandbox exec events", root, "/Users/", "TOKEN", "secret")
}

func TestM24SandboxProcessLoopSmoke(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	if err := syscall.Mkfifo(filepath.Join(root, "stream.txt"), 0o600); err != nil {
		t.Fatal(err)
	}

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:         "default",
		Name:         "Default",
		Description:  "Default",
		SystemPrompt: "Use approved sandbox process tools only.",
		ModelRoute:   productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{
			productdata.ToolNameSandboxStartProcess,
			productdata.ToolNameSandboxContinueProcess,
			productdata.ToolNameSandboxTerminateProcess,
		},
		ReasoningMode: "balanced",
		BudgetSummary: "small",
		Version:       "1",
		IsDefault:     true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m24SandboxProcessLoopProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m24_sandbox_process_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M24 sandbox process smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Run sandbox process","client_message_id":"m24-process-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	for _, step := range []struct {
		id   string
		name string
	}{
		{"tc_process_start_1", productdata.ToolNameSandboxStartProcess},
		{"tc_process_continue_2", productdata.ToolNameSandboxContinueProcess},
		{"tc_process_terminate_3", productdata.ToolNameSandboxTerminateProcess},
	} {
		if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
			t.Fatalf("%s ProcessOne ok=%v err=%v", step.id, ok, err)
		}
		if step.id == "tc_process_continue_2" {
			wrote := make(chan struct{})
			go func() {
				_ = os.WriteFile(filepath.Join(root, "stream.txt"), []byte("hello process\n"), 0o600)
				close(wrote)
			}()
			select {
			case <-wrote:
			case <-time.After(2 * time.Second):
				t.Fatal("timed out writing sandbox process input")
			}
		}
		call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, step.id)
		if err != nil {
			t.Fatal(err)
		}
		if call.ToolName != step.name || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
			t.Fatalf("blocked call = %+v", call)
		}
		approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+step.id+"/approve", "")
		assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	}

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("final ProcessOne ok=%v err=%v", ok, err)
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	continueCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_process_continue_2")
	if err != nil {
		t.Fatal(err)
	}
	if continueCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || continueCall.ResultSummary["operation"] != "continue_process" || !strings.Contains(stringValue(continueCall.ResultSummary, "stdout"), "hello process") {
		t.Fatalf("continue call = %+v", continueCall)
	}
	terminateCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_process_terminate_3")
	if err != nil {
		t.Fatal(err)
	}
	if terminateCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || terminateCall.ResultSummary["operation"] != "terminate_process" || terminateCall.ResultSummary["scope"] != "bounded_process" {
		t.Fatalf("terminate call = %+v", terminateCall)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{`"tool_name":"sandbox.start_process"`, `"tool_name":"sandbox.continue_process"`, `"tool_name":"sandbox.terminate_process"`, `"scope":"bounded_process"`, `"loop_index":3`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m24 sandbox process events", root, "/Users/", "TOKEN", "secret")
}

type m24SandboxProcessLoopProvider struct {
	calls int
}

func (p *m24SandboxProcessLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m24SandboxProcessLoopProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameSandboxStartProcess, Metadata: map[string]any{"tool_call_id": "tc_process_start_1", "arguments_summary": map[string]any{"argv": []any{"cat", "stream.txt"}, "cwd": ".", "timeout_ms": 100000, "max_output_bytes": 4096}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameSandboxContinueProcess, Metadata: map[string]any{"tool_call_id": "tc_process_continue_2", "arguments_summary": map[string]any{"process_id": lastM24ProcessID(request)}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameSandboxTerminateProcess, Metadata: map[string]any{"tool_call_id": "tc_process_terminate_3", "arguments_summary": map[string]any{"process_id": lastM24ProcessID(request)}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "Sandbox process "}, {Type: productruntime.ProviderEventCompleted, Text: "Sandbox process loop complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func lastM24ProcessID(request productruntime.ProviderRequest) string {
	for index := len(request.Messages) - 1; index >= 0; index-- {
		message := request.Messages[index]
		if message.Role != productruntime.ProviderMessageRoleToolResult {
			continue
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(message.Content), &result); err == nil {
			if processID, ok := result["process_id"].(string); ok {
				return processID
			}
		}
	}
	return ""
}
