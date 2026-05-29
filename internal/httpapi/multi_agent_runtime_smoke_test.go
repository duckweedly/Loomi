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

func TestM98CodeAgentParallelReadDelegateChildFinalSmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := syncM98CodeAgentPersona(svc, ident); err != nil {
		t.Fatal(err)
	}
	provider := &m98ParallelDelegateProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m98_parallel_delegate"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M98 parallel delegate","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	provider.parentThreadID = threadID
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Inspect files in parallel, delegate review, then summarize","client_message_id":"m98-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		run, _ := svc.GetRun(context.Background(), ident, runID)
		events, _ := svc.ListRunEvents(context.Background(), ident, runID, 0)
		t.Fatalf("parallel read ProcessOne ok=%v err=%v run=%+v events=%+v", ok, err, run, events)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m98_grep_1", productdata.ToolNameWorkspaceGrep)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m98_read_2", productdata.ToolNameWorkspaceRead)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_m98_spawn_3", productdata.ToolNameAgentSpawn)
	if provider.parentCalls != 2 {
		t.Fatalf("parent calls after parallel read = %d", provider.parentCalls)
	}
	if len(provider.parentRequests) < 2 || !providerRequestHasToolResults(provider.parentRequests[1], "tc_m98_grep_1", "tc_m98_read_2") {
		t.Fatalf("parent continuation did not receive both parallel results: %+v", provider.parentRequests)
	}
	approveM75Tool(t, srv, threadID, runID, "tc_m98_spawn_3")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("spawn ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m98_spawn_3", productdata.ToolNameAgentSpawn)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_m98_delegate_4", productdata.ToolNameAgentDelegate)
	approveM75Tool(t, srv, threadID, runID, "tc_m98_delegate_4")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("delegate ProcessOne ok=%v err=%v", ok, err)
	}
	delegateCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_m98_delegate_4")
	if err != nil {
		t.Fatal(err)
	}
	if delegateCall.ExecutionStatus != productdata.ToolCallExecutionExecuting {
		t.Fatalf("delegate should wait for child run: %+v", delegateCall)
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, productdata.ListAgentTasksInput{ThreadID: threadID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ChildRunID == "" || tasks[0].ChildThreadID == "" || tasks[0].ParentToolCallID != "tc_m98_delegate_4" {
		t.Fatalf("tasks = %+v", tasks)
	}
	if provider.parentCalls != 3 || provider.childCalls != 0 {
		t.Fatalf("unexpected provider calls after delegate: parent=%d child=%d", provider.parentCalls, provider.childCalls)
	}

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("child ProcessOne ok=%v err=%v", ok, err)
	}
	childRun, err := svc.GetRun(context.Background(), ident, tasks[0].ChildRunID)
	if err != nil {
		t.Fatal(err)
	}
	if childRun.Status != productdata.RunStatusCompleted || provider.childCalls != 1 {
		t.Fatalf("child run=%+v childCalls=%d", childRun, provider.childCalls)
	}

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("parent resume ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m98_delegate_4", productdata.ToolNameAgentDelegate)
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted || provider.parentCalls != 4 {
		t.Fatalf("parent run=%+v parentCalls=%d", run, provider.parentCalls)
	}
	messages, err := svc.ListMessages(context.Background(), ident, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Content != "Parallel read and child review complete." {
		t.Fatalf("messages = %+v", messages)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, runID, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertM76TerminalToolEventsOnce(t, events, []string{"tc_m98_grep_1", "tc_m98_read_2", "tc_m98_spawn_3", "tc_m98_delegate_4"})
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{`"tool_call_id":"tc_m98_grep_1"`, `"tool_call_id":"tc_m98_read_2"`, `"tool_call_id":"tc_m98_delegate_4"`, productdata.EventAgentChildRunStarted, `"child_run_id":"` + tasks[0].ChildRunID + `"`, `"parent_tool_call_id":"tc_m98_delegate_4"`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m98 parallel delegate events", root, "sk-secret", "Authorization", "/Users/", "raw_result")
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

func syncM98CodeAgentPersona(svc *productdata.MemoryService, ident identity.LocalIdentity) (productdata.PersonaSyncResult, error) {
	return svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:         "default",
		Name:         "Default",
		Description:  "Default",
		SystemPrompt: "Use parallel read-only workspace tools, then delegate a child review before final summary.",
		ModelRoute:   productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{
			productdata.ToolNameWorkspaceGrep,
			productdata.ToolNameWorkspaceRead,
			productdata.ToolNameAgentSpawn,
			productdata.ToolNameAgentDelegate,
		},
		ReasoningMode: "balanced",
		BudgetSummary: "small",
		Version:       "1",
		IsDefault:     true,
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

type m98ParallelDelegateProvider struct {
	parentThreadID string
	parentCalls    int
	childCalls     int
	taskID         string
	parentRequests []productruntime.ProviderRequest
}

func (p *m98ParallelDelegateProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m98ParallelDelegateProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	if request.ThreadID != p.parentThreadID {
		p.childCalls++
		events := []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "Child review: "}, {Type: productruntime.ProviderEventCompleted, Text: "Child review: no issues."}}
		return providerEvents(events), nil
	}
	p.parentCalls++
	p.parentRequests = append(p.parentRequests, request)
	if p.parentCalls > 1 && p.taskID == "" {
		p.taskID = agentTaskIDFromProviderRequest(request)
	}
	events := []productruntime.ProviderEvent{}
	switch p.parentCalls {
	case 1:
		events = []productruntime.ProviderEvent{
			{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceGrep, Metadata: map[string]any{"tool_call_id": "tc_m98_grep_1", "arguments_summary": map[string]any{"query": "needle", "path": "src", "limit": 10}}},
			{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_m98_read_2", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 128}}},
		}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameAgentSpawn, Metadata: map[string]any{"tool_call_id": "tc_m98_spawn_3", "arguments_summary": map[string]any{"role": "reviewer", "goal": "Review the parallel workspace findings"}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameAgentDelegate, Metadata: map[string]any{"tool_call_id": "tc_m98_delegate_4", "arguments_summary": map[string]any{"task_id": p.taskID}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "Parallel read and child review "}, {Type: productruntime.ProviderEventCompleted, Text: "Parallel read and child review complete."}}
	}
	return providerEvents(events), nil
}

func providerEvents(events []productruntime.ProviderEvent) <-chan productruntime.ProviderEvent {
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch
}

func providerRequestHasToolResults(request productruntime.ProviderRequest, toolCallIDs ...string) bool {
	seen := map[string]bool{}
	for _, message := range request.Messages {
		if message.Role == productruntime.ProviderMessageRoleToolResult {
			seen[message.ToolCallID] = true
		}
	}
	for _, toolCallID := range toolCallIDs {
		if !seen[toolCallID] {
			return false
		}
	}
	return true
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
