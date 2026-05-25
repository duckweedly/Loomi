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
	"strings"

	"github.com/sheridiany/loomi/internal/config"
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
	ProviderStatusAvailable     ProviderStatus = "available"
	ProviderStatusUnavailable   ProviderStatus = "unavailable"
	ProviderStatusMisconfigured ProviderStatus = "misconfigured"
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
	ID      string         `json:"id"`
	Family  ProviderFamily `json:"family"`
	BaseURL string         `json:"base_url,omitempty"`
	Model   string         `json:"model"`
	Status  ProviderStatus `json:"status"`
	Message string         `json:"message,omitempty"`
}

type ProviderRequest struct {
	ThreadID  string
	MessageID string
	Messages  []ProviderMessage
	Model     string
}

type ProviderMessage struct {
	Role    string
	Content string
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
	if capability.Status != ProviderStatusAvailable {
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
		return singleProviderEvent(eventForHTTPStatus(response.StatusCode)), nil
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
	body := anthropicRequestBody{Model: selectedModel(request.Model, p.providerConfig.Model), MaxTokens: 4096, Stream: true, Messages: anthropicMessages(request.Messages)}
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
	body := openAIRequestBody{Model: selectedModel(request.Model, p.providerConfig.Model), Stream: true, Messages: openAIMessages(request.Messages)}
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
			if capability.Status == ProviderStatusAvailable {
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
		Status:  ProviderStatusAvailable,
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

func parseProviderSSE(ctx context.Context, family ProviderFamily, body io.Reader, ch chan<- ProviderEvent) bool {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var eventName string
	var dataLines []string
	terminal := false
	dispatch := func() {
		if len(dataLines) == 0 {
			eventName = ""
			return
		}
		if dispatchProviderEvent(ctx, family, eventName, strings.Join(dataLines, "\n"), ch) {
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
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventToolCall, ToolName: event.ContentBlock.Name})
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

func dispatchOpenAIEvent(ctx context.Context, data string, ch chan<- ProviderEvent) bool {
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
			if toolCall.Function.Name != "" {
				metadata := map[string]any{}
					if toolCall.ID != "" {
						metadata["tool_call_id"] = toolCall.ID
					}
					if toolCall.Function.Arguments != "" {
						metadata["arguments_summary"] = parseToolArgumentsSummary(toolCall.Function.Arguments)
					}
					_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventToolCall, ToolName: toolCall.Function.Name, Metadata: metadata})
			}
		}
		if choice.Delta.FunctionCall.Name != "" {
			metadata := map[string]any{}
				if choice.Delta.FunctionCall.Arguments != "" {
					metadata["arguments_summary"] = parseToolArgumentsSummary(choice.Delta.FunctionCall.Arguments)
				}
				_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventToolCall, ToolName: choice.Delta.FunctionCall.Name, Metadata: metadata})
		}
		if choice.FinishReason == "stop" || choice.FinishReason == "length" {
			_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventCompleted, FinishInfo: choice.FinishReason})
			return true
		}
	}
	return false
}

func parseToolArgumentsSummary(raw string) map[string]any {
	var arguments map[string]any
	if err := json.Unmarshal([]byte(raw), &arguments); err != nil {
		return map[string]any{"_invalid_json": true}
	}
	return arguments
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
				_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventToolCall, ToolName: part.FunctionCall.Name})
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

func eventForHTTPStatus(status int) ProviderEvent {
	switch status {
	case http.StatusTooManyRequests:
		return ProviderEvent{Type: ProviderEventRateLimited, Message: "Provider rate limit reached."}
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return ProviderEvent{Type: ProviderEventMisconfigured, Message: "Provider configuration is incomplete."}
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return ProviderEvent{Type: ProviderEventTimeout, Message: "Provider request timed out."}
	default:
		return ProviderEvent{Type: ProviderEventError, ErrorCode: "provider_error", Message: "Provider request failed."}
	}
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
	Model    string          `json:"model"`
	Stream   bool            `json:"stream"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func openAIMessages(messages []ProviderMessage) []openAIMessage {
	result := make([]openAIMessage, 0, len(messages))
	for _, message := range messages {
		role := message.Role
		if role != "assistant" {
			role = "user"
		}
		result = append(result, openAIMessage{Role: role, Content: message.Content})
	}
	return result
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
	Type         string `json:"type"`
	Delta        struct {
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
