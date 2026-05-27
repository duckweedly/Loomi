package runtime

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestWebSearchDefinitionDoesNotRequireManualApproval(t *testing.T) {
	defs := WebToolDefinitions()
	for _, def := range defs {
		if def.Name != productdata.ToolNameWebSearch {
			continue
		}
		if def.ApprovalPolicy != ToolApprovalNotRequired || def.SafetyClass != ToolSafetyPublicNetworkRead {
			t.Fatalf("web search definition = %+v", def)
		}
		return
	}
	t.Fatal("web.search definition missing")
}

func TestWebFetchExecutesBoundedTextFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><title>Fixture Docs</title><body>Hello from docs.</body></html>"))
	}))
	defer server.Close()
	executor := WebToolExecutor{AllowPrivateHosts: true}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": server.URL, "max_bytes": 64, "timeout_ms": 1000}})
	if err != nil {
		t.Fatal(err)
	}
	if result["operation"] != "fetch" || result["scope"] != "web" || result["status_code"] != 200 || result["title"] != "Fixture Docs" {
		t.Fatalf("web fetch result = %+v", result)
	}
	if !strings.Contains(fmt.Sprint(result["text_excerpt"]), "Hello from docs") {
		t.Fatalf("web fetch excerpt = %+v", result)
	}
	if strings.Contains(fmt.Sprint(result), "Set-Cookie") {
		t.Fatalf("web fetch leaked header detail: %+v", result)
	}
}

func TestWebFetchTruncatesAndSkipsUnsupportedContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/binary":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte{0, 1, 2, 3})
		default:
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(strings.Repeat("a", 128)))
		}
	}))
	defer server.Close()
	executor := WebToolExecutor{AllowPrivateHosts: true}

	text, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": server.URL, "max_bytes": 32}})
	if err != nil {
		t.Fatal(err)
	}
	if text["truncated"] != true || text["bytes_read"] != 32 {
		t.Fatalf("truncated result = %+v", text)
	}
	binary, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": server.URL + "/binary"}})
	if err != nil {
		t.Fatal(err)
	}
	if binary["unsupported_content"] != true {
		t.Fatalf("binary result = %+v", binary)
	}
	if _, leaked := binary["text_excerpt"]; leaked {
		t.Fatalf("binary result leaked body excerpt: %+v", binary)
	}
}

func TestWebFetchRejectsUnsafeTargetsBeforeDial(t *testing.T) {
	executor := WebToolExecutor{Resolver: func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.2")}, nil
	}}
	for _, raw := range []string{
		"file:///etc/passwd",
		"data:text/plain,secret",
		"ftp://example.com/file",
		"https://user:pass@example.com/",
		"http://localhost:8080/",
		"http://127.0.0.1:8080/",
		"http://169.254.169.254/latest/meta-data",
		"https://example.com/private",
	} {
		_, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": raw}})
		if err == nil {
			t.Fatalf("Execute(%q) err = nil", raw)
		}
	}
}

func TestWebFetchRejectsBlockedRedirectBeforeReadingBody(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "example.com" {
			return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": []string{"http://127.0.0.1/private"}}, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
		}
		t.Fatalf("blocked target body was read")
		return nil, nil
	})}
	executor := WebToolExecutor{Client: client, Resolver: func(_ context.Context, host string) ([]net.IP, error) {
		if host == "example.com" {
			return []net.IP{net.ParseIP("93.184.216.34")}, nil
		}
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}}

	_, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": "https://example.com"}})
	if err == nil {
		t.Fatal("redirect to blocked host err = nil")
	}
}

func TestWebSearchExecutesTavilyWithSafeSummary(t *testing.T) {
	var auth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if r.Method != http.MethodPost || r.URL.Path != "/search" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"query":"latest ai news","results":[{"title":"AI News","url":"https://example.com/ai","content":"fresh public snippet","score":0.9,"raw_content":"secret raw body"}]}`))
	}))
	defer server.Close()
	executor := WebToolExecutor{TavilyAPIKey: "tvly-secret", TavilyEndpoint: server.URL + "/search"}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebSearch, ArgumentsSummary: map[string]any{"query": "latest ai news", "provider": "tavily", "limit": 3}})
	if err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer tvly-secret" {
		t.Fatalf("Authorization = %q", auth)
	}
	if result["operation"] != "search" || result["provider"] != "tavily" || result["scope"] != "web" || result["result_count"] != 1 {
		t.Fatalf("search result = %+v", result)
	}
	if strings.Contains(fmt.Sprint(result), "tvly-secret") || strings.Contains(fmt.Sprint(result), "secret raw body") {
		t.Fatalf("search result leaked sensitive data: %+v", result)
	}
}

func TestWebSearchExecutesBraveWithSafeSummary(t *testing.T) {
	var token string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = r.Header.Get("X-Subscription-Token")
		if r.Method != http.MethodGet || r.URL.Path != "/res/v1/web/search" || r.URL.Query().Get("q") != "latest ai news" || r.URL.Query().Get("count") != "2" {
			t.Fatalf("request = %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"search","web":{"results":[{"title":"Brave AI","url":"https://example.com/brave","description":"brave public snippet"}]}}`))
	}))
	defer server.Close()
	executor := WebToolExecutor{BraveAPIKey: "brave-secret", BraveEndpoint: server.URL + "/res/v1/web/search"}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebSearch, ArgumentsSummary: map[string]any{"query": "latest ai news", "provider": "brave", "limit": 2}})
	if err != nil {
		t.Fatal(err)
	}
	if token != "brave-secret" {
		t.Fatalf("X-Subscription-Token = %q", token)
	}
	if result["operation"] != "search" || result["provider"] != "brave" || result["result_count"] != 1 {
		t.Fatalf("search result = %+v", result)
	}
	if strings.Contains(fmt.Sprint(result), "brave-secret") {
		t.Fatalf("search result leaked token: %+v", result)
	}
}

func TestWebSearchRejectsMissingProviderKey(t *testing.T) {
	t.Setenv("LOOMI_TAVILY_API_KEY", "")
	t.Setenv("LOOMI_BRAVE_SEARCH_API_KEY", "")
	executor := WebToolExecutor{}
	_, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebSearch, ArgumentsSummary: map[string]any{"query": "news", "provider": "tavily"}})
	if err == nil {
		t.Fatal("missing tavily key err = nil")
	}
}

func TestWebSearchUsesProcessEnvKeyWhenExecutorWasStartedWithoutKey(t *testing.T) {
	t.Setenv("LOOMI_TAVILY_API_KEY", "tvly-env")
	var auth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"Env result","url":"https://example.com","content":"env snippet"}]}`))
	}))
	defer server.Close()
	executor := WebToolExecutor{TavilyEndpoint: server.URL + "/search"}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWebSearch, ArgumentsSummary: map[string]any{"query": "env"}})
	if err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer tvly-env" || result["provider"] != "tavily" {
		t.Fatalf("auth=%q result=%v", auth, result)
	}
}
