package runtime

import (
	"context"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestAgentSpawnListAndComplete(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread, run := agentTestThreadRun(t, svc)
	executor := AgentToolExecutor{Tasks: svc}

	spawned, err := executor.Execute(context.Background(), ToolInvocation{
		ThreadID:         thread.ID,
		RunID:            run.ID,
		ToolCallID:       "tc_agent_spawn",
		ToolName:         productdata.ToolNameAgentSpawn,
		ArgumentsSummary: map[string]any{"role": "reviewer", "goal": "Review implementation"},
		ApprovalStatus:   productdata.ToolCallApprovalApproved,
		ExecutionStatus:  productdata.ToolCallExecutionExecuting,
		Catalog:          productdata.ToolCatalogFromEvents(nil),
		EnabledTools:     ToolResolutionsForPersona([]string{productdata.ToolNameAgentSpawn, productdata.ToolNameAgentList, productdata.ToolNameAgentComplete}),
	})
	if err != nil {
		t.Fatal(err)
	}
	taskID, _ := spawned["task_id"].(string)
	if !strings.HasPrefix(taskID, "agt_") || spawned["operation"] != "spawn" || spawned["status"] != string(productdata.AgentTaskStatusSpawned) || spawned["autonomous_execution"] != false {
		t.Fatalf("spawned = %+v", spawned)
	}

	list, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameAgentList, ArgumentsSummary: map[string]any{"limit": 10}})
	if err != nil {
		t.Fatal(err)
	}
	tasks, _ := list["tasks"].([]map[string]any)
	if list["operation"] != "list" || len(tasks) != 1 || tasks[0]["task_id"] != taskID {
		t.Fatalf("list = %+v", list)
	}

	completed, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameAgentComplete, ArgumentsSummary: map[string]any{"task_id": taskID, "result_summary": "No safety issue found"}})
	if err != nil {
		t.Fatal(err)
	}
	if completed["operation"] != "complete" || completed["status"] != string(productdata.AgentTaskStatusCompleted) || completed["result_summary"] != "No safety issue found" {
		t.Fatalf("completed = %+v", completed)
	}
}

func TestAgentRejectsUnsafeInputsAndScope(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread, run := agentTestThreadRun(t, svc)
	executor := AgentToolExecutor{Tasks: svc}

	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "owner", "goal": "Review"}}); err == nil {
		t.Fatal("expected unsupported role to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "reviewer", "goal": strings.Repeat("x", 4001)}}); err == nil {
		t.Fatal("expected oversized goal to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameAgentComplete, ArgumentsSummary: map[string]any{"task_id": "agt_missing", "result_summary": "done"}}); err == nil {
		t.Fatal("expected unknown task to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: "agent.execute", ArgumentsSummary: map[string]any{}}); err == nil {
		t.Fatal("expected unsupported agent tool to fail")
	}
}

func agentTestThreadRun(t *testing.T, svc *productdata.MemoryService) (productdata.Thread, productdata.Run) {
	t.Helper()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Agents", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	return thread, run
}
