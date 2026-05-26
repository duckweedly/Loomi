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
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_glob_1", productdata.ToolNameWorkspaceGlob)
	approveGlob := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/tc_glob_1/approve", "")
	assertStatus(t, approveGlob.Code, http.StatusOK, approveGlob.Body.String())

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("glob resume ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_read_2", productdata.ToolNameWorkspaceRead)
	approveRead := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/tc_read_2/approve", "")
	assertStatus(t, approveRead.Code, http.StatusOK, approveRead.Body.String())

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("read resume ProcessOne ok=%v err=%v", ok, err)
	}
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
	for _, expected := range []string{productdata.EventToolCallApprovalRequired, productdata.EventToolCallSucceeded, `"tool_call_id":"tc_read_2"`, `"loop_index":2`, `"loop_max":3`, `"model_phase":"continuation"`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m22 bounded loop events", root, "fixture-secret", ".env", "secrets/token")
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
