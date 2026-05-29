package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

func TestWorkerReturnsBackgroundJobFailPersistenceError(t *testing.T) {
	runnerErr := errors.New("runner failed")
	persistErr := errors.New("fail job write failed")
	svc := &workerOrderService{failErr: persistErr}
	worker := NewWorker(svc, nil, workerRunnerFunc(func(context.Context, productdata.Run, productdata.BackgroundJob) error {
		return runnerErr
	}))
	worker.WorkerID = "worker_fail_persist"
	worker.LeaseSeconds = 1

	ok, err := worker.ProcessOne(context.Background())
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	if !errors.Is(err, persistErr) {
		t.Fatalf("ProcessOne() err = %v, want fail persistence error", err)
	}
	expected := []string{"recover", "claim", "renew", "fail"}
	if len(svc.calls) != len(expected) {
		t.Fatalf("calls = %+v", svc.calls)
	}
	for index, call := range expected {
		if svc.calls[index] != call {
			t.Fatalf("calls = %+v", svc.calls)
		}
	}
}

func TestWorkerRenewsLeaseWhileRunnerIsStillRunning(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker lease", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{ScriptName: "long_worker_run"}); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)
	firstWorker := NewWorker(svc, nil, workerRunnerFunc(func(ctx context.Context, run productdata.Run, job productdata.BackgroundJob) error {
		close(started)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-release:
			return nil
		}
	}))
	firstWorker.WorkerID = "worker_long_owner"
	firstWorker.LeaseSeconds = 1
	go func() {
		_, err := firstWorker.ProcessOne(context.Background())
		firstDone <- err
	}()
	<-started
	time.Sleep(1200 * time.Millisecond)

	diagnostics, err := svc.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics.StaleCount != 0 {
		t.Fatalf("worker lease went stale while runner was still active: %+v", diagnostics)
	}
	close(release)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
}

func TestWorkerCancelsRunnerSoonAfterStopRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Worker stop", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{ScriptName: "long_worker_run"})
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	cancelled := make(chan struct{})
	worker := NewWorker(svc, nil, workerRunnerFunc(func(ctx context.Context, run productdata.Run, job productdata.BackgroundJob) error {
		close(started)
		<-ctx.Done()
		close(cancelled)
		return ctx.Err()
	}))
	worker.WorkerID = "worker_stop_watch"
	worker.LeaseSeconds = 30
	done := make(chan error, 1)
	go func() {
		_, err := worker.ProcessOne(context.Background())
		done <- err
	}()
	<-started
	if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-cancelled:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runner context was not cancelled soon after StopRun")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not return after StopRun cancellation")
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
	fail     productdata.FailBackgroundJobInput
	failErr  error
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

func (s *workerOrderService) FailBackgroundJob(_ context.Context, _ identity.LocalIdentity, input productdata.FailBackgroundJobInput) (productdata.BackgroundJob, bool, error) {
	s.calls = append(s.calls, "fail")
	s.fail = input
	if s.failErr != nil {
		return productdata.BackgroundJob{}, false, s.failErr
	}
	return productdata.BackgroundJob{ID: input.JobID}, true, nil
}

type workerProjectionEnsureOrderService struct {
	workerOrderService
	ensuredRunID string
}

func (s *workerProjectionEnsureOrderService) EnsureRunStepStateProjection(_ context.Context, _ identity.LocalIdentity, runID string) error {
	s.calls = append(s.calls, "ensure_projection")
	s.ensuredRunID = runID
	return nil
}

func TestWorkerEnsuresClaimedRunStepStateProjectionBeforeRenew(t *testing.T) {
	svc := &workerProjectionEnsureOrderService{}
	worker := NewWorker(svc, nil, workerRunnerFunc(func(context.Context, productdata.Run, productdata.BackgroundJob) error { return nil }))
	worker.WorkerID = "worker_projection_ensure"
	worker.LeaseSeconds = 1

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	expected := []string{"recover", "claim", "ensure_projection", "renew", "complete"}
	if len(svc.calls) != len(expected) {
		t.Fatalf("calls = %+v", svc.calls)
	}
	for index, call := range expected {
		if svc.calls[index] != call {
			t.Fatalf("calls = %+v", svc.calls)
		}
	}
	if svc.ensuredRunID != "run_1" {
		t.Fatalf("ensured run id = %q", svc.ensuredRunID)
	}
}

func TestWorkerDoesNotReplayFullRunHistoryOnClaim(t *testing.T) {
	svc := &workerProjectedPublishService{state: productdata.RunStepState{LastEventSequence: 100}}
	worker := NewWorker(svc, NewBroadcaster(), workerRunnerFunc(func(context.Context, productdata.Run, productdata.BackgroundJob) error { return nil }))
	worker.WorkerID = "worker_incremental_publish"
	worker.LeaseSeconds = 1

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	if len(svc.listAfterSequences) == 0 {
		t.Fatal("ListRunEvents was not called")
	}
	for _, afterSequence := range svc.listAfterSequences {
		if afterSequence == 0 {
			t.Fatalf("worker replayed full run history: after sequences = %+v", svc.listAfterSequences)
		}
	}
	if svc.listAfterSequences[0] != 99 {
		t.Fatalf("first publish cursor = %d, want 99; all = %+v", svc.listAfterSequences[0], svc.listAfterSequences)
	}
}

type workerProjectedPublishService struct {
	workerProjectionEnsureOrderService
	state              productdata.RunStepState
	listAfterSequences []int
}

func (s *workerProjectedPublishService) GetRunStepState(_ context.Context, _ identity.LocalIdentity, _ string) (productdata.RunStepState, error) {
	s.calls = append(s.calls, "get_projection")
	return s.state, nil
}

func (s *workerProjectedPublishService) ListRunEvents(_ context.Context, _ identity.LocalIdentity, runID string, afterSequence int) ([]productdata.RunEvent, error) {
	s.listAfterSequences = append(s.listAfterSequences, afterSequence)
	if runID != "run_1" || afterSequence >= s.state.LastEventSequence {
		return nil, nil
	}
	return []productdata.RunEvent{{ID: "evt_job_claimed", RunID: runID, ThreadID: "thread_1", Sequence: s.state.LastEventSequence, Type: productdata.EventJobClaimed}}, nil
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
	deadline := time.After(time.Second)
	for {
		select {
		case event := <-live:
			if event.Type == productdata.EventJobClaimed {
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for live event")
		}
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
		events, _ := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
		t.Fatalf("run = %+v events=%+v provider=%+v", got, events, provider.request)
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

func TestWorkerWaitsForDelegatedChildRunBeforeParentContinuation(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Agent delegate", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "delegate reviewer"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(context.Background(), ident, productdata.SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review the patch"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_agent_delegate", ToolName: productdata.ToolNameAgentDelegate, ArgumentsSummary: map[string]any{"task_id": task.ID}, ArgumentsHash: "hash_agent_delegate", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_agent_delegate"); err != nil {
		t.Fatal(err)
	}
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{{Type: ProviderEventCompleted, Text: "Parent resumed after child result."}}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_agent_delegate"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_agent_delegate")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionExecuting {
		t.Fatalf("delegate call completed before child run result: %+v", call)
	}
	if len(provider.requests) != 0 {
		t.Fatalf("parent provider continued before child result: %+v", provider.requests)
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, productdata.ListAgentTasksInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ChildRunID == "" {
		t.Fatalf("tasks = %+v", tasks)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var handoff map[string]any
	for _, event := range events {
		if event.Type == productdata.EventAgentChildRunStarted {
			handoff = event.Metadata
			break
		}
	}
	if handoff == nil || handoff["child_run_id"] != tasks[0].ChildRunID || handoff["child_thread_id"] != tasks[0].ChildThreadID || handoff["parent_tool_call_id"] != "tc_agent_delegate" {
		t.Fatalf("handoff event = %+v, events = %+v", handoff, events)
	}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, tasks[0].ChildThreadID, productdata.AppendAssistantMessageInput{Content: "Child review result: no issues."}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, tasks[0].ChildRunID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: productdata.EventRunCompleted, Summary: "Child run completed"}); err != nil {
		t.Fatal(err)
	}

	ok, err = worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() did not pick resumed parent job")
	}
	call, err = svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_agent_delegate")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["child_status"] != string(productdata.RunStatusCompleted) || !strings.Contains(fmt.Sprint(call.ResultSummary["result_summary"]), "no issues") {
		t.Fatalf("delegate call after child result = %+v", call)
	}
	if len(provider.requests) != 1 {
		t.Fatalf("provider requests = %d, want 1", len(provider.requests))
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("parent run = %+v", got)
	}
}

func TestPostgresWorkerWaitsForDelegatedChildRunBeforeParentContinuation(t *testing.T) {
	databaseURL := os.Getenv("LOOMI_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("LOOMI_TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	svc := productdata.NewPostgresRepository(pool)
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(ctx, ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(ctx, ident, productdata.CreateThreadInput{Title: "PG agent delegate " + productdata.NewThreadID(), Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(ctx, ident, thread.ID, productdata.CreateMessageInput{Content: "delegate reviewer"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(ctx, ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(ctx, ident, productdata.SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review the patch in Postgres"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(ctx, ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_agent_delegate_pg_worker", ToolName: productdata.ToolNameAgentDelegate, ArgumentsSummary: map[string]any{"task_id": task.ID}, ArgumentsHash: "hash_agent_delegate_pg_worker", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(ctx, ident, thread.ID, run.ID, "tc_agent_delegate_pg_worker"); err != nil {
		t.Fatal(err)
	}
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{{Type: ProviderEventCompleted, Text: "Parent resumed after PG child result."}}}}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_agent_delegate_pg"

	ok, err := worker.ProcessOne(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	call, err := svc.GetToolCall(ctx, ident, thread.ID, run.ID, "tc_agent_delegate_pg_worker")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionExecuting {
		t.Fatalf("delegate call completed before child run result: %+v", call)
	}
	if len(provider.requests) != 0 {
		t.Fatalf("parent provider continued before child result: %+v", provider.requests)
	}
	tasks, err := svc.ListAgentTasks(ctx, ident, productdata.ListAgentTasksInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ChildRunID == "" || tasks[0].ChildThreadID == "" {
		t.Fatalf("tasks = %+v", tasks)
	}
	events, err := svc.ListRunEvents(ctx, ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var handoff map[string]any
	for _, event := range events {
		if event.Type == productdata.EventAgentChildRunStarted {
			handoff = event.Metadata
			break
		}
	}
	if handoff == nil || handoff["child_run_id"] != tasks[0].ChildRunID || handoff["child_thread_id"] != tasks[0].ChildThreadID || handoff["parent_tool_call_id"] != "tc_agent_delegate_pg_worker" {
		t.Fatalf("handoff event = %+v, events = %+v", handoff, events)
	}
	if _, err := svc.AppendAssistantMessage(ctx, ident, tasks[0].ChildThreadID, productdata.AppendAssistantMessageInput{Content: "PG child review result: no issues."}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(ctx, ident, tasks[0].ChildRunID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: productdata.EventRunCompleted, Summary: "Child run completed"}); err != nil {
		t.Fatal(err)
	}

	ok, err = worker.ProcessOne(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() did not pick resumed parent job")
	}
	call, err = svc.GetToolCall(ctx, ident, thread.ID, run.ID, "tc_agent_delegate_pg_worker")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["child_status"] != string(productdata.RunStatusCompleted) || !strings.Contains(fmt.Sprint(call.ResultSummary["result_summary"]), "no issues") {
		t.Fatalf("delegate call after child result = %+v", call)
	}
	if len(provider.requests) != 1 {
		t.Fatalf("provider requests = %d, want 1", len(provider.requests))
	}
	got, err := svc.GetRun(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("parent run = %+v", got)
	}
}

func TestPostgresAgentDelegateChildWorkerTerminalResumesParent(t *testing.T) {
	databaseURL := os.Getenv("LOOMI_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("LOOMI_TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	svc := productdata.NewPostgresRepository(pool)
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(ctx, ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(ctx, ident, productdata.CreateThreadInput{Title: "PG agent delegate terminal " + productdata.NewThreadID(), Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(ctx, ident, thread.ID, productdata.CreateMessageInput{Content: "delegate reviewer and wait for child"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(ctx, ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(ctx, ident, productdata.SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review the patch through a real child worker"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(ctx, ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_agent_delegate_pg_terminal", ToolName: productdata.ToolNameAgentDelegate, ArgumentsSummary: map[string]any{"task_id": task.ID}, ArgumentsHash: "hash_agent_delegate_pg_terminal", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(ctx, ident, thread.ID, run.ID, "tc_agent_delegate_pg_terminal"); err != nil {
		t.Fatal(err)
	}
	provider := &sequencedProvider{
		config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true},
		eventSets: [][]ProviderEvent{
			{{Type: ProviderEventCompleted, Text: "PG child review result: no issues."}},
			{{Type: ProviderEventCompleted, Text: "Parent resumed after PG child result."}},
		},
	}
	worker := NewWorker(svc, nil, QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})})
	worker.WorkerID = "worker_agent_delegate_pg_terminal"

	ok, err := worker.ProcessOne(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() did not execute parent delegate")
	}
	if len(provider.requests) != 0 {
		t.Fatalf("provider continued before child worker terminal: %+v", provider.requests)
	}
	tasks, err := svc.ListAgentTasks(ctx, ident, productdata.ListAgentTasksInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ChildRunID == "" || tasks[0].ChildThreadID == "" {
		t.Fatalf("tasks = %+v", tasks)
	}
	call, err := svc.GetToolCall(ctx, ident, thread.ID, run.ID, "tc_agent_delegate_pg_terminal")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionExecuting {
		t.Fatalf("delegate call after parent step = %+v", call)
	}

	ok, err = worker.ProcessOne(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() did not execute child run")
	}
	if len(provider.requests) != 1 || provider.requests[0].ThreadID != tasks[0].ChildThreadID {
		t.Fatalf("child provider request = %+v, child thread = %s", provider.requests, tasks[0].ChildThreadID)
	}
	childRun, err := svc.GetRun(ctx, ident, tasks[0].ChildRunID)
	if err != nil {
		t.Fatal(err)
	}
	if childRun.Status != productdata.RunStatusCompleted {
		t.Fatalf("child run = %+v", childRun)
	}
	call, err = svc.GetToolCall(ctx, ident, thread.ID, run.ID, "tc_agent_delegate_pg_terminal")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionExecuting {
		t.Fatalf("delegate call completed before reconciliation: %+v", call)
	}

	ok, err = worker.ProcessOne(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() did not reconcile child and resume parent")
	}
	if len(provider.requests) != 2 || provider.requests[1].ThreadID != thread.ID {
		t.Fatalf("parent continuation request = %+v, parent thread = %s", provider.requests, thread.ID)
	}
	call, err = svc.GetToolCall(ctx, ident, thread.ID, run.ID, "tc_agent_delegate_pg_terminal")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["child_status"] != string(productdata.RunStatusCompleted) || !strings.Contains(fmt.Sprint(call.ResultSummary["result_summary"]), "no issues") {
		t.Fatalf("delegate call after reconciliation = %+v", call)
	}
	got, err := svc.GetRun(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("parent run = %+v", got)
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
	if call.ExecutionStatus != productdata.ToolCallExecutionCancelled {
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

func TestQueuedRunRouterDoesNotExecuteToolWhenStartReturnsCancelled(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Tool race", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_race", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "safe.txt"}, ArgumentsHash: "hash_race", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_race"); err != nil {
		t.Fatal(err)
	}
	cancelled := productdata.ToolCall{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_race", ToolName: productdata.ToolNameWorkspaceRead, ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionCancelled}
	service := &startCancelledService{Service: svc, call: cancelled}
	executor := &countingToolExecutor{}
	router := QueuedRunRouter{Gateway: NewGateway(service, nil, nil), toolExecutor: executor}

	if err := router.runApprovedTool(context.Background(), run, productdata.BackgroundJob{}, "tc_race", nil, false); err != nil {
		t.Fatal(err)
	}
	if executor.calls != 0 {
		t.Fatalf("executor calls = %d", executor.calls)
	}
	if service.failCalls != 0 {
		t.Fatalf("fail calls = %d", service.failCalls)
	}
}

func TestQueuedRunRouterDropsToolResultWhenJobOwnershipIsLost(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Tool owner", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_owner", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "safe.txt"}, ArgumentsHash: "hash_owner", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_owner"); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_old", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	service := &lostOwnershipService{Service: svc}
	router := QueuedRunRouter{Gateway: NewGateway(service, nil, nil), toolExecutor: scriptedToolExecutor{results: map[string]map[string]any{"tc_owner": {"ok": true}}}}

	if err := router.runApprovedTool(context.Background(), run, job, "tc_owner", nil, false); err != nil {
		t.Fatal(err)
	}
	if service.completeCalls != 0 || service.failCalls != 0 {
		t.Fatalf("complete=%d fail=%d", service.completeCalls, service.failCalls)
	}
}

func TestQueuedRunRouterDropsToolFailureWhenJobOwnershipIsLost(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Tool owner failure", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_owner_fail", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "safe.txt"}, ArgumentsHash: "hash_owner_fail", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_owner_fail"); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_old", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	service := &lostOwnershipService{Service: svc}
	router := QueuedRunRouter{Gateway: NewGateway(service, nil, nil), toolExecutor: failingToolExecutor{err: errors.New("workspace read failed")}}

	if err := router.runApprovedTool(context.Background(), run, job, "tc_owner_fail", nil, false); err != nil {
		t.Fatal(err)
	}
	if service.completeCalls != 0 || service.failCalls != 0 {
		t.Fatalf("complete=%d fail=%d", service.completeCalls, service.failCalls)
	}
}

type startCancelledService struct {
	productdata.Service
	call      productdata.ToolCall
	failCalls int
}

func (s *startCancelledService) StartToolCallExecution(context.Context, identity.LocalIdentity, string, string, string) (productdata.ToolCall, []productdata.RunEvent, error) {
	return s.call, nil, nil
}

func (s *startCancelledService) FailToolCallExecution(context.Context, identity.LocalIdentity, string, string, string, string, string) (productdata.ToolCall, []productdata.RunEvent, error) {
	s.failCalls++
	return productdata.ToolCall{}, nil, nil
}

type countingToolExecutor struct {
	calls int
}

func (e *countingToolExecutor) ExecuteTool(_ context.Context, invocation ToolInvocation) (ToolResult, error) {
	e.calls++
	return ToolResult{ToolName: invocation.ToolName, ResultSummary: map[string]any{"ok": true}}, nil
}

type failingToolExecutor struct {
	err error
}

func (e failingToolExecutor) ExecuteTool(_ context.Context, _ ToolInvocation) (ToolResult, error) {
	return ToolResult{}, e.err
}

type lostOwnershipService struct {
	productdata.Service
	completeCalls int
	failCalls     int
}

func (s *lostOwnershipService) RenewBackgroundJobLease(context.Context, identity.LocalIdentity, productdata.RenewBackgroundJobLeaseInput) (productdata.BackgroundJob, bool, error) {
	return productdata.BackgroundJob{}, false, nil
}

func (s *lostOwnershipService) CompleteToolCallSuccess(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, result map[string]any) (productdata.ToolCall, []productdata.RunEvent, error) {
	s.completeCalls++
	return s.Service.CompleteToolCallSuccess(ctx, ident, threadID, runID, toolCallID, result)
}

func (s *lostOwnershipService) FailToolCallExecution(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, code string, message string) (productdata.ToolCall, []productdata.RunEvent, error) {
	s.failCalls++
	return s.Service.FailToolCallExecution(ctx, ident, threadID, runID, toolCallID, code, message)
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

func TestQueuedToolExecutionFailureContinuesAsToolResult(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Recover failed read", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Read the project and summarize."})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_missing", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "missing.ts"}, ArgumentsHash: "hash_missing", ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionNotStarted}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Recovered after missing file."}}}
	prepared := &productdata.RunContext{
		ProviderRoute: productdata.ProviderRoute{ProviderID: "custom", Model: "model", Available: true},
		EnabledTools: []productdata.ToolResolution{{
			Name:           productdata.ToolNameWorkspaceRead,
			Source:         string(productdata.ToolCatalogSourceBuiltin),
			Group:          string(productdata.ToolCatalogGroupWorkspace),
			ExecutionState: string(productdata.ToolExecutionStateExecutable),
		}},
	}
	router := QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider}), toolExecutor: failingToolExecutor{err: errors.New("workspace path is unavailable")}}

	if err := router.runApprovedTool(context.Background(), run, productdata.BackgroundJob{}, "tc_missing", prepared, true); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		events, _ := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
		t.Fatalf("run = %+v events=%+v provider=%+v", got, events, provider.request)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_missing")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != productdata.ToolCallExecutionFailed {
		t.Fatalf("call = %+v", call)
	}
	if provider.request.ThreadID != thread.ID {
		t.Fatalf("provider was not called for recovery continuation: %+v", provider.request)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Content != "Recovered after missing file." {
		t.Fatalf("messages = %+v", messages)
	}
}
