package productdata

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
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
	if _, _, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "notes.txt"}, ArgumentsHash: "hash_conflict", ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("conflicting duplicate err = %v", err)
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
	second, secondEvents, err := repo.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_2", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_2", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("second tool err = %v", err)
	}
	if second.ToolCallID != "tc_2" || len(secondEvents) != 2 || secondEvents[0].Type != EventToolCallRequested || secondEvents[1].Type != EventToolCallApprovalRequired {
		t.Fatalf("second=%+v events=%+v", second, secondEvents)
	}
	diagnostics, err := repo.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics.BlockedToolApprovalCount != 2 {
		t.Fatalf("BlockedToolApprovalCount = %d, want 2", diagnostics.BlockedToolApprovalCount)
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

func TestRepositoryContractCancelsUnresolvedToolCallsWhenRunStops(t *testing.T) {
	var repo Repository = NewMemoryService()
	ctx := context.Background()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Tool cancellation", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_deny", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_deny", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_sibling", ToolName: ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": "https://example.com"}, ArgumentsHash: "hash_sibling", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.DenyToolCall(ctx, ident, thread.ID, run.ID, "tc_deny"); err != nil {
		t.Fatal(err)
	}
	sibling, err := repo.GetToolCall(ctx, ident, thread.ID, run.ID, "tc_sibling")
	if err != nil {
		t.Fatal(err)
	}
	if sibling.ExecutionStatus != ToolCallExecutionCancelled {
		t.Fatalf("sibling execution status = %s", sibling.ExecutionStatus)
	}
	state, err := repo.GetRunStepState(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.PendingToolCalls) != 0 || state.NextAction != RunStepNextActionTerminal {
		t.Fatalf("state = %+v", state)
	}

	thread2, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Stop executing", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run2, err := repo.StartRun(ctx, ident, thread2.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_2", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(ctx, ident, run2.ID, RecordToolCallRequestInput{ToolCallID: "tc_exec", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_exec", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.ApproveToolCall(ctx, ident, thread2.ID, run2.ID, "tc_exec"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.StartToolCallExecution(ctx, ident, thread2.ID, run2.ID, "tc_exec"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.StopRun(ctx, ident, run2.ID); err != nil {
		t.Fatal(err)
	}
	executing, err := repo.GetToolCall(ctx, ident, thread2.ID, run2.ID, "tc_exec")
	if err != nil {
		t.Fatal(err)
	}
	if executing.ExecutionStatus != ToolCallExecutionCancelled {
		t.Fatalf("executing status = %s", executing.ExecutionStatus)
	}
	state2, err := repo.GetRunStepState(ctx, ident, run2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(state2.PendingToolCalls) != 0 || state2.NextAction != RunStepNextActionTerminal {
		t.Fatalf("state2 = %+v", state2)
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

func TestRepositoryContractRejectsDenyAfterApproveBeforeExecution(t *testing.T) {
	var repo Repository = NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := repo.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 approve then deny", Mode: ThreadModeChat})
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
	approved, approvedEvents, err := repo.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if approved.ApprovalStatus != ToolCallApprovalApproved || approved.ExecutionStatus != ToolCallExecutionNotStarted || len(approvedEvents) != 1 {
		t.Fatalf("approved=%+v events=%+v", approved, approvedEvents)
	}

	if _, _, err := repo.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("deny after approve err = %v", err)
	}
	got, err := repo.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ApprovalStatus != ToolCallApprovalApproved || got.ExecutionStatus != ToolCallExecutionNotStarted {
		t.Fatalf("tool call changed after rejected deny: %+v", got)
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

func TestRepositoryContractRecoversExecutingToolCallAfterExpiredLease(t *testing.T) {
	repo := NewMemoryService()
	var contract Repository = repo
	ident := identity.LocalDevIdentity()
	base := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	repo.now = func() time.Time { return base }
	thread, err := contract.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M6 tool recovery", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := contract.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := contract.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_stale_tool", LeaseSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	if _, _, err := contract.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_read", ToolName: ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "notes.txt"}, ArgumentsHash: "hash_read", ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := contract.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_read"); err != nil {
		t.Fatal(err)
	}
	base = base.Add(2 * time.Second)
	recoveries, err := contract.RecoverBackgroundJobs(context.Background(), ident, RecoverBackgroundJobsInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveries) != 1 || recoveries[0].Job.ID != job.ID || recoveries[0].Exhausted {
		t.Fatalf("recoveries = %+v", recoveries)
	}
	call, err := contract.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != ToolCallExecutionNotStarted {
		t.Fatalf("tool call = %+v", call)
	}
	state, err := contract.GetRunStepState(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.NextAction != RunStepNextActionExecuteTool || len(state.PendingToolCalls) != 1 || state.PendingToolCalls[0].Status != RunStepStatusApproved {
		t.Fatalf("step state = %+v", state)
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

func TestRepositoryContractRecordsMultipleAutoApprovedToolCallsInOneTurn(t *testing.T) {
	repositoryContractRecordsMultipleAutoApprovedToolCallsInOneTurn(t, NewMemoryService())
}

func TestRepositoryContractRecordsMixedAutoApprovedAndApprovalRequiredToolCallsInOneTurn(t *testing.T) {
	repositoryContractRecordsMixedAutoApprovedAndApprovalRequiredToolCallsInOneTurn(t, NewMemoryService())
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
	state, err := repo.GetRunStepState(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt := RebuildRunStepState(events)
	if state.LastEventSequence != rebuilt.LastEventSequence || len(state.Steps) != len(rebuilt.Steps) {
		t.Fatalf("state = %+v, rebuilt = %+v", state, rebuilt)
	}
}

func TestPostgresRecordsMultipleAutoApprovedToolCallsInOneTurn(t *testing.T) {
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

	repositoryContractRecordsMultipleAutoApprovedToolCallsInOneTurn(t, NewPostgresRepository(pool))
}

func TestPostgresRecordsMixedAutoApprovedAndApprovalRequiredToolCallsInOneTurn(t *testing.T) {
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

	repositoryContractRecordsMixedAutoApprovedAndApprovalRequiredToolCallsInOneTurn(t, NewPostgresRepository(pool))
}

func TestPostgresRejectsConflictingDuplicateToolCallID(t *testing.T) {
	databaseURL := os.Getenv("LOOMI_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("LOOMI_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewPostgresRepository(pool)
	ident := postgresTestIdentity()
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Duplicate tool id", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	call, events, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_dup", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_same", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatal(err)
	}
	again, againEvents, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_dup", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_same", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != call.ID || len(events) != 2 || len(againEvents) != 0 {
		t.Fatalf("call=%+v again=%+v events=%+v againEvents=%+v", call, again, events, againEvents)
	}
	if _, _, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_dup", ToolName: ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "notes.txt"}, ArgumentsHash: "hash_conflict", ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("conflicting duplicate err = %v", err)
	}
}

func TestPostgresClaimToolContinuationUsesRunStepProjection(t *testing.T) {
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
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Continuation claim projection", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_projection", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_projection", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_projection", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.ApproveToolCall(ctx, ident, thread.ID, run.ID, "tc_projection"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.StartToolCallExecution(ctx, ident, thread.ID, run.ID, "tc_projection"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.CompleteToolCallSuccess(ctx, ident, thread.ID, run.ID, "tc_projection", map[string]any{"iso_time": "2026-05-29T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	before, err := repo.GetRunStepState(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if before.NextAction != RunStepNextActionContinueModel {
		t.Fatalf("before state = %+v", before)
	}

	claimed, ok, err := repo.ClaimToolContinuation(ctx, ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_projection", JobID: "job_projection", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || claimed.Type != "model_request_started" || claimed.Sequence <= before.LastEventSequence {
		t.Fatalf("claim ok=%v event=%+v before=%+v", ok, claimed, before)
	}
	after, err := repo.GetRunStepState(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.LastEventSequence != claimed.Sequence || after.LastContinuationSequence != claimed.Sequence {
		t.Fatalf("after state = %+v, claimed = %+v", after, claimed)
	}
	if _, ok, err := repo.ClaimToolContinuation(ctx, ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_projection", JobID: "job_projection", ProviderID: "custom", Model: "model"}); err != nil || ok {
		t.Fatalf("duplicate claim ok=%v err=%v", ok, err)
	}
}

func TestPostgresRunStepProjectionRebuildsSemanticCorruption(t *testing.T) {
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
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Corrupt projection", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := repo.CreateMessage(ctx, ident, thread.ID, CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetRunStepState(ctx, ident, run.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `update run_step_state_projections set last_sequence=2, state='{}'::jsonb where run_id=$1`, run.ID); err != nil {
		t.Fatal(err)
	}

	state, err := repo.GetRunStepState(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.LastEventSequence != 2 || state.TriggerMessageID != message.ID || state.ProviderID != "custom" {
		t.Fatalf("state = %+v", state)
	}
}

func repositoryContractRecordsMixedAutoApprovedAndApprovalRequiredToolCallsInOneTurn(t *testing.T, repo Repository) {
	t.Helper()
	ctx := context.Background()
	ident := identity.LocalIdentity{UserID: "user_" + NewThreadID(), DisplayName: "Tool Contract", Source: "test"}
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Mixed tools", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := repo.CreateMessage(ctx, ident, thread.ID, CreateMessageInput{Content: "read then ask approval"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	autoCall, _, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{
		ToolCallID:       "tc_read",
		ToolName:         ToolNameWorkspaceRead,
		ArgumentsSummary: map[string]any{"path": "notes.txt", "limit": 128},
		ApprovalStatus:   ToolCallApprovalApproved,
		ExecutionStatus:  ToolCallExecutionNotStarted,
	})
	if err != nil {
		t.Fatal(err)
	}
	blockedCall, _, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{
		ToolCallID:       "tc_time",
		ToolName:         ToolNameCurrentTime,
		ArgumentsSummary: map[string]any{"timezone": "UTC"},
		ApprovalStatus:   ToolCallApprovalRequired,
		ExecutionStatus:  ToolCallExecutionBlocked,
	})
	if err != nil {
		t.Fatal(err)
	}
	if autoCall.ExecutionStatus != ToolCallExecutionNotStarted || blockedCall.ExecutionStatus != ToolCallExecutionBlocked {
		t.Fatalf("calls = %+v %+v", autoCall, blockedCall)
	}
	state, err := repo.GetRunStepState(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.PendingToolCalls) != 2 {
		t.Fatalf("pending = %+v", state.PendingToolCalls)
	}
}

func TestPostgresConcurrentRunEventInsertsSerializeSequence(t *testing.T) {
	databaseURL := os.Getenv("LOOMI_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("LOOMI_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	repo := NewPostgresRepository(pool)
	ident := postgresTestIdentity()
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Repository concurrent run events", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{ScriptName: "repository_concurrent_sequence"})
	if err != nil {
		t.Fatal(err)
	}

	const workers = 24
	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			tx, err := pool.Begin(ctx)
			if err != nil {
				errs <- err
				return
			}
			defer tx.Rollback(ctx)
			<-start
			if _, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, "concurrent_event", "Concurrent event", nil, map[string]any{"index": index}); err != nil {
				errs <- err
				return
			}
			if err := tx.Commit(ctx); err != nil {
				errs <- err
				return
			}
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	events, err := repo.ListRunEvents(ctx, ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != workers+2 {
		t.Fatalf("events = %d, want %d", len(events), workers+2)
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

func TestPostgresApproveMemoryWriteRollsBackWhenAuditFails(t *testing.T) {
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
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "PG memory rollback", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{ScriptName: "pg_memory_rollback"})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := repo.ProposeMemoryWrite(ctx, ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeThread, ScopeID: thread.ID, Title: "Rollback", Content: "Keep audit atomic", SourceThreadID: thread.ID, SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}

	const triggerName = "loomi_test_fail_memory_audit_insert"
	const functionName = "loomi_test_fail_memory_audit_insert"
	_, _ = pool.Exec(ctx, `drop trigger if exists `+triggerName+` on memory_audit_events`)
	_, _ = pool.Exec(ctx, `drop function if exists `+functionName+`()`)
	defer pool.Exec(context.Background(), `drop trigger if exists `+triggerName+` on memory_audit_events`)
	defer pool.Exec(context.Background(), `drop function if exists `+functionName+`()`)
	_, err = pool.Exec(ctx, fmt.Sprintf(`
create function %s() returns trigger language plpgsql as $$
begin
	if new.type = %q and new.metadata->>'memory_proposal_id' = %q then
		raise exception 'memory audit insert failed for rollback test';
	end if;
	return new;
end;
$$`, functionName, EventMemoryWriteApproved, proposal.ID))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `create trigger `+triggerName+` before insert on memory_audit_events for each row execute function `+functionName+`() `); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.ApproveMemoryWrite(ctx, ident, proposal.ID, MemoryWriteDecisionInput{Reason: "approve"}); err == nil {
		t.Fatal("ApproveMemoryWrite succeeded, want audit failure")
	}
	var entryCount int
	if err := pool.QueryRow(ctx, `select count(*) from memory_entries where user_id=$1 and source_run_id=$2`, ident.UserID, run.ID).Scan(&entryCount); err != nil {
		t.Fatal(err)
	}
	if entryCount != 0 {
		t.Fatalf("memory_entries count = %d, want rollback to 0", entryCount)
	}
	var status string
	var createdEntryID string
	if err := pool.QueryRow(ctx, `select status, coalesce(created_entry_id,'') from memory_write_proposals where id=$1`, proposal.ID).Scan(&status, &createdEntryID); err != nil {
		t.Fatal(err)
	}
	if status != string(MemoryWritePending) || createdEntryID != "" {
		t.Fatalf("proposal status=%q created_entry_id=%q, want pending without entry", status, createdEntryID)
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
	startedTask, err := repo.StartAgentTask(ctx, ident, StartAgentTaskInput{ThreadID: threadA.ID, TaskID: task.ID})
	if err != nil {
		t.Fatal(err)
	}
	if startedTask.Status != AgentTaskStatusInProgress {
		t.Fatalf("started task = %+v", startedTask)
	}
	completed, err := repo.CompleteAgentTask(ctx, ident, CompleteAgentTaskInput{ThreadID: threadA.ID, TaskID: task.ID, ResultSummary: "PG path works"})
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != AgentTaskStatusCompleted || completed.ResultSummary != "PG path works" {
		t.Fatalf("completed = %+v", completed)
	}
	if _, err := repo.StartAgentTask(ctx, ident, StartAgentTaskInput{ThreadID: threadA.ID, TaskID: task.ID}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("terminal start err = %v", err)
	}
	failedTask, err := repo.SpawnAgentTask(ctx, ident, SpawnAgentTaskInput{ThreadID: threadA.ID, RunID: run.ID, Role: "researcher", Goal: "Research PG path"})
	if err != nil {
		t.Fatal(err)
	}
	failed, err := repo.FailAgentTask(ctx, ident, FailAgentTaskInput{ThreadID: threadA.ID, TaskID: failedTask.ID, ResultSummary: "PG path blocked"})
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != AgentTaskStatusFailed || failed.ResultSummary != "PG path blocked" {
		t.Fatalf("failed = %+v", failed)
	}
	if _, err := repo.CompleteAgentTask(ctx, ident, CompleteAgentTaskInput{ThreadID: threadA.ID, TaskID: failed.ID, ResultSummary: "Wrong terminal transition"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("terminal complete err = %v", err)
	}
	if _, err := repo.CompleteAgentTask(ctx, ident, CompleteAgentTaskInput{ThreadID: threadB.ID, TaskID: task.ID, ResultSummary: "Wrong thread"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("cross-thread task err = %v", err)
	}
}

func TestRepositoryContractDelegateAgentTaskRetryIsIdempotent(t *testing.T) {
	repositoryContractDelegateAgentTaskRetryIsIdempotent(t, NewMemoryService())
}

func TestPostgresDelegateAgentTaskRetryIsIdempotent(t *testing.T) {
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

	repositoryContractDelegateAgentTaskRetryIsIdempotent(t, NewPostgresRepository(pool))
}

func TestPostgresReconcilesDelegatedAgentTaskAfterChildRunCompletes(t *testing.T) {
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

	repositoryContractReconcilesDelegatedAgentTaskAfterChildRunCompletes(t, NewPostgresRepository(pool))
}

func repositoryContractReconcilesDelegatedAgentTaskAfterChildRunCompletes(t *testing.T, repo interface {
	Repository
	AgentTaskService
}) {
	t.Helper()
	ctx := context.Background()
	ident := identity.LocalIdentity{UserID: "user_" + NewThreadID(), DisplayName: "Delegate Reconcile", Source: "test"}
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Delegate reconcile", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := repo.CreateMessage(ctx, ident, thread.ID, CreateMessageInput{Content: "Delegate a child review"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := repo.SpawnAgentTask(ctx, ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review delegate reconcile"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_delegate_pg", ToolName: ToolNameAgentDelegate, ArgumentsSummary: map[string]any{"task_id": task.ID}, ArgumentsHash: "hash_delegate_pg", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.ApproveToolCall(ctx, ident, thread.ID, run.ID, "tc_delegate_pg"); err != nil {
		t.Fatal(err)
	}
	parentJob, _, ok, err := repo.ClaimBackgroundJob(ctx, ident, ClaimBackgroundJobInput{WorkerID: "worker_parent_delegate_pg", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || parentJob.RunID != run.ID {
		t.Fatalf("parent claim ok=%v job=%+v", ok, parentJob)
	}
	if _, _, err := repo.StartToolCallExecution(ctx, ident, thread.ID, run.ID, "tc_delegate_pg"); err != nil {
		t.Fatal(err)
	}
	if _, changed, err := repo.CompleteBackgroundJob(ctx, ident, CompleteBackgroundJobInput{JobID: parentJob.ID, WorkerID: "worker_parent_delegate_pg", OwnershipVersion: parentJob.OwnershipVersion}); err != nil || !changed {
		t.Fatalf("complete parent job changed=%v err=%v", changed, err)
	}
	delegated, err := repo.DelegateAgentTask(ctx, ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ParentToolCallID: "tc_delegate_pg"})
	if err != nil {
		t.Fatal(err)
	}
	if delegated.ChildThreadID == "" || delegated.ChildRunID == "" || delegated.ChildThreadID == thread.ID || delegated.ChildRunID == run.ID {
		t.Fatalf("delegated = %+v", delegated)
	}
	if _, err := repo.AppendAssistantMessage(ctx, ident, delegated.ChildThreadID, AppendAssistantMessageInput{Content: "Postgres child review: no issues."}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AppendRunEvent(ctx, ident, delegated.ChildRunID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: EventRunCompleted, Summary: "Child run completed"}); err != nil {
		t.Fatal(err)
	}

	reconciled, err := repo.ReconcileAgentTaskChildRuns(ctx, ident, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(reconciled) != 1 || reconciled[0].Run.ID != run.ID || reconciled[0].Task.ID != task.ID {
		t.Fatalf("reconciled = %+v", reconciled)
	}
	if len(reconciled[0].Events) != 2 || reconciled[0].Events[0].Type != EventToolCallSucceeded || reconciled[0].Events[1].Type != EventRunQueued {
		t.Fatalf("reconciled events = %+v", reconciled[0].Events)
	}
	call, err := repo.GetToolCall(ctx, ident, thread.ID, run.ID, "tc_delegate_pg")
	if err != nil {
		t.Fatal(err)
	}
	if call.ExecutionStatus != ToolCallExecutionSucceeded || call.ResultSummary["child_status"] != string(RunStatusCompleted) || !strings.Contains(fmt.Sprint(call.ResultSummary["result_summary"]), "no issues") {
		t.Fatalf("call = %+v", call)
	}
	tasks, err := repo.ListAgentTasks(ctx, ident, ListAgentTasksInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Status != AgentTaskStatusCompleted || !strings.Contains(tasks[0].ResultSummary, "no issues") {
		t.Fatalf("tasks = %+v", tasks)
	}
	resumeJob, _, ok, err := repo.ClaimBackgroundJob(ctx, ident, ClaimBackgroundJobInput{WorkerID: "worker_parent_resume_pg", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || resumeJob.RunID != run.ID || resumeJob.Metadata["resume_reason"] != "agent_child_run_completed" || resumeJob.Metadata["child_run_id"] != delegated.ChildRunID {
		t.Fatalf("resume claim ok=%v job=%+v", ok, resumeJob)
	}
}

func repositoryContractDelegateAgentTaskRetryIsIdempotent(t *testing.T, repo interface {
	Repository
	AgentTaskService
}) {
	t.Helper()
	ctx := context.Background()
	ident := identity.LocalIdentity{UserID: "user_" + NewThreadID(), DisplayName: "Delegate Retry", Source: "test"}
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Delegate retry", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_delegate_retry", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := repo.SpawnAgentTask(ctx, ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review retry idempotency"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := repo.DelegateAgentTask(ctx, ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ParentToolCallID: "tc_delegate_retry"})
	if err != nil {
		t.Fatal(err)
	}
	retry, err := repo.DelegateAgentTask(ctx, ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ParentToolCallID: "tc_delegate_retry"})
	if err != nil {
		t.Fatalf("same parent tool-call delegate retry err = %v", err)
	}
	if retry.ChildThreadID != first.ChildThreadID || retry.ChildRunID != first.ChildRunID || retry.ParentToolCallID != "tc_delegate_retry" {
		t.Fatalf("retry delegate = %+v, first = %+v", retry, first)
	}
	if _, err := repo.DelegateAgentTask(ctx, ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ParentToolCallID: "tc_delegate_other"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("different parent tool-call duplicate err = %v", err)
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

func repositoryContractRecordsMultipleAutoApprovedToolCallsInOneTurn(t *testing.T, repo Repository) {
	t.Helper()
	ctx := context.Background()
	ident := identity.LocalIdentity{UserID: "user_" + NewThreadID(), DisplayName: "Tool Contract", Source: "test"}
	thread, err := repo.CreateThread(ctx, ident, CreateThreadInput{Title: "Parallel tools", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := repo.CreateMessage(ctx, ident, thread.ID, CreateMessageInput{Content: "read several things"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	first, firstEvents, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{
		ToolCallID:       "tc_search",
		ToolName:         ToolNameWebSearch,
		ArgumentsSummary: map[string]any{"query": "agent runtime"},
		ApprovalStatus:   ToolCallApprovalApproved,
		ExecutionStatus:  ToolCallExecutionNotStarted,
	})
	if err != nil {
		t.Fatal(err)
	}
	second, secondEvents, err := repo.RecordToolCallRequest(ctx, ident, run.ID, RecordToolCallRequestInput{
		ToolCallID:       "tc_fetch",
		ToolName:         ToolNameWebFetch,
		ArgumentsSummary: map[string]any{"url": "https://example.com"},
		ApprovalStatus:   ToolCallApprovalApproved,
		ExecutionStatus:  ToolCallExecutionNotStarted,
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.ToolCallID != "tc_search" || second.ToolCallID != "tc_fetch" {
		t.Fatalf("tool calls = %+v %+v", first, second)
	}
	if len(firstEvents) != 2 || len(secondEvents) != 2 {
		t.Fatalf("events = %+v %+v", firstEvents, secondEvents)
	}
	got, err := repo.GetRun(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != RunStatusQueued {
		t.Fatalf("run = %+v", got)
	}
	state, err := repo.GetRunStepState(ctx, ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.NextAction != RunStepNextActionExecuteTool || len(state.PendingToolCalls) != 2 {
		t.Fatalf("state = %+v", state)
	}
	pending := map[string]string{}
	for _, call := range state.PendingToolCalls {
		pending[call.ToolCallID] = eventValueString(call.SafeMetadata["execution_status"])
	}
	if pending["tc_search"] != string(ToolCallExecutionNotStarted) || pending["tc_fetch"] != string(ToolCallExecutionNotStarted) {
		t.Fatalf("pending = %+v", state.PendingToolCalls)
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
