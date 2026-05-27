package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	defaultWebFetchMaxBytes  = 32 * 1024
	maxWebFetchMaxBytes      = 128 * 1024
	defaultWebFetchTimeoutMS = 5000
	maxWebFetchTimeoutMS     = 15000
	maxWebFetchRedirects     = 5
	maxWebFetchExcerptRunes  = 4000
	defaultWebSearchLimit    = 5
	maxWebSearchLimit        = 10
	defaultBraveEndpoint     = "https://api.search.brave.com/res/v1/web/search"
	defaultTavilyEndpoint    = "https://api.tavily.com/search"
)

var htmlTitlePattern = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
var htmlTagPattern = regexp.MustCompile(`(?is)<[^>]+>`)
var htmlWhitespacePattern = regexp.MustCompile(`\s+`)

type WebToolExecutor struct {
	Client            *http.Client
	Resolver          func(context.Context, string) ([]net.IP, error)
	AllowPrivateHosts bool
	BraveAPIKey       string
	BraveEndpoint     string
	TavilyAPIKey      string
	TavilyEndpoint    string
}

func WebToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameWebFetch, ApprovalPolicy: ToolApprovalNotRequired, SafetyClass: ToolSafetyPublicNetworkRead, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameWebSearch, ApprovalPolicy: ToolApprovalNotRequired, SafetyClass: ToolSafetyPublicNetworkRead, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

func WebSearchProviderToolDefinition() ProviderToolDefinition {
	return ProviderToolDefinition{
		Name:         productdata.ToolNameWebSearch,
		ProviderName: "web_search",
		Description:  "Search the public web through the configured Brave or Tavily provider. Use for current news, recent facts, or external information that may have changed.",
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Search query."},
				"provider": map[string]any{
					"type":        "string",
					"description": "Optional search provider. Leave unset unless the user asks for a provider.",
					"enum":        []string{"tavily", "brave"},
				},
				"limit":      map[string]any{"type": "integer", "minimum": 1, "maximum": maxWebSearchLimit},
				"timeout_ms": map[string]any{"type": "integer", "minimum": 1000, "maximum": maxWebFetchTimeoutMS},
			},
			"required": []string{"query"},
		},
	}
}

func (e WebToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	switch invocation.ToolName {
	case productdata.ToolNameWebFetch:
		return e.fetch(ctx, invocation.ArgumentsSummary)
	case productdata.ToolNameWebSearch:
		return e.search(ctx, invocation.ArgumentsSummary)
	default:
		return nil, errors.New("web tool is not supported")
	}
}

func (e WebToolExecutor) fetch(ctx context.Context, args map[string]any) (map[string]any, error) {
	target, err := normalizeWebFetchURL(args)
	if err != nil {
		return nil, err
	}
	if err := e.validateURL(ctx, target); err != nil {
		return nil, err
	}
	maxBytes := boundedInt(args, "max_bytes", defaultWebFetchMaxBytes, maxWebFetchMaxBytes)
	timeoutMS := boundedInt(args, "timeout_ms", defaultWebFetchTimeoutMS, maxWebFetchTimeoutMS)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()

	client := e.client()
	redirects := 0
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		redirects++
		if redirects > maxWebFetchRedirects {
			return errors.New("web fetch redirect limit exceeded")
		}
		return e.validateURL(req.Context(), req.URL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, errors.New("web fetch request could not be created")
	}
	req.Header.Set("User-Agent", "LoomiWebFetch/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("web fetch request failed")
	}
	defer resp.Body.Close()

	contentType := safeContentType(resp.Header.Get("Content-Type"))
	result := map[string]any{
		"tool":              productdata.ToolNameWebFetch,
		"scope":             "web",
		"operation":         "fetch",
		"url":               target.String(),
		"final_url":         resp.Request.URL.String(),
		"status_code":       resp.StatusCode,
		"content_type":      contentType,
		"bytes_read":        0,
		"byte_limit":        maxBytes,
		"truncated":         false,
		"redaction_applied": false,
	}
	if !isTextLikeContentType(contentType) {
		result["unsupported_content"] = true
		return result, nil
	}
	raw, truncated, err := readBoundedBody(resp.Body, maxBytes)
	if err != nil {
		return nil, errors.New("web fetch response could not be read")
	}
	result["bytes_read"] = len(raw)
	result["truncated"] = truncated
	text := string(raw)
	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, "")
	}
	if strings.Contains(contentType, "html") {
		if title := extractHTMLTitle(text); title != "" {
			result["title"] = title
		}
		text = htmlTagPattern.ReplaceAllString(text, " ")
	}
	result["text_excerpt"] = boundedWebExcerpt(htmlWhitespacePattern.ReplaceAllString(strings.TrimSpace(text), " "))
	return result, nil
}

func (e WebToolExecutor) client() *http.Client {
	if e.Client == nil {
		return &http.Client{Timeout: time.Duration(maxWebFetchTimeoutMS) * time.Millisecond}
	}
	copy := *e.Client
	return &copy
}

func normalizeWebFetchURL(args map[string]any) (*url.URL, error) {
	allowed := map[string]struct{}{"url": {}, "max_bytes": {}, "timeout_ms": {}}
	for key := range args {
		if _, ok := allowed[key]; !ok {
			return nil, errors.New("web fetch argument is not supported")
		}
	}
	raw := strings.TrimSpace(stringArg(args, "url", ""))
	if raw == "" {
		return nil, errors.New("web fetch url is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return nil, errors.New("web fetch url is invalid")
	}
	if parsed.User != nil {
		return nil, errors.New("web fetch url must not contain credentials")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("web fetch scheme is not supported")
	}
	parsed.Fragment = ""
	return parsed, nil
}

func (e WebToolExecutor) validateURL(ctx context.Context, target *url.URL) error {
	if target == nil || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil {
		return errors.New("web fetch url is not allowed")
	}
	host := strings.TrimSpace(target.Hostname())
	if host == "" {
		return errors.New("web fetch host is required")
	}
	if e.AllowPrivateHosts {
		return nil
	}
	if isBlockedHostLiteral(host) {
		return errors.New("web fetch host is not allowed")
	}
	ips, err := e.resolveHost(ctx, host)
	if err != nil || len(ips) == 0 {
		return errors.New("web fetch host could not be resolved")
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return errors.New("web fetch host is not allowed")
		}
	}
	return nil
}

func (e WebToolExecutor) resolveHost(ctx context.Context, host string) ([]net.IP, error) {
	if resolver := e.Resolver; resolver != nil {
		return resolver(ctx, host)
	}
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}
	return ips, nil
}

func isBlockedHostLiteral(host string) bool {
	trimmed := strings.Trim(strings.ToLower(host), "[]")
	return trimmed == "localhost" || strings.HasSuffix(trimmed, ".localhost")
}

func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified()
}

func safeContentType(value string) string {
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil || strings.TrimSpace(mediaType) == "" {
		return "application/octet-stream"
	}
	return strings.ToLower(mediaType)
}

func isTextLikeContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml") || strings.Contains(contentType, "javascript")
}

func readBoundedBody(body io.Reader, maxBytes int) ([]byte, bool, error) {
	var buf bytes.Buffer
	_, err := io.CopyN(&buf, body, int64(maxBytes)+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, false, err
	}
	raw := buf.Bytes()
	if len(raw) > maxBytes {
		return raw[:maxBytes], true, nil
	}
	return raw, false, nil
}

func extractHTMLTitle(text string) string {
	match := htmlTitlePattern.FindStringSubmatch(text)
	if len(match) != 2 {
		return ""
	}
	return boundedWebExcerpt(htmlWhitespacePattern.ReplaceAllString(htmlTagPattern.ReplaceAllString(strings.TrimSpace(match[1]), " "), " "))
}

func boundedWebExcerpt(text string) string {
	runes := []rune(text)
	if len(runes) <= maxWebFetchExcerptRunes {
		return text
	}
	return string(runes[:maxWebFetchExcerptRunes])
}

func (e WebToolExecutor) search(ctx context.Context, args map[string]any) (map[string]any, error) {
	query := strings.TrimSpace(stringArg(args, "query", ""))
	if query == "" {
		return nil, errors.New("web search query is required")
	}
	provider := strings.TrimSpace(strings.ToLower(stringArg(args, "provider", "")))
	if provider == "" {
		provider = e.defaultSearchProvider()
	}
	limit := boundedInt(args, "limit", defaultWebSearchLimit, maxWebSearchLimit)
	timeoutMS := boundedInt(args, "timeout_ms", defaultWebFetchTimeoutMS, maxWebFetchTimeoutMS)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()
	switch provider {
	case "tavily":
		return e.searchTavily(ctx, query, limit)
	case "brave":
		return e.searchBrave(ctx, query, limit)
	default:
		return nil, errors.New("web search provider is not supported")
	}
}

func (e WebToolExecutor) defaultSearchProvider() string {
	if strings.TrimSpace(e.tavilyAPIKey()) != "" {
		return "tavily"
	}
	if strings.TrimSpace(e.braveAPIKey()) != "" {
		return "brave"
	}
	return "tavily"
}

func (e WebToolExecutor) tavilyAPIKey() string {
	if strings.TrimSpace(e.TavilyAPIKey) != "" {
		return e.TavilyAPIKey
	}
	return os.Getenv("LOOMI_TAVILY_API_KEY")
}

func (e WebToolExecutor) braveAPIKey() string {
	if strings.TrimSpace(e.BraveAPIKey) != "" {
		return e.BraveAPIKey
	}
	return os.Getenv("LOOMI_BRAVE_SEARCH_API_KEY")
}

func (e WebToolExecutor) searchTavily(ctx context.Context, query string, limit int) (map[string]any, error) {
	apiKey := strings.TrimSpace(e.tavilyAPIKey())
	if apiKey == "" {
		return nil, errors.New("tavily search api key is not configured")
	}
	body := map[string]any{
		"query":               query,
		"max_results":         limit,
		"search_depth":        "basic",
		"include_answer":      false,
		"include_raw_content": false,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, errors.New("tavily search request could not be created")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, firstNonEmptyWebEndpoint(e.TavilyEndpoint, defaultTavilyEndpoint), bytes.NewReader(raw))
	if err != nil {
		return nil, errors.New("tavily search request could not be created")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := e.client().Do(req)
	if err != nil {
		return nil, errors.New("tavily search request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("tavily search request failed with HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, int64(maxWebFetchMaxBytes))).Decode(&payload); err != nil {
		return nil, errors.New("tavily search response could not be parsed")
	}
	items := make([]map[string]any, 0, len(payload.Results))
	for _, item := range payload.Results {
		items = append(items, safeSearchItem(item.Title, item.URL, item.Content))
		if len(items) >= limit {
			break
		}
	}
	return webSearchResult("tavily", query, items), nil
}

func (e WebToolExecutor) searchBrave(ctx context.Context, query string, limit int) (map[string]any, error) {
	apiKey := strings.TrimSpace(e.braveAPIKey())
	if apiKey == "" {
		return nil, errors.New("brave search api key is not configured")
	}
	endpoint := firstNonEmptyWebEndpoint(e.BraveEndpoint, defaultBraveEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.New("brave search request could not be created")
	}
	values := req.URL.Query()
	values.Set("q", query)
	values.Set("count", strconv.Itoa(limit))
	req.URL.RawQuery = values.Encode()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)
	resp, err := e.client().Do(req)
	if err != nil {
		return nil, errors.New("brave search request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("brave search request failed with HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, int64(maxWebFetchMaxBytes))).Decode(&payload); err != nil {
		return nil, errors.New("brave search response could not be parsed")
	}
	items := make([]map[string]any, 0, len(payload.Web.Results))
	for _, item := range payload.Web.Results {
		items = append(items, safeSearchItem(item.Title, item.URL, item.Description))
		if len(items) >= limit {
			break
		}
	}
	return webSearchResult("brave", query, items), nil
}

func safeSearchItem(title string, itemURL string, snippet string) map[string]any {
	return map[string]any{
		"title":   boundedWebExcerpt(htmlWhitespacePattern.ReplaceAllString(strings.TrimSpace(title), " ")),
		"url":     strings.TrimSpace(itemURL),
		"snippet": boundedWebExcerpt(htmlWhitespacePattern.ReplaceAllString(strings.TrimSpace(snippet), " ")),
	}
}

func webSearchResult(provider string, query string, items []map[string]any) map[string]any {
	return map[string]any{
		"tool":              productdata.ToolNameWebSearch,
		"scope":             "web",
		"operation":         "search",
		"provider":          provider,
		"query":             query,
		"result_count":      len(items),
		"items":             items,
		"redaction_applied": false,
	}
}

func firstNonEmptyWebEndpoint(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}
