package runtime

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type fakeMCPExecutor struct {
	calls  int
	result map[string]any
	err    error
}

func (f *fakeMCPExecutor) ExecuteMCPTool(ctx context.Context, call productdata.ToolCall) (map[string]any, error) {
	f.calls++
	return f.result, f.err
}

func TestWorkerExecutesApprovedMCPToolOnceAndContinuesWithRedactedResult(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, run := approvedMCPRun(t, svc, "tc_mcp_1")
	executor := &fakeMCPExecutor{result: map[string]any{"summary": "safe", "path": "/Users/xuean/.ssh/id_ed25519", "token": "sk-secret"}}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "MCP result used."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider}), MCPExecutor: executor})
	worker.WorkerID = "worker_mcp_exec"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	if executor.calls != 1 {
		t.Fatalf("executor calls = %d", executor.calls)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_mcp_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("call = %+v", call)
	}
	rendered := stringify(call.ResultSummary)
	for _, leaked := range []string{"sk-secret", "/Users/xuean", "id_ed25519"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("result leaked %q in %+v", leaked, call.ResultSummary)
		}
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation messages = %+v", provider.request.Messages)
	}
	if strings.Contains(provider.request.Messages[2].Content, "sk-secret") || strings.Contains(provider.request.Messages[2].Content, "/Users/xuean") {
		t.Fatalf("continuation leaked raw result: %+v", provider.request.Messages[2])
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
}

func TestWorkerDoesNotReexecuteMCPToolAfterExecutionStarted(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, run := approvedMCPRun(t, svc, "tc_mcp_1")
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_mcp_1"); err != nil {
		t.Fatal(err)
	}
	executor := &fakeMCPExecutor{result: map[string]any{"summary": "should not run"}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{}), MCPExecutor: executor})
	worker.WorkerID = "worker_mcp_retry"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	if executor.calls != 0 {
		t.Fatalf("executor calls = %d, want no duplicate execution", executor.calls)
	}
}

func TestWorkerDoesNotStartMCPToolAfterRunStopped(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	_, run := approvedMCPRun(t, svc, "tc_mcp_1")
	if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
		t.Fatal(err)
	}
	executor := &fakeMCPExecutor{result: map[string]any{"summary": "should not run"}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{}), MCPExecutor: executor})
	worker.WorkerID = "worker_mcp_stopped"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("ProcessOne() ok = true, want no claim after stopped run")
	}
	if executor.calls != 0 {
		t.Fatalf("executor calls = %d, want stopped run to prevent process startup", executor.calls)
	}
}

func approvedMCPRun(t *testing.T, svc *productdata.MemoryService, toolCallID string) (productdata.Thread, productdata.Run) {
	t.Helper()
	ident := identity.LocalDevIdentity()
	syncPersonaWithMCPTool(t, svc, "mcp.local-search.search")
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Approved MCP", Mode: productdata.ThreadModeChat})
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
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: toolCallID, ToolName: "mcp.local-search.search", CandidateSchemaHash: "sha256:test-local-search", ArgumentsSummary: map[string]any{"query": "status"}, ArgumentsHash: "hash_mcp", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, toolCallID); err != nil {
		t.Fatal(err)
	}
	return thread, run
}

func stringify(value any) string {
	return fmt.Sprintf("%+v", value)
}
