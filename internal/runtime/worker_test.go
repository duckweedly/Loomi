package runtime

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestQueuedRunRouterPreparesContextBeforeRuntimeInvocation(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Context pipeline", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	runner := NewLocalRunner(svc, nil)
	runner.StepDelay = 0
	worker := NewWorker(svc, nil, QueuedRunRouter{Local: runner})
	worker.WorkerID = "worker_context"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	prepareCompleted := -1
	invokeStarted := -1
	resolveCompleted := false
	finalizeCompleted := false
	for index, event := range events {
		if event.Type == productdata.EventPipelineStepCompleted && event.Metadata["step"] == string(productdata.PipelineStepPrepareContext) {
			prepareCompleted = index
		}
		if event.Type == productdata.EventPipelineStepCompleted && event.Metadata["step"] == string(productdata.PipelineStepResolveTools) {
			resolveCompleted = true
		}
		if event.Type == productdata.EventPipelineStepStarted && event.Metadata["step"] == string(productdata.PipelineStepInvokeRuntime) && invokeStarted == -1 {
			invokeStarted = index
		}
		if event.Type == productdata.EventPipelineStepCompleted && event.Metadata["step"] == string(productdata.PipelineStepFinalize) {
			finalizeCompleted = true
		}
	}
	if prepareCompleted == -1 || invokeStarted == -1 || prepareCompleted > invokeStarted || !resolveCompleted || !finalizeCompleted {
		t.Fatalf("events = %+v", events)
	}
}

func TestQueuedRunRouterFailsBeforeRuntimeWhenContextIsMissing(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Missing context", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "should not run"}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_context"

	if _, err := worker.ProcessOne(context.Background()); err == nil {
		t.Fatal("ProcessOne() err = nil")
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var failedPrepare bool
	for _, event := range events {
		if event.Type == productdata.EventPipelineStepFailed && event.Metadata["step"] == string(productdata.PipelineStepPrepareContext) {
			failedPrepare = true
		}
		if event.Type == EventModelRequestStarted {
			t.Fatalf("runtime invoked despite missing context: %+v", events)
		}
	}
	if !failedPrepare {
		t.Fatalf("prepare_context failure not recorded: %+v", events)
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

func TestWorkerExecutesApprovedCurrentTimeToolAndContinuesModel(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Approved tool", Mode: productdata.ThreadModeChat})
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
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: "It is "}, {Type: ProviderEventCompleted, Text: "It is runtime time."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_tool"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["timezone"] != "UTC" {
		t.Fatalf("call = %+v", call)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var executing, succeeded int
	for _, event := range events {
		if event.Type == productdata.EventToolCallExecuting {
			executing++
		}
		if event.Type == productdata.EventToolCallSucceeded {
			succeeded++
		}
	}
	if executing != 1 || succeeded != 1 {
		t.Fatalf("events = %+v", events)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Content != "It is runtime time." {
		t.Fatalf("messages = %+v", messages)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}

	ok, err = worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("second ProcessOne() ok = true, want no duplicate job")
	}
}

func TestWorkerExecutesApprovedWorkspaceWriteFileAndContinuesModel(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace write", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "create file"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_write_1", ToolName: productdata.ToolNameWorkspaceWriteFile, ArgumentsSummary: map[string]any{"path": "created.txt", "content": "created\n"}, ArgumentsHash: "hash_write_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "created.txt")); err == nil {
		t.Fatal("file was written before approval")
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_write_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Created the file."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_workspace_write"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	written, err := os.ReadFile(filepath.Join(root, "created.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != "created\n" {
		t.Fatalf("written = %q", string(written))
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_write_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "write_file" || call.ResultSummary["path"] != "created.txt" {
		t.Fatalf("call = %+v", call)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
}

func TestWorkerExecutesApprovedWorkspaceEditAndContinuesModel(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("alpha\nbeta\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace edit", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "edit file"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_edit_1", ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "notes.txt", "old_text": "beta\n", "new_text": "gamma\n"}, ArgumentsHash: "hash_edit_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, err := (WorkspaceToolExecutor{Root: root}).Execute(context.Background(), ToolInvocation{RunID: run.ID, ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "notes.txt"}}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != "alpha\nbeta\n" {
		t.Fatalf("file changed before approval: %q", string(before))
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_edit_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Edited the file."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_workspace_edit"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	written, err := os.ReadFile(filepath.Join(root, "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != "alpha\ngamma\n" {
		t.Fatalf("written = %q", string(written))
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_edit_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "edit" || call.ResultSummary["path"] != "notes.txt" {
		t.Fatalf("call = %+v", call)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	ok, err = worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("second ProcessOne() ok = true, want no duplicate job")
	}
	afterRetry, err := os.ReadFile(filepath.Join(root, "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(afterRetry) != "alpha\ngamma\n" {
		t.Fatalf("retry changed file: %q", string(afterRetry))
	}
}

func TestWorkerExecutesApprovedSandboxExecCommandAndContinuesModel(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Sandbox exec", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "run command"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_exec_1", ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"ls", "."}, "cwd": "."}, ArgumentsHash: "hash_exec_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_exec_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Ran the command."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_sandbox_exec"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_exec_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "exec_command" || call.ResultSummary["scope"] != "bounded_command" || call.ResultSummary["exit_code"] != 0 {
		t.Fatalf("call = %+v", call)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
	ok, err = worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("second ProcessOne() ok = true, want no duplicate job")
	}
}

func TestWorkerExecutesApprovedLSPToolAndContinuesModel(t *testing.T) {
	root := createLSPFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "LSP symbols", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "find symbols"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_lsp_1", ToolName: productdata.ToolNameLSPSymbols, ArgumentsSummary: map[string]any{"path": "src/main.go", "query": "Tool"}, ArgumentsHash: "hash_lsp_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_lsp_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Found the symbol."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_lsp_symbols"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_lsp_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "symbols" || call.ResultSummary["scope"] != "lsp" || call.ResultSummary["count"] != 1 {
		t.Fatalf("call = %+v", call)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var executing, succeeded int
	for _, event := range events {
		if event.Type == productdata.EventToolCallExecuting {
			executing++
		}
		if event.Type == productdata.EventToolCallSucceeded {
			succeeded++
		}
	}
	if executing != 1 || succeeded != 1 {
		t.Fatalf("events = %+v", events)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
	ok, err = worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("second ProcessOne() ok = true, want no duplicate job")
	}
}

func TestWorkerExecutesApprovedWebFetchAndContinuesModel(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("worker web fetch result"))
	}))
	defer target.Close()
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Web fetch", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "fetch page"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_web_1", ToolName: productdata.ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": target.URL}, ArgumentsHash: "hash_web_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_web_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Fetched the page."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider}), WebExecutor: WebToolExecutor{AllowPrivateHosts: true}})
	worker.WorkerID = "worker_web_fetch"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_web_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "fetch" || call.ResultSummary["scope"] != "web" || call.ResultSummary["status_code"] != 200 {
		t.Fatalf("call = %+v", call)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
}

func TestWorkerExecutesApprovedWebSearchAndContinuesModel(t *testing.T) {
	searchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tvly-secret" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"Current AI News","url":"https://example.com/news","content":"public result snippet"}]}`))
	}))
	defer searchServer.Close()
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Search", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search latest ai news"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_search_1", ToolName: productdata.ToolNameWebSearch, ArgumentsSummary: map[string]any{"query": "latest ai news", "provider": "tavily", "limit": 2}, ArgumentsHash: "hash_search_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_search_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Found current news."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider}), WebExecutor: WebToolExecutor{TavilyAPIKey: "tvly-secret", TavilyEndpoint: searchServer.URL}})
	worker.WorkerID = "worker_web_search"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_search_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "search" || call.ResultSummary["provider"] != "tavily" || call.ResultSummary["result_count"] != 1 {
		t.Fatalf("call = %+v", call)
	}
	if strings.Contains(fmt.Sprint(call.ResultSummary), "tvly-secret") {
		t.Fatalf("result leaked key: %+v", call.ResultSummary)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].ToolName != productdata.ToolNameWebSearch || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
}

func TestWorkerExecutesApprovedBrowserOpenAndContinuesModel(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><title>Worker Browser</title><body>browser open</body></html>"))
	}))
	defer target.Close()
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Browser open", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "open page"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_browser_1", ToolName: productdata.ToolNameBrowserOpen, ArgumentsSummary: map[string]any{"url": target.URL}, ArgumentsHash: "hash_browser_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_browser_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Opened the page."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider}), BrowserExecutor: BrowserToolExecutor{Store: NewBrowserSessionStore(), AllowPrivateHosts: true}})
	worker.WorkerID = "worker_browser_open"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_browser_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "open" || call.ResultSummary["scope"] != "browser" || call.ResultSummary["title"] != "Worker Browser" {
		t.Fatalf("call = %+v", call)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
}

func TestWorkerExecutesApprovedArtifactCreateAndContinuesModel(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Artifact create", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "create artifact"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_artifact_1", ToolName: productdata.ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{"title": "Notes", "content": "hello artifact"}, ArgumentsHash: "hash_artifact_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_artifact_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Created the artifact."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_artifact_create"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_artifact_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "create_text" || call.ResultSummary["scope"] != "artifact" || call.ResultSummary["title"] != "Notes" {
		t.Fatalf("call = %+v", call)
	}
	artifacts, err := svc.ListArtifacts(context.Background(), ident, productdata.ListArtifactsInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 1 || artifacts[0].Title != "Notes" {
		t.Fatalf("artifacts = %+v", artifacts)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
}

func TestWorkerExecutesApprovedMemorySearchAndContinuesModel(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Memory search", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search memory"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateMemoryEntry(context.Background(), ident, productdata.CreateMemoryEntryInput{ScopeType: productdata.MemoryScopeThread, ScopeID: thread.ID, Title: "Memory Tool", Content: "Memory search results stay redacted.", SourceThreadID: thread.ID, SourceRunID: run.ID}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_memory_search_1", ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": "memory search", "limit": 5}, ArgumentsHash: "hash_memory_search_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_memory_search_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Used memory."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_memory_search"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_memory_search_1")
	if err != nil {
		t.Fatal(err)
	}
	items, _ := call.ResultSummary["items"].([]map[string]any)
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "search" || call.ResultSummary["scope"] != "memory" || call.ResultSummary["content"] != nil || len(items) != 1 {
		t.Fatalf("call = %+v", call)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
}

func TestWorkerExecutesApprovedAgentSpawnAndContinuesModel(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Agent spawn", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "spawn reviewer"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_agent_1", ToolName: productdata.ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "reviewer", "goal": "Review artifact runtime"}, ArgumentsHash: "hash_agent_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_agent_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Spawned reviewer task."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_agent_spawn"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_agent_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "spawn" || call.ResultSummary["scope"] != "agent" || call.ResultSummary["role"] != "reviewer" {
		t.Fatalf("call = %+v", call)
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, productdata.ListAgentTasksInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Role != "reviewer" || tasks[0].Status != productdata.AgentTaskStatusSpawned {
		t.Fatalf("tasks = %+v", tasks)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
}

func TestWorkerExecutesApprovedTodoWriteAndContinuesModel(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Todo write", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "update plan"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	items := []any{map[string]any{"id": "todo-1", "title": "Review patch", "status": "running", "summary": "Check tests"}}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_todo_1", ToolName: productdata.ToolNameTodoWrite, ArgumentsSummary: map[string]any{"items": items}, ArgumentsHash: "hash_todo_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_todo_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Updated the plan."}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_todo_write"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_todo_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "todo_write" {
		t.Fatalf("call = %+v", call)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var todoEvent productdata.RunEvent
	for _, event := range events {
		if event.Type == productdata.EventWorkTodoUpdated {
			todoEvent = event
		}
	}
	if todoEvent.Type == "" || todoEvent.Metadata["updated_by"] != "provider" {
		t.Fatalf("todo events = %+v", events)
	}
	todoItems := todoEvent.Metadata["todo_items"].([]any)
	todo := todoItems[0].(map[string]any)
	if todo["title"] != "Review patch" || todo["status"] != "running" {
		t.Fatalf("todo = %+v", todo)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("continuation request = %+v", provider.request.Messages)
	}
}

func TestWorkerDoesNotCreateArtifactAfterStopOrDenied(t *testing.T) {
	for _, tc := range []struct {
		name string
		stop bool
	}{
		{name: "stopped", stop: true},
		{name: "denied", stop: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := productdata.NewMemoryService()
			ident := identity.LocalDevIdentity()
			if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
				t.Fatal(err)
			}
			thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Artifact no exec", Mode: productdata.ThreadModeWork})
			if err != nil {
				t.Fatal(err)
			}
			message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "create artifact"})
			if err != nil {
				t.Fatal(err)
			}
			run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_artifact_blocked", ToolName: productdata.ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{"title": "Blocked", "content": "blocked"}, ArgumentsHash: "hash_artifact_blocked", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
				t.Fatal(err)
			}
			if tc.stop {
				if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_artifact_blocked"); err != nil {
					t.Fatal(err)
				}
				job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_artifact_stop", LeaseSeconds: 30})
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal("claim ok = false")
				}
				if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, _, err := svc.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_artifact_blocked"); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
				worker.WorkerID = "worker_artifact_denied"
				if ok, err := worker.ProcessOne(context.Background()); err != nil || ok {
					t.Fatalf("ProcessOne ok=%v err=%v", ok, err)
				}
			}
			artifacts, err := svc.ListArtifacts(context.Background(), ident, productdata.ListArtifactsInput{ThreadID: thread.ID})
			if err != nil {
				t.Fatal(err)
			}
			if len(artifacts) != 0 {
				t.Fatalf("artifacts created after %s: %+v", tc.name, artifacts)
			}
		})
	}
}

func TestWorkerDoesNotSpawnAgentTaskAfterStopOrDenied(t *testing.T) {
	for _, tc := range []struct {
		name string
		stop bool
	}{
		{name: "stopped", stop: true},
		{name: "denied", stop: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := productdata.NewMemoryService()
			ident := identity.LocalDevIdentity()
			if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
				t.Fatal(err)
			}
			thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Agent no exec", Mode: productdata.ThreadModeWork})
			if err != nil {
				t.Fatal(err)
			}
			message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "spawn agent"})
			if err != nil {
				t.Fatal(err)
			}
			run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_agent_blocked", ToolName: productdata.ToolNameAgentSpawn, ArgumentsSummary: map[string]any{"role": "reviewer", "goal": "Review implementation"}, ArgumentsHash: "hash_agent_blocked", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
				t.Fatal(err)
			}
			if tc.stop {
				if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_agent_blocked"); err != nil {
					t.Fatal(err)
				}
				job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_agent_stop", LeaseSeconds: 30})
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal("claim ok = false")
				}
				if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, _, err := svc.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_agent_blocked"); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
				worker.WorkerID = "worker_agent_denied"
				if ok, err := worker.ProcessOne(context.Background()); err != nil || ok {
					t.Fatalf("ProcessOne ok=%v err=%v", ok, err)
				}
			}
			tasks, err := svc.ListAgentTasks(context.Background(), ident, productdata.ListAgentTasksInput{ThreadID: thread.ID})
			if err != nil {
				t.Fatal(err)
			}
			if len(tasks) != 0 {
				t.Fatalf("agent tasks created after %s: %+v", tc.name, tasks)
			}
		})
	}
}

func TestWorkerDoesNotMutateWorkspaceAfterStopOrDenied(t *testing.T) {
	for _, tc := range []struct {
		name string
		stop bool
	}{
		{name: "stopped", stop: true},
		{name: "denied", stop: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("LOOMI_WORKSPACE_ROOT", root)
			svc := productdata.NewMemoryService()
			ident := identity.LocalDevIdentity()
			if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
				t.Fatal(err)
			}
			thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace mutation", Mode: productdata.ThreadModeWork})
			if err != nil {
				t.Fatal(err)
			}
			message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "create file"})
			if err != nil {
				t.Fatal(err)
			}
			run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_write_1", ToolName: productdata.ToolNameWorkspaceWriteFile, ArgumentsSummary: map[string]any{"path": "blocked.txt", "content": "blocked\n"}, ArgumentsHash: "hash_write_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
				t.Fatal(err)
			}
			if tc.stop {
				if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_write_1"); err != nil {
					t.Fatal(err)
				}
				job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_workspace_stop", LeaseSeconds: 30})
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal("claim ok = false")
				}
				if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, _, err := svc.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_write_1"); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
				worker.WorkerID = "worker_workspace_denied"
				ok, err := worker.ProcessOne(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if ok {
					t.Fatal("ProcessOne() ok = true, want no queued continuation")
				}
			}
			if _, err := os.Stat(filepath.Join(root, "blocked.txt")); err == nil {
				t.Fatal("mutation wrote file")
			}
		})
	}
}

func TestWorkerDoesNotExecuteSandboxCommandAfterStopOrDenied(t *testing.T) {
	for _, tc := range []struct {
		name string
		stop bool
	}{
		{name: "stopped", stop: true},
		{name: "denied", stop: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("LOOMI_WORKSPACE_ROOT", root)
			svc := productdata.NewMemoryService()
			ident := identity.LocalDevIdentity()
			if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
				t.Fatal(err)
			}
			thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Sandbox exec", Mode: productdata.ThreadModeWork})
			if err != nil {
				t.Fatal(err)
			}
			message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "run command"})
			if err != nil {
				t.Fatal(err)
			}
			run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_exec_1", ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"pwd"}}, ArgumentsHash: "hash_exec_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
				t.Fatal(err)
			}
			if tc.stop {
				if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_exec_1"); err != nil {
					t.Fatal(err)
				}
				job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_sandbox_stop", LeaseSeconds: 30})
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal("claim ok = false")
				}
				if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, _, err := svc.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_exec_1"); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
				worker.WorkerID = "worker_sandbox_denied"
				ok, err := worker.ProcessOne(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if ok {
					t.Fatal("ProcessOne() ok = true, want no queued continuation")
				}
			}
		})
	}
}

func TestWorkerDoesNotExecuteLSPToolAfterStopOrDenied(t *testing.T) {
	for _, tc := range []struct {
		name string
		stop bool
	}{
		{name: "stopped", stop: true},
		{name: "denied", stop: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := createLSPFixture(t)
			t.Setenv("LOOMI_WORKSPACE_ROOT", root)
			svc := productdata.NewMemoryService()
			ident := identity.LocalDevIdentity()
			if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
				t.Fatal(err)
			}
			thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "LSP no exec", Mode: productdata.ThreadModeWork})
			if err != nil {
				t.Fatal(err)
			}
			message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "find symbols"})
			if err != nil {
				t.Fatal(err)
			}
			run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_lsp_2", ToolName: productdata.ToolNameLSPSymbols, ArgumentsSummary: map[string]any{"path": "missing.go"}, ArgumentsHash: "hash_lsp_2", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
				t.Fatal(err)
			}
			if tc.stop {
				if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_lsp_2"); err != nil {
					t.Fatal(err)
				}
				job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_lsp_stop", LeaseSeconds: 30})
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal("claim ok = false")
				}
				if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, _, err := svc.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_lsp_2"); err != nil {
					t.Fatal(err)
				}
				provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
				worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
				worker.WorkerID = "worker_lsp_denied"
				ok, err := worker.ProcessOne(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if ok {
					t.Fatal("ProcessOne() ok = true, want no queued continuation")
				}
			}
			call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_lsp_2")
			if err != nil {
				t.Fatal(err)
			}
			if call.ExecutionStatus == productdata.ToolCallExecutionExecuting || call.ExecutionStatus == productdata.ToolCallExecutionFailed || call.ExecutionStatus == productdata.ToolCallExecutionSucceeded {
				t.Fatalf("lsp tool executed unexpectedly: %+v", call)
			}
		})
	}
}

func TestQueuedRunRouterDoesNotExecuteApprovedToolAfterStop(t *testing.T) {
	svc, ident, thread, _, run := setupWorkspaceContinuationRun(t)
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_read_2", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "safe.txt", "limit": 128}, ArgumentsHash: "hash_read_2", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_2"); err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_tool", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "should not continue"}}}
	router := QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}

	if err := router.Run(context.Background(), claimedRun, job); err != nil {
		t.Fatal(err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_2")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
		t.Fatalf("call = %+v", call)
	}
	if provider.request.ThreadID != "" {
		t.Fatalf("provider was called: %+v", provider.request)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusStopped {
		t.Fatalf("run = %+v", got)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == productdata.EventToolCallExecuting || event.Type == productdata.EventRunFailed || event.Type == productdata.EventJobAttemptFailed {
			t.Fatalf("stopped approved tool path wrote execution/failure event: %+v", events)
		}
	}
}

func TestWorkerDoesNotContinueAfterDeniedToolCall(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Denied tool", Mode: productdata.ThreadModeChat})
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
	if _, _, err := svc.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_denied"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("ProcessOne() ok = true, want no queued continuation")
	}
	if provider.request.ThreadID != "" {
		t.Fatalf("provider was called: %+v", provider.request)
	}
}

func TestWorkerDoesNotContinueAfterToolExecutionFailure(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Failed tool", Mode: productdata.ThreadModeChat})
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
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.FailToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_1", "tool_execution_failed", "Tool execution failed."); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_failed_tool"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("ProcessOne() ok = true, want no queued continuation")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed {
		t.Fatalf("run = %+v", got)
	}
	if provider.request.ThreadID != "" {
		t.Fatalf("provider was called: %+v", provider.request)
	}
}
