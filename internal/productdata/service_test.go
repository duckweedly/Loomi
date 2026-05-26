package productdata

import (
	"context"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
)

func TestGeneratedIDsDoNotUseProcessLocalSequence(t *testing.T) {
	first := NewThreadID()
	second := NewThreadID()
	if first == "thr_1" || second == "thr_2" {
		t.Fatalf("ids use process-local sequence: %q %q", first, second)
	}
}

func TestCurrentIdentityEnsuresLocalUser(t *testing.T) {
	svc := NewMemoryService()
	user, err := svc.CurrentIdentity(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatalf("CurrentIdentity() error = %v", err)
	}
	if user.ID != "user_local_dev" || user.DisplayName != "Local Developer" {
		t.Fatalf("user = %+v", user)
	}
	again, err := svc.CurrentIdentity(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatalf("CurrentIdentity() second error = %v", err)
	}
	if again.ID != user.ID || !again.CreatedAt.Equal(user.CreatedAt) {
		t.Fatalf("second user = %+v, first = %+v", again, user)
	}
}

func TestThreadLifecycleForCurrentIdentity(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "  First thread  ", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if thread.Title != "First thread" || thread.Mode != ThreadModeChat || thread.LifecycleStatus != ThreadLifecycleActive {
		t.Fatalf("thread = %+v", thread)
	}

	threads, err := svc.ListThreads(context.Background(), ident, false)
	if err != nil {
		t.Fatalf("ListThreads() error = %v", err)
	}
	if len(threads) != 1 || threads[0].ID != thread.ID {
		t.Fatalf("threads = %+v", threads)
	}

	updated, err := svc.UpdateThread(context.Background(), ident, thread.ID, UpdateThreadInput{Title: ptr("Renamed"), Mode: ptr(ThreadModeWork)})
	if err != nil {
		t.Fatalf("UpdateThread() error = %v", err)
	}
	if updated.Title != "Renamed" || updated.Mode != ThreadModeWork {
		t.Fatalf("updated = %+v", updated)
	}

	archived, err := svc.ArchiveThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("ArchiveThread() error = %v", err)
	}
	if archived.LifecycleStatus != ThreadLifecycleArchived || archived.ArchivedAt == nil {
		t.Fatalf("archived = %+v", archived)
	}
	active, err := svc.ListThreads(context.Background(), ident, false)
	if err != nil {
		t.Fatalf("ListThreads(active) error = %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("active = %+v", active)
	}
	got, err := svc.GetThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("GetThread(archived) error = %v", err)
	}
	if got.ID != thread.ID {
		t.Fatalf("got = %+v", got)
	}
}

func TestThreadValidation(t *testing.T) {
	svc := NewMemoryService()
	_, err := svc.CreateThread(context.Background(), identity.LocalDevIdentity(), CreateThreadInput{Title: " ", Mode: ThreadModeChat})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("empty title err = %v", err)
	}
	_, err = svc.CreateThread(context.Background(), identity.LocalDevIdentity(), CreateThreadInput{Title: "Thread", Mode: ThreadMode("run")})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid mode err = %v", err)
	}
}

func TestMessageCreationIsIdempotent(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Messages", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	msg, created, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: " hello ", ClientMessageID: "client-1"})
	if err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}
	if !created {
		t.Fatal("CreateMessage() created = false, want true")
	}
	if msg.Role != MessageRoleUser || msg.Content != "hello" {
		t.Fatalf("msg = %+v", msg)
	}
	threadAfterFirst, err := svc.GetThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("GetThread() error = %v", err)
	}
	dup, created, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: " hello again ", ClientMessageID: "client-1"})
	if err != nil {
		t.Fatalf("CreateMessage(duplicate) error = %v", err)
	}
	if created {
		t.Fatal("CreateMessage(duplicate) created = true, want false")
	}
	if dup.ID != msg.ID || dup.Content != msg.Content {
		t.Fatalf("dup = %+v, msg = %+v", dup, msg)
	}
	threadAfterDuplicate, err := svc.GetThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("GetThread() duplicate error = %v", err)
	}
	if !threadAfterDuplicate.UpdatedAt.Equal(threadAfterFirst.UpdatedAt) {
		t.Fatalf("duplicate changed updated_at: first=%s duplicate=%s", threadAfterFirst.UpdatedAt, threadAfterDuplicate.UpdatedAt)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].ID != msg.ID {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestAppendAssistantMessagePersistsAssistantRole(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Assistant", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	message, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, AppendAssistantMessageInput{Content: "  hello from model  ", Metadata: map[string]any{"api_key": "secret", "run_id": "run_1"}})
	if err != nil {
		t.Fatalf("AppendAssistantMessage() error = %v", err)
	}
	if message.Role != MessageRoleAssistant || message.Content != "hello from model" {
		t.Fatalf("message = %+v", message)
	}
	if message.Metadata["api_key"] != "[redacted]" {
		t.Fatalf("metadata = %+v", message.Metadata)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].Role != MessageRoleAssistant {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestAppendAssistantMessageRejectsDuplicateRunMessage(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Assistant", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	input := AppendAssistantMessageInput{Content: "hello from model", Metadata: map[string]any{"run_id": "run_1"}}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, input); err != nil {
		t.Fatalf("AppendAssistantMessage() error = %v", err)
	}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, input); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("duplicate err = %v", err)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestMessageValidationAndThreadNotFound(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	_, _, err := svc.CreateMessage(context.Background(), ident, "thr_missing", CreateMessageInput{Content: "hello"})
	if err == nil || ErrorCode(err) != CodeThreadNotFound {
		t.Fatalf("missing thread err = %v", err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Messages", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	_, _, err = svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "   "})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("empty message err = %v", err)
	}
}

func TestRedactEventMetadataRedactsSensitiveKeys(t *testing.T) {
	metadata := RedactEventMetadata(map[string]any{"api_key": "sk-live-123", "nested": map[string]any{"password": "abc123"}, "timezone": "UTC"})
	if metadata["api_key"] != "[redacted]" {
		t.Fatalf("api_key was not redacted: %+v", metadata)
	}
	nested := metadata["nested"].(map[string]any)
	if nested["password"] != "[redacted]" {
		t.Fatalf("nested password was not redacted: %+v", metadata)
	}
	if metadata["timezone"] != "UTC" {
		t.Fatalf("safe metadata was changed: %+v", metadata)
	}
}

func TestRunValidation(t *testing.T) {
	if err := ValidateRunStatus(RunStatusRunning); err != nil {
		t.Fatalf("ValidateRunStatus(running) error = %v", err)
	}
	if err := ValidateRunStatus(RunStatusBlockedOnToolApproval); err != nil {
		t.Fatalf("ValidateRunStatus(blocked_on_tool_approval) error = %v", err)
	}
	if err := ValidateRunEventCategory(RunEventCategoryFinal); err != nil {
		t.Fatalf("ValidateRunEventCategory(final) error = %v", err)
	}
	if err := ValidateRunStatus(RunStatusQueued); err != nil {
		t.Fatalf("ValidateRunStatus(queued) error = %v", err)
	}
	if err := ValidateToolCallApprovalStatus(ToolCallApprovalRequired); err != nil {
		t.Fatalf("ValidateToolCallApprovalStatus(required) error = %v", err)
	}
	if err := ValidateToolCallExecutionStatus(ToolCallExecutionBlocked); err != nil {
		t.Fatalf("ValidateToolCallExecutionStatus(blocked) error = %v", err)
	}
	if err := ValidateRunStatus(RunStatus("unknown")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid status err = %v", err)
	}
	if err := ValidateRunEventCategory(RunEventCategory("tool")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid category err = %v", err)
	}
	if err := ValidateToolCallApprovalStatus(ToolCallApprovalStatus("unknown")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid approval status err = %v", err)
	}
	if err := ValidateToolCallExecutionStatus(ToolCallExecutionStatus("unknown")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid execution status err = %v", err)
	}
	if !IsRunActive(RunStatusRunning) || !IsRunActive(RunStatusBlockedOnToolApproval) || IsRunActive(RunStatusCompleted) {
		t.Fatalf("active status helpers returned wrong result")
	}
	if !IsRunTerminal(RunStatusStopped) || IsRunTerminal(RunStatusPending) {
		t.Fatalf("terminal status helpers returned wrong result")
	}
}

func TestRecordToolCallRequestValidatesM7SafetyBoundary(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 tool safety", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	call, events, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest() error = %v", err)
	}
	if call.ArgumentsSummary["timezone"] != "UTC" || events[0].Metadata["arguments_summary"].(map[string]any)["timezone"] != "UTC" {
		t.Fatalf("tool metadata: call=%+v events=%+v", call, events)
	}
	again, againEvents, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest(duplicate) error = %v", err)
	}
	if again.ID != call.ID || len(againEvents) != 0 {
		t.Fatalf("duplicate call = %+v events = %+v", again, againEvents)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_2", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "Asia/Shanghai"}, ArgumentsHash: "hash_2", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("non-UTC timezone err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_2", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC", "api_key": "sk-live-123"}, ArgumentsHash: "hash_2", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("unknown argument err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_2", ToolName: "runtime.unknown", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_2", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("unsupported tool err = %v", err)
	}
	diagnostics, err := svc.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatalf("WorkerQueueDiagnostics() error = %v", err)
	}
	if diagnostics.BlockedToolApprovalCount != 1 {
		t.Fatalf("BlockedToolApprovalCount = %d, want 1", diagnostics.BlockedToolApprovalCount)
	}
}

func TestRecordWorkspaceReadToolRequestsRequireApprovalAndValidateArguments(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M8 workspace reads", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	call, events, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_glob", ToolName: ToolNameWorkspaceGlob, ArgumentsSummary: map[string]any{"pattern": "**/*.go", "limit": 10}, ArgumentsHash: "hash_glob", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest(workspace.glob) error = %v", err)
	}
	if call.ToolName != ToolNameWorkspaceGlob || call.ArgumentsSummary["pattern"] != "**/*.go" || len(events) != 2 {
		t.Fatalf("call=%+v events=%+v", call, events)
	}

	secondThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M8 rejected reads", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread(second) error = %v", err)
	}
	secondRun, err := svc.StartRun(context.Background(), ident, secondThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_2", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun(second) error = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_bad", ToolName: ToolNameWorkspaceReadFile, ArgumentsSummary: map[string]any{"path": ".env"}, ArgumentsHash: "hash_bad", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("sensitive path err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_auto", ToolName: ToolNameWorkspaceGrep, ArgumentsSummary: map[string]any{"query": "TODO"}, ArgumentsHash: "hash_auto", ApprovalStatus: ToolCallApprovalNotRequired, ExecutionStatus: ToolCallExecutionNotStarted}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("not-required approval err = %v", err)
	}
}

func TestRecordWorkspaceWriteToolRequestsRequireApprovalAndValidateArguments(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M9 workspace writes", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	call, events, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_write", ToolName: ToolNameWorkspaceWriteFile, ArgumentsSummary: map[string]any{"path": "internal/generated.txt", "content": "hello"}, ArgumentsHash: "hash_write", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest(workspace.write_file) error = %v", err)
	}
	if call.ToolName != ToolNameWorkspaceWriteFile || call.ArgumentsSummary["path"] != "internal/generated.txt" || len(events) != 2 {
		t.Fatalf("call=%+v events=%+v", call, events)
	}

	secondThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M9 rejected writes", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread(second) error = %v", err)
	}
	secondRun, err := svc.StartRun(context.Background(), ident, secondThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_2", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun(second) error = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_bad", ToolName: ToolNameWorkspaceWriteFile, ArgumentsSummary: map[string]any{"path": ".env", "content": "SECRET=sk-live"}, ArgumentsHash: "hash_bad", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("sensitive write path err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_auto", ToolName: ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "file.txt", "old_text": "a", "new_text": "b"}, ArgumentsHash: "hash_auto", ApprovalStatus: ToolCallApprovalNotRequired, ExecutionStatus: ToolCallExecutionNotStarted}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("not-required approval err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_edit_bad", ToolName: ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "file.txt", "old_text": "", "new_text": "b"}, ArgumentsHash: "hash_edit_bad", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("empty old_text err = %v", err)
	}
}

func TestRecordWorkspaceExecCommandRequiresApprovalAndValidatesArguments(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M10 workspace exec", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	call, events, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_exec", ToolName: ToolNameWorkspaceExecCommand, ArgumentsSummary: map[string]any{"command": []any{"printf", "hello"}, "cwd": ".", "timeout_seconds": 5}, ArgumentsHash: "hash_exec", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest(workspace.exec_command) error = %v", err)
	}
	if call.ToolName != ToolNameWorkspaceExecCommand || len(events) != 2 {
		t.Fatalf("call=%+v events=%+v", call, events)
	}

	secondThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M10 rejected exec", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread(second) error = %v", err)
	}
	secondRun, err := svc.StartRun(context.Background(), ident, secondThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_2", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun(second) error = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_shell", ToolName: ToolNameWorkspaceExecCommand, ArgumentsSummary: map[string]any{"command": []any{"sh", "-c", "echo no"}}, ArgumentsHash: "hash_shell", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("shell wrapper err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_escape", ToolName: ToolNameWorkspaceExecCommand, ArgumentsSummary: map[string]any{"command": []any{"printf", "x"}, "cwd": "../"}, ArgumentsHash: "hash_escape", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("cwd escape err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_auto", ToolName: ToolNameWorkspaceExecCommand, ArgumentsSummary: map[string]any{"command": []any{"printf", "x"}}, ArgumentsHash: "hash_auto", ApprovalStatus: ToolCallApprovalNotRequired, ExecutionStatus: ToolCallExecutionNotStarted}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("not-required approval err = %v", err)
	}
}

func TestRecordTodoWriteToolRequestsRequireApprovalAndValidateArguments(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M12 todo write", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	call, events, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_todo", ToolName: ToolNameTodoWrite, ArgumentsSummary: map[string]any{"items": []any{map[string]any{"title": "Inspect tool registry", "status": "completed"}, map[string]any{"title": "Add todo tool", "status": "in_progress"}, map[string]any{"title": "Validate docs"}}}, ArgumentsHash: "hash_todo", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest(runtime.todo_write) error = %v", err)
	}
	items, ok := call.ArgumentsSummary["items"].([]any)
	if call.ToolName != ToolNameTodoWrite || !ok || len(items) != 3 || len(events) != 2 {
		t.Fatalf("call=%+v events=%+v", call, events)
	}

	secondThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M12 rejected todo", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread(second) error = %v", err)
	}
	secondRun, err := svc.StartRun(context.Background(), ident, secondThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_2", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun(second) error = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_empty", ToolName: ToolNameTodoWrite, ArgumentsSummary: map[string]any{"items": []any{}}, ArgumentsHash: "hash_empty", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("empty items err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_status", ToolName: ToolNameTodoWrite, ArgumentsSummary: map[string]any{"items": []any{map[string]any{"title": "Bad status", "status": "blocked"}}}, ArgumentsHash: "hash_status", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("bad status err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_auto", ToolName: ToolNameTodoWrite, ArgumentsSummary: map[string]any{"items": []any{map[string]any{"title": "Auto"}}}, ArgumentsHash: "hash_auto", ApprovalStatus: ToolCallApprovalNotRequired, ExecutionStatus: ToolCallExecutionNotStarted}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("not-required approval err = %v", err)
	}
}

func TestRecordMCPCallToolRequestsRequireApprovalAndValidateArguments(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M13 MCP call", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	call, events, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_mcp", ToolName: ToolNameMCPCallTool, ArgumentsSummary: map[string]any{"server": "local", "tool": "echo", "arguments": map[string]any{"message": "hello mcp"}}, ArgumentsHash: "hash_mcp", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest(mcp.call_tool) error = %v", err)
	}
	if call.ToolName != ToolNameMCPCallTool || call.ArgumentsSummary["server"] != "local" || call.ArgumentsSummary["tool"] != "echo" || len(events) != 2 {
		t.Fatalf("call=%+v events=%+v", call, events)
	}

	secondThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M13 rejected MCP", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread(second) error = %v", err)
	}
	secondRun, err := svc.StartRun(context.Background(), ident, secondThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_2", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun(second) error = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_server", ToolName: ToolNameMCPCallTool, ArgumentsSummary: map[string]any{"server": "remote", "tool": "echo", "arguments": map[string]any{"message": "hello"}}, ArgumentsHash: "hash_server", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("unknown server err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_tool", ToolName: ToolNameMCPCallTool, ArgumentsSummary: map[string]any{"server": "local", "tool": "shell", "arguments": map[string]any{"message": "hello"}}, ArgumentsHash: "hash_tool", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("unknown tool err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_secret", ToolName: ToolNameMCPCallTool, ArgumentsSummary: map[string]any{"server": "local", "tool": "echo", "arguments": map[string]any{"message": "secret=sk-live"}}, ArgumentsHash: "hash_secret", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("secret message err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, secondRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_auto", ToolName: ToolNameMCPCallTool, ArgumentsSummary: map[string]any{"server": "local", "tool": "echo", "arguments": map[string]any{"message": "hello"}}, ArgumentsHash: "hash_auto", ApprovalStatus: ToolCallApprovalNotRequired, ExecutionStatus: ToolCallExecutionNotStarted}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("not-required approval err = %v", err)
	}
}

func TestToolCallApprovalDecisionsAreIdempotent(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 approval", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_approve", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_approve", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatalf("RecordToolCallRequest(approve) error = %v", err)
	}

	approved, events, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_approve")
	if err != nil {
		t.Fatalf("ApproveToolCall() error = %v", err)
	}
	if approved.ApprovalStatus != ToolCallApprovalApproved || approved.ExecutionStatus != ToolCallExecutionNotStarted {
		t.Fatalf("approved call = %+v", approved)
	}
	if len(events) != 1 || events[0].Type != EventToolCallApproved {
		t.Fatalf("approve events = %+v", events)
	}
	again, againEvents, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_approve")
	if err != nil {
		t.Fatalf("ApproveToolCall(retry) error = %v", err)
	}
	if again.ID != approved.ID || len(againEvents) != 0 {
		t.Fatalf("approve retry call=%+v events=%+v", again, againEvents)
	}
	if _, _, err := svc.DenyToolCall(context.Background(), ident, thread.ID, run.ID, "tc_approve"); err == nil || ErrorCode(err) != CodeConflict {
		t.Fatalf("deny after approve err = %v", err)
	}

	denyThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 denial", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread(deny) error = %v", err)
	}
	denyRun, err := svc.StartRun(context.Background(), ident, denyThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_2", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun(deny) error = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, denyRun.ID, RecordToolCallRequestInput{ToolCallID: "tc_deny", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_deny", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatalf("RecordToolCallRequest(deny) error = %v", err)
	}
	denied, denyEvents, err := svc.DenyToolCall(context.Background(), ident, denyThread.ID, denyRun.ID, "tc_deny")
	if err != nil {
		t.Fatalf("DenyToolCall() error = %v", err)
	}
	if denied.ApprovalStatus != ToolCallApprovalDenied || denied.ExecutionStatus != ToolCallExecutionCancelled {
		t.Fatalf("denied call = %+v", denied)
	}
	if len(denyEvents) != 2 || denyEvents[0].Type != EventToolCallDenied || denyEvents[1].Type != EventRunStopped {
		t.Fatalf("deny events = %+v", denyEvents)
	}
	deniedAgain, deniedAgainEvents, err := svc.DenyToolCall(context.Background(), ident, denyThread.ID, denyRun.ID, "tc_deny")
	if err != nil {
		t.Fatalf("DenyToolCall(retry) error = %v", err)
	}
	if deniedAgain.ID != denied.ID || len(deniedAgainEvents) != 0 {
		t.Fatalf("deny retry call=%+v events=%+v", deniedAgain, deniedAgainEvents)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, denyThread.ID, denyRun.ID, "tc_deny"); err == nil || ErrorCode(err) != CodeConflict {
		t.Fatalf("approve after deny err = %v", err)
	}
}

func TestStopRunCancelsPendingApprovedAndExecutingToolCalls(t *testing.T) {
	for _, tc := range []struct {
		name    string
		prepare func(*MemoryService, identity.LocalIdentity, Thread, Run)
	}{
		{name: "pending", prepare: func(*MemoryService, identity.LocalIdentity, Thread, Run) {}},
		{name: "approved", prepare: func(svc *MemoryService, ident identity.LocalIdentity, thread Thread, run Run) {
			if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_cancel"); err != nil {
				t.Fatalf("ApproveToolCall() error = %v", err)
			}
		}},
		{name: "executing", prepare: func(svc *MemoryService, ident identity.LocalIdentity, thread Thread, run Run) {
			if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_cancel"); err != nil {
				t.Fatalf("ApproveToolCall() error = %v", err)
			}
			if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_cancel"); err != nil {
				t.Fatalf("StartToolCallExecution() error = %v", err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMemoryService()
			ident := identity.LocalDevIdentity()
			thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 cancel", Mode: ThreadModeChat})
			if err != nil {
				t.Fatalf("CreateThread() error = %v", err)
			}
			run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_cancel", ProviderID: "custom", Model: "model"})
			if err != nil {
				t.Fatalf("StartRun() error = %v", err)
			}
			if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_cancel", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_cancel", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
				t.Fatalf("RecordToolCallRequest() error = %v", err)
			}
			tc.prepare(svc, ident, thread, run)

			stopped, err := svc.StopRun(context.Background(), ident, run.ID)
			if err != nil {
				t.Fatalf("StopRun() error = %v", err)
			}
			if stopped.Run.Status != RunStatusStopped {
				t.Fatalf("stopped = %+v", stopped)
			}
			call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_cancel")
			if err != nil {
				t.Fatalf("GetToolCall() error = %v", err)
			}
			if call.ApprovalStatus != ToolCallApprovalCancelled || call.ExecutionStatus != ToolCallExecutionCancelled {
				t.Fatalf("call = %+v", call)
			}
			if _, _, err := svc.CompleteToolCallSuccess(context.Background(), ident, thread.ID, run.ID, "tc_cancel", map[string]any{"timezone": "UTC"}); err != nil {
				t.Fatalf("CompleteToolCallSuccess(after cancel) error = %v", err)
			}
			if _, _, err := svc.CompleteToolCallFailure(context.Background(), ident, thread.ID, run.ID, "tc_cancel", "tool_execution_failed", "failed"); err != nil {
				t.Fatalf("CompleteToolCallFailure(after cancel) error = %v", err)
			}
			events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
			if err != nil {
				t.Fatalf("ListRunEvents() error = %v", err)
			}
			var cancelled, succeeded, failed int
			for _, event := range events {
				switch event.Type {
				case EventToolCallCancelled:
					cancelled++
				case EventToolCallSucceeded:
					succeeded++
				case EventToolCallFailed:
					failed++
				}
			}
			if cancelled != 1 || succeeded != 0 || failed != 0 {
				t.Fatalf("cancelled=%d succeeded=%d failed=%d events=%+v", cancelled, succeeded, failed, events)
			}
		})
	}
}

func TestStartRunCreatesInitialLifecycleEvent(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{ScriptName: "m4_smoke token"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if run.ThreadID != thread.ID || run.Status != RunStatusQueued || run.Source != RunSourceLocalSimulated {
		t.Fatalf("run = %+v", run)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if len(events) != 2 || events[0].Sequence != 1 || events[0].Type != "run_created" || events[1].Type != EventRunQueued {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Metadata["script_name"] != "[redacted]" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
}

func TestStartRunSupportsModelGatewaySource(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "gpt-5.5"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if run.Source != RunSourceModelGateway || run.Title != "Model gateway run" {
		t.Fatalf("run = %+v", run)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if events[0].Metadata["provider_id"] != "custom" || events[0].Metadata["model"] != "gpt-5.5" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
}

func TestStartRunRejectsSecondActiveRunForThread(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{}); err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	_, err = svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err == nil || ErrorCode(err) != CodeActiveRunExists {
		t.Fatalf("second active run err = %v", err)
	}
}

func TestStartRunAndJobCreationAreAtomicFromServiceBoundary(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Jobs", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	diagnostics, err := svc.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics.QueuedCount != 1 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || job.RunID != run.ID || claimedRun.Status != RunStatusRunning {
		t.Fatalf("job=%+v run=%+v ok=%v", job, claimedRun, ok)
	}
}

func TestFailBackgroundJobRedactsFailureAndTerminalEvents(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Fail", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	failed, changed, err := svc.FailBackgroundJob(context.Background(), ident, FailBackgroundJobInput{JobID: job.ID, WorkerID: "worker_test", OwnershipVersion: job.OwnershipVersion, ErrorCode: "provider_failed", ErrorMessage: "token secret leaked"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed || failed.Status != BackgroundJobStatusFailed || failed.LastError == nil || *failed.LastError != "[redacted]" {
		t.Fatalf("failed = %+v", failed)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != RunStatusFailed || got.ErrorMessage == nil || *got.ErrorMessage != "[redacted]" {
		t.Fatalf("run = %+v", got)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if events[len(events)-2].Type != EventJobAttemptFailed || events[len(events)-1].Type != EventRunFailed || events[len(events)-1].Summary != "[redacted]" {
		t.Fatalf("events = %+v", events)
	}
}

func TestRecoverBackgroundJobsReschedulesExpiredLeaseAndRejectsStaleOwner(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	base := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return base }
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Recover", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_stale", LeaseSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || job.OwnershipVersion != 1 {
		t.Fatalf("job = %+v ok=%v", job, ok)
	}
	base = base.Add(2 * time.Second)
	recoveries, err := svc.RecoverBackgroundJobs(context.Background(), ident, RecoverBackgroundJobsInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveries) != 1 || recoveries[0].Exhausted || recoveries[0].Job.Status != BackgroundJobStatusQueued || recoveries[0].Run.Status != RunStatusRecovering {
		t.Fatalf("recoveries = %+v", recoveries)
	}
	if recoveries[0].Events[0].Type != EventJobRecovering || recoveries[0].Events[1].Type != EventJobRetryScheduled {
		t.Fatalf("events = %+v", recoveries[0].Events)
	}
	if _, changed, err := svc.CompleteBackgroundJob(context.Background(), ident, CompleteBackgroundJobInput{JobID: job.ID, WorkerID: "worker_stale", OwnershipVersion: job.OwnershipVersion}); err != nil || changed {
		t.Fatalf("stale completion changed=%v err=%v", changed, err)
	}
	claimed, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_fresh", LeaseSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || claimed.OwnershipVersion <= job.OwnershipVersion || claimed.AttemptCount != 2 {
		t.Fatalf("fresh claim = %+v ok=%v", claimed, ok)
	}
}

func TestRecoverBackgroundJobsExhaustsRetriesWithRedactedFailure(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	base := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return base }
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Recover", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 1; attempt <= 3; attempt++ {
		if _, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_retry", LeaseSeconds: 1}); err != nil || !ok {
			t.Fatalf("claim attempt %d ok=%v err=%v", attempt, ok, err)
		}
		base = base.Add(2 * time.Second)
		recoveries, err := svc.RecoverBackgroundJobs(context.Background(), ident, RecoverBackgroundJobsInput{ErrorMessage: "token secret leaked"})
		if err != nil {
			t.Fatal(err)
		}
		if len(recoveries) != 1 {
			t.Fatalf("attempt %d recoveries = %+v", attempt, recoveries)
		}
		if attempt < 3 && recoveries[0].Exhausted {
			t.Fatalf("attempt %d exhausted early: %+v", attempt, recoveries[0])
		}
		if attempt == 3 {
			if !recoveries[0].Exhausted || recoveries[0].Job.Status != BackgroundJobStatusDead || recoveries[0].Run.Status != RunStatusFailed {
				t.Fatalf("final recovery = %+v", recoveries[0])
			}
			if recoveries[0].Run.ErrorMessage == nil || *recoveries[0].Run.ErrorMessage != "[redacted]" || recoveries[0].Events[0].Summary != "[redacted]" {
				t.Fatalf("final recovery did not redact = %+v", recoveries[0])
			}
		}
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != RunStatusFailed || got.CompletedAt == nil {
		t.Fatalf("run = %+v", got)
	}
}

func TestStopRunCancelsQueuedJobAndPreventsClaim(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	stopped, err := svc.StopRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("StopRun() error = %v", err)
	}
	if stopped.Run.StopRequestedAt == nil || stopped.Run.Status != RunStatusStopped {
		t.Fatalf("stopped = %+v", stopped)
	}
	if _, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 1}); err != nil || ok {
		t.Fatalf("claim after stop ok=%v err=%v", ok, err)
	}
	diagnostics, err := svc.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics.QueuedCount != 0 || diagnostics.LeasedCount != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
}

func TestStopRunRecordsStoppedTerminalEvents(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	stopped, err := svc.StopRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("StopRun() error = %v", err)
	}
	if stopped.Result != StopRunResultStopped || stopped.Run.Status != RunStatusStopped || stopped.Run.CompletedAt == nil {
		t.Fatalf("stopped = %+v", stopped)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if len(events) != 4 || events[2].Type != EventStopRequested || events[3].Category != RunEventCategoryFinal {
		t.Fatalf("events = %+v", events)
	}
}

func TestStopRunReturnsAlreadyTerminalWithoutChangingOutcome(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"}); err != nil {
		t.Fatalf("AppendRunEvent(final) error = %v", err)
	}
	output, err := svc.StopRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("StopRun() error = %v", err)
	}
	if output.Result != StopRunResultAlreadyTerminal || output.Run.Status != RunStatusCompleted {
		t.Fatalf("output = %+v", output)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("events = %+v", events)
	}
}

func TestAppendRunEventRejectsTerminalRun(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"}); err != nil {
		t.Fatalf("AppendRunEvent(final) error = %v", err)
	}
	_, err = svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: "late", Summary: "Late"})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("late append err = %v", err)
	}
}

func TestRunEventRedactsSecretText(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	content := "postgres://loomi:secret@localhost/db"
	event, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryError, Type: "run_failed", Summary: "token leaked", Content: &content, Metadata: map[string]any{"database_url": "postgresql://user:password=secret@localhost/db", "nested": map[string]any{"bearer": "Bearer abc"}}})
	if err != nil {
		t.Fatalf("AppendRunEvent(error) error = %v", err)
	}
	if event.Summary != "[redacted]" || event.Content == nil || *event.Content != "[redacted]" {
		t.Fatalf("event = %+v", event)
	}
	if event.Metadata["database_url"] != "[redacted]" {
		t.Fatalf("metadata = %+v", event.Metadata)
	}
}

func TestAppendRunEventOrdersPersistedEvents(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: "context_loaded", Summary: "Context loaded"}); err != nil {
		t.Fatalf("AppendRunEvent(progress) error = %v", err)
	}
	final, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"})
	if err != nil {
		t.Fatalf("AppendRunEvent(final) error = %v", err)
	}
	if final.Sequence != 4 {
		t.Fatalf("final sequence = %d", final.Sequence)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 1)
	if err != nil {
		t.Fatalf("ListRunEvents(after=1) error = %v", err)
	}
	if len(events) != 3 || events[0].Sequence != 2 || events[1].Sequence != 3 || events[2].Sequence != 4 {
		t.Fatalf("events = %+v", events)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Status != RunStatusCompleted || got.CompletedAt == nil {
		t.Fatalf("run after final = %+v", got)
	}
}

func ptr[T any](v T) *T { return &v }
