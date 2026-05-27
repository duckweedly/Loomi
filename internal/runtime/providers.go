package runtime

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
	"sort"
	"strings"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/productdata"
)

type ProviderFamily string

const (
	ProviderFamilyAnthropic        ProviderFamily = "anthropic"
	ProviderFamilyOpenAI           ProviderFamily = "openai"
	ProviderFamilyGemini           ProviderFamily = "gemini"
	ProviderFamilyOpenAICompatible ProviderFamily = "openai_compatible"
)

type ProviderStatus string

const (
	ProviderStatusConfigured       ProviderStatus = "configured"
	ProviderStatusReachable        ProviderStatus = "reachable"
	ProviderStatusCompletionOK     ProviderStatus = "completion-ok"
	ProviderStatusCompletionFailed ProviderStatus = "completion-failed"
	ProviderStatusAvailable        ProviderStatus = "available"
	ProviderStatusUnavailable      ProviderStatus = "unavailable"
	ProviderStatusMisconfigured    ProviderStatus = "misconfigured"
)

type ProviderEventType string

const (
	ProviderEventTextDelta     ProviderEventType = "text_delta"
	ProviderEventCompleted     ProviderEventType = "completed"
	ProviderEventRefusal       ProviderEventType = "refusal"
	ProviderEventToolCall      ProviderEventType = "tool_call"
	ProviderEventError         ProviderEventType = "error"
	ProviderEventTimeout       ProviderEventType = "timeout"
	ProviderEventRateLimited   ProviderEventType = "rate_limited"
	ProviderEventEmptyResponse ProviderEventType = "empty_response"
	ProviderEventMisconfigured ProviderEventType = "misconfigured"
)

type ProviderConfig struct {
	ID      string
	Family  ProviderFamily
	BaseURL string
	APIKey  string
	Model   string
	Enabled bool
}

type ProviderCapability struct {
	ID                  string         `json:"id"`
	Family              ProviderFamily `json:"family"`
	BaseURL             string         `json:"base_url,omitempty"`
	Model               string         `json:"model"`
	Status              ProviderStatus `json:"status"`
	Message             string         `json:"message,omitempty"`
	LocalProvider       bool           `json:"local_provider,omitempty"`
	SessionLocal        bool           `json:"session_local,omitempty"`
	CredentialReference string         `json:"credential_reference,omitempty"`
	ExecutionState      string         `json:"execution_state,omitempty"`
	CheckStage          string         `json:"check_stage,omitempty"`
	CheckCode           string         `json:"check_code,omitempty"`
	HTTPStatus          int            `json:"http_status,omitempty"`
}

type ProviderRequest struct {
	ThreadID     string
	MessageID    string
	SystemPrompt string
	Messages     []ProviderMessage
	Model        string
	Tools        []ProviderToolDefinition
}

type ProviderToolDefinition struct {
	Name         string
	ProviderName string
	Description  string
	Parameters   map[string]any
}

const (
	ProviderMessageRoleAssistantToolCall = "assistant_tool_call"
	ProviderMessageRoleToolResult        = "tool_result"
)

type ProviderMessage struct {
	Role             string
	Content          string
	ToolCallID       string
	ToolName         string
	ArgumentsSummary map[string]any
}

type ProviderEvent struct {
	Type       ProviderEventType
	Text       string
	ToolName   string
	ErrorCode  string
	Message    string
	Metadata   map[string]any
	FinishInfo string
}

type Provider interface {
	Config() ProviderConfig
	Stream(context.Context, ProviderRequest) (<-chan ProviderEvent, error)
}

type StaticProvider struct {
	ProviderConfig ProviderConfig
	Events         []ProviderEvent
	Err            error
}

func (p StaticProvider) Config() ProviderConfig { return p.ProviderConfig }

func (p StaticProvider) Stream(ctx context.Context, _ ProviderRequest) (<-chan ProviderEvent, error) {
	if p.Err != nil {
		return nil, p.Err
	}
	ch := make(chan ProviderEvent, len(p.Events))
	go func() {
		defer close(ch)
		for _, event := range p.Events {
			select {
			case <-ctx.Done():
				return
			case ch <- event:
			}
		}
	}()
	return ch, nil
}

type HTTPProvider struct {
	providerConfig ProviderConfig
	client         *http.Client
}

func NewHTTPProvider(provider ProviderConfig, client *http.Client) Provider {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPProvider{providerConfig: provider, client: client}
}

func NewHTTPProviders(providers []ProviderConfig, client *http.Client) []Provider {
	result := make([]Provider, 0, len(providers))
	for _, provider := range providers {
		result = append(result, NewHTTPProvider(provider, client))
	}
	return result
}

func (p *HTTPProvider) Config() ProviderConfig { return p.providerConfig }

func (p *HTTPProvider) Stream(ctx context.Context, request ProviderRequest) (<-chan ProviderEvent, error) {
	capability := p.providerConfig.Capability()
	if !providerStatusCanRun(capability.Status) {
		return nil, errors.New(capability.Message)
	}
	httpRequest, err := p.buildRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	response, err := p.client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		defer response.Body.Close()
		return singleProviderEvent(eventForHTTPResponse(response.StatusCode, response.Body)), nil
	}
	ch := make(chan ProviderEvent)
	go func() {
		defer close(ch)
		defer response.Body.Close()
		terminal := parseProviderSSE(ctx, p.providerConfig.Family, response.Body, ch)
		if p.providerConfig.Family == ProviderFamilyGemini && !terminal {
			_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventCompleted})
		}
	}()
	return ch, nil
}

func (p *HTTPProvider) buildRequest(ctx context.Context, request ProviderRequest) (*http.Request, error) {
	switch p.providerConfig.Family {
	case ProviderFamilyAnthropic:
		return p.buildAnthropicRequest(ctx, request)
	case ProviderFamilyOpenAI, ProviderFamilyOpenAICompatible:
		return p.buildOpenAICompatibleRequest(ctx, request)
	case ProviderFamilyGemini:
		return p.buildGeminiRequest(ctx, request)
	default:
		return nil, errors.New("Provider family is unsupported.")
	}
}

func (p *HTTPProvider) buildAnthropicRequest(ctx context.Context, request ProviderRequest) (*http.Request, error) {
	body := anthropicRequestBody{Model: selectedModel(request.Model, p.providerConfig.Model), MaxTokens: 4096, Stream: true, System: request.SystemPrompt, Messages: anthropicMessages(request.Messages)}
	if len(body.Messages) == 0 {
		return nil, errors.New("Provider request messages are required.")
	}
	endpoint := strings.TrimRight(defaultProviderBaseURL(p.providerConfig), "/") + "/v1/messages"
	httpRequest, err := jsonRequest(ctx, endpoint, body)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("x-api-key", p.providerConfig.APIKey)
	httpRequest.Header.Set("anthropic-version", "2023-06-01")
	return httpRequest, nil
}

func (p *HTTPProvider) buildOpenAICompatibleRequest(ctx context.Context, request ProviderRequest) (*http.Request, error) {
	body := openAIRequestBody{Model: selectedModel(request.Model, p.providerConfig.Model), Stream: true, Messages: openAIMessages(request.Messages), Tools: openAITools(request.Tools)}
	if strings.TrimSpace(request.SystemPrompt) != "" {
		system := request.SystemPrompt
		body.Messages = append([]openAIMessage{{Role: "system", Content: &system}}, body.Messages...)
	}
	if len(body.Messages) == 0 {
		return nil, errors.New("Provider request messages are required.")
	}
	endpoint := strings.TrimRight(defaultProviderBaseURL(p.providerConfig), "/") + "/chat/completions"
	httpRequest, err := jsonRequest(ctx, endpoint, body)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+p.providerConfig.APIKey)
	return httpRequest, nil
}

func (p *HTTPProvider) buildGeminiRequest(ctx context.Context, request ProviderRequest) (*http.Request, error) {
	body := geminiRequestBody{Contents: geminiContents(request.Messages)}
	if len(body.Contents) == 0 {
		return nil, errors.New("Provider request messages are required.")
	}
	baseURL := strings.TrimRight(defaultProviderBaseURL(p.providerConfig), "/")
	endpoint := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse", baseURL, url.PathEscape(selectedModel(request.Model, p.providerConfig.Model)))
	httpRequest, err := jsonRequest(ctx, endpoint, body)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("x-goog-api-key", p.providerConfig.APIKey)
	return httpRequest, nil
}

func ProviderConfigsFromConfig(cfg config.Config) []ProviderConfig {
	providers := make([]ProviderConfig, 0, len(cfg.ModelProviders))
	for _, provider := range cfg.ModelProviders {
		providers = append(providers, ProviderConfig{ID: provider.ID, Family: ProviderFamily(provider.Family), BaseURL: provider.BaseURL, APIKey: provider.APIKey, Model: provider.Model, Enabled: provider.Enabled})
	}
	return providers
}

func ProviderCapabilities(providers []ProviderConfig) []ProviderCapability {
	capabilities := make([]ProviderCapability, 0, len(providers))
	for _, provider := range providers {
		capabilities = append(capabilities, provider.Capability())
	}
	return capabilities
}

func SelectProvider(providers []ProviderConfig, providerID string) (ProviderConfig, error) {
	for _, provider := range providers {
		if provider.ID == providerID {
			capability := provider.Capability()
			if providerStatusCanRun(capability.Status) {
				return provider, nil
			}
			return ProviderConfig{}, errors.New(capability.Message)
		}
	}
	return ProviderConfig{}, errors.New("Provider is not configured.")
}

func (c ProviderConfig) Capability() ProviderCapability {
	capability := ProviderCapability{
		ID:      c.ID,
		Family:  c.Family,
		BaseURL: redactProviderBaseURL(c.BaseURL),
		Model:   c.Model,
		Status:  ProviderStatusConfigured,
	}
	if !c.Enabled {
		capability.Status = ProviderStatusUnavailable
		capability.Message = "Provider is disabled."
		return capability
	}
	if c.ID == "" || c.Model == "" || c.APIKey == "" {
		capability.Status = ProviderStatusMisconfigured
		capability.Message = "Provider configuration is incomplete."
		return capability
	}
	switch c.Family {
	case ProviderFamilyAnthropic, ProviderFamilyOpenAI, ProviderFamilyGemini:
	case ProviderFamilyOpenAICompatible:
		if c.BaseURL == "" {
			capability.Status = ProviderStatusMisconfigured
			capability.Message = "Custom provider base URL is required."
			return capability
		}
		if _, err := url.ParseRequestURI(c.BaseURL); err != nil {
			capability.Status = ProviderStatusMisconfigured
			capability.Message = "Custom provider base URL is invalid."
			return capability
		}
	default:
		capability.Status = ProviderStatusMisconfigured
		capability.Message = "Provider family is unsupported."
	}
	return capability
}

func providerStatusCanRun(status ProviderStatus) bool {
	switch status {
	case ProviderStatusAvailable, ProviderStatusConfigured, ProviderStatusReachable, ProviderStatusCompletionOK:
		return true
	default:
		return false
	}
}

func CheckProviderCompletion(ctx context.Context, provider ProviderConfig, client *http.Client) ProviderCapability {
	if client == nil {
		client = http.DefaultClient
	}
	capability := provider.Capability()
	if !providerStatusCanRun(capability.Status) {
		return capability
	}
	checker := &HTTPProvider{providerConfig: provider, client: client}
	req, err := checker.buildCompletionCheckRequest(ctx)
	if err != nil {
		capability.Status = ProviderStatusCompletionFailed
		capability.CheckStage = "completion"
		capability.CheckCode = "completion-failed-request"
		capability.Message = "Provider completion check could not be prepared."
		return capability
	}
	resp, err := client.Do(req)
	if err != nil {
		capability.Status = ProviderStatusCompletionFailed
		capability.CheckStage = "completion"
		capability.CheckCode = "completion-failed-network"
		capability.Message = "Provider completion check failed before receiving a response."
		return capability
	}
	defer resp.Body.Close()
	capability.CheckStage = "completion"
	capability.HTTPStatus = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
		capability.Status = ProviderStatusCompletionFailed
		capability.CheckCode, capability.Message = completionHTTPFailure(resp.StatusCode)
		return capability
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
	capability.Status = ProviderStatusCompletionOK
	capability.CheckCode = "completion-ok"
	capability.Message = "Provider completion check succeeded."
	return capability
}

func (p *HTTPProvider) buildCompletionCheckRequest(ctx context.Context) (*http.Request, error) {
	switch p.providerConfig.Family {
	case ProviderFamilyAnthropic:
		content := "ok"
		body := anthropicRequestBody{Model: p.providerConfig.Model, MaxTokens: 8, Stream: false, Messages: []anthropicMessage{{Role: "user", Content: []anthropicContentBlock{{Type: "text", Text: content}}}}}
		httpRequest, err := jsonRequest(ctx, strings.TrimRight(defaultProviderBaseURL(p.providerConfig), "/")+"/v1/messages", body)
		if err != nil {
			return nil, err
		}
		httpRequest.Header.Set("x-api-key", p.providerConfig.APIKey)
		httpRequest.Header.Set("anthropic-version", "2023-06-01")
		return httpRequest, nil
	case ProviderFamilyOpenAI, ProviderFamilyOpenAICompatible:
		content := "ok"
		body := openAIRequestBody{Model: p.providerConfig.Model, Stream: false, Messages: []openAIMessage{{Role: "user", Content: &content}}}
		httpRequest, err := jsonRequest(ctx, strings.TrimRight(defaultProviderBaseURL(p.providerConfig), "/")+"/chat/completions", body)
		if err != nil {
			return nil, err
		}
		httpRequest.Header.Set("Authorization", "Bearer "+p.providerConfig.APIKey)
		return httpRequest, nil
	case ProviderFamilyGemini:
		body := geminiRequestBody{Contents: []geminiContent{{Role: "user", Parts: []geminiPart{{Text: "ok"}}}}}
		baseURL := strings.TrimRight(defaultProviderBaseURL(p.providerConfig), "/")
		httpRequest, err := jsonRequest(ctx, fmt.Sprintf("%s/v1beta/models/%s:generateContent", baseURL, url.PathEscape(p.providerConfig.Model)), body)
		if err != nil {
			return nil, err
		}
		httpRequest.Header.Set("x-goog-api-key", p.providerConfig.APIKey)
		return httpRequest, nil
	default:
		return nil, errors.New("Provider family is unsupported.")
	}
}

func completionHTTPFailure(status int) (string, string) {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return "completion-failed-auth", "Provider token was rejected by the upstream completion endpoint."
	case http.StatusTooManyRequests:
		return "completion-failed-rate-limited", "Provider completion check was rate limited by the upstream endpoint."
	default:
		return fmt.Sprintf("completion-failed-%d", status), fmt.Sprintf("Provider completion check failed with HTTP %d.", status)
	}
}

func parseProviderSSE(ctx context.Context, family ProviderFamily, body io.Reader, ch chan<- ProviderEvent) bool {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var eventName string
	var dataLines []string
	terminal := false
	openAIState := openAIStreamAccumulator{}
	dispatch := func() {
		if len(dataLines) == 0 {
			eventName = ""
			return
		}
		data := strings.Join(dataLines, "\n")
		if (family == ProviderFamilyOpenAI || family == ProviderFamilyOpenAICompatible) && data == "[DONE]" {
			if openAIState.flushToolCall(ctx, ch) {
				terminal = true
			} else {
				_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventCompleted})
				terminal = true
			}
			eventName = ""
			dataLines = nil
			return
		}
		if (family == ProviderFamilyOpenAI || family == ProviderFamilyOpenAICompatible) && openAIState.dispatch(ctx, data, ch) {
			terminal = true
		} else if family != ProviderFamilyOpenAI && family != ProviderFamilyOpenAICompatible && dispatchProviderEvent(ctx, family, eventName, data, ch) {
			terminal = true
		}
		eventName = ""
		dataLines = nil
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			dispatch()
			continue
		}
		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	dispatch()
	if err := scanner.Err(); err != nil {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventError, ErrorCode: "stream_error", Message: "Provider stream failed."})
		return true
	}
	return terminal
}

func dispatchProviderEvent(ctx context.Context, family ProviderFamily, eventName string, data string, ch chan<- ProviderEvent) bool {
	if data == "" {
		return false
	}
	if data == "[DONE]" {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventCompleted})
		return true
	}
	switch family {
	case ProviderFamilyAnthropic:
		return dispatchAnthropicEvent(ctx, eventName, data, ch)
	case ProviderFamilyOpenAI, ProviderFamilyOpenAICompatible:
		return dispatchOpenAIEvent(ctx, data, ch)
	case ProviderFamilyGemini:
		return dispatchGeminiEvent(ctx, data, ch)
	default:
		return false
	}
}

type openAIStreamAccumulator struct {
	toolCalls      map[int]*openAIToolCallAccumulator
	functionName   string
	functionArgs   strings.Builder
	functionCallID string
}

type openAIToolCallAccumulator struct {
	ID   string
	Name string
	Args strings.Builder
}

func dispatchAnthropicEvent(ctx context.Context, eventName string, data string, ch chan<- ProviderEvent) bool {
	var event anthropicStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventError, ErrorCode: "parse_error", Message: "Provider stream could not be parsed."})
		return true
	}
	if event.Type == "error" || eventName == "error" {
		_ = sendProviderEvent(ctx, ch, eventForProviderError(event.Error.Type))
		return true
	}
	if event.Type == "content_block_start" && event.ContentBlock.Type == "tool_use" {
		_ = sendProviderEvent(ctx, ch, providerToolCallEvent(event.ContentBlock.Name, nil))
		return false
	}
	if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventTextDelta, Text: event.Delta.Text})
		return false
	}
	if event.Type == "message_delta" && event.Delta.StopReason == "refusal" {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventRefusal, Message: "Model response was refused."})
		return true
	}
	if event.Type == "message_stop" || eventName == "message_stop" {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventCompleted})
		return true
	}
	return false
}

func (s *openAIStreamAccumulator) dispatch(ctx context.Context, data string, ch chan<- ProviderEvent) bool {
	var event openAIStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventError, ErrorCode: "parse_error", Message: "Provider stream could not be parsed."})
		return true
	}
	if event.Error.Message != "" || event.Error.Type != "" {
		_ = sendProviderEvent(ctx, ch, eventForProviderError(event.Error.Type))
		return true
	}
	for _, choice := range event.Choices {
		if choice.Delta.Content != "" {
			_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventTextDelta, Text: choice.Delta.Content})
		}
		if choice.Delta.Refusal != "" || choice.FinishReason == "content_filter" {
			_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventRefusal, Message: fallbackMessage(choice.Delta.Refusal, "Model response was refused.")})
			return true
		}
		for _, toolCall := range choice.Delta.ToolCalls {
			if s.toolCalls == nil {
				s.toolCalls = map[int]*openAIToolCallAccumulator{}
			}
			acc := s.toolCalls[toolCall.Index]
			if acc == nil {
				acc = &openAIToolCallAccumulator{}
				s.toolCalls[toolCall.Index] = acc
			}
			if toolCall.ID != "" {
				acc.ID = toolCall.ID
			}
			if toolCall.Function.Name != "" {
				acc.Name = toolCall.Function.Name
			}
			if toolCall.Function.Arguments != "" {
				acc.Args.WriteString(toolCall.Function.Arguments)
			}
		}
		if choice.Delta.FunctionCall.Name != "" {
			s.functionName = choice.Delta.FunctionCall.Name
		}
		if choice.Delta.FunctionCall.Arguments != "" {
			s.functionArgs.WriteString(choice.Delta.FunctionCall.Arguments)
		}
		if choice.FinishReason == "tool_calls" || choice.FinishReason == "function_call" {
			return s.flushToolCall(ctx, ch)
		}
		if choice.FinishReason == "stop" || choice.FinishReason == "length" {
			_ = s.flushToolCall(ctx, ch)
			_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventCompleted, FinishInfo: choice.FinishReason})
			return true
		}
	}
	return false
}

func dispatchOpenAIEvent(ctx context.Context, data string, ch chan<- ProviderEvent) bool {
	state := openAIStreamAccumulator{}
	return state.dispatch(ctx, data, ch)
}

func (s *openAIStreamAccumulator) flushToolCall(ctx context.Context, ch chan<- ProviderEvent) bool {
	if s.toolCalls != nil {
		indexes := make([]int, 0, len(s.toolCalls))
		for index := range s.toolCalls {
			indexes = append(indexes, index)
		}
		sort.Ints(indexes)
		for _, index := range indexes {
			acc := s.toolCalls[index]
			if acc == nil || acc.Name == "" {
				continue
			}
			metadata := map[string]any{}
			if acc.ID != "" {
				metadata["tool_call_id"] = acc.ID
			}
			if args := acc.Args.String(); args != "" {
				metadata["arguments_summary"] = parseToolArgumentsSummary(args)
			}
			_ = sendProviderEvent(ctx, ch, providerToolCallEvent(acc.Name, metadata))
			delete(s.toolCalls, index)
			return true
		}
	}
	if s.functionName != "" {
		metadata := map[string]any{}
		if args := s.functionArgs.String(); args != "" {
			metadata["arguments_summary"] = parseToolArgumentsSummary(args)
		}
		_ = sendProviderEvent(ctx, ch, providerToolCallEvent(s.functionName, metadata))
		s.functionName = ""
		s.functionArgs.Reset()
		return true
	}
	return false
}

func providerToolCallEvent(providerName string, metadata map[string]any) ProviderEvent {
	normalized := internalProviderToolName(providerName)
	next := map[string]any{}
	for key, value := range metadata {
		next[key] = value
	}
	if raw := strings.TrimSpace(providerName); raw != "" && raw != normalized {
		next["provider_tool_name"] = raw
	}
	return ProviderEvent{Type: ProviderEventToolCall, ToolName: normalized, Metadata: next}
}

func parseToolArgumentsSummary(raw string) map[string]any {
	var arguments map[string]any
	if err := json.Unmarshal([]byte(raw), &arguments); err != nil {
		return map[string]any{"_invalid_json": true}
	}
	return arguments
}

func internalProviderToolName(name string) string {
	switch strings.TrimSpace(name) {
	case "tool_load_tools":
		return productdata.ToolNameLoadTools
	case "skill_load_skill":
		return productdata.ToolNameLoadSkill
	case "web_search":
		return "web.search"
	case "web.search", "search":
		return productdata.ToolNameWebSearch
	case "workspace_glob":
		return productdata.ToolNameWorkspaceGlob
	case "workspace_grep":
		return productdata.ToolNameWorkspaceGrep
	case "workspace_read":
		return productdata.ToolNameWorkspaceRead
	case "workspace_write_file":
		return productdata.ToolNameWorkspaceWriteFile
	case "workspace_edit":
		return productdata.ToolNameWorkspaceEdit
	case "workspace_patch_preview":
		return productdata.ToolNameWorkspacePatchPreview
	case "workspace_patch_apply":
		return productdata.ToolNameWorkspacePatchApply
	case "sandbox_exec_command":
		return productdata.ToolNameSandboxExecCommand
	case "sandbox_start_process":
		return productdata.ToolNameSandboxStartProcess
	case "sandbox_continue_process":
		return productdata.ToolNameSandboxContinueProcess
	case "sandbox_terminate_process":
		return productdata.ToolNameSandboxTerminateProcess
	case "lsp_diagnostics":
		return productdata.ToolNameLSPDiagnostics
	case "lsp_symbols":
		return productdata.ToolNameLSPSymbols
	case "lsp_references":
		return productdata.ToolNameLSPReferences
	case "lsp_definition":
		return productdata.ToolNameLSPDefinition
	case "lsp_hover":
		return productdata.ToolNameLSPHover
	case "web_fetch", "web.fetch", "fetch":
		return productdata.ToolNameWebFetch
	case "browser_open":
		return productdata.ToolNameBrowserOpen
	case "browser_snapshot":
		return productdata.ToolNameBrowserSnapshot
	case "browser_click_link":
		return productdata.ToolNameBrowserClickLink
	case "browser_screenshot":
		return productdata.ToolNameBrowserScreenshot
	case "browser_type":
		return productdata.ToolNameBrowserType
	case "browser_press":
		return productdata.ToolNameBrowserPress
	case "memory_search":
		return productdata.ToolNameMemorySearch
	case "memory_list":
		return productdata.ToolNameMemoryList
	case "memory_read":
		return productdata.ToolNameMemoryRead
	case "memory_write":
		return productdata.ToolNameMemoryWrite
	case "memory_edit":
		return productdata.ToolNameMemoryEdit
	case "memory_forget":
		return productdata.ToolNameMemoryForget
	case "memory_context":
		return productdata.ToolNameMemoryContext
	case "memory_timeline":
		return productdata.ToolNameMemoryTimeline
	case "memory_connections":
		return productdata.ToolNameMemoryConnections
	case "memory_thread_search":
		return productdata.ToolNameMemoryThreadSearch
	case "memory_thread_fetch":
		return productdata.ToolNameMemoryThreadFetch
	case "memory_status":
		return productdata.ToolNameMemoryStatus
	case "notebook_read":
		return productdata.ToolNameNotebookRead
	case "notebook_write":
		return productdata.ToolNameNotebookWrite
	case "notebook_edit":
		return productdata.ToolNameNotebookEdit
	case "notebook_forget":
		return productdata.ToolNameNotebookForget
	case "todo_write":
		return productdata.ToolNameTodoWrite
	default:
		return name
	}
}

func dispatchGeminiEvent(ctx context.Context, data string, ch chan<- ProviderEvent) bool {
	var event geminiStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventError, ErrorCode: "parse_error", Message: "Provider stream could not be parsed."})
		return true
	}
	for _, candidate := range event.Candidates {
		if candidate.FinishReason == "SAFETY" || candidate.FinishReason == "BLOCKED" || candidate.FinishReason == "RECITATION" {
			_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventRefusal, Message: "Model response was refused."})
			return true
		}
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventTextDelta, Text: part.Text})
			}
			if part.FunctionCall.Name != "" {
				_ = sendProviderEvent(ctx, ch, providerToolCallEvent(part.FunctionCall.Name, nil))
			}
		}
	}
	return false
}

func jsonRequest(ctx context.Context, endpoint string, body any) (*http.Request, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func singleProviderEvent(event ProviderEvent) <-chan ProviderEvent {
	ch := make(chan ProviderEvent, 1)
	ch <- event
	close(ch)
	return ch
}

func eventForHTTPResponse(status int, body io.Reader) ProviderEvent {
	metadata := map[string]any{"http_status": status}
	if errorType, errorCode := providerHTTPErrorFields(body); errorType != "" || errorCode != "" {
		if errorType != "" {
			metadata["provider_error_type"] = errorType
		}
		if errorCode != "" {
			metadata["provider_error_code"] = errorCode
		}
	}
	switch status {
	case http.StatusTooManyRequests:
		return ProviderEvent{Type: ProviderEventRateLimited, Message: "Provider rate limit reached.", Metadata: metadata}
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return ProviderEvent{Type: ProviderEventMisconfigured, Message: fmt.Sprintf("Provider rejected the credential or endpoint (HTTP %d).", status), Metadata: metadata}
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return ProviderEvent{Type: ProviderEventTimeout, Message: "Provider request timed out.", Metadata: metadata}
	default:
		return ProviderEvent{Type: ProviderEventError, ErrorCode: "provider_error", Message: fmt.Sprintf("Provider request failed with HTTP %d.", status), Metadata: metadata}
	}
}

func providerHTTPErrorFields(body io.Reader) (string, string) {
	raw, err := io.ReadAll(io.LimitReader(body, 64*1024))
	if err != nil || len(raw) == 0 {
		return "", ""
	}
	var parsed struct {
		Error struct {
			Type string `json:"type"`
			Code any    `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", ""
	}
	return sanitizeProviderErrorField(parsed.Error.Type), sanitizeProviderErrorField(fmt.Sprint(parsed.Error.Code))
}

func sanitizeProviderErrorField(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "<nil>" || strings.Contains(value, "/") || strings.Contains(value, "\\") || strings.Contains(strings.ToLower(value), "token") || strings.Contains(strings.ToLower(value), "key") {
		return ""
	}
	if len(value) > 80 {
		return value[:80]
	}
	return value
}

func eventForProviderError(errorType string) ProviderEvent {
	switch providerErrorCode(errorType) {
	case string(ProviderEventRateLimited):
		return ProviderEvent{Type: ProviderEventRateLimited, Message: "Provider rate limit reached."}
	case string(ProviderEventTimeout):
		return ProviderEvent{Type: ProviderEventTimeout, Message: "Provider request timed out."}
	case string(ProviderEventMisconfigured):
		return ProviderEvent{Type: ProviderEventMisconfigured, Message: "Provider configuration is incomplete."}
	default:
		return ProviderEvent{Type: ProviderEventError, ErrorCode: "provider_error", Message: "Provider request failed."}
	}
}

func sendProviderEvent(ctx context.Context, ch chan<- ProviderEvent, event ProviderEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- event:
		return true
	}
}

func defaultProviderBaseURL(provider ProviderConfig) string {
	if provider.BaseURL != "" {
		return provider.BaseURL
	}
	switch provider.Family {
	case ProviderFamilyAnthropic:
		return "https://api.anthropic.com"
	case ProviderFamilyOpenAI:
		return "https://api.openai.com/v1"
	case ProviderFamilyGemini:
		return "https://generativelanguage.googleapis.com"
	default:
		return provider.BaseURL
	}
}

func providerErrorCode(value string) string {
	switch strings.ToLower(value) {
	case "rate_limit_error", "rate_limited", "overloaded_error":
		return string(ProviderEventRateLimited)
	case "timeout", "request_timeout":
		return string(ProviderEventTimeout)
	case "authentication_error", "permission_error", "invalid_request_error":
		return string(ProviderEventMisconfigured)
	default:
		return string(ProviderEventError)
	}
}

func redactProviderBaseURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "[redacted]"
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = ""
	u.RawPath = ""
	return u.String()
}

type anthropicRequestBody struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type         string             `json:"type"`
	Text         string             `json:"text"`
	CacheControl *anthropicCacheCtl `json:"cache_control,omitempty"`
}

type anthropicCacheCtl struct {
	Type string `json:"type"`
}

func anthropicMessages(messages []ProviderMessage) []anthropicMessage {
	result := make([]anthropicMessage, 0, len(messages))
	for _, message := range messages {
		role := "user"
		if message.Role == "assistant" {
			role = "assistant"
		}
		result = append(result, anthropicMessage{Role: role, Content: []anthropicContentBlock{{Type: "text", Text: message.Content}}})
	}
	if len(result) > 0 {
		lastMessage := len(result) - 1
		lastBlock := len(result[lastMessage].Content) - 1
		result[lastMessage].Content[lastBlock].CacheControl = &anthropicCacheCtl{Type: "ephemeral"}
	}
	return result
}

type openAIRequestBody struct {
	Model    string           `json:"model"`
	Stream   bool             `json:"stream"`
	Messages []openAIMessage  `json:"messages"`
	Tools    []openAIToolSpec `json:"tools,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    *string          `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIToolSpec struct {
	Type     string               `json:"type"`
	Function openAIToolDefinition `json:"function"`
}

type openAIToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

func openAIMessages(messages []ProviderMessage) []openAIMessage {
	result := make([]openAIMessage, 0, len(messages))
	for _, message := range messages {
		if message.Role == ProviderMessageRoleAssistantToolCall {
			arguments, _ := json.Marshal(message.ArgumentsSummary)
			result = append(result, openAIMessage{Role: "assistant", ToolCalls: []openAIToolCall{{ID: message.ToolCallID, Type: "function", Function: openAIToolFunction{Name: providerToolName(message.ToolName), Arguments: string(arguments)}}}})
			continue
		}
		if message.Role == ProviderMessageRoleToolResult {
			content := message.Content
			result = append(result, openAIMessage{Role: "tool", Content: &content, ToolCallID: message.ToolCallID})
			continue
		}
		role := message.Role
		if role != "assistant" && role != "system" {
			role = "user"
		}
		content := message.Content
		result = append(result, openAIMessage{Role: role, Content: &content})
	}
	return result
}

func openAITools(tools []ProviderToolDefinition) []openAIToolSpec {
	result := make([]openAIToolSpec, 0, len(tools))
	for _, tool := range tools {
		if tool.ProviderName == "" || len(tool.Parameters) == 0 {
			continue
		}
		result = append(result, openAIToolSpec{Type: "function", Function: openAIToolDefinition{Name: tool.ProviderName, Description: tool.Description, Parameters: tool.Parameters}})
	}
	return result
}

func providerToolName(name string) string {
	switch strings.TrimSpace(name) {
	case productdata.ToolNameLoadTools:
		return "tool_load_tools"
	case productdata.ToolNameLoadSkill:
		return "skill_load_skill"
	case "web.search":
		return "web_search"
	case productdata.ToolNameWorkspaceGlob:
		return "workspace_glob"
	case productdata.ToolNameWorkspaceGrep:
		return "workspace_grep"
	case productdata.ToolNameWorkspaceRead:
		return "workspace_read"
	case productdata.ToolNameWorkspaceWriteFile:
		return "workspace_write_file"
	case productdata.ToolNameWorkspaceEdit:
		return "workspace_edit"
	case productdata.ToolNameWorkspacePatchPreview:
		return "workspace_patch_preview"
	case productdata.ToolNameWorkspacePatchApply:
		return "workspace_patch_apply"
	case productdata.ToolNameSandboxExecCommand:
		return "sandbox_exec_command"
	case productdata.ToolNameSandboxStartProcess:
		return "sandbox_start_process"
	case productdata.ToolNameSandboxContinueProcess:
		return "sandbox_continue_process"
	case productdata.ToolNameSandboxTerminateProcess:
		return "sandbox_terminate_process"
	case productdata.ToolNameLSPDiagnostics:
		return "lsp_diagnostics"
	case productdata.ToolNameLSPSymbols:
		return "lsp_symbols"
	case productdata.ToolNameLSPReferences:
		return "lsp_references"
	case productdata.ToolNameLSPDefinition:
		return "lsp_definition"
	case productdata.ToolNameLSPHover:
		return "lsp_hover"
	case productdata.ToolNameWebFetch:
		return "web_fetch"
	case productdata.ToolNameBrowserOpen:
		return "browser_open"
	case productdata.ToolNameBrowserSnapshot:
		return "browser_snapshot"
	case productdata.ToolNameBrowserClickLink:
		return "browser_click_link"
	case productdata.ToolNameBrowserScreenshot:
		return "browser_screenshot"
	case productdata.ToolNameBrowserType:
		return "browser_type"
	case productdata.ToolNameBrowserPress:
		return "browser_press"
	case productdata.ToolNameMemorySearch:
		return "memory_search"
	case productdata.ToolNameMemoryList:
		return "memory_list"
	case productdata.ToolNameMemoryRead:
		return "memory_read"
	case productdata.ToolNameMemoryWrite:
		return "memory_write"
	case productdata.ToolNameMemoryEdit:
		return "memory_edit"
	case productdata.ToolNameMemoryForget:
		return "memory_forget"
	case productdata.ToolNameMemoryContext:
		return "memory_context"
	case productdata.ToolNameMemoryTimeline:
		return "memory_timeline"
	case productdata.ToolNameMemoryConnections:
		return "memory_connections"
	case productdata.ToolNameMemoryThreadSearch:
		return "memory_thread_search"
	case productdata.ToolNameMemoryThreadFetch:
		return "memory_thread_fetch"
	case productdata.ToolNameMemoryStatus:
		return "memory_status"
	case productdata.ToolNameNotebookRead:
		return "notebook_read"
	case productdata.ToolNameNotebookWrite:
		return "notebook_write"
	case productdata.ToolNameNotebookEdit:
		return "notebook_edit"
	case productdata.ToolNameNotebookForget:
		return "notebook_forget"
	case productdata.ToolNameTodoWrite:
		return "todo_write"
	default:
		return name
	}
}

type geminiRequestBody struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text,omitempty"`
}

func geminiContents(messages []ProviderMessage) []geminiContent {
	result := make([]geminiContent, 0, len(messages))
	for _, message := range messages {
		role := "user"
		if message.Role == "assistant" {
			role = "model"
		}
		result = append(result, geminiContent{Role: role, Parts: []geminiPart{{Text: message.Content}}})
	}
	return result
}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Type       string `json:"type"`
		Text       string `json:"text"`
		StopReason string `json:"stop_reason"`
	} `json:"delta"`
	ContentBlock struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"content_block"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

type openAIStreamEvent struct {
	Choices []struct {
		Delta struct {
			Content      string `json:"content"`
			Refusal      string `json:"refusal"`
			FunctionCall struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function_call"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

type geminiStreamEvent struct {
	Candidates []struct {
		FinishReason string `json:"finishReason"`
		Content      struct {
			Parts []struct {
				Text         string `json:"text"`
				FunctionCall struct {
					Name string `json:"name"`
				} `json:"functionCall"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}
