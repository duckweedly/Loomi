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

func TestToolCatalogIncludesDiscoveryTools(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ToolNameLoadTools, ToolNameLoadSkill} {
		tool := catalogToolByName(tools, name)
		if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupDiscovery || tool.RiskLevel != ToolRiskLow || tool.ApprovalPolicy != ToolApprovalReadOnly {
			t.Fatalf("%s metadata = %+v", name, tool)
		}
		if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["read_only"] != true {
			t.Fatalf("%s safe metadata = %+v", name, tool)
		}
	}
}

func TestToolCatalogIncludesWorkspaceReadOnlyTools(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ToolNameWorkspaceGlob, ToolNameWorkspaceGrep, ToolNameWorkspaceRead} {
		tool := catalogToolByName(tools, name)
		if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupWorkspace || tool.RiskLevel != ToolRiskLow || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("%s metadata = %+v", name, tool)
		}
		if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "workspace" || tool.SafeMetadata["read_only"] != true {
			t.Fatalf("%s safe metadata = %+v", name, tool)
		}
		if strings.Contains(fmt.Sprint(tool), "/Users/") {
			t.Fatalf("workspace catalog leaked host path: %+v", tool)
		}
	}
}

func TestToolCatalogIncludesWorkspaceMutationTools(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ToolNameWorkspaceWriteFile, ToolNameWorkspaceEdit, ToolNameWorkspacePatchPreview, ToolNameWorkspacePatchApply} {
		tool := catalogToolByName(tools, name)
		if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupWorkspace || tool.RiskLevel != ToolRiskHigh || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("%s metadata = %+v", name, tool)
		}
		if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "workspace" {
			t.Fatalf("%s safe metadata = %+v", name, tool)
		}
		if name == ToolNameWorkspacePatchPreview {
			if tool.SafeMetadata["read_only"] != true || tool.SafeMetadata["write_capable"] != false || tool.SafeMetadata["requires_read_before_preview"] != true || tool.SafeMetadata["preview_only"] != true || tool.SafeMetadata["returns_diff"] != true {
				t.Fatalf("%s preview metadata = %+v", name, tool)
			}
			continue
		}
		if tool.SafeMetadata["read_only"] != false || tool.SafeMetadata["write_capable"] != true {
			t.Fatalf("%s mutation metadata = %+v", name, tool)
		}
		if name == ToolNameWorkspaceEdit && (tool.SafeMetadata["requires_read_before_edit"] != true || tool.SafeMetadata["returns_diff"] != true || tool.SafeMetadata["normalizes_line_endings"] != true || tool.SafeMetadata["preserves_indentation"] != true || tool.SafeMetadata["strips_trailing_whitespace_except_markdown"] != true) {
			t.Fatalf("%s edit metadata = %+v", name, tool)
		}
		if name == ToolNameWorkspacePatchApply && (tool.SafeMetadata["requires_patch_preview"] != true || tool.SafeMetadata["returns_diff"] != true || tool.SafeMetadata["normalizes_line_endings"] != true || tool.SafeMetadata["preserves_indentation"] != true || tool.SafeMetadata["strips_trailing_whitespace_except_markdown"] != true) {
			t.Fatalf("%s apply metadata = %+v", name, tool)
		}
		if strings.Contains(fmt.Sprint(tool), "/Users/") {
			t.Fatalf("workspace mutation catalog leaked host path: %+v", tool)
		}
	}
}

func TestToolCatalogIncludesSandboxExecCommand(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ToolNameSandboxExecCommand, ToolNameSandboxStartProcess, ToolNameSandboxContinueProcess, ToolNameSandboxTerminateProcess} {
		tool := catalogToolByName(tools, name)
		if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupSandbox || tool.RiskLevel != ToolRiskHigh || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("%s metadata = %+v", name, tool)
		}
		if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["exec_capable"] != true || tool.SafeMetadata["argv_only"] != true || tool.SafeMetadata["validation_capable"] != true || tool.SafeMetadata["isolated_sandbox"] != false {
			t.Fatalf("%s safe metadata = %+v", name, tool)
		}
		if strings.Contains(fmt.Sprint(tool), "/Users/") {
			t.Fatalf("%s catalog leaked host path: %+v", name, tool)
		}
	}
}

func TestToolCatalogIncludesLSPReadOnlyTools(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ToolNameLSPDiagnostics, ToolNameLSPSymbols, ToolNameLSPReferences, ToolNameLSPDefinition, ToolNameLSPHover} {
		tool := catalogToolByName(tools, name)
		if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupLSP || tool.RiskLevel != ToolRiskLow || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("%s metadata = %+v", name, tool)
		}
		if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "lsp" || tool.SafeMetadata["read_only"] != true {
			t.Fatalf("%s safe metadata = %+v", name, tool)
		}
		if strings.Contains(fmt.Sprint(tool), "/Users/") {
			t.Fatalf("lsp catalog leaked host path: %+v", tool)
		}
	}
}

func TestToolCatalogIncludesWebFetchTool(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	tool := catalogToolByName(tools, ToolNameWebFetch)
	if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupWeb || tool.RiskLevel != ToolRiskMedium || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("web fetch metadata = %+v", tool)
	}
	if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "web" || tool.SafeMetadata["read_only"] != true || tool.SafeMetadata["network_access"] != "public_http_only" {
		t.Fatalf("web fetch safe metadata = %+v", tool)
	}
	if strings.Contains(fmt.Sprint(tool), "/Users/") {
		t.Fatalf("web fetch catalog leaked host path: %+v", tool)
	}
}

func TestToolCatalogIncludesWebSearchTool(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	tool := catalogToolByName(tools, ToolNameWebSearch)
	if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupWeb || tool.RiskLevel != ToolRiskMedium || tool.ApprovalPolicy != ToolApprovalReadOnly {
		t.Fatalf("web search metadata = %+v", tool)
	}
	if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "web" || tool.SafeMetadata["read_only"] != true || tool.SafeMetadata["network_access"] != "search_provider_api" {
		t.Fatalf("web search safe metadata = %+v", tool)
	}
	if strings.Contains(fmt.Sprint(tool), "/Users/") || strings.Contains(fmt.Sprint(tool), "tvly-") {
		t.Fatalf("web search catalog leaked sensitive data: %+v", tool)
	}
}

func TestWebSearchIsAvailableInChatRunContext(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default persona",
		SystemPrompt:     "prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{ToolNameCurrentTime, ToolNameWebSearch, ToolNameWebFetch},
		ReasoningMode:    "balanced",
		BudgetSummary:    "budget",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Chat search", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_chat_search", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	ctxData, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if catalogResolutionByName(ctxData.EnabledTools, ToolNameWebSearch).Name == "" {
		t.Fatalf("chat run missing web.search: %+v", ctxData.EnabledTools)
	}
	if catalogResolutionByName(ctxData.EnabledTools, ToolNameWebFetch).Name != "" {
		t.Fatalf("chat run should not enable web.fetch: %+v", ctxData.EnabledTools)
	}
}

func TestToolCatalogIncludesBrowserAutomationTools(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ToolNameBrowserOpen, ToolNameBrowserSnapshot, ToolNameBrowserClickLink, ToolNameBrowserScreenshot, ToolNameBrowserType, ToolNameBrowserPress} {
		tool := catalogToolByName(tools, name)
		if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupBrowser || tool.RiskLevel != ToolRiskMedium || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("%s metadata = %+v", name, tool)
		}
		if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "browser" || tool.SafeMetadata["network_access"] != "public_http_only" || tool.SafeMetadata["stateful"] != true {
			t.Fatalf("%s safe metadata = %+v", name, tool)
		}
		if strings.Contains(fmt.Sprint(tool), "/Users/") {
			t.Fatalf("browser catalog leaked host path: %+v", tool)
		}
	}
}

func TestToolCatalogIncludesArtifactRuntimeTools(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ToolNameArtifactCreateText, ToolNameArtifactRead, ToolNameArtifactList} {
		tool := catalogToolByName(tools, name)
		if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupArtifact || tool.RiskLevel != ToolRiskMedium || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("%s metadata = %+v", name, tool)
		}
		if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "artifact" || tool.SafeMetadata["non_executable"] != true {
			t.Fatalf("%s safe metadata = %+v", name, tool)
		}
		if name == ToolNameArtifactCreateText && tool.SafeMetadata["read_only"] != false {
			t.Fatalf("create_text should be mutation-like artifact tool: %+v", tool)
		}
		if name != ToolNameArtifactCreateText && tool.SafeMetadata["read_only"] != true {
			t.Fatalf("%s should be read-only artifact tool: %+v", name, tool)
		}
		if strings.Contains(fmt.Sprint(tool), "/Users/") {
			t.Fatalf("artifact catalog leaked host path: %+v", tool)
		}
	}
}

func TestToolCatalogIncludesAgentRuntimeTools(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{ToolNameAgentSpawn, ToolNameAgentList, ToolNameAgentComplete} {
		tool := catalogToolByName(tools, name)
		if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupAgent || tool.RiskLevel != ToolRiskMedium || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("%s metadata = %+v", name, tool)
		}
		if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "agent" || tool.SafeMetadata["coordination_only"] != true || tool.SafeMetadata["autonomous_execution"] != false {
			t.Fatalf("%s safe metadata = %+v", name, tool)
		}
		if name == ToolNameAgentList && tool.SafeMetadata["read_only"] != true {
			t.Fatalf("agent.list should be read-only: %+v", tool)
		}
		if name != ToolNameAgentList && tool.SafeMetadata["read_only"] != false {
			t.Fatalf("%s should mutate coordination records only: %+v", name, tool)
		}
		if strings.Contains(fmt.Sprint(tool), "/Users/") {
			t.Fatalf("agent catalog leaked host path: %+v", tool)
		}
	}
}

func TestToolCatalogIncludesTodoWriteTool(t *testing.T) {
	svc := NewMemoryService()
	tools, err := svc.ListToolCatalog(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatal(err)
	}
	tool := catalogToolByName(tools, ToolNameTodoWrite)
	if tool.Source != ToolCatalogSourceBuiltin || tool.Group != ToolCatalogGroupTodo || tool.RiskLevel != ToolRiskLow || tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("todo.write metadata = %+v", tool)
	}
	if !tool.Enabled || tool.ExecutionState != ToolExecutionStateExecutable || tool.SafeMetadata["scope"] != "work_todo" || tool.SafeMetadata["updates_plan_ui"] != true || tool.SafeMetadata["read_only"] != false {
		t.Fatalf("todo.write safe metadata = %+v", tool)
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
