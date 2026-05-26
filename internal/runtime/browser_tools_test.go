package runtime

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestBrowserOpenSnapshotAndClickLink(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><title>Home</title><body>Home page <a href="/docs">Docs</a><a href="file:///etc/passwd">Blocked</a></body></html>`))
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><title>Docs</title><body>Docs page</body></html>`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	store := NewBrowserSessionStore()
	executor := BrowserToolExecutor{Store: store, AllowPrivateHosts: true}

	open, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_browser", ToolCallID: "tc_open", ToolName: productdata.ToolNameBrowserOpen, ArgumentsSummary: map[string]any{"url": server.URL, "max_bytes": 4096}})
	if err != nil {
		t.Fatal(err)
	}
	sessionID, _ := open["session_id"].(string)
	if sessionID == "" || open["operation"] != "open" || open["scope"] != "browser" || open["title"] != "Home" {
		t.Fatalf("open = %+v", open)
	}
	if !strings.Contains(fmt.Sprint(open["text_excerpt"]), "Home page") || strings.Contains(fmt.Sprint(open), "<html>") {
		t.Fatalf("open leaked raw html or missed text: %+v", open)
	}
	links, ok := open["links"].([]map[string]any)
	if !ok || len(links) != 2 || links[0]["text"] != "Docs" || links[1]["blocked"] != true {
		t.Fatalf("links = %#v", open["links"])
	}

	snapshot, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_browser", ToolCallID: "tc_snapshot", ToolName: productdata.ToolNameBrowserSnapshot, ArgumentsSummary: map[string]any{"session_id": sessionID}})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot["operation"] != "snapshot" || snapshot["session_id"] != sessionID || snapshot["title"] != "Home" {
		t.Fatalf("snapshot = %+v", snapshot)
	}

	click, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_browser", ToolCallID: "tc_click", ToolName: productdata.ToolNameBrowserClickLink, ArgumentsSummary: map[string]any{"session_id": sessionID, "link_index": 0}})
	if err != nil {
		t.Fatal(err)
	}
	if click["operation"] != "click_link" || click["previous_url"] != server.URL || click["title"] != "Docs" {
		t.Fatalf("click = %+v", click)
	}
}

func TestBrowserRejectsUnsafeTargetsAndUnknownSessions(t *testing.T) {
	executor := BrowserToolExecutor{Store: NewBrowserSessionStore(), Resolver: func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.2")}, nil
	}}
	for _, raw := range []string{
		"file:///etc/passwd",
		"ftp://example.com/file",
		"https://user:pass@example.com/",
		"http://localhost:8080/",
		"http://127.0.0.1:8080/",
		"https://example.com/private",
	} {
		_, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_browser", ToolCallID: "tc_open", ToolName: productdata.ToolNameBrowserOpen, ArgumentsSummary: map[string]any{"url": raw}})
		if err == nil {
			t.Fatalf("open(%q) err = nil", raw)
		}
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_browser", ToolName: productdata.ToolNameBrowserSnapshot, ArgumentsSummary: map[string]any{"session_id": "missing"}}); err == nil {
		t.Fatal("snapshot unknown session err = nil")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_browser", ToolName: productdata.ToolNameBrowserClickLink, ArgumentsSummary: map[string]any{"session_id": "missing", "link_index": 0}}); err == nil {
		t.Fatal("click unknown session err = nil")
	}
}

func TestBrowserBoundsSnapshotAndRejectsBlockedLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><title>Many</title><body>` + strings.Repeat("x", 256) + `<a href="http://127.0.0.1/private">Private</a></body></html>`))
	}))
	defer server.Close()
	executor := BrowserToolExecutor{Store: NewBrowserSessionStore(), AllowPrivateHosts: true}
	open, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_browser", ToolCallID: "tc_open", ToolName: productdata.ToolNameBrowserOpen, ArgumentsSummary: map[string]any{"url": server.URL, "max_bytes": 64}})
	if err != nil {
		t.Fatal(err)
	}
	if open["truncated"] != true {
		t.Fatalf("open should be truncated: %+v", open)
	}
	sessionID := open["session_id"].(string)
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_browser", ToolCallID: "tc_click", ToolName: productdata.ToolNameBrowserClickLink, ArgumentsSummary: map[string]any{"session_id": sessionID, "link_index": 0}}); err == nil {
		t.Fatal("blocked link click err = nil")
	}
}
