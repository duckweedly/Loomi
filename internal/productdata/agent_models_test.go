package productdata

import (
	"context"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
)

func TestValidateAgentToolCallArguments(t *testing.T) {
	spawn := RecordToolCallRequestInput{ToolCallID: "tc_agent", ToolName: ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": " reviewer ", "goal": "Review implementation"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	input, err := ValidateToolCallRequestInput(spawn)
	if err != nil {
		t.Fatal(err)
	}
	if input.ArgumentsSummary["role"] != "reviewer" {
		t.Fatalf("role was not normalized: %+v", input.ArgumentsSummary)
	}

	list := RecordToolCallRequestInput{ToolCallID: "tc_agent_list", ToolName: ToolNameAgentList, ArgumentsSummary: map[string]any{"limit": 10}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(list); err != nil {
		t.Fatal(err)
	}

	complete := RecordToolCallRequestInput{ToolCallID: "tc_agent_complete", ToolName: ToolNameAgentComplete, ArgumentsSummary: map[string]any{"task_id": "agt_123", "result_summary": "Done"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(complete); err != nil {
		t.Fatal(err)
	}

	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_agent", ToolName: ToolNameAgentSpawn, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent", ToolName: ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "", "goal": "Review"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent", ToolName: ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "reviewer", "goal": "", "api_key": "sk-secret"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent_complete", ToolName: ToolNameAgentComplete, ArgumentsSummary: map[string]any{"task_id": "agt_123"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent_list", ToolName: ToolNameAgentList, ArgumentsSummary: map[string]any{"query": "secret"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
	} {
		if _, err := ValidateToolCallRequestInput(input); err == nil || ErrorCode(err) != CodeInvalidRequest {
			t.Fatalf("ValidateToolCallRequestInput(%+v) err = %v", input, err)
		}
	}
}

func TestMemoryServiceAgentTaskLifecycle(t *testing.T) {
	svc := NewMemoryService()
	var _ AgentTaskService = svc
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Agents", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}

	task, err := svc.SpawnAgentTask(context.Background(), ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review implementation"})
	if err != nil {
		t.Fatal(err)
	}
	if task.ID == "" || task.ThreadID != thread.ID || task.RunID != run.ID || task.Status != AgentTaskStatusSpawned {
		t.Fatalf("task = %+v", task)
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, ListAgentTasksInput{ThreadID: thread.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID || tasks[0].Goal != "Review implementation" {
		t.Fatalf("tasks = %+v", tasks)
	}
	completed, err := svc.CompleteAgentTask(context.Background(), ident, CompleteAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ResultSummary: "Looks safe"})
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != AgentTaskStatusCompleted || completed.ResultSummary != "Looks safe" {
		t.Fatalf("completed = %+v", completed)
	}
	otherThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Other", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CompleteAgentTask(context.Background(), ident, CompleteAgentTaskInput{ThreadID: otherThread.ID, TaskID: task.ID, ResultSummary: "Wrong scope"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("out-of-thread complete err = %v", err)
	}
}
