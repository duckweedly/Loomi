package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestMemoryHandlersListSearchApproveAndDelete(t *testing.T) {
	svc := productdata.NewMemoryService()
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	propose := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals", `{"scope_type":"user","title":"Preference","content":"Prefers safe memory snapshots","idempotency_key":"p1"}`)
	if propose.Code != http.StatusCreated || strings.Contains(propose.Body.String(), "sk-secret") {
		t.Fatalf("propose status=%d body=%s", propose.Code, propose.Body.String())
	}
	proposalID := decodeStringField(t, propose.Body.Bytes(), "proposal", "id")

	beforeApproval := requestJSON(t, srv, http.MethodPost, "/v1/memory/search", `{"query":"safe","limit":10}`)
	if beforeApproval.Code != http.StatusOK || !strings.Contains(beforeApproval.Body.String(), `"items":[]`) {
		t.Fatalf("before approval status=%d body=%s", beforeApproval.Code, beforeApproval.Body.String())
	}

	approve := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals/"+proposalID+"/approve", `{"idempotency_key":"a1"}`)
	if approve.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", approve.Code, approve.Body.String())
	}
	entryID := decodeStringField(t, approve.Body.Bytes(), "entry", "id")

	list := requestJSON(t, srv, http.MethodGet, "/v1/memory", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), entryID) || strings.Contains(list.Body.String(), "sk-secret") {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}

	deleteRes := requestJSON(t, srv, http.MethodDelete, "/v1/memory/"+entryID, `{"reason":"user_request"}`)
	if deleteRes.Code != http.StatusOK || !strings.Contains(deleteRes.Body.String(), `"status":"tombstoned"`) {
		t.Fatalf("delete status=%d body=%s", deleteRes.Code, deleteRes.Body.String())
	}

	afterDelete := requestJSON(t, srv, http.MethodPost, "/v1/memory/search", `{"query":"safe","limit":10}`)
	if afterDelete.Code != http.StatusOK || !strings.Contains(afterDelete.Body.String(), `"items":[]`) {
		t.Fatalf("after delete status=%d body=%s", afterDelete.Code, afterDelete.Body.String())
	}
}

func TestMemoryHandlersRequireScopeForThreadEntryReadDelete(t *testing.T) {
	svc := productdata.NewMemoryService()
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	ident := identity.LocalDevIdentity()
	ctx := context.Background()
	threadA, err := svc.CreateThread(ctx, ident, productdata.CreateThreadInput{Title: "Thread A", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	threadB, err := svc.CreateThread(ctx, ident, productdata.CreateThreadInput{Title: "Thread B", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := svc.CreateMemoryEntry(ctx, ident, productdata.CreateMemoryEntryInput{ScopeType: productdata.MemoryScopeThread, ScopeID: threadA.ID, Title: "Thread note", Content: "Thread A only", SourceThreadID: threadA.ID})
	if err != nil {
		t.Fatal(err)
	}

	outOfScopeRead := requestJSON(t, srv, http.MethodGet, "/v1/memory/entries/"+entry.ID+"?scope_type=thread&scope_id="+threadB.ID, "")
	if outOfScopeRead.Code != http.StatusNotFound || strings.Contains(outOfScopeRead.Body.String(), entry.ID) {
		t.Fatalf("out-of-scope read status=%d body=%s", outOfScopeRead.Code, outOfScopeRead.Body.String())
	}
	unscopedRead := requestJSON(t, srv, http.MethodGet, "/v1/memory/entries/"+entry.ID, "")
	if unscopedRead.Code != http.StatusNotFound || strings.Contains(unscopedRead.Body.String(), entry.ID) {
		t.Fatalf("unscoped read status=%d body=%s", unscopedRead.Code, unscopedRead.Body.String())
	}
	outOfScopeDelete := requestJSON(t, srv, http.MethodDelete, "/v1/memory/entries/"+entry.ID, `{"scope_type":"thread","scope_id":"`+threadB.ID+`","reason":"wrong thread"}`)
	if outOfScopeDelete.Code != http.StatusNotFound || strings.Contains(outOfScopeDelete.Body.String(), entry.ID) {
		t.Fatalf("out-of-scope delete status=%d body=%s", outOfScopeDelete.Code, outOfScopeDelete.Body.String())
	}

	detail := requestJSON(t, srv, http.MethodGet, "/v1/memory/entries/"+entry.ID+"?scope_type=thread&scope_id="+threadA.ID, "")
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), entry.ID) || strings.Contains(detail.Body.String(), `"content"`) {
		t.Fatalf("scoped detail status=%d body=%s", detail.Code, detail.Body.String())
	}
	deleteRes := requestJSON(t, srv, http.MethodDelete, "/v1/memory/entries/"+entry.ID, `{"scope_type":"thread","scope_id":"`+threadA.ID+`","reason":"right thread"}`)
	if deleteRes.Code != http.StatusOK {
		t.Fatalf("scoped delete status=%d body=%s", deleteRes.Code, deleteRes.Body.String())
	}
}

func TestMemoryHandlersRejectThreadSearchWithoutScopeID(t *testing.T) {
	svc := productdata.NewMemoryService()
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	list := requestJSON(t, srv, http.MethodGet, "/v1/memory?scope_type=thread&q=management", "")
	if list.Code != http.StatusBadRequest || !strings.Contains(list.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}

	search := requestJSON(t, srv, http.MethodPost, "/v1/memory/search", `{"scope_type":"thread","query":"management"}`)
	if search.Code != http.StatusBadRequest || !strings.Contains(search.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("search status=%d body=%s", search.Code, search.Body.String())
	}
}

func TestMemoryHandlersManagementFiltersDetailAndAudit(t *testing.T) {
	svc := productdata.NewMemoryService()
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	ident := identity.LocalDevIdentity()
	ctx := context.Background()
	thread, err := svc.CreateThread(ctx, ident, productdata.CreateThreadInput{Title: "Memory audit", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(ctx, ident, thread.ID, productdata.StartRunInput{ScriptName: "m14_memory"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(ctx, ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventMemorySnapshotLoaded, Summary: "Memory snapshot loaded", Metadata: map[string]any{"status": "empty", "entry_count": 0, "redaction_applied": true}}); err != nil {
		t.Fatal(err)
	}

	propose := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals", `{"scope_type":"thread","scope_id":"`+thread.ID+`","title":"Project style","content":"Prefers short memory management UX","source_thread_id":"`+thread.ID+`","source_run_id":"`+run.ID+`","idempotency_key":"m14-propose"}`)
	if propose.Code != http.StatusCreated {
		t.Fatalf("propose status=%d body=%s", propose.Code, propose.Body.String())
	}
	proposalID := decodeStringField(t, propose.Body.Bytes(), "proposal", "id")
	approve := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals/"+proposalID+"/approve", `{"idempotency_key":"m14-approve"}`)
	if approve.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", approve.Code, approve.Body.String())
	}
	entryID := decodeStringField(t, approve.Body.Bytes(), "entry", "id")

	denied := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals", `{"scope_type":"thread","scope_id":"`+thread.ID+`","title":"Denied","content":"Do not store this temporary note","source_thread_id":"`+thread.ID+`","source_run_id":"`+run.ID+`","idempotency_key":"m14-deny-propose"}`)
	if denied.Code != http.StatusCreated {
		t.Fatalf("denied proposal status=%d body=%s", denied.Code, denied.Body.String())
	}
	deniedID := decodeStringField(t, denied.Body.Bytes(), "proposal", "id")
	deny := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals/"+deniedID+"/deny", `{"idempotency_key":"m14-deny"}`)
	if deny.Code != http.StatusOK {
		t.Fatalf("deny status=%d body=%s", deny.Code, deny.Body.String())
	}

	filtered := requestJSON(t, srv, http.MethodGet, "/v1/memory?scope_type=thread&scope_id="+thread.ID+"&source_run_id="+run.ID+"&source_type=run&q=management", "")
	if filtered.Code != http.StatusOK || !strings.Contains(filtered.Body.String(), entryID) || strings.Contains(filtered.Body.String(), "content") {
		t.Fatalf("filtered list status=%d body=%s", filtered.Code, filtered.Body.String())
	}
	if !strings.Contains(filtered.Body.String(), `"source_type":"run"`) || !strings.Contains(filtered.Body.String(), `"status":"approved"`) {
		t.Fatalf("filtered metadata missing: %s", filtered.Body.String())
	}
	threadFiltered := requestJSON(t, srv, http.MethodGet, "/v1/memory?scope_type=thread&scope_id="+thread.ID+"&source_thread_id="+thread.ID+"&q=management", "")
	if threadFiltered.Code != http.StatusOK || !strings.Contains(threadFiltered.Body.String(), entryID) || !strings.Contains(threadFiltered.Body.String(), `"source_thread_id":"`+thread.ID+`"`) {
		t.Fatalf("source_thread_id filter status=%d body=%s", threadFiltered.Code, threadFiltered.Body.String())
	}

	detail := requestJSON(t, srv, http.MethodGet, "/v1/memory/entries/"+entryID+"?scope_type=thread&scope_id="+thread.ID, "")
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"scope_id":"`+thread.ID+`"`) || strings.Contains(detail.Body.String(), `"content"`) {
		t.Fatalf("detail status=%d body=%s", detail.Code, detail.Body.String())
	}

	deleteRes := requestJSON(t, srv, http.MethodDelete, "/v1/memory/entries/"+entryID, `{"reason":"confirmed in settings","scope_type":"thread","scope_id":"`+thread.ID+`"}`)
	if deleteRes.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleteRes.Code, deleteRes.Body.String())
	}
	tombstoned := requestJSON(t, srv, http.MethodGet, "/v1/memory/entries/"+entryID+"?scope_type=thread&scope_id="+thread.ID, "")
	if tombstoned.Code != http.StatusOK || !strings.Contains(tombstoned.Body.String(), `"status":"tombstoned"`) || strings.Contains(tombstoned.Body.String(), `"content"`) {
		t.Fatalf("tombstoned detail status=%d body=%s", tombstoned.Code, tombstoned.Body.String())
	}

	audit := requestJSON(t, srv, http.MethodGet, "/v1/memory/audit?source_run_id="+run.ID, "")
	auditBody := audit.Body.String()
	for _, eventType := range []string{"memory_write_proposed", "memory_write_approved", "memory_write_denied", "memory_deleted"} {
		if audit.Code != http.StatusOK || !strings.Contains(auditBody, eventType) {
			t.Fatalf("audit missing %s status=%d body=%s", eventType, audit.Code, auditBody)
		}
	}
	if strings.Contains(auditBody, "Prefers short memory management UX") || strings.Contains(auditBody, "/Users/") || strings.Contains(auditBody, "sk-secret") {
		t.Fatalf("audit leaked unsafe metadata: %s", auditBody)
	}
}

func TestMemoryHandlersAuditAfterTerminalRunAndRedaction(t *testing.T) {
	svc := productdata.NewMemoryService()
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	ident := identity.LocalDevIdentity()
	ctx := context.Background()
	thread, err := svc.CreateThread(ctx, ident, productdata.CreateThreadInput{Title: "Terminal audit", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(ctx, ident, thread.ID, productdata.StartRunInput{ScriptName: "terminal_audit"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(ctx, ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: productdata.EventRunCompleted, Summary: "Run completed"}); err != nil {
		t.Fatal(err)
	}

	body := `{"scope_type":"thread","scope_id":"` + thread.ID + `","title":"Terminal audit","content":"Keep safe terminal audit","source_thread_id":"` + thread.ID + `","source_run_id":"` + run.ID + `","idempotency_key":"terminal-propose"}`
	propose := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals", body)
	if propose.Code != http.StatusCreated {
		t.Fatalf("propose status=%d body=%s", propose.Code, propose.Body.String())
	}
	proposalID := decodeStringField(t, propose.Body.Bytes(), "proposal", "id")
	deny := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals/"+proposalID+"/deny", `{"reason":"provider trace marker","idempotency_key":"terminal-deny"}`)
	if deny.Code != http.StatusOK {
		t.Fatalf("deny status=%d body=%s", deny.Code, deny.Body.String())
	}
	denyAgain := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals/"+proposalID+"/deny", `{"reason":"retry","idempotency_key":"terminal-deny"}`)
	if denyAgain.Code != http.StatusOK {
		t.Fatalf("deny again status=%d body=%s", denyAgain.Code, denyAgain.Body.String())
	}

	audit := requestJSON(t, srv, http.MethodGet, "/v1/memory/audit?source_run_id="+run.ID, "")
	auditBody := audit.Body.String()
	if audit.Code != http.StatusOK || !strings.Contains(auditBody, "memory_write_proposed") || !strings.Contains(auditBody, "memory_write_denied") {
		t.Fatalf("audit status=%d body=%s", audit.Code, auditBody)
	}
	if strings.Count(auditBody, "memory_write_denied") != 1 {
		t.Fatalf("deny audit duplicated: %s", auditBody)
	}
	for _, leaked := range []string{"/home/xuean", "Authorization", "sk-secret", "stdout:", "provider trace"} {
		if strings.Contains(auditBody, leaked) {
			t.Fatalf("audit leaked %q: %s", leaked, auditBody)
		}
	}
}

func TestMemoryAuditKeepsTerminalRunAndRedactsExpandedSensitiveText(t *testing.T) {
	svc := productdata.NewMemoryService()
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	ident := identity.LocalDevIdentity()
	ctx := context.Background()
	thread, err := svc.CreateThread(ctx, ident, productdata.CreateThreadInput{Title: "Terminal audit", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(ctx, ident, thread.ID, productdata.StartRunInput{ScriptName: "m14_terminal"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(ctx, ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: productdata.EventRunCompleted, Summary: "Run completed", Metadata: map[string]any{}}); err != nil {
		t.Fatal(err)
	}

	propose := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals", `{"scope_type":"thread","scope_id":"`+thread.ID+`","title":"Terminal proposal","content":"Safe terminal audit survives","source_thread_id":"`+thread.ID+`","source_run_id":"`+run.ID+`","idempotency_key":"terminal-audit"}`)
	if propose.Code != http.StatusCreated {
		t.Fatalf("propose status=%d body=%s", propose.Code, propose.Body.String())
	}
	audit := requestJSON(t, srv, http.MethodGet, "/v1/memory/audit?source_run_id="+run.ID, "")
	if audit.Code != http.StatusOK || !strings.Contains(audit.Body.String(), "memory_write_proposed") {
		t.Fatalf("terminal audit status=%d body=%s", audit.Code, audit.Body.String())
	}

	for _, raw := range []string{"/home/xuean/.ssh/id_ed25519", `C:\Users\xuean\.env`, "stdout provider trace sk-secret", "stderr Authorization: Bearer token"} {
		if got := productdata.RedactEventText(raw); got != "[redacted]" {
			t.Fatalf("RedactEventText(%q) = %q", raw, got)
		}
	}
}

func decodeStringField(t *testing.T, raw []byte, objectKey string, field string) string {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	object, ok := body[objectKey].(map[string]any)
	if !ok {
		t.Fatalf("%s missing in %s", objectKey, string(raw))
	}
	value, ok := object[field].(string)
	if !ok || value == "" {
		t.Fatalf("%s.%s missing in %s", objectKey, field, string(raw))
	}
	return value
}
