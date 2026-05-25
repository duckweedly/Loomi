package productdata

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
)

func TestMemorySearchExcludesPendingDeniedDeletedUnsafeAndOutOfScopeEntries(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	other := identity.LocalIdentity{UserID: "user_other", DisplayName: "Other", Source: "test"}

	approved, err := svc.CreateMemoryEntry(context.Background(), ident, CreateMemoryEntryInput{ScopeType: MemoryScopeUser, Title: "Project taste", Content: "Prefers compact implementation slices", SourceThreadID: "thr_1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateMemoryEntry(context.Background(), ident, CreateMemoryEntryInput{ScopeType: MemoryScopeUser, Title: "Unsafe", Content: "token sk-secret", SourceThreadID: "thr_1"}); err != nil {
		t.Fatal(err)
	}
	deleted, err := svc.CreateMemoryEntry(context.Background(), ident, CreateMemoryEntryInput{ScopeType: MemoryScopeUser, Title: "Delete me", Content: "old context"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeleteMemoryEntry(context.Background(), ident, deleted.ID, DeleteMemoryEntryInput{Reason: "user_request"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateMemoryEntry(context.Background(), other, CreateMemoryEntryInput{ScopeType: MemoryScopeUser, Title: "Other", Content: "other user memory"}); err != nil {
		t.Fatal(err)
	}
	proposal, err := svc.ProposeMemoryWrite(context.Background(), ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeUser, Title: "Pending", Content: "pending memory", IdempotencyKey: "pending-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DenyMemoryWrite(context.Background(), ident, proposal.ID, MemoryWriteDecisionInput{IdempotencyKey: "deny-1"}); err != nil {
		t.Fatal(err)
	}

	results, err := svc.SearchMemory(context.Background(), ident, MemorySearchInput{Query: "memory implementation", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(results.Items) != 1 || results.Items[0].ID != approved.ID {
		t.Fatalf("results = %+v, want only %s", results.Items, approved.ID)
	}
	if results.Items[0].Summary == "" || results.Items[0].Content != "" {
		t.Fatalf("result should expose safe summary only: %+v", results.Items[0])
	}
}

func TestMemoryEntryReadDeleteRequiresMatchingThreadScope(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	threadA := "thr_scope_a"
	threadB := "thr_scope_b"
	entry, err := svc.CreateMemoryEntry(context.Background(), ident, CreateMemoryEntryInput{ScopeType: MemoryScopeThread, ScopeID: threadA, Title: "Thread scoped", Content: "Visible only in thread A", SourceThreadID: threadA})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.GetMemoryEntry(context.Background(), ident, entry.ID, MemoryEntryAccessInput{ScopeType: MemoryScopeThread, ScopeID: threadB}); err == nil || ErrorCode(err) != CodeMemoryNotFound {
		t.Fatalf("thread B read err = %v", err)
	}
	if _, err := svc.DeleteMemoryEntry(context.Background(), ident, entry.ID, DeleteMemoryEntryInput{Reason: "wrong thread", ScopeType: MemoryScopeThread, ScopeID: threadB}); err == nil || ErrorCode(err) != CodeMemoryNotFound {
		t.Fatalf("thread B delete err = %v", err)
	}
	if _, err := svc.GetMemoryEntry(context.Background(), ident, entry.ID, MemoryEntryAccessInput{}); err == nil || ErrorCode(err) != CodeMemoryNotFound {
		t.Fatalf("unscoped read err = %v", err)
	}

	got, err := svc.GetMemoryEntry(context.Background(), ident, entry.ID, MemoryEntryAccessInput{ScopeType: MemoryScopeThread, ScopeID: threadA})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != entry.ID || got.Content != "" {
		t.Fatalf("got = %+v", got)
	}
	if _, err := svc.DeleteMemoryEntry(context.Background(), ident, entry.ID, DeleteMemoryEntryInput{Reason: "right thread", ScopeType: MemoryScopeThread, ScopeID: threadA}); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareRunContextIncludesSafeMemorySnapshot(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Memory", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "Use my previous preferences"})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := svc.CreateMemoryEntry(context.Background(), ident, CreateMemoryEntryInput{ScopeType: MemoryScopeUser, Title: "Preference", Content: "Prefers PostgreSQL first memory", SourceThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateMemoryEntry(context.Background(), ident, CreateMemoryEntryInput{ScopeType: MemoryScopeThread, ScopeID: "thr_other", Title: "Other thread", Content: "not visible here"}); err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_memory", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}

	ctxData, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if ctxData.Run.ID != run.ID || len(ctxData.MemorySnapshot.Entries) != 1 || ctxData.MemorySnapshot.Entries[0].ID != entry.ID {
		t.Fatalf("memory snapshot = %+v", ctxData.MemorySnapshot)
	}
	summary := ctxData.SafeSummary()
	if summary["memory_entry_count"] != 1 || summary["memory_status"] != "loaded" {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestMemoryAuditPersistsAfterTerminalRun(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Terminal audit", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{ScriptName: "terminal_memory_audit"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: EventRunCompleted, Summary: "Run completed"}); err != nil {
		t.Fatal(err)
	}

	proposal, err := svc.ProposeMemoryWrite(context.Background(), ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeThread, ScopeID: thread.ID, Title: "Terminal proposal", Content: "Keep safe terminal audit", SourceThreadID: thread.ID, SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := svc.ApproveMemoryWrite(context.Background(), ident, proposal.ID, MemoryWriteDecisionInput{Reason: "approve after terminal"})
	if err != nil {
		t.Fatal(err)
	}
	denied, err := svc.ProposeMemoryWrite(context.Background(), ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeThread, ScopeID: thread.ID, Title: "Terminal deny", Content: "Deny after terminal", SourceThreadID: thread.ID, SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DenyMemoryWrite(context.Background(), ident, denied.ID, MemoryWriteDecisionInput{Reason: "deny after terminal"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DenyMemoryWrite(context.Background(), ident, denied.ID, MemoryWriteDecisionInput{Reason: "deny retry"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeleteMemoryEntry(context.Background(), ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "delete after terminal", ScopeType: MemoryScopeThread, ScopeID: thread.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeleteMemoryEntry(context.Background(), ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "delete retry", ScopeType: MemoryScopeThread, ScopeID: thread.ID}); err != nil {
		t.Fatal(err)
	}

	audit, err := svc.ListMemoryAudit(context.Background(), ident, MemoryAuditInput{SourceRunID: run.ID, Limit: 10})
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

func TestMemoryRedactionCoversPathsAndRuntimePayloadMarkers(t *testing.T) {
	cases := []string{
		"/home/xuean/private/project/file.txt",
		`C:\Users\xuean\secret\file.txt`,
		"stdout: raw terminal payload",
		"stderr: raw terminal payload",
		"tool output: raw tool payload",
		"provider trace: raw provider payload",
		"LOOMI_TOKEN=secret",
		"Authorization: Bearer sk-secret",
	}
	for _, value := range cases {
		if got := RedactEventText(value); got != "[redacted]" {
			t.Fatalf("RedactEventText(%q) = %q", value, got)
		}
	}

	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Redaction", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "Check safe summary"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "/home/xuean/private/provider", Model: `C:\Users\xuean\model`})
	if err != nil {
		t.Fatal(err)
	}
	_ = run
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_redaction", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	ctxData, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	summary := ctxData.SafeSummary()
	encoded := fmt.Sprint(summary)
	for _, leaked := range []string{"/home/xuean", `C:\Users\xuean`, "provider trace", "tool output", "Authorization"} {
		if strings.Contains(encoded, leaked) {
			t.Fatalf("safe summary leaked %q: %+v", leaked, summary)
		}
	}
}

func TestMemoryWriteApprovalDeleteAndIdempotency(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()

	proposal, err := svc.ProposeMemoryWrite(context.Background(), ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeUser, Title: "Remember", Content: "Keep approval gated memory", IdempotencyKey: "proposal-1", SourceRunID: "run_1"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := svc.ProposeMemoryWrite(context.Background(), ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeUser, Title: "Remember duplicate", Content: "different", IdempotencyKey: "proposal-1"})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != proposal.ID {
		t.Fatalf("duplicate proposal = %+v, first = %+v", again, proposal)
	}
	beforeApproval, err := svc.SearchMemory(context.Background(), ident, MemorySearchInput{Query: "approval", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(beforeApproval.Items) != 0 {
		t.Fatalf("pending proposal appeared in search: %+v", beforeApproval.Items)
	}

	decision, err := svc.ApproveMemoryWrite(context.Background(), ident, proposal.ID, MemoryWriteDecisionInput{IdempotencyKey: "approve-1"})
	if err != nil {
		t.Fatal(err)
	}
	decisionAgain, err := svc.ApproveMemoryWrite(context.Background(), ident, proposal.ID, MemoryWriteDecisionInput{IdempotencyKey: "approve-1"})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Entry.ID == "" || decisionAgain.Entry.ID != decision.Entry.ID {
		t.Fatalf("approval decision = %+v again=%+v", decision, decisionAgain)
	}
	found, err := svc.SearchMemory(context.Background(), ident, MemorySearchInput{Query: "approval", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(found.Items) != 1 || found.Items[0].ID != decision.Entry.ID {
		t.Fatalf("approved search = %+v", found.Items)
	}

	tombstone, err := svc.DeleteMemoryEntry(context.Background(), ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "no_longer_needed"})
	if err != nil {
		t.Fatal(err)
	}
	tombstoneAgain, err := svc.DeleteMemoryEntry(context.Background(), ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "retry"})
	if err != nil {
		t.Fatal(err)
	}
	if tombstone.EntryID != decision.Entry.ID || tombstoneAgain.EntryID != tombstone.EntryID {
		t.Fatalf("tombstone=%+v again=%+v", tombstone, tombstoneAgain)
	}
	afterDelete, err := svc.SearchMemory(context.Background(), ident, MemorySearchInput{Query: "approval", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(afterDelete.Items) != 0 {
		t.Fatalf("deleted memory appeared in search: %+v", afterDelete.Items)
	}
}

func TestMemoryWriteAppendsSourceRunAuditEvents(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()

	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Memory audit", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "Remember this later"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := svc.ProposeMemoryWrite(context.Background(), ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeUser, Title: "Audit", Content: "Keep a concise audit trail", SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := svc.ApproveMemoryWrite(context.Background(), ident, proposal.ID, MemoryWriteDecisionInput{Reason: "approved"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeleteMemoryEntry(context.Background(), ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "remove"}); err != nil {
		t.Fatal(err)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	for _, event := range events {
		if event.Type == EventMemoryWriteProposed || event.Type == EventMemoryWriteApproved || event.Type == EventMemoryEntryDeleted {
			found = append(found, event.Type)
		}
	}
	if len(found) != 3 {
		t.Fatalf("memory audit events = %v", found)
	}
}

func TestMemoryDecisionAndDeleteDoNotDuplicateAuditEvents(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Memory idempotency", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "Remember only once"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}

	denied, err := svc.ProposeMemoryWrite(context.Background(), ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeUser, Title: "Deny once", Content: "Do not duplicate deny audit", SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DenyMemoryWrite(context.Background(), ident, denied.ID, MemoryWriteDecisionInput{Reason: "first deny"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DenyMemoryWrite(context.Background(), ident, denied.ID, MemoryWriteDecisionInput{Reason: "second deny"}); err != nil {
		t.Fatal(err)
	}

	approved, err := svc.ProposeMemoryWrite(context.Background(), ident, ProposeMemoryWriteInput{ScopeType: MemoryScopeUser, Title: "Delete once", Content: "Do not duplicate delete audit", SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := svc.ApproveMemoryWrite(context.Background(), ident, approved.ID, MemoryWriteDecisionInput{Reason: "approve"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeleteMemoryEntry(context.Background(), ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "first delete"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeleteMemoryEntry(context.Background(), ident, decision.Entry.ID, DeleteMemoryEntryInput{Reason: "second delete"}); err != nil {
		t.Fatal(err)
	}

	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, event := range events {
		counts[event.Type]++
	}
	if counts[EventMemoryWriteDenied] != 1 || counts[EventMemoryEntryDeleted] != 1 {
		t.Fatalf("audit counts = %+v", counts)
	}
}
