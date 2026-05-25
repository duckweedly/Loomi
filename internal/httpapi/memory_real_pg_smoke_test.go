package httpapi

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestM13MemoryRealPGHTTPAPISmoke(t *testing.T) {
	databaseURL := os.Getenv("LOOMI_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("LOOMI_TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	assertM13MemoryTablesExist(t, ctx, pool)

	repo := productdata.NewPostgresRepository(pool)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, repo)
	ident := identity.LocalDevIdentity()
	unique := productdata.NewThreadID()

	thread, err := repo.CreateThread(ctx, ident, productdata.CreateThreadInput{Title: "M13.5 memory smoke " + unique, Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := repo.CreateMessage(ctx, ident, thread.ID, productdata.CreateMessageInput{Content: "Please remember the closeout preference."})
	if err != nil {
		t.Fatal(err)
	}
	run, err := repo.StartRun(ctx, ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := repo.ClaimBackgroundJob(ctx, ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_m13_5_" + unique, LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}

	safeContent := "M13.5 closeout prefers durable Postgres memory smoke " + unique
	propose := requestJSON(t, srv, "POST", "/v1/memory/write-proposals", `{"scope_type":"user","title":"M13.5 closeout","content":"`+safeContent+`","source_thread_id":"`+thread.ID+`","source_run_id":"`+run.ID+`","idempotency_key":"proposal-`+unique+`"}`)
	assertStatus(t, propose.Code, 201, propose.Body.String())
	assertBodyExcludes(t, propose.Body.String(), "proposal response", "sk-secret", "Authorization", "/Users/")
	proposalID := decodeStringField(t, propose.Body.Bytes(), "proposal", "id")

	pendingSearch := requestJSON(t, srv, "POST", "/v1/memory/search", `{"query":"`+unique+`","limit":10}`)
	assertStatus(t, pendingSearch.Code, 200, pendingSearch.Body.String())
	assertJSONItemCount(t, pendingSearch.Body.Bytes(), 0)

	approve := requestJSON(t, srv, "POST", "/v1/memory/write-proposals/"+proposalID+"/approve", `{"reason":"closeout approve","idempotency_key":"approve-`+unique+`"}`)
	assertStatus(t, approve.Code, 200, approve.Body.String())
	entryID := decodeStringField(t, approve.Body.Bytes(), "entry", "id")
	approveAgain := requestJSON(t, srv, "POST", "/v1/memory/write-proposals/"+proposalID+"/approve", `{"reason":"retry approve","idempotency_key":"approve-`+unique+`"}`)
	assertStatus(t, approveAgain.Code, 200, approveAgain.Body.String())
	if againEntryID := decodeStringField(t, approveAgain.Body.Bytes(), "entry", "id"); againEntryID != entryID {
		t.Fatalf("duplicate approve entry id = %s, want %s", againEntryID, entryID)
	}
	assertPGCount(t, ctx, pool, `select count(*) from memory_entries where id=$1`, entryID, 1)

	list := requestJSON(t, srv, "GET", "/v1/memory", "")
	assertStatus(t, list.Code, 200, list.Body.String())
	if !strings.Contains(list.Body.String(), entryID) {
		t.Fatalf("list body missing approved entry %s: %s", entryID, list.Body.String())
	}
	assertBodyExcludes(t, list.Body.String(), "list response", "sk-secret", "Authorization", "/Users/")
	search := requestJSON(t, srv, "POST", "/v1/memory/search", `{"query":"`+unique+`","limit":10}`)
	assertStatus(t, search.Code, 200, search.Body.String())
	assertJSONItemCount(t, search.Body.Bytes(), 1)

	runContext, err := repo.PrepareRunContext(ctx, ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if len(runContext.MemorySnapshot.Entries) != 1 || runContext.MemorySnapshot.Entries[0].ID != entryID {
		t.Fatalf("memory snapshot = %+v, want entry %s", runContext.MemorySnapshot, entryID)
	}
	if summary := runContext.SafeSummary(); summary["memory_entry_count"] != 1 || summary["memory_status"] != "loaded" {
		t.Fatalf("safe summary = %+v", summary)
	}

	sensitive := requestJSON(t, srv, "POST", "/v1/memory/write-proposals", `{"scope_type":"user","title":"Sensitive","content":"Authorization: Bearer sk-secret-`+unique+`","source_run_id":"`+run.ID+`","idempotency_key":"sensitive-`+unique+`"}`)
	assertStatus(t, sensitive.Code, 201, sensitive.Body.String())
	assertBodyExcludes(t, sensitive.Body.String(), "sensitive proposal response", "sk-secret", "Authorization: Bearer")
	sensitiveID := decodeStringField(t, sensitive.Body.Bytes(), "proposal", "id")
	blockedApprove := requestJSON(t, srv, "POST", "/v1/memory/write-proposals/"+sensitiveID+"/approve", `{}`)
	if blockedApprove.Code != 400 {
		t.Fatalf("blocked approve status=%d body=%s", blockedApprove.Code, blockedApprove.Body.String())
	}

	otherIdent := identity.LocalIdentity{UserID: "user_other_" + unique, DisplayName: "Other", Source: "test"}
	otherEntry, err := repo.CreateMemoryEntry(ctx, otherIdent, productdata.CreateMemoryEntryInput{ScopeType: productdata.MemoryScopeUser, Title: "Other", Content: "Other user memory " + unique})
	if err != nil {
		t.Fatal(err)
	}
	outOfScopeDelete := requestJSON(t, srv, "DELETE", "/v1/memory/"+otherEntry.ID, `{"reason":"closeout"}`)
	if outOfScopeDelete.Code != 404 {
		t.Fatalf("out-of-scope delete status=%d body=%s", outOfScopeDelete.Code, outOfScopeDelete.Body.String())
	}
	if strings.Contains(outOfScopeDelete.Body.String(), otherEntry.ID) {
		t.Fatalf("out-of-scope response leaked entry id: %s", outOfScopeDelete.Body.String())
	}

	denied, err := repo.ProposeMemoryWrite(ctx, ident, productdata.ProposeMemoryWriteInput{ScopeType: productdata.MemoryScopeThread, ScopeID: thread.ID, Title: "Deny once", Content: "Do not keep denied " + unique, SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DenyMemoryWrite(ctx, ident, denied.ID, productdata.MemoryWriteDecisionInput{Reason: "first deny"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DenyMemoryWrite(ctx, ident, denied.ID, productdata.MemoryWriteDecisionInput{Reason: "retry deny"}); err != nil {
		t.Fatal(err)
	}

	deleteRes := requestJSON(t, srv, "DELETE", "/v1/memory/"+entryID, `{"reason":"closeout delete"}`)
	assertStatus(t, deleteRes.Code, 200, deleteRes.Body.String())
	if !strings.Contains(deleteRes.Body.String(), `"status":"tombstoned"`) {
		t.Fatalf("delete body = %s", deleteRes.Body.String())
	}
	deleteAgain := requestJSON(t, srv, "DELETE", "/v1/memory/"+entryID, `{"reason":"retry delete"}`)
	assertStatus(t, deleteAgain.Code, 200, deleteAgain.Body.String())

	afterDelete := requestJSON(t, srv, "POST", "/v1/memory/search", `{"query":"`+unique+`","limit":10}`)
	assertStatus(t, afterDelete.Code, 200, afterDelete.Body.String())
	assertJSONItemCount(t, afterDelete.Body.Bytes(), 0)
	afterDeleteContext, err := repo.PrepareRunContext(ctx, ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if len(afterDeleteContext.MemorySnapshot.Entries) != 0 || afterDeleteContext.MemorySnapshot.LoadStatus != "empty" {
		t.Fatalf("after delete memory snapshot = %+v", afterDeleteContext.MemorySnapshot)
	}

	events, err := repo.ListRunEvents(ctx, ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	encodedEvents, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	assertBodyExcludes(t, string(encodedEvents), "memory run events", "sk-secret", "Authorization: Bearer", "/Users/")
	for _, event := range events {
		counts[event.Type]++
	}
	if counts[productdata.EventMemoryWriteApproved] != 1 || counts[productdata.EventMemoryWriteDenied] != 1 || counts[productdata.EventMemoryEntryDeleted] != 1 {
		t.Fatalf("memory audit event counts = %+v", counts)
	}
}

func assertM13MemoryTablesExist(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var version int
	var dirty bool
	if err := pool.QueryRow(ctx, `select version, dirty from schema_migrations order by version desc limit 1`).Scan(&version, &dirty); err != nil {
		t.Fatal(err)
	}
	if version != 10 || dirty {
		t.Fatalf("schema_migrations version=%d dirty=%v; apply migrations through clean M14 version 10 first", version, dirty)
	}
	for _, table := range []string{"memory_entries", "memory_write_proposals", "memory_audit_events"} {
		var exists bool
		if err := pool.QueryRow(ctx, `select to_regclass($1) is not null`, table).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatalf("%s table missing; apply migrations through M13 first", table)
		}
	}
}

func assertStatus(t *testing.T, got int, want int, body string) {
	t.Helper()
	if got != want {
		t.Fatalf("status=%d want=%d body=%s", got, want, body)
	}
}

func assertJSONItemCount(t *testing.T, raw []byte, want int) {
	t.Helper()
	var body struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != want {
		t.Fatalf("items=%d want=%d body=%s", len(body.Items), want, string(raw))
	}
}

func assertPGCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, arg string, want int) {
	t.Helper()
	var got int
	if err := pool.QueryRow(ctx, query, arg).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("count=%d want=%d for %s", got, want, arg)
	}
}

func assertBodyExcludes(t *testing.T, body string, label string, disallowed ...string) {
	t.Helper()
	for _, value := range disallowed {
		if strings.Contains(body, value) {
			t.Fatalf("%s leaked %q: %s", label, value, body)
		}
	}
}
