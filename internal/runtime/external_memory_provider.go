package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type memoryProviderConfigReader interface {
	GetMemoryProviderConfig(context.Context, identity.LocalIdentity) (productdata.MemoryProviderConfig, error)
}

type externalMemoryHit struct {
	URI         string
	Title       string
	Summary     string
	Score       float64
	RankReason  string
	SourceType  string
	Provider    productdata.MemoryProviderID
	RedactedRaw bool
}

func (e MemoryToolExecutor) externalMemorySearch(ctx context.Context, query string, limit int) ([]externalMemoryHit, bool, error) {
	reader, ok := e.Service.(memoryProviderConfigReader)
	if !ok {
		return nil, false, nil
	}
	config, err := reader.GetMemoryProviderConfig(ctx, e.ident())
	if err != nil {
		return nil, false, err
	}
	status := productdata.MemoryProviderStatus{}
	if e.Service != nil {
		status, _ = e.Service.GetMemoryProviderStatus(ctx, e.ident())
	}
	if !status.Enabled || !status.Configured {
		return nil, false, nil
	}
	switch status.Provider {
	case productdata.MemoryProviderOpenViking:
		hits, err := searchOpenVikingMemory(ctx, config.OpenViking, e.ident(), query, limit)
		return hits, true, err
	case productdata.MemoryProviderNowledge:
		hits, err := searchNowledgeMemory(ctx, config.Nowledge, e.ident(), query, limit)
		return hits, true, err
	default:
		return nil, false, nil
	}
}

func (e MemoryToolExecutor) externalMemoryRead(ctx context.Context, uri string) (map[string]any, bool, error) {
	reader, ok := e.Service.(memoryProviderConfigReader)
	if !ok {
		return nil, false, nil
	}
	config, err := reader.GetMemoryProviderConfig(ctx, e.ident())
	if err != nil {
		return nil, false, err
	}
	status, err := e.Service.GetMemoryProviderStatus(ctx, e.ident())
	if err != nil || !status.Enabled || !status.Configured {
		return nil, false, err
	}
	switch {
	case status.Provider == productdata.MemoryProviderOpenViking && strings.HasPrefix(uri, "viking://"):
		text, err := readOpenVikingMemory(ctx, config.OpenViking, e.ident(), uri)
		if err != nil {
			return nil, true, err
		}
		return externalReadResult(productdata.MemoryProviderOpenViking, uri, text), true, nil
	case status.Provider == productdata.MemoryProviderNowledge && strings.HasPrefix(uri, "nowledge://memory/"):
		text, err := readNowledgeMemory(ctx, config.Nowledge, e.ident(), uri)
		if err != nil {
			return nil, true, err
		}
		return externalReadResult(productdata.MemoryProviderNowledge, uri, text), true, nil
	default:
		return nil, false, nil
	}
}

func (e MemoryToolExecutor) externalMemoryWrite(ctx context.Context, title string, content string) (map[string]any, bool, error) {
	reader, ok := e.Service.(memoryProviderConfigReader)
	if !ok {
		return nil, false, nil
	}
	config, err := reader.GetMemoryProviderConfig(ctx, e.ident())
	if err != nil {
		return nil, false, err
	}
	status, err := e.Service.GetMemoryProviderStatus(ctx, e.ident())
	if err != nil || !status.Enabled || !status.Configured {
		return nil, false, err
	}
	switch status.Provider {
	case productdata.MemoryProviderOpenViking:
		if err := writeOpenVikingMemory(ctx, config.OpenViking, e.ident(), title, content); err != nil {
			return nil, true, err
		}
		return map[string]any{"tool": productdata.ToolNameMemoryWrite, "scope": "memory", "operation": "write_provider", "provider": string(productdata.MemoryProviderOpenViking), "status": "accepted", "title": productdata.RedactEventText(title), "content": nil, "redaction_applied": true}, true, nil
	case productdata.MemoryProviderNowledge:
		uri, err := writeNowledgeMemory(ctx, config.Nowledge, e.ident(), title, content)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"tool": productdata.ToolNameMemoryWrite, "scope": "memory", "operation": "write_provider", "provider": string(productdata.MemoryProviderNowledge), "entry_id": uri, "status": "accepted", "title": productdata.RedactEventText(title), "content": nil, "redaction_applied": true}, true, nil
	default:
		return nil, false, nil
	}
}

func (e MemoryToolExecutor) externalMemoryEdit(ctx context.Context, uri string, title string, content string) (map[string]any, bool, error) {
	reader, ok := e.Service.(memoryProviderConfigReader)
	if !ok {
		return nil, false, nil
	}
	config, err := reader.GetMemoryProviderConfig(ctx, e.ident())
	if err != nil {
		return nil, false, err
	}
	status, err := e.Service.GetMemoryProviderStatus(ctx, e.ident())
	if err != nil || !status.Enabled || !status.Configured {
		return nil, false, err
	}
	if status.Provider != productdata.MemoryProviderOpenViking || !strings.HasPrefix(uri, "viking://") {
		return nil, false, nil
	}
	if err := editOpenVikingMemory(ctx, config.OpenViking, e.ident(), uri, title, content); err != nil {
		return nil, true, err
	}
	return map[string]any{"tool": productdata.ToolNameMemoryEdit, "scope": "memory", "operation": "edit_provider", "provider": string(productdata.MemoryProviderOpenViking), "entry_id": uri, "status": "accepted", "title": productdata.RedactEventText(title), "content": nil, "redaction_applied": true}, true, nil
}

func (e MemoryToolExecutor) externalMemoryForget(ctx context.Context, uri string) (map[string]any, bool, error) {
	reader, ok := e.Service.(memoryProviderConfigReader)
	if !ok {
		return nil, false, nil
	}
	config, err := reader.GetMemoryProviderConfig(ctx, e.ident())
	if err != nil {
		return nil, false, err
	}
	status, err := e.Service.GetMemoryProviderStatus(ctx, e.ident())
	if err != nil || !status.Enabled || !status.Configured {
		return nil, false, err
	}
	switch {
	case status.Provider == productdata.MemoryProviderOpenViking && strings.HasPrefix(uri, "viking://"):
		if err := deleteOpenVikingMemory(ctx, config.OpenViking, e.ident(), uri); err != nil {
			return nil, true, err
		}
		return externalForgetResult(productdata.MemoryProviderOpenViking, uri), true, nil
	case status.Provider == productdata.MemoryProviderNowledge && strings.HasPrefix(uri, "nowledge://memory/"):
		if err := deleteNowledgeMemory(ctx, config.Nowledge, e.ident(), uri); err != nil {
			return nil, true, err
		}
		return externalForgetResult(productdata.MemoryProviderNowledge, uri), true, nil
	default:
		return nil, false, nil
	}
}

func externalForgetResult(provider productdata.MemoryProviderID, uri string) map[string]any {
	return map[string]any{"tool": productdata.ToolNameMemoryForget, "scope": "memory", "operation": "forget_provider", "provider": string(provider), "entry_id": uri, "status": "deleted", "redaction_applied": true}
}

func (e MemoryToolExecutor) externalOpenVikingConnections(ctx context.Context, uri string, limit int) (map[string]any, bool, error) {
	reader, ok := e.Service.(memoryProviderConfigReader)
	if !ok {
		return nil, false, nil
	}
	uri = strings.TrimSpace(uri)
	if !strings.HasPrefix(uri, "viking://") {
		return nil, false, nil
	}
	config, err := reader.GetMemoryProviderConfig(ctx, e.ident())
	if err != nil {
		return nil, false, err
	}
	status, err := e.Service.GetMemoryProviderStatus(ctx, e.ident())
	if err != nil || !status.Enabled || !status.Configured {
		return nil, false, err
	}
	if status.Provider != productdata.MemoryProviderOpenViking {
		return nil, false, nil
	}
	children, err := listOpenVikingMemoryDir(ctx, config.OpenViking, e.ident(), uri)
	if err != nil {
		return nil, true, err
	}
	maxItems := boundedLimit(limit)
	items := make([]map[string]any, 0, len(children))
	for _, child := range children {
		if len(items) >= maxItems {
			break
		}
		items = append(items, map[string]any{
			"entry_id":          productdata.RedactEventText(child.URI),
			"title":             productdata.RedactEventText(openVikingURITitle(child.URI)),
			"node_type":         child.Type,
			"relation":          "child",
			"redaction_applied": true,
		})
	}
	return map[string]any{"tool": productdata.ToolNameMemoryConnections, "scope": "memory", "operation": "connections", "provider": string(productdata.MemoryProviderOpenViking), "target_entry_id": uri, "items": items, "count": len(items), "redaction_applied": true}, true, nil
}

func (e MemoryToolExecutor) externalNowledgeConnections(ctx context.Context, uri string, limit int) (map[string]any, bool, error) {
	config, ok, err := e.nowledgeConfig(ctx)
	if !ok || err != nil {
		return nil, ok, err
	}
	memoryID := strings.TrimPrefix(strings.TrimSpace(uri), "nowledge://memory/")
	if memoryID == "" || memoryID == strings.TrimSpace(uri) {
		return nil, false, nil
	}
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return nil, true, err
	}
	var response struct {
		Neighbors []struct {
			ID          string `json:"id"`
			Label       string `json:"label"`
			Title       string `json:"title"`
			NodeType    string `json:"node_type"`
			Type        string `json:"type"`
			Content     string `json:"content"`
			Description string `json:"description"`
			Summary     string `json:"summary"`
		} `json:"neighbors"`
		Edges []struct {
			Source   string  `json:"source"`
			Target   string  `json:"target"`
			EdgeType string  `json:"edge_type"`
			Type     string  `json:"type"`
			Weight   float64 `json:"weight"`
			Label    string  `json:"label"`
		} `json:"edges"`
	}
	path := base + "/graph/expand/" + url.PathEscape(memoryID) + "?depth=1&limit=" + fmt.Sprintf("%d", boundedLimit(limit))
	if err := doProviderJSON(ctx, http.MethodGet, path, nil, nowledgeHeaders(config.APIKey, e.ident()), &response); err != nil {
		return nil, true, err
	}
	nodes := map[string]map[string]any{}
	for _, node := range response.Neighbors {
		nodes[node.ID] = map[string]any{"node_id": node.ID, "title": productdata.RedactEventText(firstNonEmpty(node.Label, node.Title)), "node_type": productdata.RedactEventText(firstNonEmpty(node.NodeType, node.Type, "memory")), "summary": productdata.RedactEventText(firstNonEmpty(node.Summary, node.Description, node.Content)), "redaction_applied": true}
	}
	items := []map[string]any{}
	for _, edge := range response.Edges {
		neighborID := edge.Target
		if neighborID == memoryID {
			neighborID = edge.Source
		}
		node, ok := nodes[neighborID]
		if !ok {
			continue
		}
		node["edge_type"] = productdata.RedactEventText(firstNonEmpty(edge.EdgeType, edge.Type, "RELATED"))
		node["relation"] = productdata.RedactEventText(edge.Label)
		node["weight"] = edge.Weight
		items = append(items, node)
	}
	return map[string]any{"tool": productdata.ToolNameMemoryConnections, "scope": "memory", "operation": "connections", "provider": string(productdata.MemoryProviderNowledge), "target_entry_id": uri, "items": items, "count": len(items), "redaction_applied": true}, true, nil
}

func (e MemoryToolExecutor) externalNowledgeTimeline(ctx context.Context, limit int) (map[string]any, bool, error) {
	config, ok, err := e.nowledgeConfig(ctx)
	if !ok || err != nil {
		return nil, ok, err
	}
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return nil, true, err
	}
	values := url.Values{}
	values.Set("last_n_days", "7")
	values.Set("limit", fmt.Sprintf("%d", boundedLimit(limit)))
	var response struct {
		Events []struct {
			ID          string `json:"id"`
			EventType   string `json:"event_type"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Content     string `json:"content"`
			CreatedAt   string `json:"created_at"`
			Timestamp   string `json:"timestamp"`
			MemoryID    string `json:"memory_id"`
		} `json:"events"`
	}
	if err := doProviderJSON(ctx, http.MethodGet, base+"/agent/feed/events?"+values.Encode(), nil, nowledgeHeaders(config.APIKey, e.ident()), &response); err != nil {
		return nil, true, err
	}
	items := make([]map[string]any, 0, len(response.Events))
	for _, item := range response.Events {
		items = append(items, map[string]any{"audit_id": productdata.RedactEventText(item.ID), "event_type": productdata.RedactEventText(item.EventType), "summary": productdata.RedactEventText(firstNonEmpty(item.Title, item.Description, item.Content)), "memory_entry_id": productdata.RedactEventText(item.MemoryID), "source_type": "nowledge", "occurred_at": productdata.RedactEventText(firstNonEmpty(item.CreatedAt, item.Timestamp)), "redaction_applied": true})
	}
	return map[string]any{"tool": productdata.ToolNameMemoryTimeline, "scope": "memory", "operation": "timeline", "provider": string(productdata.MemoryProviderNowledge), "items": items, "count": len(items), "redaction_applied": true}, true, nil
}

func (e MemoryToolExecutor) externalNowledgeThreadSearch(ctx context.Context, query string, limit int) (map[string]any, bool, error) {
	config, ok, err := e.nowledgeConfig(ctx)
	if !ok || err != nil {
		return nil, ok, err
	}
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return nil, true, err
	}
	values := url.Values{}
	values.Set("query", strings.TrimSpace(query))
	values.Set("mode", "full")
	values.Set("limit", fmt.Sprintf("%d", boundedLimit(limit)))
	var response struct {
		Threads []struct {
			ID           string   `json:"id"`
			ThreadID     string   `json:"thread_id"`
			Title        string   `json:"title"`
			Source       string   `json:"source"`
			MessageCount int      `json:"message_count"`
			Score        float64  `json:"score"`
			Snippets     []string `json:"snippets"`
		} `json:"threads"`
	}
	if err := doProviderJSON(ctx, http.MethodGet, base+"/threads/search?"+values.Encode(), nil, nowledgeHeaders(config.APIKey, e.ident()), &response); err != nil {
		return nil, true, err
	}
	items := make([]map[string]any, 0, len(response.Threads))
	for _, thread := range response.Threads {
		items = append(items, map[string]any{"thread_id": firstNonEmpty(thread.ThreadID, thread.ID), "title": productdata.RedactEventText(thread.Title), "source": productdata.RedactEventText(thread.Source), "message_count": thread.MessageCount, "score": thread.Score, "excerpt": productdata.RedactEventText(firstNonEmpty(thread.Snippets...)), "redaction_applied": true})
	}
	return map[string]any{"tool": productdata.ToolNameMemoryThreadSearch, "scope": "memory", "operation": "thread_search", "provider": string(productdata.MemoryProviderNowledge), "items": items, "count": len(items), "redaction_applied": true}, true, nil
}

func (e MemoryToolExecutor) externalNowledgeThreadFetch(ctx context.Context, threadID string, limit int) (map[string]any, bool, error) {
	config, ok, err := e.nowledgeConfig(ctx)
	if !ok || err != nil {
		return nil, ok, err
	}
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return nil, true, err
	}
	values := url.Values{}
	values.Set("limit", fmt.Sprintf("%d", boundedLimit(limit)))
	var response struct {
		ID           string `json:"id"`
		ThreadID     string `json:"thread_id"`
		Title        string `json:"title"`
		MessageCount int    `json:"message_count"`
		Messages     []struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			Timestamp string `json:"timestamp"`
		} `json:"messages"`
	}
	if err := doProviderJSON(ctx, http.MethodGet, base+"/threads/"+url.PathEscape(strings.TrimSpace(threadID))+"?"+values.Encode(), nil, nowledgeHeaders(config.APIKey, e.ident()), &response); err != nil {
		return nil, true, err
	}
	items := make([]map[string]any, 0, len(response.Messages))
	for _, message := range response.Messages {
		items = append(items, map[string]any{"role": productdata.RedactEventText(message.Role), "excerpt": productdata.RedactEventText(safeMemoryExcerpt(message.Content)), "created_at": productdata.RedactEventText(message.Timestamp), "redaction_applied": true})
	}
	return map[string]any{"tool": productdata.ToolNameMemoryThreadFetch, "scope": "memory", "operation": "thread_fetch", "provider": string(productdata.MemoryProviderNowledge), "thread_id": firstNonEmpty(response.ThreadID, response.ID, threadID), "title": productdata.RedactEventText(response.Title), "items": items, "count": len(items), "message_count": response.MessageCount, "redaction_applied": true}, true, nil
}

func (e MemoryToolExecutor) nowledgeConfig(ctx context.Context) (productdata.NowledgeMemoryConfig, bool, error) {
	reader, ok := e.Service.(memoryProviderConfigReader)
	if !ok {
		return productdata.NowledgeMemoryConfig{}, false, nil
	}
	config, err := reader.GetMemoryProviderConfig(ctx, e.ident())
	if err != nil {
		return productdata.NowledgeMemoryConfig{}, false, err
	}
	status, err := e.Service.GetMemoryProviderStatus(ctx, e.ident())
	if err != nil || !status.Enabled || !status.Configured || status.Provider != productdata.MemoryProviderNowledge {
		return productdata.NowledgeMemoryConfig{}, false, err
	}
	return config.Nowledge, true, nil
}

func externalReadResult(provider productdata.MemoryProviderID, uri string, text string) map[string]any {
	title, summary := splitExternalMemoryText(productdata.RedactEventText(text))
	return map[string]any{"tool": productdata.ToolNameMemoryRead, "scope": "memory", "operation": "read", "provider": string(provider), "entry_id": uri, "title": title, "summary": summary, "content": nil, "redaction_applied": true}
}

func externalSearchItems(hits []externalMemoryHit) []map[string]any {
	items := make([]map[string]any, 0, len(hits))
	for _, hit := range hits {
		items = append(items, map[string]any{
			"id":                hit.URI,
			"entry_id":          hit.URI,
			"title":             hit.Title,
			"summary":           hit.Summary,
			"source_type":       hit.SourceType,
			"rank_reason":       hit.RankReason,
			"score":             hit.Score,
			"provider":          string(hit.Provider),
			"content":           nil,
			"redaction_applied": true,
		})
	}
	return items
}

func searchOpenVikingMemory(ctx context.Context, config productdata.OpenVikingMemoryConfig, ident identity.LocalIdentity, query string, limit int) ([]externalMemoryHit, error) {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return nil, err
	}
	body := map[string]any{"query": strings.TrimSpace(query), "limit": boundedLimit(limit)}
	var response struct {
		Result struct {
			Memories  []openVikingMatchedContext `json:"memories"`
			Resources []openVikingMatchedContext `json:"resources"`
			Skills    []openVikingMatchedContext `json:"skills"`
		} `json:"result"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	if err := doProviderJSON(ctx, http.MethodPost, base+"/api/v1/search/find", body, openVikingHeaders(config.RootAPIKey, ident), &response); err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("openviking search: %s", productdata.RedactEventText(response.Error.Code))
	}
	contexts := append([]openVikingMatchedContext{}, response.Result.Memories...)
	contexts = append(contexts, response.Result.Resources...)
	contexts = append(contexts, response.Result.Skills...)
	hits := make([]externalMemoryHit, 0, len(contexts))
	for _, item := range contexts {
		title, summary := splitExternalMemoryText(item.Abstract)
		hits = append(hits, externalMemoryHit{URI: strings.TrimSpace(item.URI), Title: title, Summary: summary, Score: item.Score, RankReason: productdata.RedactEventText(item.MatchReason), SourceType: "openviking", Provider: productdata.MemoryProviderOpenViking, RedactedRaw: true})
	}
	return hits, nil
}

type openVikingMatchedContext struct {
	URI         string  `json:"uri"`
	Abstract    string  `json:"abstract"`
	Score       float64 `json:"score"`
	MatchReason string  `json:"match_reason"`
}

func readOpenVikingMemory(ctx context.Context, config productdata.OpenVikingMemoryConfig, ident identity.LocalIdentity, uri string) (string, error) {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return "", err
	}
	var response struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code string `json:"code"`
		} `json:"error,omitempty"`
	}
	path := base + "/api/v1/content/read?uri=" + url.QueryEscape(strings.TrimSpace(uri))
	if err := doProviderJSON(ctx, http.MethodGet, path, nil, openVikingHeaders(config.RootAPIKey, ident), &response); err != nil {
		return "", err
	}
	if response.Error != nil {
		return "", fmt.Errorf("openviking read: %s", productdata.RedactEventText(response.Error.Code))
	}
	var text string
	if err := json.Unmarshal(response.Result, &text); err == nil {
		return text, nil
	}
	return string(response.Result), nil
}

func writeOpenVikingMemory(ctx context.Context, config productdata.OpenVikingMemoryConfig, ident identity.LocalIdentity, title string, content string) error {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return err
	}
	text := titledContent(title, content)
	if strings.TrimSpace(text) == "" {
		return errors.New("memory content is empty")
	}
	var created struct {
		Result struct {
			SessionID string `json:"session_id"`
		} `json:"result"`
	}
	if err := doProviderJSON(ctx, http.MethodPost, base+"/api/v1/sessions", nil, openVikingHeaders(config.RootAPIKey, ident), &created); err != nil {
		return err
	}
	sessionID := strings.TrimSpace(created.Result.SessionID)
	if sessionID == "" {
		return errors.New("openviking session id is empty")
	}
	messagePath := base + "/api/v1/sessions/" + url.PathEscape(sessionID) + "/messages"
	for _, message := range []map[string]any{{"role": "user", "content": text}, {"role": "assistant", "content": "Noted."}} {
		if err := doProviderJSON(ctx, http.MethodPost, messagePath, message, openVikingHeaders(config.RootAPIKey, ident), nil); err != nil {
			return err
		}
	}
	return doProviderJSON(ctx, http.MethodPost, base+"/api/v1/sessions/"+url.PathEscape(sessionID)+"/commit", nil, openVikingHeaders(config.RootAPIKey, ident), nil)
}

func editOpenVikingMemory(ctx context.Context, config productdata.OpenVikingMemoryConfig, ident identity.LocalIdentity, uri string, title string, content string) error {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return err
	}
	body := map[string]any{"uri": strings.TrimSpace(uri), "content": titledContent(title, content), "mode": "replace", "wait": true}
	return doProviderJSON(ctx, http.MethodPost, base+"/api/v1/content/write", body, openVikingHeaders(config.RootAPIKey, ident), nil)
}

func deleteOpenVikingMemory(ctx context.Context, config productdata.OpenVikingMemoryConfig, ident identity.LocalIdentity, uri string) error {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return err
	}
	return doProviderJSON(ctx, http.MethodDelete, base+"/api/v1/fs?uri="+url.QueryEscape(strings.TrimSpace(uri))+"&recursive=false", nil, openVikingHeaders(config.RootAPIKey, ident), nil)
}

type openVikingDirEntry struct {
	URI  string
	Type string
}

func listOpenVikingMemoryDir(ctx context.Context, config productdata.OpenVikingMemoryConfig, ident identity.LocalIdentity, uri string) ([]openVikingDirEntry, error) {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return nil, err
	}
	var response struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code string `json:"code"`
		} `json:"error,omitempty"`
	}
	path := base + "/api/v1/fs/ls?uri=" + url.QueryEscape(strings.TrimSpace(uri))
	if err := doProviderJSON(ctx, http.MethodGet, path, nil, openVikingHeaders(config.RootAPIKey, ident), &response); err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("openviking ls: %s", productdata.RedactEventText(response.Error.Code))
	}
	var rawEntries []struct {
		URI   string `json:"uri"`
		IsDir bool   `json:"isDir"`
	}
	if err := json.Unmarshal(response.Result, &rawEntries); err == nil {
		entries := make([]openVikingDirEntry, 0, len(rawEntries))
		for _, entry := range rawEntries {
			if strings.TrimSpace(entry.URI) == "" {
				continue
			}
			nodeType := "memory"
			if entry.IsDir {
				nodeType = "directory"
			}
			entries = append(entries, openVikingDirEntry{URI: strings.TrimSpace(entry.URI), Type: nodeType})
		}
		return entries, nil
	}
	var rawURIs []string
	if err := json.Unmarshal(response.Result, &rawURIs); err != nil {
		return nil, err
	}
	entries := make([]openVikingDirEntry, 0, len(rawURIs))
	for _, rawURI := range rawURIs {
		rawURI = strings.TrimSpace(rawURI)
		if rawURI == "" {
			continue
		}
		nodeType := "memory"
		if strings.HasSuffix(rawURI, "/") {
			nodeType = "directory"
		}
		entries = append(entries, openVikingDirEntry{URI: rawURI, Type: nodeType})
	}
	return entries, nil
}

func openVikingURITitle(uri string) string {
	trimmed := strings.Trim(strings.TrimSpace(uri), "/")
	if trimmed == "" {
		return "OpenViking memory"
	}
	parts := strings.Split(trimmed, "/")
	return firstNonEmpty(parts[len(parts)-1], trimmed)
}

func searchNowledgeMemory(ctx context.Context, config productdata.NowledgeMemoryConfig, ident identity.LocalIdentity, query string, limit int) ([]externalMemoryHit, error) {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return nil, err
	}
	values := url.Values{}
	values.Set("query", strings.TrimSpace(query))
	values.Set("limit", fmt.Sprintf("%d", boundedLimit(limit)))
	var response struct {
		Memories []struct {
			ID              string  `json:"id"`
			Title           string  `json:"title"`
			Content         string  `json:"content"`
			Score           float64 `json:"score"`
			Confidence      float64 `json:"confidence"`
			RelevanceReason string  `json:"relevance_reason"`
			Metadata        struct {
				SimilarityScore float64 `json:"similarity_score"`
				RelevanceReason string  `json:"relevance_reason"`
			} `json:"metadata"`
		} `json:"memories"`
	}
	if err := doProviderJSON(ctx, http.MethodGet, base+"/memories/search?"+values.Encode(), nil, nowledgeHeaders(config.APIKey, ident), &response); err != nil {
		return nil, err
	}
	hits := make([]externalMemoryHit, 0, len(response.Memories))
	for _, item := range response.Memories {
		score := item.Score
		if score == 0 {
			score = item.Confidence
		}
		if score == 0 {
			score = item.Metadata.SimilarityScore
		}
		reason := firstNonEmpty(item.RelevanceReason, item.Metadata.RelevanceReason)
		summary := productdata.RedactEventText(firstNonEmpty(item.Content, item.Title))
		hits = append(hits, externalMemoryHit{URI: "nowledge://memory/" + strings.TrimSpace(item.ID), Title: productdata.RedactEventText(item.Title), Summary: summary, Score: score, RankReason: productdata.RedactEventText(reason), SourceType: "nowledge", Provider: productdata.MemoryProviderNowledge, RedactedRaw: true})
	}
	return hits, nil
}

func readNowledgeMemory(ctx context.Context, config productdata.NowledgeMemoryConfig, ident identity.LocalIdentity, uri string) (string, error) {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return "", err
	}
	id := strings.TrimPrefix(strings.TrimSpace(uri), "nowledge://memory/")
	if id == "" {
		return "", errors.New("nowledge memory id is empty")
	}
	var response struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := doProviderJSON(ctx, http.MethodGet, base+"/memories/"+url.PathEscape(id), nil, nowledgeHeaders(config.APIKey, ident), &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.Title) == "" {
		return response.Content, nil
	}
	return response.Title + "\n\n" + response.Content, nil
}

func writeNowledgeMemory(ctx context.Context, config productdata.NowledgeMemoryConfig, ident identity.LocalIdentity, title string, content string) (string, error) {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return "", err
	}
	body := map[string]any{"content": productdata.RedactEventText(strings.TrimSpace(content))}
	if strings.TrimSpace(title) != "" {
		body["title"] = productdata.RedactEventText(strings.TrimSpace(title))
	}
	var response struct {
		ID string `json:"id"`
	}
	if err := doProviderJSON(ctx, http.MethodPost, base+"/memories", body, nowledgeHeaders(config.APIKey, ident), &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.ID) == "" {
		return "", nil
	}
	return "nowledge://memory/" + strings.TrimSpace(response.ID), nil
}

func deleteNowledgeMemory(ctx context.Context, config productdata.NowledgeMemoryConfig, ident identity.LocalIdentity, uri string) error {
	base, err := normalizedHTTPBase(config.BaseURL)
	if err != nil {
		return err
	}
	id := strings.TrimPrefix(strings.TrimSpace(uri), "nowledge://memory/")
	if id == "" {
		return errors.New("nowledge memory id is empty")
	}
	return doProviderJSON(ctx, http.MethodDelete, base+"/memories/"+url.PathEscape(id), nil, nowledgeHeaders(config.APIKey, ident), nil)
}

func doProviderJSON(ctx context.Context, method string, target string, body any, headers map[string]string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		requestBody = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		if strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("memory provider request failed")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("memory provider returned status %d", resp.StatusCode)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("memory provider response decode failed")
	}
	return nil
}

func normalizedHTTPBase(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == nil || parsed.Host == "" {
		return "", errors.New("memory provider base url is invalid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("memory provider base url must be http or https")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func openVikingHeaders(rootKey string, ident identity.LocalIdentity) map[string]string {
	return map[string]string{"X-API-Key": rootKey, "X-OpenViking-Account": "local", "X-OpenViking-User": ident.UserID, "X-OpenViking-Agent": "loomi-default"}
}

func nowledgeHeaders(apiKey string, ident identity.LocalIdentity) map[string]string {
	headers := map[string]string{"X-Arkloop-Account": "local", "X-Arkloop-User": ident.UserID, "X-Arkloop-Agent": "loomi-default", "X-Arkloop-App": "loomi"}
	if strings.TrimSpace(apiKey) != "" {
		headers["Authorization"] = "Bearer " + strings.TrimSpace(apiKey)
		headers["x-nmem-api-key"] = strings.TrimSpace(apiKey)
	}
	return headers
}

func boundedLimit(limit int) int {
	if limit <= 0 {
		return 5
	}
	if limit > 20 {
		return 20
	}
	return limit
}

func splitExternalMemoryText(value string) (string, string) {
	text := productdata.RedactEventText(strings.TrimSpace(value))
	parts := strings.SplitN(text, "\n", 2)
	title := strings.TrimSpace(parts[0])
	summary := title
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		summary = strings.TrimSpace(parts[1])
	}
	return title, summary
}

func titledContent(title string, content string) string {
	title = strings.TrimSpace(productdata.RedactEventText(title))
	content = strings.TrimSpace(productdata.RedactEventText(content))
	if title == "" {
		return content
	}
	if content == "" {
		return title
	}
	return title + "\n\n" + content
}
