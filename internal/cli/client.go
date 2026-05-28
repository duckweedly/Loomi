package cli

import (
	"bufio"
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
)

const DefaultBaseURL = "http://127.0.0.1:18080"

type Client struct {
	baseURL     string
	bearerToken string
	httpClient  *http.Client
	sseClient   *http.Client
}

type Thread struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Mode            string  `json:"mode"`
	PersonaID       *string `json:"persona_id"`
	LifecycleStatus string  `json:"lifecycle_status"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	ArchivedAt      *string `json:"archived_at"`
}

type Message struct {
	ID       string         `json:"id"`
	ThreadID string         `json:"thread_id"`
	Role     string         `json:"role"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
	RunID    string         `json:"run_id"`
}

type Run struct {
	ID       string `json:"id"`
	ThreadID string `json:"thread_id"`
	Status   string `json:"status"`
}

type StopRunResult struct {
	Run    Run    `json:"run"`
	Result string `json:"result"`
}

type RunEvent struct {
	ID       string         `json:"id"`
	RunID    string         `json:"run_id"`
	ThreadID string         `json:"thread_id"`
	Sequence int            `json:"sequence"`
	Type     string         `json:"type"`
	Summary  string         `json:"summary"`
	Content  *string        `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

type ToolCall struct {
	ToolCallID      string         `json:"tool_call_id"`
	ToolName        string         `json:"tool_name"`
	ApprovalStatus  string         `json:"approval_status"`
	ExecutionStatus string         `json:"execution_status"`
	Arguments       map[string]any `json:"arguments_summary"`
}

type ToolCatalogEntry struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Group          string `json:"group"`
	RiskLevel      string `json:"risk_level"`
	ApprovalPolicy string `json:"approval_policy"`
	Enabled        bool   `json:"enabled"`
	ExecutionState string `json:"execution_state"`
}

type MCPServerStatus struct {
	ServerSafeID      string   `json:"server_safe_id"`
	ServerSlug        string   `json:"server_slug"`
	DisplayName       string   `json:"display_name"`
	Transport         string   `json:"transport"`
	Enabled           bool     `json:"enabled"`
	ConfigSource      string   `json:"config_source"`
	DiscoveryStatus   string   `json:"discovery_status"`
	CandidateCount    int      `json:"candidate_count"`
	CandidateNames    []string `json:"candidate_names"`
	ExecutionMode     string   `json:"execution_mode"`
	RedactedErrorCode string   `json:"redacted_error_code,omitempty"`
	LastDiscoveredAt  string   `json:"last_discovered_at,omitempty"`
}

type Artifact struct {
	ID           string `json:"id"`
	ThreadID     string `json:"thread_id"`
	RunID        string `json:"run_id"`
	Title        string `json:"title"`
	ArtifactType string `json:"artifact_type"`
	ContentBytes int    `json:"content_bytes"`
	TextExcerpt  string `json:"text_excerpt"`
	Truncated    bool   `json:"truncated"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type MemorySearchResult struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Summary          string `json:"summary"`
	ScopeType        string `json:"scope_type"`
	ScopeID          string `json:"scope_id"`
	Status           string `json:"status"`
	SafetyState      string `json:"safety_state"`
	SourceThreadID   string `json:"source_thread_id,omitempty"`
	SourceRunID      string `json:"source_run_id,omitempty"`
	SourceEventID    string `json:"source_event_id,omitempty"`
	SourceType       string `json:"source_type"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	DeletedAt        string `json:"deleted_at,omitempty"`
	RankReason       string `json:"rank_reason,omitempty"`
	RedactionApplied bool   `json:"redaction_applied"`
}

type MemoryAuditItem struct {
	ID               string `json:"id"`
	EventType        string `json:"event_type"`
	Summary          string `json:"summary"`
	ThreadID         string `json:"thread_id,omitempty"`
	RunID            string `json:"run_id,omitempty"`
	MemoryEntryID    string `json:"memory_entry_id,omitempty"`
	MemoryProposalID string `json:"memory_proposal_id,omitempty"`
	Status           string `json:"status,omitempty"`
	ScopeType        string `json:"scope_type,omitempty"`
	SourceType       string `json:"source_type,omitempty"`
	RedactionApplied bool   `json:"redaction_applied"`
	OccurredAt       string `json:"occurred_at"`
}

type AgentTask struct {
	ID            string `json:"id"`
	ThreadID      string `json:"thread_id"`
	RunID         string `json:"run_id"`
	Role          string `json:"role"`
	Goal          string `json:"goal"`
	Status        string `json:"status"`
	ResultSummary string `json:"result_summary,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type MemoryFilters struct {
	Query             string
	ScopeType         string
	ScopeID           string
	SourceThreadID    string
	SourceRunID       string
	SourceType        string
	IncludeTombstoned bool
	Limit             int
}

type Persona struct {
	ID            string `json:"id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Source        string `json:"source"`
	IsDefault     bool   `json:"is_default"`
	IsActive      bool   `json:"is_active"`
	ActiveVersion int    `json:"active_version"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type ProviderCapability struct {
	ID                  string `json:"id"`
	Family              string `json:"family"`
	BaseURL             string `json:"base_url"`
	Model               string `json:"model"`
	Status              string `json:"status"`
	Message             string `json:"message"`
	LocalProvider       bool   `json:"local_provider"`
	SessionLocal        bool   `json:"session_local"`
	CredentialReference string `json:"credential_reference"`
	ExecutionState      string `json:"execution_state"`
	CheckStage          string `json:"check_stage"`
	CheckCode           string `json:"check_code"`
	HTTPStatus          int    `json:"http_status"`
}

type LocalProviderCapability struct {
	ProviderID string `json:"provider_id"`
	AuthMode   string `json:"auth_mode"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

type WorkspaceRootConfig struct {
	Configured  bool   `json:"configured"`
	DisplayName string `json:"display_name"`
}

type Readiness struct {
	Status string           `json:"status"`
	Checks []ReadinessCheck `json:"checks"`
}

type ReadinessCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		sseClient: &http.Client{
			Transport: &http.Transport{ResponseHeaderTimeout: 15 * time.Second},
		},
	}
}

func NewClientFromConfig(cfg Config) *Client {
	client := NewClient(cfg.Host)
	client.SetBearerToken(cfg.APIToken)
	return client
}

func (c *Client) SetBearerToken(token string) {
	if c != nil {
		c.bearerToken = strings.TrimSpace(token)
	}
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) CheckReady(ctx context.Context) error {
	ready, err := c.GetReadiness(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(ready.Status) != "" && ready.Status != "ready" {
		return fmt.Errorf("readyz status=%s", ready.Status)
	}
	return nil
}

func (c *Client) GetReadiness(ctx context.Context) (Readiness, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/readyz", nil)
	if err != nil {
		return Readiness{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Readiness{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Readiness{}, err
	}
	var ready Readiness
	if len(bytes.TrimSpace(raw)) > 0 {
		_ = json.Unmarshal(raw, &ready)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if strings.TrimSpace(ready.Status) != "" || len(ready.Checks) > 0 {
			return ready, nil
		}
		return Readiness{}, fmt.Errorf("readyz returned %d", resp.StatusCode)
	}
	if strings.TrimSpace(ready.Status) == "" {
		ready.Status = "ready"
	}
	return ready, nil
}

func (c *Client) CreateThread(ctx context.Context, mode string, title string) (Thread, error) {
	body := map[string]string{"mode": strings.TrimSpace(mode), "title": strings.TrimSpace(title)}
	var resp struct {
		Thread Thread `json:"thread"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/threads", body, &resp); err != nil {
		return Thread{}, err
	}
	return resp.Thread, nil
}

func (c *Client) ListThreads(ctx context.Context) ([]Thread, error) {
	var resp struct {
		Threads []Thread `json:"threads"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/threads", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Threads, nil
}

func (c *Client) AddMessage(ctx context.Context, threadID string, content string) (Message, error) {
	var resp struct {
		Message Message `json:"message"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/threads/"+url.PathEscape(threadID)+"/messages", map[string]string{"content": content}, &resp); err != nil {
		return Message{}, err
	}
	return resp.Message, nil
}

func (c *Client) ListMessages(ctx context.Context, threadID string) ([]Message, error) {
	var resp struct {
		Messages []Message `json:"messages"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/threads/"+url.PathEscape(threadID)+"/messages", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func (c *Client) SaveWorkspaceRoot(ctx context.Context, path string) (WorkspaceRootConfig, error) {
	var resp struct {
		Config WorkspaceRootConfig `json:"config"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/workspace/root", map[string]string{"path": strings.TrimSpace(path)}, &resp); err != nil {
		return WorkspaceRootConfig{}, err
	}
	return resp.Config, nil
}

func (c *Client) GetWorkspaceRoot(ctx context.Context) (WorkspaceRootConfig, error) {
	var resp struct {
		Config WorkspaceRootConfig `json:"config"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/workspace/root", nil, &resp); err != nil {
		return WorkspaceRootConfig{}, err
	}
	return resp.Config, nil
}

func (c *Client) StartRun(ctx context.Context, threadID string, input StartRunInput) (Run, error) {
	var resp struct {
		Run Run `json:"run"`
	}
	body := map[string]string{}
	if input.MessageID != "" {
		body["message_id"] = input.MessageID
	}
	if input.Source != "" {
		body["source"] = input.Source
	}
	if input.ProviderID != "" {
		body["provider_id"] = input.ProviderID
	}
	if input.Model != "" {
		body["model"] = input.Model
	}
	if input.PersonaID != "" {
		body["persona_id"] = input.PersonaID
	}
	if input.ScriptName != "" {
		body["script_name"] = input.ScriptName
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/threads/"+url.PathEscape(threadID)+"/runs", body, &resp); err != nil {
		return Run{}, err
	}
	return resp.Run, nil
}

func (c *Client) GetRun(ctx context.Context, runID string) (Run, error) {
	var resp struct {
		Run Run `json:"run"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/runs/"+url.PathEscape(runID), nil, &resp); err != nil {
		return Run{}, err
	}
	return resp.Run, nil
}

func (c *Client) StopRun(ctx context.Context, runID string) (StopRunResult, error) {
	var resp StopRunResult
	if err := c.doJSON(ctx, http.MethodPost, "/v1/runs/"+url.PathEscape(runID)+"/stop", nil, &resp); err != nil {
		return StopRunResult{}, err
	}
	return resp, nil
}

func (c *Client) ListTools(ctx context.Context) ([]ToolCatalogEntry, error) {
	var resp struct {
		Tools []ToolCatalogEntry `json:"tools"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/tools/catalog", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Tools, nil
}

func (c *Client) ListMCPServers(ctx context.Context) ([]MCPServerStatus, error) {
	var resp struct {
		Servers []MCPServerStatus `json:"servers"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/mcp/servers", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Servers, nil
}

func (c *Client) ListArtifacts(ctx context.Context, threadID string, limit int) ([]Artifact, error) {
	path := fmt.Sprintf("/v1/threads/%s/artifacts", url.PathEscape(threadID))
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	var resp struct {
		Artifacts []Artifact `json:"artifacts"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Artifacts, nil
}

func (c *Client) ReadArtifact(ctx context.Context, threadID string, artifactID string, maxBytes int) (Artifact, error) {
	path := fmt.Sprintf("/v1/threads/%s/artifacts/%s", url.PathEscape(threadID), url.PathEscape(artifactID))
	if maxBytes > 0 {
		path += fmt.Sprintf("?max_bytes=%d", maxBytes)
	}
	var resp struct {
		Artifact Artifact `json:"artifact"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return Artifact{}, err
	}
	return resp.Artifact, nil
}

func (c *Client) ListMemory(ctx context.Context, filters MemoryFilters) ([]MemorySearchResult, error) {
	values := memoryQuery(filters)
	path := "/v1/memory"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var resp struct {
		Items []MemorySearchResult `json:"items"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) SearchMemory(ctx context.Context, filters MemoryFilters) ([]MemorySearchResult, error) {
	body := map[string]any{}
	if filters.Query != "" {
		body["query"] = filters.Query
	}
	if filters.ScopeType != "" {
		body["scope_type"] = filters.ScopeType
	}
	if filters.ScopeID != "" {
		body["scope_id"] = filters.ScopeID
	}
	if filters.SourceThreadID != "" {
		body["source_thread_id"] = filters.SourceThreadID
	}
	if filters.SourceRunID != "" {
		body["source_run_id"] = filters.SourceRunID
	}
	if filters.SourceType != "" {
		body["source_type"] = filters.SourceType
	}
	if filters.IncludeTombstoned {
		body["include_tombstoned"] = true
	}
	if filters.Limit > 0 {
		body["limit"] = filters.Limit
	}
	var resp struct {
		Items []MemorySearchResult `json:"items"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/memory/search", body, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetMemory(ctx context.Context, entryID string, filters MemoryFilters) (MemorySearchResult, error) {
	values := memoryQuery(filters)
	path := "/v1/memory/entries/" + url.PathEscape(entryID)
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var resp struct {
		Entry MemorySearchResult `json:"entry"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return MemorySearchResult{}, err
	}
	return resp.Entry, nil
}

func (c *Client) ListMemoryAudit(ctx context.Context, filters MemoryFilters, eventType string) ([]MemoryAuditItem, error) {
	values := memoryQuery(filters)
	if eventType != "" {
		values.Set("event_type", eventType)
	}
	path := "/v1/memory/audit"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var resp struct {
		Items []MemoryAuditItem `json:"items"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) ListAgentTasks(ctx context.Context, threadID string, limit int) ([]AgentTask, error) {
	path := fmt.Sprintf("/v1/threads/%s/agent-tasks", url.PathEscape(threadID))
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	var resp struct {
		Tasks []AgentTask `json:"tasks"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

func memoryQuery(filters MemoryFilters) url.Values {
	values := url.Values{}
	if filters.Query != "" {
		values.Set("q", filters.Query)
	}
	if filters.ScopeType != "" {
		values.Set("scope_type", filters.ScopeType)
	}
	if filters.ScopeID != "" {
		values.Set("scope_id", filters.ScopeID)
	}
	if filters.SourceThreadID != "" {
		values.Set("source_thread_id", filters.SourceThreadID)
	}
	if filters.SourceRunID != "" {
		values.Set("source_run_id", filters.SourceRunID)
	}
	if filters.SourceType != "" {
		values.Set("source_type", filters.SourceType)
	}
	if filters.IncludeTombstoned {
		values.Set("include_tombstoned", "true")
	}
	if filters.Limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", filters.Limit))
	}
	return values
}

func (c *Client) ListPersonas(ctx context.Context) ([]Persona, error) {
	var resp struct {
		Personas []Persona `json:"personas"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/personas", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Personas, nil
}

func (c *Client) ListModelProviders(ctx context.Context) ([]ProviderCapability, error) {
	var resp struct {
		Providers []ProviderCapability `json:"providers"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/model-providers", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Providers, nil
}

func (c *Client) CheckModelProvider(ctx context.Context, providerID string) (ProviderCapability, error) {
	var resp struct {
		Provider ProviderCapability `json:"provider"`
	}
	body := map[string]string{"provider_id": strings.TrimSpace(providerID)}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/model-providers/check", body, &resp); err != nil {
		return ProviderCapability{}, err
	}
	return resp.Provider, nil
}

func (c *Client) ListLocalProviderDetections(ctx context.Context) ([]LocalProviderCapability, error) {
	var resp struct {
		Providers []LocalProviderCapability `json:"providers"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/local-provider-detections", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Providers, nil
}

func (c *Client) ListEvents(ctx context.Context, runID string, afterSequence int) ([]RunEvent, error) {
	path := fmt.Sprintf("/v1/runs/%s/events", url.PathEscape(runID))
	if afterSequence > 0 {
		path += fmt.Sprintf("?after_sequence=%d", afterSequence)
	}
	var resp struct {
		Events []RunEvent `json:"events"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Events, nil
}

func (c *Client) StreamEvents(ctx context.Context, runID string, afterSequence int, onEvent func(RunEvent)) error {
	path := fmt.Sprintf("%s/v1/runs/%s/events/stream", c.baseURL, url.PathEscape(runID))
	if afterSequence > 0 {
		path += fmt.Sprintf("?after_sequence=%d", afterSequence)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	c.authorize(req)
	resp, err := c.sseClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("event stream returned %d", resp.StatusCode)
	}
	return readSSE(resp.Body, onEvent)
}

func (c *Client) DecideToolCall(ctx context.Context, threadID string, runID string, toolCallID string, action string) (ToolCall, error) {
	action = strings.TrimSpace(action)
	if action != "approve" && action != "deny" {
		return ToolCall{}, errors.New("tool call action must be approve or deny")
	}
	path := fmt.Sprintf("/v1/threads/%s/runs/%s/tool-calls/%s/%s", url.PathEscape(threadID), url.PathEscape(runID), url.PathEscape(toolCallID), action)
	var resp struct {
		ToolCall ToolCall `json:"tool_call"`
	}
	if err := c.doJSON(ctx, http.MethodPost, path, map[string]string{}, &resp); err != nil {
		return ToolCall{}, err
	}
	return resp.ToolCall, nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.authorize(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s returned %d%s", method, path, resp.StatusCode, safeErrorSuffix(raw))
	}
	if out == nil || len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return err
	}
	return nil
}

func safeErrorSuffix(raw []byte) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return ""
	}
	var body any
	if err := json.Unmarshal(raw, &body); err != nil {
		return ""
	}
	code, message := safeErrorFields(body)
	if strings.TrimSpace(code) != "" {
		return " code=" + strings.TrimSpace(code)
	}
	if strings.Contains(strings.ToLower(message), "missing bearer token") {
		return " missing bearer token"
	}
	return ""
}

func safeErrorFields(body any) (string, string) {
	object, ok := body.(map[string]any)
	if !ok {
		return "", ""
	}
	if value, ok := object["error"].(string); ok {
		return "", value
	}
	if nested, ok := object["error"].(map[string]any); ok {
		code, _ := nested["code"].(string)
		message, _ := nested["message"].(string)
		return code, message
	}
	code, _ := object["code"].(string)
	message, _ := object["message"].(string)
	return code, message
}

func (c *Client) authorize(req *http.Request) {
	if c == nil || req == nil || c.bearerToken == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
}

func readSSE(r io.Reader, onEvent func(RunEvent)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 16*1024), 1024*1024)
	eventName := ""
	var data strings.Builder
	flush := func() error {
		if data.Len() == 0 {
			eventName = ""
			return nil
		}
		raw := strings.TrimSpace(data.String())
		name := eventName
		eventName = ""
		data.Reset()
		if name == "stream_closed" || name == "close" {
			return io.EOF
		}
		if name != "" && name != "run_event" {
			return nil
		}
		event, err := decodeRunEvent([]byte(raw))
		if err != nil {
			return err
		}
		if onEvent != nil {
			onEvent(event)
		}
		return nil
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := flush(); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
			continue
		}
		if value, ok := strings.CutPrefix(line, "event:"); ok {
			eventName = strings.TrimSpace(value)
			continue
		}
		if value, ok := strings.CutPrefix(line, "data:"); ok {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(value))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return flush()
}

func decodeRunEvent(raw []byte) (RunEvent, error) {
	var wrapped struct {
		Event RunEvent `json:"event"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Event.Type != "" {
		return wrapped.Event, nil
	}
	var event RunEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return RunEvent{}, err
	}
	return event, nil
}
