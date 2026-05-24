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
