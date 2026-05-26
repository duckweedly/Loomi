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
var browserInputPattern = regexp.MustCompile(`(?is)<(?:input|textarea)\s+([^>]*)>`)
var browserAttrPattern = regexp.MustCompile(`(?is)(name|id|placeholder|type|aria-label)\s*=\s*["']([^"']*)["']`)

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
	Inputs      []map[string]any
	FormValues  map[string]string
	LastKey     string
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
		{Name: productdata.ToolNameBrowserScreenshot, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyPublicNetworkRead, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameBrowserType, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyPublicNetworkRead, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameBrowserPress, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyPublicNetworkRead, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
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
	case productdata.ToolNameBrowserScreenshot:
		return e.screenshot(invocation)
	case productdata.ToolNameBrowserType:
		return e.typeText(invocation)
	case productdata.ToolNameBrowserPress:
		return e.press(invocation)
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

func (e BrowserToolExecutor) screenshot(invocation ToolInvocation) (map[string]any, error) {
	sessionID := strings.TrimSpace(stringArg(invocation.ArgumentsSummary, "session_id", ""))
	session, ok := e.store().get(invocation.RunID, sessionID)
	if !ok {
		return nil, errors.New("browser session is unavailable")
	}
	result := browserSessionResult(productdata.ToolNameBrowserScreenshot, "screenshot", session, "")
	result["format"] = "text"
	result["screenshot_text"] = boundedWebExcerpt(strings.TrimSpace(session.Title + "\n" + session.TextExcerpt))
	return result, nil
}

func (e BrowserToolExecutor) typeText(invocation ToolInvocation) (map[string]any, error) {
	sessionID := strings.TrimSpace(stringArg(invocation.ArgumentsSummary, "session_id", ""))
	session, ok := e.store().get(invocation.RunID, sessionID)
	if !ok {
		return nil, errors.New("browser session is unavailable")
	}
	target := strings.TrimSpace(stringArg(invocation.ArgumentsSummary, "target", ""))
	text := boundedWebExcerpt(stringArg(invocation.ArgumentsSummary, "text", ""))
	if target == "" || text == "" {
		return nil, errors.New("browser type target and text are required")
	}
	if !browserInputTargetAllowed(session, target) {
		return nil, errors.New("browser input target is unavailable")
	}
	if session.FormValues == nil {
		session.FormValues = map[string]string{}
	}
	session.FormValues[target] = text
	e.store().save(session)
	result := browserSessionResult(productdata.ToolNameBrowserType, "type", session, "")
	result["target"] = target
	result["text_length"] = len([]rune(text))
	return result, nil
}

func (e BrowserToolExecutor) press(invocation ToolInvocation) (map[string]any, error) {
	sessionID := strings.TrimSpace(stringArg(invocation.ArgumentsSummary, "session_id", ""))
	session, ok := e.store().get(invocation.RunID, sessionID)
	if !ok {
		return nil, errors.New("browser session is unavailable")
	}
	key := strings.TrimSpace(stringArg(invocation.ArgumentsSummary, "key", ""))
	if !browserKeyAllowed(key) {
		return nil, errors.New("browser key is not allowed")
	}
	session.LastKey = key
	e.store().save(session)
	result := browserSessionResult(productdata.ToolNameBrowserPress, "press", session, "")
	result["key"] = key
	result["submitted"] = key == "Enter"
	return result, nil
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
	inputs := []map[string]any{}
	if isTextLikeContentType(contentType) {
		if strings.Contains(contentType, "html") {
			title = extractHTMLTitle(text)
			links = e.extractLinks(ctx, resp.Request.URL, text)
			inputs = extractBrowserInputs(text)
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
		Inputs:      inputs,
		FormValues:  map[string]string{},
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
		"inputs":            session.Inputs,
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
	if len(session.FormValues) > 0 {
		result["form_value_count"] = len(session.FormValues)
	}
	if session.LastKey != "" {
		result["last_key"] = session.LastKey
	}
	if previousURL != "" {
		result["previous_url"] = previousURL
	}
	return result
}

func extractBrowserInputs(document string) []map[string]any {
	matches := browserInputPattern.FindAllStringSubmatch(document, -1)
	inputs := make([]map[string]any, 0, min(len(matches), 20))
	for _, match := range matches {
		attrs := map[string]string{}
		for _, attr := range browserAttrPattern.FindAllStringSubmatch(match[1], -1) {
			if len(attr) == 3 {
				attrs[strings.ToLower(attr[1])] = html.UnescapeString(strings.TrimSpace(attr[2]))
			}
		}
		target := attrs["name"]
		if target == "" {
			target = attrs["id"]
		}
		if target == "" {
			target = attrs["placeholder"]
		}
		if target == "" {
			target = attrs["aria-label"]
		}
		if strings.TrimSpace(target) == "" {
			continue
		}
		inputs = append(inputs, map[string]any{"index": len(inputs), "target": boundedWebExcerpt(target), "type": boundedWebExcerpt(attrs["type"]), "label": boundedWebExcerpt(firstBrowserNonEmpty(attrs["aria-label"], attrs["placeholder"], target))})
	}
	return inputs
}

func browserInputTargetAllowed(session browserSession, target string) bool {
	for _, input := range session.Inputs {
		if value, _ := input["target"].(string); value == target {
			return true
		}
	}
	return false
}

func browserKeyAllowed(key string) bool {
	switch key {
	case "Enter", "Escape", "Tab", "ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight":
		return true
	default:
		return false
	}
}

func firstBrowserNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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
