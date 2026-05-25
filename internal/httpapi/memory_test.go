package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
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
