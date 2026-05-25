package runtime

import (
	"context"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

type recordingBrokerExecutor struct {
	names []string
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
