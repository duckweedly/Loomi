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

func ptr[T any](v T) *T { return &v }
