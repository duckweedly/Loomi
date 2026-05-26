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

const DefaultBaseURL = "http://127.0.0.1:8080"

type Client struct {
	baseURL    string
	httpClient *http.Client
	sseClient  *http.Client
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
	ID string `json:"id"`
}

type Run struct {
	ID       string `json:"id"`
	ThreadID string `json:"thread_id"`
	Status   string `json:"status"`
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

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) CheckReady(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/readyz", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("readyz returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) CreateThread(ctx context.Context, mode string) (Thread, error) {
	body := map[string]string{"mode": strings.TrimSpace(mode)}
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

func (c *Client) ListTools(ctx context.Context) ([]ToolCatalogEntry, error) {
	var resp struct {
		Tools []ToolCatalogEntry `json:"tools"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/tools/catalog", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Tools, nil
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
		return fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if out == nil || len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return err
	}
	return nil
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
