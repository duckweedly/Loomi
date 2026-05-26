package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

type recordingBrokerExecutor struct {
	names []string
}

func TestToolBrokerExecutesWorkspaceToolThroughOneEntrypoint(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("needle\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	broker := ToolBroker{Executor: DefaultToolExecutor{WorkspaceExecutor: WorkspaceToolExecutor{Root: root}}}
	call := productdata.ToolCall{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_workspace", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "notes.txt"}, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}
	catalog := []productdata.ToolCatalogEntry{{Name: productdata.ToolNameWorkspaceRead, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin, Group: productdata.ToolCatalogGroupWorkspace}}
	allowed := []productdata.ToolResolution{{Name: productdata.ToolNameWorkspaceRead, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupWorkspace)}}

	result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
	if err != nil {
		t.Fatal(err)
	}
	if result.ResultSummary["content"] != "needle\n" || result.ResultSummary["scope"] != "workspace" {
		t.Fatalf("workspace broker result = %+v", result)
	}
}

func TestToolBrokerExecutesWorkspaceMutationToolThroughOneEntrypoint(t *testing.T) {
	root := t.TempDir()
	broker := ToolBroker{Executor: DefaultToolExecutor{WorkspaceExecutor: WorkspaceToolExecutor{Root: root}}}
	call := productdata.ToolCall{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_workspace_write", ToolName: productdata.ToolNameWorkspaceWriteFile, ArgumentsSummary: map[string]any{"path": "generated.txt", "content": "created\n"}, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}
	catalog := []productdata.ToolCatalogEntry{{Name: productdata.ToolNameWorkspaceWriteFile, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin, Group: productdata.ToolCatalogGroupWorkspace, RiskLevel: productdata.ToolRiskHigh}}
	allowed := []productdata.ToolResolution{{Name: productdata.ToolNameWorkspaceWriteFile, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupWorkspace), RiskLevel: string(productdata.ToolRiskHigh)}}

	result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
	if err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(filepath.Join(root, "generated.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != "created\n" || result.ResultSummary["operation"] != "write_file" || result.ResultSummary["path"] != "generated.txt" {
		t.Fatalf("workspace mutation broker result = %+v", result)
	}
	if _, leaked := result.ResultSummary["content"]; leaked {
		t.Fatalf("workspace mutation result leaked content: %+v", result)
	}
}

func TestToolBrokerExecutesSandboxExecCommandThroughOneEntrypoint(t *testing.T) {
	root := t.TempDir()
	broker := ToolBroker{Executor: DefaultToolExecutor{SandboxExecutor: SandboxToolExecutor{Root: root}}}
	call := productdata.ToolCall{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_sandbox_exec", ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"pwd"}}, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}
	catalog := []productdata.ToolCatalogEntry{{Name: productdata.ToolNameSandboxExecCommand, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin, Group: productdata.ToolCatalogGroupSandbox, RiskLevel: productdata.ToolRiskHigh}}
	allowed := []productdata.ToolResolution{{Name: productdata.ToolNameSandboxExecCommand, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupSandbox), RiskLevel: string(productdata.ToolRiskHigh)}}

	result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
	if err != nil {
		t.Fatal(err)
	}
	if result.ResultSummary["operation"] != "exec_command" || result.ResultSummary["scope"] != "bounded_read_only_command" || result.ResultSummary["stdout"] == "" {
		t.Fatalf("sandbox broker result = %+v", result)
	}
}

func TestToolBrokerExecutesLSPToolThroughOneEntrypoint(t *testing.T) {
	root := createLSPFixture(t)
	broker := ToolBroker{Executor: DefaultToolExecutor{LSPExecutor: LSPToolExecutor{Root: root}}}
	call := productdata.ToolCall{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_lsp", ToolName: productdata.ToolNameLSPSymbols, ArgumentsSummary: map[string]any{"path": "src/main.go", "query": "Tool"}, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}
	catalog := []productdata.ToolCatalogEntry{{Name: productdata.ToolNameLSPSymbols, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin, Group: productdata.ToolCatalogGroupLSP}}
	allowed := []productdata.ToolResolution{{Name: productdata.ToolNameLSPSymbols, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupLSP)}}

	result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
	if err != nil {
		t.Fatal(err)
	}
	if result.ResultSummary["operation"] != "symbols" || result.ResultSummary["scope"] != "lsp" || result.ResultSummary["count"] != 1 {
		t.Fatalf("lsp broker result = %+v", result)
	}
}

func TestToolBrokerExecutesWebFetchThroughOneEntrypoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("web result"))
	}))
	defer server.Close()
	broker := ToolBroker{Executor: DefaultToolExecutor{WebExecutor: WebToolExecutor{AllowPrivateHosts: true}}}
	call := productdata.ToolCall{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_web", ToolName: productdata.ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": server.URL}, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}
	catalog := []productdata.ToolCatalogEntry{{Name: productdata.ToolNameWebFetch, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin, Group: productdata.ToolCatalogGroupWeb, RiskLevel: productdata.ToolRiskMedium}}
	allowed := []productdata.ToolResolution{{Name: productdata.ToolNameWebFetch, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupWeb), RiskLevel: string(productdata.ToolRiskMedium)}}

	result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
	if err != nil {
		t.Fatal(err)
	}
	if result.ResultSummary["operation"] != "fetch" || result.ResultSummary["scope"] != "web" || result.ResultSummary["status_code"] != 200 {
		t.Fatalf("web broker result = %+v", result)
	}
}

func TestToolBrokerExecutesBrowserOpenThroughOneEntrypoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><title>Browser</title><body>ok</body></html>"))
	}))
	defer server.Close()
	broker := ToolBroker{Executor: DefaultToolExecutor{BrowserExecutor: BrowserToolExecutor{Store: NewBrowserSessionStore(), AllowPrivateHosts: true}}}
	call := productdata.ToolCall{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_browser", ToolName: productdata.ToolNameBrowserOpen, ArgumentsSummary: map[string]any{"url": server.URL}, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}
	catalog := []productdata.ToolCatalogEntry{{Name: productdata.ToolNameBrowserOpen, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin, Group: productdata.ToolCatalogGroupBrowser, RiskLevel: productdata.ToolRiskMedium}}
	allowed := []productdata.ToolResolution{{Name: productdata.ToolNameBrowserOpen, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupBrowser), RiskLevel: string(productdata.ToolRiskMedium)}}

	result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
	if err != nil {
		t.Fatal(err)
	}
	if result.ResultSummary["operation"] != "open" || result.ResultSummary["scope"] != "browser" || result.ResultSummary["title"] != "Browser" {
		t.Fatalf("browser broker result = %+v", result)
	}
}

func TestToolBrokerExecutesArtifactCreateThroughOneEntrypoint(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread, run := artifactTestThreadRun(t, svc)
	broker := ToolBroker{Executor: DefaultToolExecutor{ArtifactExecutor: ArtifactToolExecutor{Artifacts: svc}}}
	call := productdata.ToolCall{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_artifact", ToolName: productdata.ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{"title": "Notes", "content": "hello artifact"}, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}
	catalog := []productdata.ToolCatalogEntry{{Name: productdata.ToolNameArtifactCreateText, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin, Group: productdata.ToolCatalogGroupArtifact, RiskLevel: productdata.ToolRiskMedium}}
	allowed := []productdata.ToolResolution{{Name: productdata.ToolNameArtifactCreateText, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupArtifact), RiskLevel: string(productdata.ToolRiskMedium)}}

	result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
	if err != nil {
		t.Fatal(err)
	}
	if result.ResultSummary["operation"] != "create_text" || result.ResultSummary["scope"] != "artifact" || result.ResultSummary["title"] != "Notes" {
		t.Fatalf("artifact broker result = %+v", result)
	}
	if _, leaked := result.ResultSummary["content"]; leaked {
		t.Fatalf("artifact result leaked raw content: %+v", result)
	}
}

func TestToolBrokerExecutesAgentSpawnThroughOneEntrypoint(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread, run := agentTestThreadRun(t, svc)
	broker := ToolBroker{Executor: DefaultToolExecutor{AgentExecutor: AgentToolExecutor{Tasks: svc}}}
	call := productdata.ToolCall{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_agent", ToolName: productdata.ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "reviewer", "goal": "Review safety"}, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}
	catalog := []productdata.ToolCatalogEntry{{Name: productdata.ToolNameAgentSpawn, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin, Group: productdata.ToolCatalogGroupAgent, RiskLevel: productdata.ToolRiskMedium}}
	allowed := []productdata.ToolResolution{{Name: productdata.ToolNameAgentSpawn, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupAgent), RiskLevel: string(productdata.ToolRiskMedium)}}

	result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
	if err != nil {
		t.Fatal(err)
	}
	if result.ResultSummary["operation"] != "spawn" || result.ResultSummary["scope"] != "agent" || result.ResultSummary["role"] != "reviewer" || result.ResultSummary["autonomous_execution"] != false {
		t.Fatalf("agent broker result = %+v", result)
	}
}

func (e *recordingBrokerExecutor) ExecuteTool(_ context.Context, invocation ToolInvocation) (ToolResult, error) {
	e.names = append(e.names, invocation.ToolName)
	return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: map[string]any{"ok": true, "api_key": "sk-secret-result"}}, nil
}

func TestToolBrokerRejectsNotAllowedDisabledAndSchemaMismatch(t *testing.T) {
	executor := &recordingBrokerExecutor{}
	broker := ToolBroker{Executor: executor}
	call := productdata.ToolCall{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_1", ToolName: "mcp.local-smoke.echo", CandidateSchemaHash: "sha256:call", ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting}

	cases := []struct {
		name    string
		catalog []productdata.ToolCatalogEntry
		allowed []productdata.ToolResolution
	}{
		{name: "not allowed", catalog: []productdata.ToolCatalogEntry{{Name: call.ToolName, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, InputSchemaHash: "sha256:call"}}, allowed: nil},
		{name: "disabled", catalog: []productdata.ToolCatalogEntry{{Name: call.ToolName, Enabled: false, ExecutionState: productdata.ToolExecutionStateDisabled, InputSchemaHash: "sha256:call"}}, allowed: []productdata.ToolResolution{{Name: call.ToolName, ExecutionState: string(productdata.ToolExecutionStateExecutable), InputSchemaHash: "sha256:call"}}},
		{name: "schema mismatch", catalog: []productdata.ToolCatalogEntry{{Name: call.ToolName, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, InputSchemaHash: "sha256:other"}}, allowed: []productdata.ToolResolution{{Name: call.ToolName, ExecutionState: string(productdata.ToolExecutionStateExecutable), InputSchemaHash: "sha256:other"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := broker.Execute(context.Background(), ToolInvocation{ThreadID: call.ThreadID, RunID: call.RunID, ToolCallID: call.ToolCallID, ToolName: call.ToolName, CandidateSchemaHash: call.CandidateSchemaHash, ApprovalStatus: call.ApprovalStatus, ExecutionStatus: call.ExecutionStatus, Catalog: tc.catalog, EnabledTools: tc.allowed})
			if err == nil {
				t.Fatal("Execute() err = nil, want rejection")
			}
		})
	}
	if len(executor.names) != 0 {
		t.Fatalf("executor called on rejected invocations: %v", executor.names)
	}
}

func TestToolBrokerUsesOneEntrypointAndRedactsResult(t *testing.T) {
	executor := &recordingBrokerExecutor{}
	broker := ToolBroker{Executor: executor}
	catalog := []productdata.ToolCatalogEntry{
		{Name: productdata.ToolNameCurrentTime, Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceBuiltin},
		{Name: "mcp.local-smoke.echo", Enabled: true, ExecutionState: productdata.ToolExecutionStateExecutable, Source: productdata.ToolCatalogSourceMCP, InputSchemaHash: "sha256:mcp"},
	}
	allowed := []productdata.ToolResolution{
		{Name: productdata.ToolNameCurrentTime, ExecutionState: string(productdata.ToolExecutionStateExecutable)},
		{Name: "mcp.local-smoke.echo", ExecutionState: string(productdata.ToolExecutionStateExecutable), InputSchemaHash: "sha256:mcp"},
	}
	for _, call := range []productdata.ToolCall{
		{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_builtin", ToolName: productdata.ToolNameCurrentTime, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting},
		{ThreadID: "thr_1", RunID: "run_1", ToolCallID: "tc_mcp", ToolName: "mcp.local-smoke.echo", CandidateSchemaHash: "sha256:mcp", ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionExecuting},
	} {
		result, err := broker.Execute(context.Background(), ToolInvocationFromCall(call, catalog, allowed))
		if err != nil {
			t.Fatal(err)
		}
		if _, leaked := result.ResultSummary["api_key"]; leaked {
			t.Fatalf("result not redacted: %+v", result)
		}
	}
	if len(executor.names) != 2 || executor.names[0] != productdata.ToolNameCurrentTime || executor.names[1] != "mcp.local-smoke.echo" {
		t.Fatalf("executor names = %v", executor.names)
	}
}

func TestToolCatalogForExecutionUsesPreparedRunContextHash(t *testing.T) {
	catalog := toolCatalogForExecution([]productdata.ToolResolution{{
		Name:            "mcp.local-smoke.echo",
		Source:          string(productdata.ToolCatalogSourceMCP),
		Group:           string(productdata.ToolCatalogGroupMCP),
		InputSchemaHash: "sha256:current-run",
		RiskLevel:       string(productdata.ToolRiskMedium),
		ApprovalPolicy:  string(productdata.ToolApprovalAlwaysRequired),
		ExecutionState:  string(productdata.ToolExecutionStateExecutable),
	}})

	entry, ok := catalogEntryByName(catalog, "mcp.local-smoke.echo")
	if !ok {
		t.Fatalf("catalog = %+v, missing mcp tool", catalog)
	}
	if entry.InputSchemaHash != "sha256:current-run" || entry.ExecutionState != productdata.ToolExecutionStateExecutable {
		t.Fatalf("execution catalog should come from prepared RunContext: %+v", entry)
	}
}
