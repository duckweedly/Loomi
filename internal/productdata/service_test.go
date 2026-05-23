package productdata

import (
	"context"
	"testing"

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

func TestRunValidation(t *testing.T) {
	if err := ValidateRunStatus(RunStatusRunning); err != nil {
		t.Fatalf("ValidateRunStatus(running) error = %v", err)
	}
	if err := ValidateRunEventCategory(RunEventCategoryFinal); err != nil {
		t.Fatalf("ValidateRunEventCategory(final) error = %v", err)
	}
	if err := ValidateRunStatus(RunStatus("queued")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid status err = %v", err)
	}
	if err := ValidateRunEventCategory(RunEventCategory("tool")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid category err = %v", err)
	}
	if !IsRunActive(RunStatusRunning) || IsRunActive(RunStatusCompleted) {
		t.Fatalf("active status helpers returned wrong result")
	}
	if !IsRunTerminal(RunStatusStopped) || IsRunTerminal(RunStatusPending) {
		t.Fatalf("terminal status helpers returned wrong result")
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
	if run.ThreadID != thread.ID || run.Status != RunStatusRunning || run.Source != RunSourceLocalSimulated {
		t.Fatalf("run = %+v", run)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if len(events) != 1 || events[0].Sequence != 1 || events[0].Type != "run_created" {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Metadata["script_name"] != "[redacted]" {
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
	if len(events) != 3 || events[1].Type != "run_stopped" || events[2].Category != RunEventCategoryFinal {
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
	if len(events) != 2 {
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
	if final.Sequence != 3 {
		t.Fatalf("final sequence = %d", final.Sequence)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 1)
	if err != nil {
		t.Fatalf("ListRunEvents(after=1) error = %v", err)
	}
	if len(events) != 2 || events[0].Sequence != 2 || events[1].Sequence != 3 {
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
