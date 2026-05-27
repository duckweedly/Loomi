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

func TestM28ArtifactCreateApproveExecuteFinalSmoke(t *testing.T) {
	const toolCallID = "tc_m28_artifact_create"
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := syncM28ArtifactPersona(svc, ident); err != nil {
		t.Fatal(err)
	}
	provider := &m21WorkspaceProvider{toolName: productdata.ToolNameArtifactCreateText, toolCallID: toolCallID, args: map[string]any{"title": "M28 Notes", "content": "artifact content"}}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m28_artifact_create"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadID, runID := startM28ArtifactRun(t, srv)
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolBlocked(t, svc, threadID, runID, toolCallID, productdata.ToolNameArtifactCreateText)
	approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
	assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("second ProcessOne ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "create_text" || call.ResultSummary["scope"] != "artifact" || call.ResultSummary["title"] != "M28 Notes" {
		t.Fatalf("call = %+v", call)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{productdata.EventToolCallApprovalRequired, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, `"tool_name":"artifact.create_text"`, `"scope":"artifact"`} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m28 artifact create events", "sk-secret", "Authorization", "/Users/", "raw_result")
}

func TestM28ArtifactCreateListReadLoopSmoke(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := syncM28ArtifactPersona(svc, ident); err != nil {
		t.Fatal(err)
	}
	provider := &m28ArtifactLoopProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m28_artifact_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadID, runID := startM28ArtifactRun(t, srv)
	for _, toolCallID := range []string{"tc_artifact_create_1", "tc_artifact_list_2", "tc_artifact_read_3"} {
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
	readCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_artifact_read_3")
	if err != nil {
		t.Fatal(err)
	}
	if readCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || readCall.ResultSummary["operation"] != "read" || readCall.ResultSummary["text_excerpt"] != "artifact content" {
		t.Fatalf("read call = %+v", readCall)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{`"tool_name":"artifact.create_text"`, `"tool_name":"artifact.list"`, `"tool_name":"artifact.read"`, `"loop_index":3`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m28 artifact loop events", "sk-secret", "Authorization", "/Users/", "raw_result")
}

func syncM28ArtifactPersona(svc *productdata.MemoryService, ident identity.LocalIdentity) (productdata.PersonaSyncResult, error) {
	return svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved artifact tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameArtifactCreateText, productdata.ToolNameArtifactList, productdata.ToolNameArtifactRead},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}})
}

func startM28ArtifactRun(t *testing.T, srv http.Handler) (string, string) {
	t.Helper()
	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M28 artifact smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Create artifact","client_message_id":"m28-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	return threadID, decodeStringField(t, runRes.Body.Bytes(), "run", "id")
}

type m28ArtifactLoopProvider struct {
	calls      int
	artifactID string
}

func (p *m28ArtifactLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m28ArtifactLoopProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	if p.calls > 1 && p.artifactID == "" {
		p.artifactID = artifactIDFromProviderRequest(request)
	}
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameArtifactCreateText, Metadata: map[string]any{"tool_call_id": "tc_artifact_create_1", "arguments_summary": map[string]any{"title": "M28 Notes", "content": "artifact content"}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameArtifactList, Metadata: map[string]any{"tool_call_id": "tc_artifact_list_2", "arguments_summary": map[string]any{"limit": 10}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameArtifactRead, Metadata: map[string]any{"tool_call_id": "tc_artifact_read_3", "arguments_summary": map[string]any{"artifact_id": p.artifactID}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M28 artifact "}, {Type: productruntime.ProviderEventCompleted, Text: "M28 artifact complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func artifactIDFromProviderRequest(request productruntime.ProviderRequest) string {
	for _, message := range request.Messages {
		if message.Role != productruntime.ProviderMessageRoleToolResult {
			continue
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(message.Content), &result); err != nil {
			continue
		}
		if id, _ := result["artifact_id"].(string); id != "" {
			return id
		}
	}
	return ""
}
