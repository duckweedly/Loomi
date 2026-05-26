package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM25LSPReadonlyApproveExecuteFinalSmoke(t *testing.T) {
	root := createM25LSPFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	const toolCallID = "tc_lsp_smoke"
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved LSP read-only tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameCurrentTime, productdata.ToolNameLSPSymbols, productdata.ToolNameLSPReferences, productdata.ToolNameLSPDiagnostics},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m25LSPProvider{toolCallID: toolCallID}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m25_lsp_smoke"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M25 LSP smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Find LSP symbols","client_message_id":"m25-lsp-user-message"}`)
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
	if call.ToolName != productdata.ToolNameLSPSymbols || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
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
	if finalCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || finalCall.ResultSummary["operation"] != "symbols" || finalCall.ResultSummary["scope"] != "lsp" || finalCall.ResultSummary["count"] != 1 {
		t.Fatalf("final call = %+v", finalCall)
	}
	runGet := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+runID, "")
	assertStatus(t, runGet.Code, http.StatusOK, runGet.Body.String())
	if !strings.Contains(runGet.Body.String(), `"status":"completed"`) {
		t.Fatalf("run response = %s", runGet.Body.String())
	}
	eventsRes := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+runID+"/events", "")
	assertStatus(t, eventsRes.Code, http.StatusOK, eventsRes.Body.String())
	eventsBody := eventsRes.Body.String()
	for _, expected := range []string{productdata.EventToolCallApprovalRequired, productdata.EventToolCallApproved, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, productdata.EventRunCompleted, `"tool_name":"lsp.symbols"`} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %q: %s", expected, eventsBody)
		}
	}
	renderedResult, err := json.Marshal(finalCall.ResultSummary)
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{root, "/Users/xuean", ".ssh", "id_ed25519"} {
		if strings.Contains(string(renderedResult), leaked) || strings.Contains(eventsBody, leaked) {
			t.Fatalf("LSP smoke leaked %q\nresult=%s\nevents=%s", leaked, string(renderedResult), eventsBody)
		}
	}
	if provider.calls != 2 || len(provider.continuationRequest.Messages) != 3 || provider.continuationRequest.Messages[2].Role != productruntime.ProviderMessageRoleToolResult {
		t.Fatalf("provider calls=%d continuation=%+v", provider.calls, provider.continuationRequest.Messages)
	}
}

type m25LSPProvider struct {
	toolCallID          string
	calls               int
	continuationRequest productruntime.ProviderRequest
}

func (p *m25LSPProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m25LSPProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameLSPSymbols, Metadata: map[string]any{"tool_call_id": p.toolCallID, "arguments_summary": map[string]any{"path": "src/main.go", "query": "Tool", "limit": 10}}}}
	if p.calls == 2 {
		p.continuationRequest = request
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M25 LSP "}, {Type: productruntime.ProviderEventCompleted, Text: "M25 LSP complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func createM25LSPFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n\ntype ToolBroker struct{}\n\nfunc useToolBroker() { _ = ToolBroker{} }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
