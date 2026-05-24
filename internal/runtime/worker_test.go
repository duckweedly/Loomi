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
