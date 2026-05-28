package productdata

import (
	"context"
	"os"
	"strings"
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

func TestRepositoryContractPreparesRunContext(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M9 context", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := repo.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := repo.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_context", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	context, err := repo.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if context.Run.ID != run.ID || context.Thread.ID != thread.ID || len(context.Messages) != 1 || context.ProviderRoute.ProviderID != "custom" {
		t.Fatalf("context = %+v", context)
	}
}

func TestRepositoryContractPreservesThreadPersonaOnMetadataUpdate(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	persona := syncContractPersona(t, repo, ident, "contract-thread-persona")
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Persona thread", Mode: ThreadModeChat, PersonaID: persona.ID})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := repo.UpdateThread(context.Background(), ident, thread.ID, UpdateThreadInput{Title: ptr("Renamed persona thread")})
	if err != nil {
		t.Fatal(err)
	}
	if updated.PersonaID != persona.ID {
		t.Fatalf("updated persona id = %q, want %q", updated.PersonaID, persona.ID)
	}
	got, err := repo.GetThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.PersonaID != persona.ID {
		t.Fatalf("got persona id = %q, want %q", got.PersonaID, persona.ID)
	}
	threads, err := repo.ListThreads(context.Background(), ident, false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, candidate := range threads {
		if candidate.ID == thread.ID {
			found = true
			if candidate.PersonaID != persona.ID {
				t.Fatalf("listed persona id = %q, want %q", candidate.PersonaID, persona.ID)
			}
		}
	}
	if !found {
		t.Fatalf("thread %s not listed", thread.ID)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{ScriptName: "persona_contract"})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := repo.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_persona_contract", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	context, err := repo.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if context.Run.ID != run.ID || context.Persona.ID != persona.ID || context.Persona.ResolvedFrom != PersonaResolvedFromThread {
		t.Fatalf("context persona = %+v", context.Persona)
	}
}

func TestRepositoryContractRejectsUnknownThreadPersona(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Unknown persona", Mode: ThreadModeChat, PersonaID: "persona_unknown"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("create err = %v", err)
	}
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Thread", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateThread(context.Background(), ident, thread.ID, UpdateThreadInput{PersonaID: ptr("persona_unknown")}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("update err = %v", err)
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

func TestRepositoryContractCoversMCPToolCallRequestProjection(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "MCP projection", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	call, events, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_mcp_1", ToolName: "mcp.local-search.search", CandidateSchemaHash: "sha256:test-local-search", ArgumentsSummary: map[string]any{"query": "status", "api_key": "sk-secret"}, ArgumentsHash: "hash_mcp", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != "mcp.local-search.search" || call.CandidateSchemaHash != "sha256:test-local-search" || call.ArgumentsSummary["api_key"] == "sk-secret" {
		t.Fatalf("call = %+v", call)
	}
	if len(events) != 2 || events[1].Metadata["tool_source"] != ToolSourceMCP || events[1].Metadata["server_slug"] != "local-search" || events[1].Metadata["candidate_schema_hash"] != "sha256:test-local-search" {
		t.Fatalf("events = %+v", events)
	}
	again, againEvents, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_mcp_1", ToolName: "mcp.local-search.search", CandidateSchemaHash: "sha256:test-local-search", ArgumentsSummary: map[string]any{"query": "status"}, ArgumentsHash: "hash_mcp", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != call.ID || len(againEvents) != 0 {
		t.Fatalf("again=%+v events=%+v", again, againEvents)
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
	ident := postgresTestIdentity()
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
	if len(events) != 4 {
		t.Fatalf("events = %+v", events)
	}
	for i, event := range events {
		if event.Sequence != i+1 {
			t.Fatalf("event[%d].Sequence = %d", i, event.Sequence)
		}
	}
}

func TestPostgresMemoryEntryScopeAndTerminalAudit(t *testing.T) {
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
	ident := postgresTestIdentity()
	threadA, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "PG Memory A", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	threadB, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "PG Memory B", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, threadA.ID, StartRunInput{ScriptName: "pg_terminal_memory_audit"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AppendRunEvent(ctx, ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: EventRunCompleted, Summary: "Run completed"}); err != nil {
		t.Fatal(err)
	}

	proposal, err := repo.ProposeMemoryWrite(ctx, ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeThread, ScopeID: threadA.ID, Title: "PG terminal", Content: "Keep terminal audit in PG", SourceThreadID: threadA.ID, SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := repo.ApproveMemoryWrite(ctx, ident, proposal.ID, MemoryWriteDecisionInput{Reason: "approve"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetMemoryEntry(ctx, ident, decision.Entry.ID, MemoryEntryAccessInput{ScopeType: MemoryScopeThread, ScopeID: threadB.ID}); err == nil || ErrorCode(err) != CodeMemoryNotFound {
		t.Fatalf("thread B read err = %v", err)
	}
	if _, err := repo.DeleteMemoryEntry(ctx, ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "wrong thread", ScopeType: MemoryScopeThread, ScopeID: threadB.ID}); err == nil || ErrorCode(err) != CodeMemoryNotFound {
		t.Fatalf("thread B delete err = %v", err)
	}
	if _, err := repo.GetMemoryEntry(ctx, ident, decision.Entry.ID, MemoryEntryAccessInput{ScopeType: MemoryScopeThread, ScopeID: threadA.ID}); err != nil {
		t.Fatal(err)
	}

	denied, err := repo.ProposeMemoryWrite(ctx, ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeThread, ScopeID: threadA.ID, Title: "PG deny", Content: "Deny once", SourceThreadID: threadA.ID, SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DenyMemoryWrite(ctx, ident, denied.ID, MemoryWriteDecisionInput{Reason: "deny"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DenyMemoryWrite(ctx, ident, denied.ID, MemoryWriteDecisionInput{Reason: "retry deny"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DeleteMemoryEntry(ctx, ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "delete", ScopeType: MemoryScopeThread, ScopeID: threadA.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DeleteMemoryEntry(ctx, ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "retry delete", ScopeType: MemoryScopeThread, ScopeID: threadA.ID}); err != nil {
		t.Fatal(err)
	}

	audit, err := repo.ListMemoryAudit(ctx, ident, MemoryAuditInput{SourceRunID: run.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, item := range audit.Items {
		counts[item.EventType]++
	}
	if counts[EventMemoryWriteProposed] != 2 || counts[EventMemoryWriteApproved] != 1 || counts[EventMemoryWriteDenied] != 1 || counts["memory_deleted"] != 1 {
		t.Fatalf("audit counts = %+v items=%+v", counts, audit.Items)
	}
}

func TestPostgresArtifactsAndAgentTasksUseThreadScope(t *testing.T) {
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
	var _ ArtifactService = repo
	var _ AgentTaskService = repo
	ident := postgresTestIdentity()
	threadA, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "PG runtime A", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	threadB, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "PG runtime B", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, threadA.ID, StartRunInput{ScriptName: "pg_artifact_agent_scope"})
	if err != nil {
		t.Fatal(err)
	}

	artifact, err := repo.CreateArtifact(ctx, ident, CreateArtifactInput{ThreadID: threadA.ID, RunID: run.ID, Title: " Notes ", Content: "hello postgres artifact", MaxBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	if artifact.ThreadID != threadA.ID || artifact.Title != "Notes" || artifact.ContentBytes != len("hello postgres artifact") {
		t.Fatalf("artifact = %+v", artifact)
	}
	read, err := repo.ReadArtifact(ctx, ident, ReadArtifactInput{ThreadID: threadA.ID, ArtifactID: artifact.ID, MaxBytes: 5})
	if err != nil {
		t.Fatal(err)
	}
	if read.TextExcerpt != "hello" || !read.Truncated || read.Content != "hello postgres artifact" {
		t.Fatalf("read = %+v", read)
	}
	list, err := repo.ListArtifacts(ctx, ident, ListArtifactsInput{ThreadID: threadA.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != artifact.ID || list[0].Content != "" {
		t.Fatalf("list = %+v", list)
	}
	if _, err := repo.ReadArtifact(ctx, ident, ReadArtifactInput{ThreadID: threadB.ID, ArtifactID: artifact.ID}); err == nil || ErrorCode(err) != CodeArtifactNotFound {
		t.Fatalf("cross-thread artifact err = %v", err)
	}

	task, err := repo.SpawnAgentTask(ctx, ident, SpawnAgentTaskInput{ThreadID: threadA.ID, RunID: run.ID, Role: "reviewer", Goal: "Review PG path"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, err := repo.ListAgentTasks(ctx, ident, ListAgentTasksInput{ThreadID: threadA.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID || tasks[0].Goal != "Review PG path" {
		t.Fatalf("tasks = %+v", tasks)
	}
	completed, err := repo.CompleteAgentTask(ctx, ident, CompleteAgentTaskInput{ThreadID: threadA.ID, TaskID: task.ID, ResultSummary: "PG path works"})
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != AgentTaskStatusCompleted || completed.ResultSummary != "PG path works" {
		t.Fatalf("completed = %+v", completed)
	}
	if _, err := repo.CompleteAgentTask(ctx, ident, CompleteAgentTaskInput{ThreadID: threadB.ID, TaskID: task.ID, ResultSummary: "Wrong thread"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("cross-thread task err = %v", err)
	}
}

func TestPostgresSandboxProcessRecordsAreDurableAndScoped(t *testing.T) {
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
	var _ SandboxProcessRepository = repo
	ident := postgresTestIdentity()
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "PG sandbox process", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{ScriptName: "pg_sandbox_process"})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	exitCode := 0
	record := SandboxProcessRecord{
		RunID:           run.ID,
		ProcessID:       "sp_pg_sandbox",
		ArgvSummary:     []string{"cat", "token=secret", "/Users/xuean/private"},
		CwdAlias:        "/Users/xuean/Repos/personal-projects/Loomi",
		Status:          SandboxProcessStatusExited,
		Cursor:          12,
		StartedAt:       now.Add(-time.Second),
		UpdatedAt:       now,
		EndedAt:         &now,
		ExitCode:        &exitCode,
		StdoutTail:      "tail token=secret /Users/xuean/private",
		StdoutCursor:    12,
		StderrTail:      "stderr token=secret",
		StderrCursor:    18,
		StdoutBytes:     34,
		StderrBytes:     18,
		TerminalSummary: "exited exit_code=0 token=secret /Users/xuean/private",
		OutputLimit:     1024,
	}
	if err := repo.SaveSandboxProcess(ctx, record); err != nil {
		t.Fatal(err)
	}
	records, err := repo.ListSandboxProcesses(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var got SandboxProcessRecord
	for _, candidate := range records {
		if candidate.ProcessID == record.ProcessID {
			got = candidate
			break
		}
	}
	if got.ProcessID == "" {
		t.Fatalf("process record not found in %+v", records)
	}
	rendered := strings.Join(append(append([]string{}, got.ArgvSummary...), got.CwdAlias, got.StdoutTail, got.StderrTail, got.TerminalSummary), "\n")
	if strings.Contains(rendered, "/Users/") || strings.Contains(rendered, "token=secret") {
		t.Fatalf("record leaked unsafe data: %+v", got)
	}
	if got.RunID != record.RunID || got.Status != SandboxProcessStatusExited || got.Cursor != 12 || got.StdoutBytes != 34 || got.StderrBytes != 18 || got.ExitCode == nil || *got.ExitCode != 0 {
		t.Fatalf("record = %+v", got)
	}
	if deleted, err := repo.DeleteSandboxProcessesUpdatedBefore(ctx, now.Add(time.Second)); err != nil || deleted < 1 {
		t.Fatalf("deleted=%d err=%v", deleted, err)
	}
}

func TestPostgresPreservesThreadPersonaOnMetadataUpdate(t *testing.T) {
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
	ident := postgresTestIdentity()
	persona := syncContractPersona(t, repo, ident, "postgres-thread-persona-"+NewThreadID())
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Postgres persona thread", Mode: ThreadModeChat, PersonaID: persona.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateThread(ctx, ident, thread.ID, UpdateThreadInput{Title: ptr("Renamed postgres persona thread")}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetThread(ctx, ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.PersonaID != persona.ID {
		t.Fatalf("got persona id = %q, want %q", got.PersonaID, persona.ID)
	}
	threads, err := repo.ListThreads(ctx, ident, false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, candidate := range threads {
		if candidate.ID == thread.ID {
			found = true
			if candidate.PersonaID != persona.ID {
				t.Fatalf("listed persona id = %q, want %q", candidate.PersonaID, persona.ID)
			}
		}
	}
	if !found {
		t.Fatalf("thread %s not listed", thread.ID)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{ScriptName: "postgres_persona_contract"})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := repo.ClaimBackgroundJob(ctx, ident, ClaimBackgroundJobInput{WorkerID: "worker_postgres_persona_contract", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	context, err := repo.PrepareRunContext(ctx, ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if context.Run.ID != run.ID || context.Persona.ID != persona.ID || context.Persona.ResolvedFrom != PersonaResolvedFromThread {
		t.Fatalf("context persona = %+v", context.Persona)
	}
}

func TestPostgresRejectsUnknownThreadPersona(t *testing.T) {
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
	ident := postgresTestIdentity()
	if _, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Unknown postgres persona", Mode: ThreadModeChat, PersonaID: "persona_unknown"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("create err = %v", err)
	}
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Postgres thread", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateThread(ctx, ident, thread.ID, UpdateThreadInput{PersonaID: ptr("persona_unknown")}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("update err = %v", err)
	}
	persona := syncContractPersona(t, repo, ident, "postgres-inactive-thread-persona-"+NewThreadID())
	if _, err := pool.Exec(ctx, `update personas set is_active=false where id=$1`, persona.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Inactive postgres persona", Mode: ThreadModeChat, PersonaID: persona.ID}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("inactive create err = %v", err)
	}
	if _, err := repo.UpdateThread(ctx, ident, thread.ID, UpdateThreadInput{PersonaID: ptr(persona.ID)}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("inactive update err = %v", err)
	}
}

func TestPostgresSyncBuiltInPersonasUpdatesExistingVersionDefinition(t *testing.T) {
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
	ident := postgresTestIdentity()
	config := BuiltInPersonaConfig{
		Slug:             "postgres-persona-update-" + NewThreadID(),
		Name:             "Postgres Persona Update",
		Description:      "Persona update fixture.",
		SystemPrompt:     "prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{ToolNameCurrentTime},
		ReasoningMode:    "balanced",
		BudgetSummary:    "budget",
		Version:          "2026-05-27.2",
		IsDefault:        true,
	}
	if _, err := repo.SyncBuiltInPersonas(ctx, ident, []BuiltInPersonaConfig{config}); err != nil {
		t.Fatal(err)
	}
	config.AllowedToolNames = []string{ToolNameCurrentTime, ToolNameWorkspaceTreeSummary, ToolNameWorkspaceListDirectory}
	if _, err := repo.SyncBuiltInPersonas(ctx, ident, []BuiltInPersonaConfig{config}); err != nil {
		t.Fatal(err)
	}
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Postgres work", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := repo.CreateMessage(ctx, ident, thread.ID, CreateMessageInput{Content: "分类当前目录"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := repo.ClaimBackgroundJob(ctx, ident, ClaimBackgroundJobInput{WorkerID: "worker_postgres_persona_update", LeaseSeconds: 5})
	if err != nil || !ok {
		t.Fatalf("claim ok=%v err=%v", ok, err)
	}
	context, err := repo.PrepareRunContext(ctx, ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if !hasToolResolution(context.EnabledTools, ToolNameWorkspaceTreeSummary) || !hasToolResolution(context.EnabledTools, ToolNameWorkspaceListDirectory) {
		t.Fatalf("enabled tools = %+v", context.EnabledTools)
	}
}

func postgresTestIdentity() identity.LocalIdentity {
	id := "user_" + NewThreadID()
	return identity.LocalIdentity{UserID: id, DisplayName: "Postgres Test", Source: "test"}
}

func syncContractPersona(t *testing.T, repo Repository, ident identity.LocalIdentity, slug string) Persona {
	t.Helper()
	_, err := repo.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{{
		Slug:             slug,
		Name:             "Contract Persona",
		Description:      "Persona contract fixture.",
		SystemPrompt:     "contract prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "contract-model"},
		AllowedToolNames: []string{ToolNameCurrentTime},
		ReasoningMode:    "balanced",
		BudgetSummary:    "contract budget",
		Version:          "1",
		IsDefault:        true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	personas, err := repo.ListPersonas(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	for _, persona := range personas {
		if persona.Slug == slug {
			return persona
		}
	}
	t.Fatalf("persona slug %q not found in %+v", slug, personas)
	return Persona{}
}
