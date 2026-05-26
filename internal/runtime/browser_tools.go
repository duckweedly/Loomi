package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"html"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	defaultBrowserMaxBytes = 32 * 1024
	maxBrowserMaxBytes     = 128 * 1024
	maxBrowserLinks        = 20
)

var browserLinkPattern = regexp.MustCompile(`(?is)<a\s+[^>]*href\s*=\s*["']([^"']+)["'][^>]*>(.*?)</a>`)

type BrowserToolExecutor struct {
	Store             *BrowserSessionStore
	Client            *http.Client
	Resolver          func(context.Context, string) ([]net.IP, error)
	AllowPrivateHosts bool
}

type BrowserSessionStore struct {
	mu       sync.Mutex
	sessions map[string]browserSession
}

type browserSession struct {
	RunID       string
	SessionID   string
	URL         string
	FinalURL    string
	StatusCode  int
	ContentType string
	Title       string
	TextExcerpt string
	Links       []map[string]any
	BytesRead   int
	ByteLimit   int
	Truncated   bool
	UpdatedAt   time.Time
}

func NewBrowserSessionStore() *BrowserSessionStore {
	return &BrowserSessionStore{sessions: map[string]browserSession{}}
}

func BrowserToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameBrowserOpen, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyPublicNetworkRead, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameBrowserSnapshot, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyPublicNetworkRead, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameBrowserClickLink, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyPublicNetworkRead, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

func (e BrowserToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	switch invocation.ToolName {
	case productdata.ToolNameBrowserOpen:
		return e.open(ctx, invocation)
	case productdata.ToolNameBrowserSnapshot:
		return e.snapshot(invocation)
	case productdata.ToolNameBrowserClickLink:
		return e.clickLink(ctx, invocation)
	default:
		return nil, errors.New("browser tool is not supported")
	}
}

func (e BrowserToolExecutor) open(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	target, err := normalizeWebFetchURL(invocation.ArgumentsSummary)
	if err != nil {
		return nil, err
	}
	sessionID := browserSessionID(invocation.RunID, invocation.ToolCallID, target.String())
	session, err := e.navigate(ctx, invocation, sessionID, target, "")
	if err != nil {
		return nil, err
	}
	e.store().save(session)
	return browserSessionResult(productdata.ToolNameBrowserOpen, "open", session, ""), nil
}

func (e BrowserToolExecutor) snapshot(invocation ToolInvocation) (map[string]any, error) {
	sessionID := strings.TrimSpace(stringArg(invocation.ArgumentsSummary, "session_id", ""))
	session, ok := e.store().get(invocation.RunID, sessionID)
	if !ok {
		return nil, errors.New("browser session is unavailable")
	}
	return browserSessionResult(productdata.ToolNameBrowserSnapshot, "snapshot", session, ""), nil
}

func (e BrowserToolExecutor) clickLink(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	sessionID := strings.TrimSpace(stringArg(invocation.ArgumentsSummary, "session_id", ""))
	session, ok := e.store().get(invocation.RunID, sessionID)
	if !ok {
		return nil, errors.New("browser session is unavailable")
	}
	index := browserLinkIndex(invocation.ArgumentsSummary, "link_index")
	if index < 0 || index >= len(session.Links) {
		return nil, errors.New("browser link index is unavailable")
	}
	if blocked, _ := session.Links[index]["blocked"].(bool); blocked {
		return nil, errors.New("browser link target is not allowed")
	}
	href, _ := session.Links[index]["href"].(string)
	target, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return nil, errors.New("browser link target is invalid")
	}
	previousURL := session.FinalURL
	next, err := e.navigate(ctx, invocation, session.SessionID, target, previousURL)
	if err != nil {
		return nil, err
	}
	e.store().save(next)
	return browserSessionResult(productdata.ToolNameBrowserClickLink, "click_link", next, previousURL), nil
}

func (e BrowserToolExecutor) navigate(ctx context.Context, invocation ToolInvocation, sessionID string, target *url.URL, previousURL string) (browserSession, error) {
	web := WebToolExecutor{Client: e.Client, Resolver: e.Resolver, AllowPrivateHosts: e.AllowPrivateHosts}
	if err := web.validateURL(ctx, target); err != nil {
		return browserSession{}, err
	}
	maxBytes := boundedInt(invocation.ArgumentsSummary, "max_bytes", defaultBrowserMaxBytes, maxBrowserMaxBytes)
	timeoutMS := boundedInt(invocation.ArgumentsSummary, "timeout_ms", defaultWebFetchTimeoutMS, maxWebFetchTimeoutMS)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()
	client := web.client()
	redirects := 0
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		redirects++
		if redirects > maxWebFetchRedirects {
			return errors.New("browser redirect limit exceeded")
		}
		return web.validateURL(req.Context(), req.URL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return browserSession{}, errors.New("browser request could not be created")
	}
	req.Header.Set("User-Agent", "LoomiBrowserFoundation/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return browserSession{}, errors.New("browser navigation failed")
	}
	defer resp.Body.Close()
	contentType := safeContentType(resp.Header.Get("Content-Type"))
	raw, truncated, err := readBoundedBody(resp.Body, maxBytes)
	if err != nil {
		return browserSession{}, errors.New("browser response could not be read")
	}
	text := strings.ToValidUTF8(string(raw), "")
	title := ""
	links := []map[string]any{}
	if isTextLikeContentType(contentType) {
		if strings.Contains(contentType, "html") {
			title = extractHTMLTitle(text)
			links = e.extractLinks(ctx, resp.Request.URL, text)
			text = htmlTagPattern.ReplaceAllString(text, " ")
		}
		text = boundedWebExcerpt(htmlWhitespacePattern.ReplaceAllString(strings.TrimSpace(text), " "))
	} else {
		text = ""
	}
	_ = previousURL
	return browserSession{
		RunID:       invocation.RunID,
		SessionID:   sessionID,
		URL:         target.String(),
		FinalURL:    resp.Request.URL.String(),
		StatusCode:  resp.StatusCode,
		ContentType: contentType,
		Title:       title,
		TextExcerpt: text,
		Links:       links,
		BytesRead:   min(len(raw), maxBytes),
		ByteLimit:   maxBytes,
		Truncated:   truncated,
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

func (e BrowserToolExecutor) extractLinks(ctx context.Context, base *url.URL, document string) []map[string]any {
	matches := browserLinkPattern.FindAllStringSubmatch(document, -1)
	links := make([]map[string]any, 0, min(len(matches), maxBrowserLinks))
	web := WebToolExecutor{Client: e.Client, Resolver: e.Resolver, AllowPrivateHosts: e.AllowPrivateHosts}
	for _, match := range matches {
		if len(links) >= maxBrowserLinks {
			break
		}
		rawHref := html.UnescapeString(strings.TrimSpace(match[1]))
		target, err := base.Parse(rawHref)
		blocked := err != nil
		href := ""
		host := ""
		if err == nil {
			target.Fragment = ""
			href = target.String()
			host = target.Hostname()
			if err := web.validateURL(ctx, target); err != nil {
				blocked = true
			}
		}
		links = append(links, map[string]any{
			"blocked": blocked,
			"host":    host,
			"href":    href,
			"index":   len(links),
			"text":    boundedWebExcerpt(htmlWhitespacePattern.ReplaceAllString(htmlTagPattern.ReplaceAllString(html.UnescapeString(strings.TrimSpace(match[2])), " "), " ")),
		})
	}
	return links
}

func (s *BrowserSessionStore) save(session browserSession) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.SessionID] = session
}

func (s *BrowserSessionStore) get(runID string, sessionID string) (browserSession, bool) {
	if s == nil {
		return browserSession{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[strings.TrimSpace(sessionID)]
	if !ok || session.RunID != runID {
		return browserSession{}, false
	}
	return session, true
}

var defaultBrowserSessionStore = NewBrowserSessionStore()

func (e BrowserToolExecutor) store() *BrowserSessionStore {
	if e.Store != nil {
		return e.Store
	}
	return defaultBrowserSessionStore
}

func browserSessionResult(toolName string, operation string, session browserSession, previousURL string) map[string]any {
	result := map[string]any{
		"byte_limit":        session.ByteLimit,
		"bytes_read":        session.BytesRead,
		"content_type":      session.ContentType,
		"final_url":         session.FinalURL,
		"links":             session.Links,
		"operation":         operation,
		"redaction_applied": false,
		"scope":             "browser",
		"session_id":        session.SessionID,
		"status_code":       session.StatusCode,
		"text_excerpt":      session.TextExcerpt,
		"title":             session.Title,
		"tool":              toolName,
		"truncated":         session.Truncated,
		"url":               session.URL,
	}
	if previousURL != "" {
		result["previous_url"] = previousURL
	}
	return result
}

func browserSessionID(runID string, toolCallID string, rawURL string) string {
	sum := sha256.Sum256([]byte(runID + "\x00" + toolCallID + "\x00" + rawURL))
	return "br_" + hex.EncodeToString(sum[:8])
}

func browserLinkIndex(args map[string]any, key string) int {
	value, ok := args[key]
	if !ok || value == nil {
		return -1
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return -1
	}
}
