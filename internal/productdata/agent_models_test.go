package productdata

import (
	"context"
	"strings"
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

	start := RecordToolCallRequestInput{ToolCallID: "tc_agent_start", ToolName: ToolNameAgentStart, ArgumentsSummary: map[string]any{"task_id": " agt_123 "}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	input, err = ValidateToolCallRequestInput(start)
	if err != nil {
		t.Fatal(err)
	}
	if input.ArgumentsSummary["task_id"] != "agt_123" {
		t.Fatalf("task id was not normalized: %+v", input.ArgumentsSummary)
	}

	delegate := RecordToolCallRequestInput{ToolCallID: "tc_agent_delegate", ToolName: ToolNameAgentDelegate, ArgumentsSummary: map[string]any{"task_id": " agt_123 "}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	input, err = ValidateToolCallRequestInput(delegate)
	if err != nil {
		t.Fatal(err)
	}
	if input.ArgumentsSummary["task_id"] != "agt_123" {
		t.Fatalf("delegate task id was not normalized: %+v", input.ArgumentsSummary)
	}

	complete := RecordToolCallRequestInput{ToolCallID: "tc_agent_complete", ToolName: ToolNameAgentComplete, ArgumentsSummary: map[string]any{"task_id": "agt_123", "result_summary": "Done"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(complete); err != nil {
		t.Fatal(err)
	}

	fail := RecordToolCallRequestInput{ToolCallID: "tc_agent_fail", ToolName: ToolNameAgentFail, ArgumentsSummary: map[string]any{"task_id": "agt_123", "result_summary": "Blocked by missing context"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(fail); err != nil {
		t.Fatal(err)
	}

	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_agent", ToolName: ToolNameAgentSpawn, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent", ToolName: ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "", "goal": "Review"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent", ToolName: ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "reviewer", "goal": "", "api_key": "sk-secret"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent_start", ToolName: ToolNameAgentStart, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent_delegate", ToolName: ToolNameAgentDelegate, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent_delegate", ToolName: ToolNameAgentDelegate, ArgumentsSummary: map[string]any{"task_id": "agt_123", "prompt": "extra"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent_complete", ToolName: ToolNameAgentComplete, ArgumentsSummary: map[string]any{"task_id": "agt_123"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_agent_fail", ToolName: ToolNameAgentFail, ArgumentsSummary: map[string]any{"task_id": "agt_123"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
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
	started, err := svc.StartAgentTask(context.Background(), ident, StartAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID})
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != AgentTaskStatusInProgress {
		t.Fatalf("started = %+v", started)
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, ListAgentTasksInput{ThreadID: thread.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID || tasks[0].Goal != "Review implementation" || tasks[0].Status != AgentTaskStatusInProgress {
		t.Fatalf("tasks = %+v", tasks)
	}
	completed, err := svc.CompleteAgentTask(context.Background(), ident, CompleteAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ResultSummary: "Looks safe"})
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != AgentTaskStatusCompleted || completed.ResultSummary != "Looks safe" {
		t.Fatalf("completed = %+v", completed)
	}
	if _, err := svc.StartAgentTask(context.Background(), ident, StartAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("terminal start err = %v", err)
	}
	failedTask, err := svc.SpawnAgentTask(context.Background(), ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "researcher", Goal: "Research edge case"})
	if err != nil {
		t.Fatal(err)
	}
	failed, err := svc.FailAgentTask(context.Background(), ident, FailAgentTaskInput{ThreadID: thread.ID, TaskID: failedTask.ID, ResultSummary: "Blocked by missing context"})
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != AgentTaskStatusFailed || failed.ResultSummary != "Blocked by missing context" {
		t.Fatalf("failed = %+v", failed)
	}
	if _, err := svc.CompleteAgentTask(context.Background(), ident, CompleteAgentTaskInput{ThreadID: thread.ID, TaskID: failed.ID, ResultSummary: "Wrong terminal transition"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("terminal complete err = %v", err)
	}
	otherThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Other", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CompleteAgentTask(context.Background(), ident, CompleteAgentTaskInput{ThreadID: otherThread.ID, TaskID: task.ID, ResultSummary: "Wrong scope"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("out-of-thread complete err = %v", err)
	}
}

func TestMemoryServiceDelegateAgentTaskCreatesChildRun(t *testing.T) {
	svc := NewMemoryService()
	var _ AgentTaskService = svc
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Parent", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "Please review this change"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(context.Background(), ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review implementation safety"})
	if err != nil {
		t.Fatal(err)
	}

	delegated, err := svc.DelegateAgentTask(context.Background(), ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID})
	if err != nil {
		t.Fatal(err)
	}
	if delegated.Status != AgentTaskStatusInProgress || delegated.ChildThreadID == "" || delegated.ChildRunID == "" {
		t.Fatalf("delegated = %+v", delegated)
	}
	if delegated.ChildThreadID == thread.ID || delegated.ChildRunID == run.ID {
		t.Fatalf("child ids reused parent ids: %+v", delegated)
	}
	childThread, err := svc.GetThread(context.Background(), ident, delegated.ChildThreadID)
	if err != nil {
		t.Fatal(err)
	}
	if childThread.Mode != ThreadModeWork {
		t.Fatalf("childThread = %+v", childThread)
	}
	childRun, err := svc.GetRun(context.Background(), ident, delegated.ChildRunID)
	if err != nil {
		t.Fatal(err)
	}
	if childRun.ThreadID != delegated.ChildThreadID || childRun.Source != RunSourceModelGateway || childRun.Status != RunStatusQueued {
		t.Fatalf("childRun = %+v", childRun)
	}
	childMessages, err := svc.ListMessages(context.Background(), ident, delegated.ChildThreadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(childMessages) != 1 || childMessages[0].Role != MessageRoleUser || childMessages[0].Content == "" {
		t.Fatalf("childMessages = %+v", childMessages)
	}
	if _, err := svc.DelegateAgentTask(context.Background(), ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("duplicate delegate err = %v", err)
	}
	retryTask, err := svc.SpawnAgentTask(context.Background(), ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Retry delegate idempotency"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := svc.DelegateAgentTask(context.Background(), ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: retryTask.ID, ParentToolCallID: "tc_delegate_retry"})
	if err != nil {
		t.Fatal(err)
	}
	retry, err := svc.DelegateAgentTask(context.Background(), ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: retryTask.ID, ParentToolCallID: "tc_delegate_retry"})
	if err != nil {
		t.Fatalf("same parent tool-call delegate retry err = %v", err)
	}
	if retry.ChildThreadID != first.ChildThreadID || retry.ChildRunID != first.ChildRunID || retry.ParentToolCallID != "tc_delegate_retry" {
		t.Fatalf("retry delegate = %+v, first = %+v", retry, first)
	}

	failedTask, err := svc.SpawnAgentTask(context.Background(), ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "researcher", Goal: "Research failure"})
	if err != nil {
		t.Fatal(err)
	}
	failedTask, err = svc.FailAgentTask(context.Background(), ident, FailAgentTaskInput{ThreadID: thread.ID, TaskID: failedTask.ID, ResultSummary: "No context"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DelegateAgentTask(context.Background(), ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: failedTask.ID}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("terminal delegate err = %v", err)
	}
}

func TestMemoryServiceReconcilesDelegatedAgentTaskAfterChildRunCompletes(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Parent", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "Please review this change"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(context.Background(), ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review implementation safety"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_delegate", ToolName: ToolNameAgentDelegate, ArgumentsSummary: map[string]any{"task_id": task.ID}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_delegate"); err != nil {
		t.Fatal(err)
	}
	parentJob, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_parent_delegate", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("parent delegate job was not claimable")
	}
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_delegate"); err != nil {
		t.Fatal(err)
	}
	if _, changed, err := svc.CompleteBackgroundJob(context.Background(), ident, CompleteBackgroundJobInput{JobID: parentJob.ID, WorkerID: "worker_parent_delegate", OwnershipVersion: parentJob.OwnershipVersion}); err != nil || !changed {
		t.Fatalf("complete parent job changed=%v err=%v", changed, err)
	}
	delegated, err := svc.DelegateAgentTask(context.Background(), ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ParentToolCallID: "tc_delegate"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, delegated.ChildThreadID, AppendAssistantMessageInput{Content: "No issue found in child review."}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, delegated.ChildRunID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: EventRunCompleted, Summary: "Child run completed"}); err != nil {
		t.Fatal(err)
	}

	reconciled, err := svc.ReconcileAgentTaskChildRuns(context.Background(), ident, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(reconciled) != 1 {
		t.Fatalf("reconciled = %+v", reconciled)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_delegate")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != ToolCallExecutionSucceeded || call.ResultSummary["child_status"] != string(RunStatusCompleted) || !strings.Contains(call.ResultSummary["result_summary"].(string), "No issue") {
		t.Fatalf("call = %+v", call)
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, ListAgentTasksInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Status != AgentTaskStatusCompleted || !strings.Contains(tasks[0].ResultSummary, "No issue") {
		t.Fatalf("tasks = %+v", tasks)
	}
	claimed, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_parent_resume", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || claimed.RunID != run.ID || claimed.Metadata["resume_reason"] != "agent_child_run_completed" {
		t.Fatalf("claimed=%+v ok=%v", claimed, ok)
	}
}

func TestStopRunStopsDelegatedChildRunAndTask(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Parent", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "Delegate and then stop"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(context.Background(), ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review stop behavior"})
	if err != nil {
		t.Fatal(err)
	}
	delegated, err := svc.DelegateAgentTask(context.Background(), ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ParentToolCallID: "tc_delegate"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, delegated.ChildRunID, RecordToolCallRequestInput{ToolCallID: "tc_child_pending", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_child_pending", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}

	stopped, err := svc.StopRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stopped.Run.Status != RunStatusStopped {
		t.Fatalf("parent stopped = %+v", stopped)
	}
	childRun, err := svc.GetRun(context.Background(), ident, delegated.ChildRunID)
	if err != nil {
		t.Fatal(err)
	}
	if childRun.Status != RunStatusStopped || childRun.StopRequestedAt == nil || childRun.CompletedAt == nil {
		t.Fatalf("child run = %+v", childRun)
	}
	childCall, err := svc.GetToolCall(context.Background(), ident, delegated.ChildThreadID, delegated.ChildRunID, "tc_child_pending")
	if err != nil {
		t.Fatal(err)
	}
	if childCall.ExecutionStatus != ToolCallExecutionCancelled {
		t.Fatalf("child call = %+v", childCall)
	}
	if _, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_after_parent_stop", LeaseSeconds: 30}); err != nil || ok {
		t.Fatalf("claim after parent stop ok=%v run=%+v err=%v", ok, claimedRun, err)
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, ListAgentTasksInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Status != AgentTaskStatusFailed || !strings.Contains(tasks[0].ResultSummary, "Parent run stopped") {
		t.Fatalf("tasks = %+v", tasks)
	}
	childEvents, err := svc.ListRunEvents(context.Background(), ident, delegated.ChildRunID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(childEvents) < 5 || childEvents[len(childEvents)-3].Type != EventToolCallCancelled || childEvents[len(childEvents)-2].Type != EventStopRequested || childEvents[len(childEvents)-1].Type != EventRunStopped {
		t.Fatalf("child events = %+v", childEvents)
	}
}
