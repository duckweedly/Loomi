package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestWorkerProposesPostRunMemoryWhenCommitAfterRunEnabled(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Post-run memory", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveMemoryProviderConfig(context.Background(), ident, productdata.MemoryProviderConfig{Enabled: true, Provider: productdata.MemoryProviderLocal, CommitAfterRun: true}); err != nil {
		t.Fatal(err)
	}
	runner := NewLocalRunner(svc, nil)
	runner.StepDelay = 0
	worker := NewWorker(svc, nil, runner)
	worker.WorkerID = "worker_memory"

	ok, err := worker.ProcessOne(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ProcessOne() ok = false")
	}
	audit, err := svc.ListMemoryAudit(context.Background(), ident, productdata.MemoryAuditInput{SourceRunID: run.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.Items) != 1 || audit.Items[0].EventType != productdata.EventMemoryWriteProposed {
		t.Fatalf("audit = %+v", audit.Items)
	}
	proposalID := audit.Items[0].MemoryProposalID
	if proposalID == "" {
		t.Fatalf("proposal id missing in audit: %+v", audit.Items[0])
	}
	proposal, err := svc.ProposeMemoryWrite(context.Background(), ident, productdata.ProposeMemoryWriteInput{ScopeType: productdata.MemoryScopeThread, ScopeID: thread.ID, Title: "duplicate", Content: "duplicate", SourceThreadID: thread.ID, SourceRunID: run.ID, IdempotencyKey: postRunMemoryIdempotencyKey(run.ID)})
	if err != nil {
		t.Fatal(err)
	}
	if proposal.ID != proposalID || proposal.Status != productdata.MemoryWritePending || proposal.ScopeType != productdata.MemoryScopeThread || proposal.ScopeID != thread.ID {
		t.Fatalf("proposal = %+v audit=%+v", proposal, audit.Items[0])
	}
	if !strings.Contains(proposal.Content, "Local simulated response is ready.") {
		t.Fatalf("proposal content did not include assistant outcome: %+v", proposal)
	}
	entries, err := svc.SearchMemory(context.Background(), ident, productdata.MemorySearchInput{ScopeType: productdata.MemoryScopeThread, ScopeID: thread.ID, Query: "Loomi", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries.Items) != 0 {
		t.Fatalf("pending proposal appeared as approved memory: %+v", entries.Items)
	}
}

func TestPostRunMemorySkipsWhenCommitAfterRunDisabled(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Post-run disabled", Mode: productdata.ThreadModeChat})
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
	worker.WorkerID = "worker_memory_disabled"

	if _, err := worker.ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	audit, err := svc.ListMemoryAudit(context.Background(), ident, productdata.MemoryAuditInput{SourceRunID: run.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.Items) != 0 {
		t.Fatalf("audit = %+v", audit.Items)
	}
}

func TestPostRunMemoryIsIdempotent(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Post-run idempotent", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveMemoryProviderConfig(context.Background(), ident, productdata.MemoryProviderConfig{Enabled: true, Provider: productdata.MemoryProviderLocal, CommitAfterRun: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, productdata.AppendAssistantMessageInput{Content: "Remember one concise outcome.", Metadata: map[string]any{"run_id": run.ID}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: productdata.EventRunCompleted, Summary: "Run completed"}); err != nil {
		t.Fatal(err)
	}

	if err := proposePostRunMemory(context.Background(), svc, ident, run.ID); err != nil {
		t.Fatal(err)
	}
	if err := proposePostRunMemory(context.Background(), svc, ident, run.ID); err != nil {
		t.Fatal(err)
	}
	audit, err := svc.ListMemoryAudit(context.Background(), ident, productdata.MemoryAuditInput{SourceRunID: run.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.Items) != 1 {
		t.Fatalf("audit = %+v", audit.Items)
	}
}

func TestPostRunMemoryCommitsToExternalProviderWhenConfigured(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	writeCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/memories" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		writeCount++
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "mem_post_run"})
	}))
	defer server.Close()

	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "External post-run", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveMemoryProviderConfig(context.Background(), ident, productdata.MemoryProviderConfig{Enabled: true, Provider: productdata.MemoryProviderNowledge, CommitAfterRun: true, Nowledge: productdata.NowledgeMemoryConfig{BaseURL: server.URL, APIKey: "secret"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, productdata.AppendAssistantMessageInput{Content: "External provider should commit this outcome.", Metadata: map[string]any{"run_id": run.ID}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: productdata.EventRunCompleted, Summary: "Run completed"}); err != nil {
		t.Fatal(err)
	}

	if err := proposePostRunMemory(context.Background(), svc, ident, run.ID); err != nil {
		t.Fatal(err)
	}
	if err := proposePostRunMemory(context.Background(), svc, ident, run.ID); err != nil {
		t.Fatal(err)
	}
	if writeCount != 1 {
		t.Fatalf("writeCount = %d", writeCount)
	}
	audit, err := svc.ListMemoryAudit(context.Background(), ident, productdata.MemoryAuditInput{SourceRunID: run.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.Items) != 0 {
		t.Fatalf("external post-run should not create local proposal audit: %+v", audit.Items)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, event := range events {
		if event.Type == eventMemoryProviderCommitCompleted {
			found = true
			if strings.Contains(event.Summary, "secret") || strings.Contains(fmt.Sprint(event.Metadata), "secret") {
				t.Fatalf("event leaked secret: %+v", event)
			}
		}
	}
	if !found {
		t.Fatalf("commit event missing: %+v", events)
	}
}
