package runtime

import (
	"context"
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
