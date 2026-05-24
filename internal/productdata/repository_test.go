package productdata

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sheridiany/loomi/internal/identity"
)

func TestRepositoryContractUsesPostgresImplementation(t *testing.T) {
	var _ Repository = (*MemoryService)(nil)
	var _ Repository = (*PostgresRepository)(nil)
}

func TestRepositoryContractCoversM5AssistantAndModelGateway(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M5", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, err := repo.AppendAssistantMessage(context.Background(), ident, thread.ID, AppendAssistantMessageInput{Content: "hello", Metadata: map[string]any{"run_id": "run_1"}})
	if err != nil {
		t.Fatal(err)
	}
	if message.Role != MessageRoleAssistant || message.Metadata["run_id"] != "run_1" {
		t.Fatalf("message = %+v", message)
	}
	if _, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AppendAssistantMessage(context.Background(), ident, thread.ID, AppendAssistantMessageInput{Content: "again", Metadata: map[string]any{"run_id": "run_1"}}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("duplicate err = %v", err)
	}
}

func TestRepositoryContractCoversM6JobCreationAndClaim(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M6", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{}); err == nil || ErrorCode(err) != CodeActiveRunExists {
		t.Fatalf("second active run err = %v", err)
	}
	job, claimedRun, ok, err := repo.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || job.RunID != run.ID || job.Status != BackgroundJobStatusLeased || claimedRun.Status != RunStatusRunning {
		t.Fatalf("job=%+v run=%+v ok=%v", job, claimedRun, ok)
	}
	if _, _, ok, err := repo.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test_2", LeaseSeconds: 5}); err != nil || ok {
		t.Fatalf("second claim ok=%v err=%v", ok, err)
	}
}

func TestRepositoryContractCoversM6RecoveryAndRetryExhaustion(t *testing.T) {
	repo := NewMemoryService()
	var contract Repository = repo
	ident := identity.LocalDevIdentity()
	base := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	repo.now = func() time.Time { return base }
	thread, err := contract.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M6 recovery", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := contract.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := contract.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_stale", LeaseSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	base = base.Add(2 * time.Second)
	recoveries, err := contract.RecoverBackgroundJobs(context.Background(), ident, RecoverBackgroundJobsInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveries) != 1 || recoveries[0].Exhausted || recoveries[0].Job.Status != BackgroundJobStatusQueued || recoveries[0].Run.Status != RunStatusRecovering {
		t.Fatalf("recoveries = %+v", recoveries)
	}
	if _, changed, err := contract.FailBackgroundJob(context.Background(), ident, FailBackgroundJobInput{JobID: job.ID, WorkerID: "worker_stale", OwnershipVersion: job.OwnershipVersion, ErrorCode: "stale", ErrorMessage: "stale"}); err != nil || changed {
		t.Fatalf("stale fail changed=%v err=%v", changed, err)
	}
	for attempt := 2; attempt <= 3; attempt++ {
		if _, _, ok, err := contract.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_retry", LeaseSeconds: 1}); err != nil || !ok {
			t.Fatalf("claim attempt %d ok=%v err=%v", attempt, ok, err)
		}
		base = base.Add(2 * time.Second)
		recoveries, err = contract.RecoverBackgroundJobs(context.Background(), ident, RecoverBackgroundJobsInput{ErrorMessage: "password=secret"})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(recoveries) != 1 || !recoveries[0].Exhausted || recoveries[0].Job.Status != BackgroundJobStatusDead || recoveries[0].Run.Status != RunStatusFailed {
		t.Fatalf("final recoveries = %+v", recoveries)
	}
	got, err := contract.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != RunStatusFailed || got.ErrorMessage == nil || *got.ErrorMessage != "[redacted]" {
		t.Fatalf("run = %+v", got)
	}
}

func TestRepositoryContractCoversM6QueueDiagnosticsStates(t *testing.T) {
	repo := NewMemoryService()
	var contract Repository = repo
	ident := identity.LocalDevIdentity()
	base := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	repo.now = func() time.Time { return base }
	thread, err := contract.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M6 diagnostics", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := contract.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	queued, err := contract.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if queued.QueueStatus != WorkerQueueStatusReady || queued.QueuedCount != 1 {
		t.Fatalf("queued diagnostics = %+v", queued)
	}
	if _, _, ok, err := contract.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 1}); err != nil || !ok {
		t.Fatalf("claim ok=%v err=%v", ok, err)
	}
	leased, err := contract.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if leased.LeasedCount != 1 || leased.StaleCount != 0 {
		t.Fatalf("leased diagnostics = %+v", leased)
	}
	base = base.Add(2 * time.Second)
	stale, err := contract.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if stale.QueueStatus != WorkerQueueStatusDegraded || stale.StaleCount != 1 {
		t.Fatalf("stale diagnostics = %+v", stale)
	}
	for attempt := 1; attempt <= 3; attempt++ {
		recoveries, err := contract.RecoverBackgroundJobs(context.Background(), ident, RecoverBackgroundJobsInput{ErrorMessage: "token secret"})
		if err != nil {
			t.Fatal(err)
		}
		if attempt == 3 {
			if len(recoveries) != 1 || !recoveries[0].Exhausted {
				t.Fatalf("recoveries = %+v", recoveries)
			}
			break
		}
		if _, _, ok, err := contract.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 1}); err != nil || !ok {
			t.Fatalf("retry claim %d ok=%v err=%v", attempt, ok, err)
		}
		base = base.Add(2 * time.Second)
	}
	dead, err := contract.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if dead.QueueStatus != WorkerQueueStatusDegraded || dead.DeadCount != 1 {
		t.Fatalf("dead diagnostics = %+v", dead)
	}
	got, err := contract.GetRun(context.Background(), ident, run.ID)
	if err == nil && got.ErrorMessage != nil && *got.ErrorMessage != "[redacted]" {
		t.Fatalf("run leaked error = %+v", got)
	}
}

func TestPostgresRunEventsUseUniqueSequenceOrdering(t *testing.T) {
	databaseURL := os.Getenv("LOOMI_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("LOOMI_TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	repo := NewPostgresRepository(pool)
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Repository run events", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{ScriptName: "repository_test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AppendRunEvent(ctx, ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: "context_loaded", Summary: "Context loaded"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AppendRunEvent(ctx, ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"}); err != nil {
		t.Fatal(err)
	}
	events, err := repo.ListRunEvents(ctx, ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("events = %+v", events)
	}
	for i, event := range events {
		if event.Sequence != i+1 {
			t.Fatalf("event[%d].Sequence = %d", i, event.Sequence)
		}
	}
}
