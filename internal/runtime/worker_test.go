package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestWorkerRecoversExpiredLeaseBeforeProcessingRetry(t *testing.T) {
	svc := &workerOrderService{}
	worker := NewWorker(svc, nil, workerRunnerFunc(func(context.Context, productdata.Run, productdata.BackgroundJob) error { return nil }))
	worker.WorkerID = "worker_fresh"
	worker.LeaseSeconds = 1

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	expected := []string{"recover", "claim", "renew", "complete"}
	if len(svc.calls) != len(expected) {
		t.Fatalf("calls = %+v", svc.calls)
	}
	for index, call := range expected {
		if svc.calls[index] != call {
			t.Fatalf("calls = %+v", svc.calls)
		}
	}
	if svc.claim.WorkerID != "worker_fresh" || svc.renew.OwnershipVersion != 2 || svc.complete.OwnershipVersion != 2 {
		t.Fatalf("claim=%+v renew=%+v complete=%+v", svc.claim, svc.renew, svc.complete)
	}
}

type workerRunnerFunc func(context.Context, productdata.Run, productdata.BackgroundJob) error

func (f workerRunnerFunc) Run(ctx context.Context, run productdata.Run, job productdata.BackgroundJob) error {
	return f(ctx, run, job)
}

type workerOrderService struct {
	productdata.Service
	calls    []string
	claim    productdata.ClaimBackgroundJobInput
	renew    productdata.RenewBackgroundJobLeaseInput
	complete productdata.CompleteBackgroundJobInput
}

func (s *workerOrderService) RecoverBackgroundJobs(_ context.Context, _ identity.LocalIdentity, _ productdata.RecoverBackgroundJobsInput) ([]productdata.BackgroundJobRecovery, error) {
	s.calls = append(s.calls, "recover")
	return []productdata.BackgroundJobRecovery{{Job: productdata.BackgroundJob{ID: "job_1", Status: productdata.BackgroundJobStatusQueued}}}, nil
}

func (s *workerOrderService) ClaimBackgroundJob(_ context.Context, _ identity.LocalIdentity, input productdata.ClaimBackgroundJobInput) (productdata.BackgroundJob, productdata.Run, bool, error) {
	s.calls = append(s.calls, "claim")
	s.claim = input
	return productdata.BackgroundJob{ID: "job_1", RunID: "run_1", ThreadID: "thread_1", Status: productdata.BackgroundJobStatusLeased, LeasedBy: &input.WorkerID, OwnershipVersion: 2}, productdata.Run{ID: "run_1", ThreadID: "thread_1", Status: productdata.RunStatusRunning}, true, nil
}

func (s *workerOrderService) RenewBackgroundJobLease(_ context.Context, _ identity.LocalIdentity, input productdata.RenewBackgroundJobLeaseInput) (productdata.BackgroundJob, bool, error) {
	s.calls = append(s.calls, "renew")
	s.renew = input
	return productdata.BackgroundJob{ID: input.JobID}, true, nil
}

func (s *workerOrderService) CompleteBackgroundJob(_ context.Context, _ identity.LocalIdentity, input productdata.CompleteBackgroundJobInput) (productdata.BackgroundJob, bool, error) {
	s.calls = append(s.calls, "complete")
	s.complete = input
	return productdata.BackgroundJob{ID: input.JobID}, true, nil
}

func TestLocalRunnerStopsWhenRunIsStoppedAtSafeBoundary(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_owner", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
		t.Fatal(err)
	}
	runner := NewLocalRunner(svc, nil)
	runner.StepDelay = 0

	if err := runner.Run(context.Background(), claimedRun, job); err != nil {
		t.Fatal(err)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == productdata.EventRunCompleted || event.Type == "assistant_message" {
			t.Fatalf("stopped runner wrote event: %+v", events)
		}
	}
}

func TestLocalRunnerStopsWhenJobOwnershipIsStale(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_owner", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	staleWorker := "worker_stale"
	job.LeasedBy = &staleWorker
	runner := NewLocalRunner(svc, nil)
	runner.StepDelay = 0

	if err := runner.Run(context.Background(), run, job); err == nil {
		t.Fatal("Run() error = nil")
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == productdata.EventRunCompleted || event.Type == "assistant_message" {
			t.Fatalf("stale runner wrote event: %+v", events)
		}
	}
}

func TestWorkerPublishesServiceCreatedJobEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	broadcaster := NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	live := broadcaster.Subscribe(ctx, run.ID)
	worker := NewWorker(svc, broadcaster, workerRunnerFunc(func(context.Context, productdata.Run, productdata.BackgroundJob) error { return nil }))
	worker.WorkerID = "worker_test"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	var found bool
	for i := 0; i < 4; i++ {
		select {
		case event := <-live:
			if event.Type == productdata.EventJobClaimed {
				found = true
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for live event")
		}
	}
	if !found {
		t.Fatal("job_claimed was not published")
	}
}

func TestWorkerProcessesQueuedLocalRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	runner := NewLocalRunner(svc, nil)
	runner.StepDelay = 0
	worker := NewWorker(svc, nil, runner)
	worker.WorkerID = "worker_test"

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
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var claimed, pipelineCompleted, completed bool
	for _, event := range events {
		if event.Type == productdata.EventJobClaimed {
			claimed = true
		}
		if event.Type == productdata.EventPipelineStepCompleted {
			pipelineCompleted = true
		}
		if event.Type == productdata.EventRunCompleted {
			completed = true
		}
	}
	if !claimed || !pipelineCompleted || !completed {
		t.Fatalf("events = %+v", events)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].Role != productdata.MessageRoleAssistant || messages[0].Metadata["run_id"] != run.ID {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestWorkerLeavesQueuedGatewayRunBlockedOnToolApproval(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "time?"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_1"}}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_gateway"

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
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == productdata.EventRunFailed || event.Type == productdata.EventJobAttemptFailed {
			t.Fatalf("blocked run failed: %+v", events)
		}
	}
}

func TestWorkerExecutesApprovedCurrentTimeToolCall(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker tool", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "time?"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_1"}}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_gateway"

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("ProcessOne(request) ok=%v err=%v", ok, err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatalf("ApproveToolCall() error = %v", err)
	}
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("ProcessOne(tool) ok=%v err=%v", ok, err)
	}

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("call = %+v", call)
	}
	if call.ResultSummary["timezone"] != "UTC" || call.ResultSummary["source"] != "runtime" {
		t.Fatalf("result summary = %+v", call.ResultSummary)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var executing, succeeded, completed bool
	for _, event := range events {
		switch event.Type {
		case productdata.EventToolCallExecuting:
			executing = true
		case productdata.EventToolCallSucceeded:
			succeeded = true
		case productdata.EventRunCompleted:
			completed = true
		}
	}
	if !executing || !succeeded || !completed {
		t.Fatalf("events = %+v", events)
	}
}

func TestWorkerExecutesApprovedWorkspaceReadToolCall(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker workspace read", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "read tools"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceReadFile, Metadata: map[string]any{"tool_call_id": "tc_read", "arguments_summary": map[string]any{"path": "tools.go", "max_bytes": 64}}}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_workspace_read"

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("ProcessOne(request) ok=%v err=%v", ok, err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read"); err != nil {
		t.Fatalf("ApproveToolCall() error = %v", err)
	}
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("ProcessOne(tool) ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("call = %+v", call)
	}
	if call.ResultSummary["path"] != "tools.go" || call.ResultSummary["preview"] == "" {
		t.Fatalf("result summary = %+v", call.ResultSummary)
	}
}

func TestWorkerExecutesApprovedWorkspaceWriteAndEditToolCalls(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "internal"), 0o700); err != nil {
		t.Fatalf("Mkdir(internal) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "note.txt"), []byte("hello old\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(note) error = %v", err)
	}

	writeCall := executeApprovedWorkspaceToolForTest(t, root, productdata.ToolNameWorkspaceWriteFile, "tc_write", map[string]any{"path": "internal/generated.txt", "content": "hello write\n"})
	if writeCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("write call = %+v", writeCall)
	}
	if writeCall.ResultSummary["path"] != "internal/generated.txt" || writeCall.ResultSummary["bytes_written"] != len("hello write\n") {
		t.Fatalf("write result summary = %+v", writeCall.ResultSummary)
	}
	written, err := os.ReadFile(filepath.Join(root, "internal", "generated.txt"))
	if err != nil {
		t.Fatalf("ReadFile(generated) error = %v", err)
	}
	if string(written) != "hello write\n" {
		t.Fatalf("written content = %q", written)
	}

	editCall := executeApprovedWorkspaceToolForTest(t, root, productdata.ToolNameWorkspaceEdit, "tc_edit", map[string]any{"path": "internal/note.txt", "old_text": "old", "new_text": "new"})
	if editCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("edit call = %+v", editCall)
	}
	if editCall.ResultSummary["path"] != "internal/note.txt" || editCall.ResultSummary["replacements"] != 1 {
		t.Fatalf("edit result summary = %+v", editCall.ResultSummary)
	}
	edited, err := os.ReadFile(filepath.Join(root, "internal", "note.txt"))
	if err != nil {
		t.Fatalf("ReadFile(note) error = %v", err)
	}
	if string(edited) != "hello new\n" {
		t.Fatalf("edited content = %q", edited)
	}
}

func TestWorkerExecutesApprovedWorkspaceExecCommandToolCall(t *testing.T) {
	root := t.TempDir()
	call := executeApprovedWorkspaceToolForTest(t, root, productdata.ToolNameWorkspaceExecCommand, "tc_exec", map[string]any{"command": []any{"printf", "hello"}, "cwd": ".", "timeout_seconds": 5})
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("exec call = %+v", call)
	}
	if call.ResultSummary["cwd"] != "." || call.ResultSummary["exit_code"] != 0 || call.ResultSummary["stdout"] != "hello" || call.ResultSummary["timed_out"] != false {
		t.Fatalf("exec result summary = %+v", call.ResultSummary)
	}
}

func TestWorkerExecutesApprovedTodoWriteToolCall(t *testing.T) {
	call := executeApprovedWorkspaceToolForTest(t, "", productdata.ToolNameTodoWrite, "tc_todo", map[string]any{"items": []any{map[string]any{"title": "Inspect tools", "status": "completed"}, map[string]any{"title": "Implement todo", "status": "in_progress"}, map[string]any{"title": "Validate"}}})
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("todo call = %+v", call)
	}
	if call.ResultSummary["total"] != 3 || call.ResultSummary["completed_count"] != 1 || call.ResultSummary["in_progress_count"] != 1 || call.ResultSummary["pending_count"] != 1 {
		t.Fatalf("todo result summary = %+v", call.ResultSummary)
	}
}

func TestWorkerExecutesApprovedMCPCallTool(t *testing.T) {
	call := executeApprovedWorkspaceToolForTest(t, "", productdata.ToolNameMCPCallTool, "tc_mcp", map[string]any{"server": "local", "tool": "echo", "arguments": map[string]any{"message": "hello mcp"}})
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("mcp call = %+v", call)
	}
	if call.ResultSummary["server"] != "local" || call.ResultSummary["tool"] != "echo" || call.ResultSummary["message"] != "hello mcp" || call.ResultSummary["side_effect"] != "none" {
		t.Fatalf("mcp result summary = %+v", call.ResultSummary)
	}
}

func executeApprovedWorkspaceToolForTest(t *testing.T, root string, toolName string, toolCallID string, arguments map[string]any) productdata.ToolCall {
	t.Helper()
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker workspace write", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "write tools"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: toolName, Metadata: map[string]any{"tool_call_id": toolCallID, "arguments_summary": arguments}}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_workspace_write"
	worker.WorkspaceRoot = root

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("ProcessOne(request) ok=%v err=%v", ok, err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, toolCallID); err != nil {
		t.Fatalf("ApproveToolCall() error = %v", err)
	}
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("ProcessOne(tool) ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	return call
}

func TestWorkerSkipsCancelledApprovedToolCall(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker tool cancel", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "time?"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: productdata.ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
		t.Fatal(err)
	}
	worker := NewWorker(svc, nil, workerRunnerFunc(func(context.Context, productdata.Run, productdata.BackgroundJob) error {
		t.Fatal("runner should not execute cancelled tool job")
		return nil
	}))
	worker.WorkerID = "worker_tool_cancel"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatalf("ProcessOne() error = %v", err)
	}
	if ok {
		t.Fatal("ProcessOne() ok = true")
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == productdata.EventToolCallSucceeded || event.Type == productdata.EventToolCallFailed || event.Type == productdata.EventJobAttemptFailed {
			t.Fatalf("cancelled tool job wrote terminal/failure event: %+v", events)
		}
	}
}
