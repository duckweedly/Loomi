package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestMemoryToolExecutorSearchReadStatusWriteAndForget(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, run := memoryTestThreadRun(t, svc)
	entry, err := svc.CreateMemoryEntry(context.Background(), ident, productdata.CreateMemoryEntryInput{ScopeType: productdata.MemoryScopeThread, ScopeID: thread.ID, Title: "Decision", Content: "Keep memory tool results safe.", SourceThreadID: thread.ID, SourceRunID: run.ID})
	if err != nil {
		t.Fatal(err)
	}
	executor := MemoryToolExecutor{Service: svc, Ident: ident}

	search, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": "memory tool", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	results, _ := search["items"].([]map[string]any)
	if search["operation"] != "search" || len(results) != 1 || results[0]["entry_id"] != entry.ID || results[0]["content"] != nil {
		t.Fatalf("search = %+v", search)
	}

	list, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryList, ArgumentsSummary: map[string]any{"limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	listResults, _ := list["items"].([]map[string]any)
	if list["operation"] != "list" || len(listResults) != 1 || listResults[0]["content"] != nil {
		t.Fatalf("list = %+v", list)
	}

	read, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryRead, ArgumentsSummary: map[string]any{"entry_id": entry.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if read["operation"] != "read" || read["entry_id"] != entry.ID || read["content"] != nil {
		t.Fatalf("read = %+v", read)
	}

	status, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryStatus, ArgumentsSummary: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if status["operation"] != "status" || status["provider"] != string(productdata.MemoryProviderLocal) || status["state"] != string(productdata.MemoryProviderStateAvailable) {
		t.Fatalf("status = %+v", status)
	}

	write, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryWrite, ArgumentsSummary: map[string]any{"title": "Follow-up", "content": "Write proposals stay pending until approval."}})
	if err != nil {
		t.Fatal(err)
	}
	proposalID, _ := write["proposal_id"].(string)
	if write["operation"] != "write_proposal" || !strings.HasPrefix(proposalID, "memprop_") || write["content"] != nil {
		t.Fatalf("write = %+v", write)
	}

	edit, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryEdit, ArgumentsSummary: map[string]any{"proposal_id": proposalID, "title": "Edited follow-up", "content": "Edited proposals still stay pending."}})
	if err != nil {
		t.Fatal(err)
	}
	if edit["operation"] != "edit_proposal" || edit["proposal_id"] != proposalID || edit["content"] != nil {
		t.Fatalf("edit = %+v", edit)
	}

	contextResult, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryContext, ArgumentsSummary: map[string]any{"query": "memory tool", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if contextResult["operation"] != "context" || contextResult["provider"] == nil || contextResult["items"] == nil {
		t.Fatalf("context = %+v", contextResult)
	}

	connections, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryConnections, ArgumentsSummary: map[string]any{"entry_id": entry.ID, "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if connections["operation"] != "connections" || connections["items"] == nil {
		t.Fatalf("connections = %+v", connections)
	}

	timeline, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryTimeline, ArgumentsSummary: map[string]any{"limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if timeline["operation"] != "timeline" || timeline["items"] == nil {
		t.Fatalf("timeline = %+v", timeline)
	}

	if _, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Thread memory search marker"}); err != nil {
		t.Fatal(err)
	}
	threadSearch, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryThreadSearch, ArgumentsSummary: map[string]any{"query": "marker", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if threadSearch["operation"] != "thread_search" || threadSearch["count"] != 1 {
		t.Fatalf("thread search = %+v", threadSearch)
	}

	threadFetch, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryThreadFetch, ArgumentsSummary: map[string]any{"thread_id": thread.ID, "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if threadFetch["operation"] != "thread_fetch" || threadFetch["count"] != 1 {
		t.Fatalf("thread fetch = %+v", threadFetch)
	}

	forget, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryForget, ArgumentsSummary: map[string]any{"entry_id": entry.ID, "reason": "obsolete"}})
	if err != nil {
		t.Fatal(err)
	}
	if forget["operation"] != "forget" || forget["entry_id"] != entry.ID || forget["status"] != string(productdata.MemoryEntryTombstoned) {
		t.Fatalf("forget = %+v", forget)
	}
}

func TestMemoryToolExecutorRejectsUnsafeInputs(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread, run := memoryTestThreadRun(t, svc)
	executor := MemoryToolExecutor{Service: svc}

	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": ""}}); err == nil {
		t.Fatal("expected empty search to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryRead, ArgumentsSummary: map[string]any{"entry_id": "mem_missing"}}); err == nil {
		t.Fatal("expected unknown memory read to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryWrite, ArgumentsSummary: map[string]any{"title": "Huge", "content": strings.Repeat("x", 5000)}}); err == nil {
		t.Fatal("expected oversized memory write to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryThreadSearch, ArgumentsSummary: map[string]any{"query": ""}}); err == nil {
		t.Fatal("expected empty thread search to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: "memory.export", ArgumentsSummary: map[string]any{}}); err == nil {
		t.Fatal("expected unsupported memory tool to fail")
	}
}

func TestMemoryToolExecutorNotebookLifecycle(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, run := memoryTestThreadRun(t, svc)
	executor := MemoryToolExecutor{Service: svc, Ident: ident}

	write, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameNotebookWrite, ArgumentsSummary: map[string]any{"title": "Project notebook", "content": "Store structured project notes separately from semantic memory."}})
	if err != nil {
		t.Fatal(err)
	}
	entryID, _ := write["entry_id"].(string)
	if write["operation"] != "notebook_write" || !strings.HasPrefix(entryID, "mem_") || write["content"] != nil {
		t.Fatalf("notebook write = %+v", write)
	}

	read, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameNotebookRead, ArgumentsSummary: map[string]any{"entry_id": entryID}})
	if err != nil {
		t.Fatal(err)
	}
	if read["operation"] != "notebook_read" || read["entry_id"] != entryID || read["content"] != nil {
		t.Fatalf("notebook read = %+v", read)
	}

	search, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": "structured", "source_type": "notebook", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	items, _ := search["items"].([]map[string]any)
	if len(items) != 1 || items[0]["entry_id"] != entryID || items[0]["source_type"] != "notebook" {
		t.Fatalf("notebook search = %+v", search)
	}

	edit, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameNotebookEdit, ArgumentsSummary: map[string]any{"entry_id": entryID, "title": "Project notebook updated", "content": "Notebook edits replace entries through tombstone audit."}})
	if err != nil {
		t.Fatal(err)
	}
	replacementID, _ := edit["entry_id"].(string)
	if edit["operation"] != "notebook_edit" || replacementID == "" || replacementID == entryID || edit["replaced_entry_id"] != entryID {
		t.Fatalf("notebook edit = %+v", edit)
	}

	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameNotebookRead, ArgumentsSummary: map[string]any{"entry_id": entryID}}); err == nil {
		t.Fatal("expected old notebook entry to be unavailable after edit")
	}

	forget, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameNotebookForget, ArgumentsSummary: map[string]any{"entry_id": replacementID, "reason": "obsolete"}})
	if err != nil {
		t.Fatal(err)
	}
	if forget["operation"] != "notebook_forget" || forget["entry_id"] != replacementID || forget["status"] != string(productdata.MemoryEntryTombstoned) {
		t.Fatalf("notebook forget = %+v", forget)
	}
}

func TestMemoryToolExecutorSearchesOpenVikingProvider(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, run := memoryTestThreadRun(t, svc)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "root-secret" {
			t.Fatalf("missing root key header")
		}
		switch r.URL.Path {
		case "/api/v1/sessions":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"session_id": "session_1"}})
		case "/api/v1/sessions/session_1/messages":
			w.WriteHeader(http.StatusNoContent)
		case "/api/v1/sessions/session_1/commit":
			w.WriteHeader(http.StatusNoContent)
		case "/api/v1/search/find":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"memories": []map[string]any{{"uri": "viking://user/user_local_dev/memories/decision", "abstract": "Decision\nKeep OpenViking recall safe.", "score": 0.91, "match_reason": "semantic"}}}})
		case "/api/v1/content/read":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "Decision\n\nKeep OpenViking recall safe."})
		case "/api/v1/content/write":
			w.WriteHeader(http.StatusNoContent)
		case "/api/v1/fs/ls":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": []map[string]any{{"uri": "viking://user/user_local_dev/memories/decision/child", "isDir": false}, {"uri": "viking://user/user_local_dev/memories/folder", "isDir": true}}})
				return
			}
			http.NotFound(w, r)
		case "/api/v1/fs":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	if _, err := svc.SaveMemoryProviderConfig(context.Background(), ident, productdata.MemoryProviderConfig{Enabled: true, Provider: productdata.MemoryProviderOpenViking, OpenViking: productdata.OpenVikingMemoryConfig{BaseURL: server.URL, RootAPIKey: "root-secret", EmbeddingModel: "embed", VLMModel: "vlm"}}); err != nil {
		t.Fatal(err)
	}
	executor := MemoryToolExecutor{Service: svc, Ident: ident}

	search, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": "decision", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	items, _ := search["items"].([]map[string]any)
	if search["provider"] != string(productdata.MemoryProviderOpenViking) || len(items) != 1 || items[0]["entry_id"] != "viking://user/user_local_dev/memories/decision" || items[0]["content"] != nil {
		t.Fatalf("openviking search = %+v", search)
	}

	read, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryRead, ArgumentsSummary: map[string]any{"entry_id": "viking://user/user_local_dev/memories/decision"}})
	if err != nil {
		t.Fatal(err)
	}
	if read["provider"] != string(productdata.MemoryProviderOpenViking) || read["content"] != nil || !strings.Contains(read["summary"].(string), "Keep OpenViking") {
		t.Fatalf("openviking read = %+v", read)
	}

	write, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryWrite, ArgumentsSummary: map[string]any{"title": "Decision", "content": "Write to OpenViking after tool approval."}})
	if err != nil {
		t.Fatal(err)
	}
	if write["operation"] != "write_provider" || write["provider"] != string(productdata.MemoryProviderOpenViking) || write["content"] != nil {
		t.Fatalf("openviking write = %+v", write)
	}

	edit, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryEdit, ArgumentsSummary: map[string]any{"entry_id": "viking://user/user_local_dev/memories/decision", "title": "Decision", "content": "Replace OpenViking memory."}})
	if err != nil {
		t.Fatal(err)
	}
	if edit["operation"] != "edit_provider" || edit["provider"] != string(productdata.MemoryProviderOpenViking) || edit["content"] != nil {
		t.Fatalf("openviking edit = %+v", edit)
	}

	forget, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryForget, ArgumentsSummary: map[string]any{"entry_id": "viking://user/user_local_dev/memories/decision"}})
	if err != nil {
		t.Fatal(err)
	}
	if forget["operation"] != "forget_provider" || forget["provider"] != string(productdata.MemoryProviderOpenViking) {
		t.Fatalf("openviking forget = %+v", forget)
	}

	connections, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryConnections, ArgumentsSummary: map[string]any{"entry_id": "viking://user/user_local_dev/memories/decision", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if connections["operation"] != "connections" || connections["provider"] != string(productdata.MemoryProviderOpenViking) || connections["count"] != 2 {
		t.Fatalf("openviking connections = %+v", connections)
	}
}

func TestMemoryToolExecutorSearchesNowledgeProvider(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, run := memoryTestThreadRun(t, svc)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer nowledge-secret" || r.Header.Get("x-nmem-api-key") != "nowledge-secret" {
			t.Fatalf("missing nowledge auth headers")
		}
		switch r.URL.Path {
		case "/memories/search":
			_ = json.NewEncoder(w).Encode(map[string]any{"memories": []map[string]any{{"id": "mem_remote", "title": "Preference", "content": "Use Nowledge recall.", "score": 0.82, "relevance_reason": "keyword"}}})
		case "/memories/mem_remote":
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "Preference", "content": "Use Nowledge recall."})
		case "/memories":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "mem_created"})
		case "/graph/expand/mem_remote":
			_ = json.NewEncoder(w).Encode(map[string]any{"neighbors": []map[string]any{{"id": "neighbor_1", "label": "Related memory", "node_type": "memory", "summary": "Related safe summary"}}, "edges": []map[string]any{{"source": "mem_remote", "target": "neighbor_1", "edge_type": "RELATED", "weight": 0.7, "label": "related"}}})
		case "/agent/feed/events":
			_ = json.NewEncoder(w).Encode(map[string]any{"events": []map[string]any{{"id": "evt_1", "event_type": "memory_created", "title": "Created", "created_at": "2026-05-27T00:00:00Z", "memory_id": "mem_remote"}}})
		case "/threads/search":
			_ = json.NewEncoder(w).Encode(map[string]any{"threads": []map[string]any{{"thread_id": "remote_thread", "title": "Remote thread", "source": "nowledge", "message_count": 2, "score": 0.8, "snippets": []string{"remote snippet"}}}})
		case "/threads/remote_thread":
			_ = json.NewEncoder(w).Encode(map[string]any{"thread_id": "remote_thread", "title": "Remote thread", "message_count": 1, "messages": []map[string]any{{"role": "user", "content": "remote content", "timestamp": "2026-05-27T00:00:00Z"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	if _, err := svc.SaveMemoryProviderConfig(context.Background(), ident, productdata.MemoryProviderConfig{Enabled: true, Provider: productdata.MemoryProviderNowledge, Nowledge: productdata.NowledgeMemoryConfig{BaseURL: server.URL, APIKey: "nowledge-secret"}}); err != nil {
		t.Fatal(err)
	}
	executor := MemoryToolExecutor{Service: svc, Ident: ident}

	search, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": "preference", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	items, _ := search["items"].([]map[string]any)
	if search["provider"] != string(productdata.MemoryProviderNowledge) || len(items) != 1 || items[0]["entry_id"] != "nowledge://memory/mem_remote" || items[0]["content"] != nil {
		t.Fatalf("nowledge search = %+v", search)
	}

	read, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryRead, ArgumentsSummary: map[string]any{"entry_id": "nowledge://memory/mem_remote"}})
	if err != nil {
		t.Fatal(err)
	}
	if read["provider"] != string(productdata.MemoryProviderNowledge) || read["content"] != nil || !strings.Contains(read["summary"].(string), "Use Nowledge") {
		t.Fatalf("nowledge read = %+v", read)
	}

	write, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryWrite, ArgumentsSummary: map[string]any{"title": "Preference", "content": "Write to Nowledge after tool approval."}})
	if err != nil {
		t.Fatal(err)
	}
	if write["operation"] != "write_provider" || write["provider"] != string(productdata.MemoryProviderNowledge) || write["entry_id"] != "nowledge://memory/mem_created" || write["content"] != nil {
		t.Fatalf("nowledge write = %+v", write)
	}

	forget, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryForget, ArgumentsSummary: map[string]any{"entry_id": "nowledge://memory/mem_remote"}})
	if err != nil {
		t.Fatal(err)
	}
	if forget["operation"] != "forget_provider" || forget["provider"] != string(productdata.MemoryProviderNowledge) {
		t.Fatalf("nowledge forget = %+v", forget)
	}

	connections, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryConnections, ArgumentsSummary: map[string]any{"entry_id": "nowledge://memory/mem_remote", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if connections["operation"] != "connections" || connections["provider"] != string(productdata.MemoryProviderNowledge) || connections["count"] != 1 {
		t.Fatalf("nowledge connections = %+v", connections)
	}

	timeline, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryTimeline, ArgumentsSummary: map[string]any{"limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if timeline["operation"] != "timeline" || timeline["provider"] != string(productdata.MemoryProviderNowledge) || timeline["count"] != 1 {
		t.Fatalf("nowledge timeline = %+v", timeline)
	}

	threadSearch, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryThreadSearch, ArgumentsSummary: map[string]any{"query": "remote", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if threadSearch["operation"] != "thread_search" || threadSearch["provider"] != string(productdata.MemoryProviderNowledge) || threadSearch["count"] != 1 {
		t.Fatalf("nowledge thread search = %+v", threadSearch)
	}

	threadFetch, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameMemoryThreadFetch, ArgumentsSummary: map[string]any{"thread_id": "remote_thread", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if threadFetch["operation"] != "thread_fetch" || threadFetch["provider"] != string(productdata.MemoryProviderNowledge) || threadFetch["count"] != 1 {
		t.Fatalf("nowledge thread fetch = %+v", threadFetch)
	}
}

func memoryTestThreadRun(t *testing.T, svc *productdata.MemoryService) (productdata.Thread, productdata.Run) {
	t.Helper()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Memory", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	return thread, run
}
