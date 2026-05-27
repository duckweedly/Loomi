package runtime

import (
	"context"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestGatewayRecordsApprovalRequiredMCPToolCallWhenDiscoveredAndPersonaAllowed(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	syncPersonaWithMCPTool(t, svc, "mcp.local-search.search")
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "MCP approval", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	appendMCPDiscovery(t, svc, run.ID, "local-search", "mcp.local-search.search")
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: "mcp.local-search.search", Metadata: map[string]any{"tool_call_id": "tc_mcp_1", "arguments_summary": map[string]any{"query": "status", "api_key": "sk-secret"}}}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_mcp_gate"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusBlockedOnToolApproval {
		t.Fatalf("run = %+v", got)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_mcp_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != "mcp.local-search.search" || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("call = %+v", call)
	}
	if call.CandidateSchemaHash != "sha256:test-local-search" {
		t.Fatalf("candidate schema hash = %q", call.CandidateSchemaHash)
	}
	if call.ArgumentsSummary["api_key"] == "sk-secret" {
		t.Fatalf("arguments leaked: %+v", call.ArgumentsSummary)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var required productdata.RunEvent
	for _, event := range events {
		if event.Type == productdata.EventToolCallApprovalRequired {
			required = event
		}
		if event.Type == productdata.EventToolCallExecuting || event.Type == productdata.EventToolCallSucceeded {
			t.Fatalf("MCP executed before approval: %+v", events)
		}
	}
	if required.Metadata["tool_source"] != "mcp" || required.Metadata["server_slug"] != "local-search" {
		t.Fatalf("approval metadata = %+v", required.Metadata)
	}
}

func TestGatewayRejectsMCPToolWithoutDiscoveryOrPersonaAllowlist(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "MCP reject", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: "mcp.local-search.search", Metadata: map[string]any{"tool_call_id": "tc_mcp_1"}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_mcp_1"); err == nil {
		t.Fatal("GetToolCall() error = nil, want rejection before projection")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || !strings.Contains(*got.ErrorCode, "tool_call") {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRejectsPersonaAllowedMCPToolWithoutDiscovery(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	syncPersonaWithMCPTool(t, svc, "mcp.local-search.search")
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "MCP reject undiscovered", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: "mcp.local-search.search", Metadata: map[string]any{"tool_call_id": "tc_mcp_1"}}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_mcp_no_discovery"

	_, _ = worker.ProcessOne(context.Background())

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_mcp_1"); err == nil {
		t.Fatal("GetToolCall() error = nil, want rejection without discovery")
	}
}

func syncPersonaWithMCPTool(t *testing.T, svc *productdata.MemoryService, toolName string) {
	t.Helper()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), identity.LocalDevIdentity(), []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use tools safely.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameCurrentTime, toolName},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
}

func appendMCPDiscovery(t *testing.T, svc productdata.Service, runID string, serverSlug string, toolName string) {
	t.Helper()
	result := MCPDiscoveryResult{
		ServerSlug: serverSlug,
		Status:     MCPDiscoverySucceeded,
		Candidates: []MCPToolCandidate{{
			ServerSlug:  serverSlug,
			MCPName:     mcpToolName(toolName),
			Name:        toolName,
			SchemaHash:  "sha256:test-local-search",
			Description: "Test MCP tool",
		}},
	}
	if _, err := svc.AppendRunEvent(context.Background(), identity.LocalDevIdentity(), runID, productdata.AppendRunEventInput{
		Category: productdata.RunEventCategoryProgress,
		Type:     "mcp_discovery_succeeded",
		Summary:  "MCP discovery succeeded",
		Metadata: MCPDiscoveryEventMetadata(result),
	}); err != nil {
		t.Fatal(err)
	}
}
