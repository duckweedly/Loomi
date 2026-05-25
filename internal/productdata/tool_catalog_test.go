package productdata

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
)

func TestToolCatalogIncludesBuiltinCurrentTime(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	tool := catalogToolByName(tools, ToolNameCurrentTime)
	if tool.Name != ToolNameCurrentTime {
		t.Fatalf("catalog = %+v, missing current time", tools)
	}
	if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupRuntime || tool.RiskLevel != ToolRiskLow || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("current time catalog metadata = %+v", tool)
	}
	if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.InputSchemaHash == "" {
		t.Fatalf("current time execution metadata = %+v", tool)
	}
}

func TestToolCatalogIncludesDiscoveredMCPCandidate(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	run := createCatalogRun(t, svc, ident)
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{
		Category: RunEventCategoryProgress,
		Type:     "mcp_discovery_succeeded",
		Summary:  "MCP discovery succeeded",
		Metadata: map[string]any{
			"server_slug":             "local-smoke",
			"status":                  "succeeded",
			"candidate_names":         []string{"mcp.local-smoke.echo"},
			"candidate_schema_hashes": map[string]any{"mcp.local-smoke.echo": "sha256:test-schema"},
			"command":                 "/home/xuean/private/bin/mcp",
			"api_key":                 "sk-secret-catalog",
		},
	}); err != nil {
		t.Fatal(err)
	}

	tools, err := svc.ListToolCatalog(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	tool := catalogToolByName(tools, "mcp.local-smoke.echo")
	if tool.Source != ToolCatalogSourceMCP || tool.Group != ToolCatalogGroupMCP || tool.InputSchemaHash != "sha256:test-schema" || tool.ExecutionState != ToolExecutionStateNonExecutable {
		t.Fatalf("mcp catalog metadata = %+v", tool)
	}
	encoded := fmt.Sprint(tool)
	for _, leaked := range []string{"sk-secret-catalog", "/home/xuean", "command"} {
		if strings.Contains(encoded, leaked) {
			t.Fatalf("catalog leaked %q: %+v", leaked, tool)
		}
	}
}

func TestToolCatalogUsesLatestSuccessfulMCPDiscoveryHash(t *testing.T) {
	oldEvent := RunEvent{ID: "evt_1", Type: "mcp_discovery_succeeded", Sequence: 1, CreatedAt: time.Date(2026, 5, 25, 1, 0, 0, 0, time.UTC), Metadata: map[string]any{"server_slug": "local-smoke", "status": "succeeded", "candidate_names": []string{"mcp.local-smoke.echo"}, "candidate_schema_hashes": map[string]any{"mcp.local-smoke.echo": "sha256:old"}}}
	newEvent := RunEvent{ID: "evt_2", Type: "mcp_discovery_succeeded", Sequence: 2, CreatedAt: time.Date(2026, 5, 25, 2, 0, 0, 0, time.UTC), Metadata: map[string]any{"server_slug": "local-smoke", "status": "succeeded", "candidate_names": []string{"mcp.local-smoke.echo"}, "candidate_schema_hashes": map[string]any{"mcp.local-smoke.echo": "sha256:new"}}}

	tools := ToolCatalogFromEvents([]RunEvent{newEvent, oldEvent})
	tool := catalogToolByName(tools, "mcp.local-smoke.echo")
	if tool.InputSchemaHash != "sha256:new" || tool.ExecutionState != ToolExecutionStateExecutable {
		t.Fatalf("execution catalog should use latest schema hash: %+v", tool)
	}

	safeTools := SafeToolCatalogFromEvents([]RunEvent{oldEvent, newEvent})
	safeTool := catalogToolByName(safeTools, "mcp.local-smoke.echo")
	if safeTool.InputSchemaHash != "sha256:new" || safeTool.ExecutionState != ToolExecutionStateNonExecutable {
		t.Fatalf("safe catalog should use latest schema hash without claiming executor availability: %+v", safeTool)
	}
}

func TestPrepareRunContextFiltersEnabledToolsThroughPersonaAndDiscovery(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Tools", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "use tools"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{{
		Slug:             "tools",
		Name:             "Tools",
		Description:      "Tool persona",
		SystemPrompt:     "Use allowed tools.",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{ToolNameCurrentTime, "mcp.local-smoke.echo", "mcp.local-smoke.blocked"},
		ReasoningMode:    "balanced",
		BudgetSummary:    "test",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: "mcp_discovery_succeeded", Summary: "MCP discovery succeeded", Metadata: map[string]any{"server_slug": "local-smoke", "status": "succeeded", "candidate_names": []string{"mcp.local-smoke.echo"}, "candidate_schema_hashes": map[string]any{"mcp.local-smoke.echo": "sha256:test-schema"}}}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_tools", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}

	contextData, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	allowed := catalogResolutionByName(contextData.EnabledTools, "mcp.local-smoke.echo")
	if allowed.ExecutionState != string(ToolExecutionStateExecutable) || allowed.InputSchemaHash != "sha256:test-schema" {
		t.Fatalf("allowed MCP resolution = %+v", allowed)
	}
	blocked := catalogResolutionByName(contextData.EnabledTools, "mcp.local-smoke.blocked")
	if blocked.Name != "" {
		t.Fatalf("undiscovered MCP should not be enabled: %+v", contextData.EnabledTools)
	}
}

func createCatalogRun(t *testing.T, svc *MemoryService, ident identity.LocalIdentity) Run {
	t.Helper()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Catalog", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "catalog"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func catalogToolByName(tools []ToolCatalogEntry, name string) ToolCatalogEntry {
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	return ToolCatalogEntry{}
}

func catalogResolutionByName(tools []ToolResolution, name string) ToolResolution {
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	return ToolResolution{}
}
