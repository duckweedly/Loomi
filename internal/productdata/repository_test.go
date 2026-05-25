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

func TestRepositoryContractCoversM7ToolCallRequestProjection(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	call, events, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatal(err)
	}
	if call.ThreadID != thread.ID || call.RunID != run.ID || call.ToolCallID != "tc_1" || call.ApprovalStatus != ToolCallApprovalRequired || call.ExecutionStatus != ToolCallExecutionBlocked {
		t.Fatalf("call = %+v", call)
	}
	if call.ArgumentsSummary["timezone"] != "UTC" {
		t.Fatalf("arguments summary = %+v", call.ArgumentsSummary)
	}
	if len(events) != 2 || events[0].Type != EventToolCallRequested || events[1].Type != EventToolCallApprovalRequired {
		t.Fatalf("events = %+v", events)
	}
	again, againEvents, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != call.ID || len(againEvents) != 0 {
		t.Fatalf("again=%+v events=%+v", again, againEvents)
	}
	got, err := repo.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != call.ID || got.ThreadID != thread.ID || got.RunID != run.ID {
		t.Fatalf("got = %+v, call = %+v", got, call)
	}
	if _, err := repo.GetToolCall(context.Background(), ident, "wrong-thread", run.ID, "tc_1"); err == nil || ErrorCode(err) != CodeRunNotFound {
		t.Fatalf("wrong scoped lookup err = %v", err)
	}
	if _, _, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_2", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_2", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("second tool err = %v", err)
	}
	diagnostics, err := repo.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics.BlockedToolApprovalCount != 1 {
		t.Fatalf("BlockedToolApprovalCount = %d, want 1", diagnostics.BlockedToolApprovalCount)
	}
}

func TestRepositoryContractRejectsToolCallsForTerminalRuns(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 terminal", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.StopRun(context.Background(), ident, run.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("terminal run tool call err = %v", err)
	}
}

func TestRepositoryContractApprovesToolCallsIdempotently(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 approve", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}

	call, events, err := repo.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ApprovalStatus != ToolCallApprovalApproved || call.ExecutionStatus != ToolCallExecutionNotStarted {
		t.Fatalf("approved call = %+v", call)
	}
	if len(events) != 1 || events[0].Type != EventToolCallApproved {
		t.Fatalf("events = %+v", events)
	}
	diagnostics, err := repo.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics.BlockedToolApprovalCount != 0 || diagnostics.ResumableToolCallCount != 1 || diagnostics.QueuedCount != 1 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}

	again, againEvents, err := repo.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != call.ID || len(againEvents) != 0 {
		t.Fatalf("again=%+v events=%+v", again, againEvents)
	}
}

func TestRepositoryContractDeniesToolCallsIdempotently(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 deny", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}

	call, events, err := repo.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ApprovalStatus != ToolCallApprovalDenied || call.ExecutionStatus != ToolCallExecutionCancelled {
		t.Fatalf("denied call = %+v", call)
	}
	if len(events) != 2 || events[0].Type != EventToolCallDenied || events[1].Type != EventRunStopped {
		t.Fatalf("events = %+v", events)
	}
	gotRun, err := repo.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotRun.Status != RunStatusStopped {
		t.Fatalf("run = %+v", gotRun)
	}

	again, againEvents, err := repo.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != call.ID || len(againEvents) != 0 {
		t.Fatalf("again=%+v events=%+v", again, againEvents)
	}
}

func TestRepositoryContractRejectsConflictingOrWrongScopeToolDecisions(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 conflicts", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.ApproveToolCall(context.Background(), ident, "wrong-thread", run.ID, "tc_1"); err == nil || ErrorCode(err) != CodeRunNotFound {
		t.Fatalf("wrong scope err = %v", err)
	}
	if _, _, err := repo.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_missing"); err == nil || ErrorCode(err) != CodeRunNotFound {
		t.Fatalf("unknown err = %v", err)
	}
	if _, _, err := repo.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("conflict err = %v", err)
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
	if !recoveries[0].Job.ScheduledAt.After(base) {
		t.Fatalf("retry was not backed off: scheduled_at=%s base=%s", recoveries[0].Job.ScheduledAt, base)
	}
	if _, changed, err := contract.FailBackgroundJob(context.Background(), ident, FailBackgroundJobInput{JobID: job.ID, WorkerID: "worker_stale", OwnershipVersion: job.OwnershipVersion, ErrorCode: "stale", ErrorMessage: "stale"}); err != nil || changed {
		t.Fatalf("stale fail changed=%v err=%v", changed, err)
	}
	for attempt := 2; attempt <= 3; attempt++ {
		base = recoveries[0].Job.ScheduledAt
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
		base = recoveries[0].Job.ScheduledAt
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
