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
