package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM29AgentSpawnApproveExecuteFinalSmoke(t *testing.T) {
	const toolCallID = "tc_m29_agent_spawn"
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := syncM29AgentPersona(svc, ident); err != nil {
		t.Fatal(err)
	}
	provider := &m21WorkspaceProvider{toolName: productdata.ToolNameAgentSpawn, toolCallID: toolCallID, args: map[string]any{"role": "reviewer", "goal": "Review artifact runtime"}}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m29_agent_spawn"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadID, runID := startM29AgentRun(t, srv)
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolBlocked(t, svc, threadID, runID, toolCallID, productdata.ToolNameAgentSpawn)
	approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
	assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("second ProcessOne ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "spawn" || call.ResultSummary["scope"] != "agent" || call.ResultSummary["role"] != "reviewer" || call.ResultSummary["autonomous_execution"] != false {
		t.Fatalf("call = %+v", call)
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, productdata.ListAgentTasksInput{ThreadID: threadID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Role != "reviewer" || tasks[0].Status != productdata.AgentTaskStatusSpawned {
		t.Fatalf("tasks = %+v", tasks)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{productdata.EventToolCallApprovalRequired, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, `"tool_name":"agent.spawn"`, `"scope":"agent"`, `"tool_group":"agent"`, `"autonomous_execution":false`} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m29 agent spawn events", "sk-secret", "Authorization", "/Users/", "raw_result")
}

func TestM29AgentSpawnListCompleteLoopSmoke(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := syncM29AgentPersona(svc, ident); err != nil {
		t.Fatal(err)
	}
	provider := &m29AgentLoopProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m29_agent_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadID, runID := startM29AgentRun(t, srv)
	for _, toolCallID := range []string{"tc_agent_spawn_1", "tc_agent_list_2", "tc_agent_complete_3"} {
		if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
			t.Fatalf("%s request/execution ProcessOne ok=%v err=%v", toolCallID, ok, err)
		}
		approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
		assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	}
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("final ProcessOne ok=%v err=%v", ok, err)
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted || provider.calls != 4 {
		t.Fatalf("run=%+v provider calls=%d", run, provider.calls)
	}
	completeCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_agent_complete_3")
	if err != nil {
		t.Fatal(err)
	}
	if completeCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || completeCall.ResultSummary["operation"] != "complete" || completeCall.ResultSummary["status"] != string(productdata.AgentTaskStatusCompleted) || completeCall.ResultSummary["result_summary"] != "No safety issue found" {
		t.Fatalf("complete call = %+v", completeCall)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{`"tool_name":"agent.spawn"`, `"tool_name":"agent.list"`, `"tool_name":"agent.complete"`, `"loop_index":3`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m29 agent loop events", "sk-secret", "Authorization", "/Users/", "raw_result")
}

func syncM29AgentPersona(svc *productdata.MemoryService, ident identity.LocalIdentity) (productdata.PersonaSyncResult, error) {
	return svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved agent coordination tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameAgentSpawn, productdata.ToolNameAgentList, productdata.ToolNameAgentComplete},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}})
}

func startM29AgentRun(t *testing.T, srv http.Handler) (string, string) {
	t.Helper()
	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M29 agent smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Coordinate agent task","client_message_id":"m29-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	return threadID, decodeStringField(t, runRes.Body.Bytes(), "run", "id")
}

type m29AgentLoopProvider struct {
	calls  int
	taskID string
}

func (p *m29AgentLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m29AgentLoopProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	if p.calls > 1 && p.taskID == "" {
		p.taskID = agentTaskIDFromProviderRequest(request)
	}
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameAgentSpawn, Metadata: map[string]any{"tool_call_id": "tc_agent_spawn_1", "arguments_summary": map[string]any{"role": "reviewer", "goal": "Review implementation"}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameAgentList, Metadata: map[string]any{"tool_call_id": "tc_agent_list_2", "arguments_summary": map[string]any{"limit": 10}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameAgentComplete, Metadata: map[string]any{"tool_call_id": "tc_agent_complete_3", "arguments_summary": map[string]any{"task_id": p.taskID, "result_summary": "No safety issue found"}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M29 agent "}, {Type: productruntime.ProviderEventCompleted, Text: "M29 agent complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func agentTaskIDFromProviderRequest(request productruntime.ProviderRequest) string {
	for _, message := range request.Messages {
		if message.Role != productruntime.ProviderMessageRoleToolResult {
			continue
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(message.Content), &result); err != nil {
			continue
		}
		if id, _ := result["task_id"].(string); id != "" {
			return id
		}
	}
	return ""
}
