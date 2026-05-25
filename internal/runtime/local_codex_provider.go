package runtime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type LocalCodexProvider struct {
	input LocalProviderDetectionInput
}

func NewLocalCodexProvider(input LocalProviderDetectionInput) Provider {
	return LocalCodexProvider{input: input}
}

func (p LocalCodexProvider) Config() ProviderConfig {
	return ProviderConfig{ID: "local_codex", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://api.openai.com/v1", APIKey: "redacted", Model: localCodexDefaultModel, Enabled: true}
}

func (p LocalCodexProvider) Stream(ctx context.Context, request ProviderRequest) (<-chan ProviderEvent, error) {
	snapshot, err := LoadLocalCodexCredentialSnapshot(p.input)
	if err != nil {
		return singleProviderEvent(ProviderEvent{Type: ProviderEventMisconfigured, Message: "Local Codex login is unavailable."}), nil
	}
	if snapshot.AuthMode == LocalProviderAuthModeOAuth {
		return streamLocalCodexResponses(ctx, snapshot, request, http.DefaultClient)
	}
	provider := NewHTTPProvider(ProviderConfig{ID: "local_codex", Family: ProviderFamilyOpenAICompatible, BaseURL: snapshot.BaseURL, APIKey: snapshot.APIKey, Model: snapshot.Model, Enabled: true}, http.DefaultClient)
	return provider.Stream(ctx, request)
}

func streamLocalCodexResponses(ctx context.Context, snapshot LocalCodexCredentialSnapshot, request ProviderRequest, client *http.Client) (<-chan ProviderEvent, error) {
	if client == nil {
		client = http.DefaultClient
	}
	httpRequest, err := buildLocalCodexResponsesRequest(ctx, snapshot, request)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(httpRequest)
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
		parseLocalCodexResponsesSSE(ctx, response.Body, ch)
	}()
	return ch, nil
}

func buildLocalCodexResponsesRequest(ctx context.Context, snapshot LocalCodexCredentialSnapshot, request ProviderRequest) (*http.Request, error) {
	body := map[string]any{
		"model":               selectedModel(request.Model, snapshot.Model),
		"instructions":        "",
		"input":               localCodexResponsesInput(request.Messages),
		"tools":               []any{},
		"tool_choice":         "auto",
		"parallel_tool_calls": false,
		"reasoning":           nil,
		"store":               false,
		"stream":              true,
		"include":             []string{},
		"prompt_cache_key":    request.ThreadID,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimRight(snapshot.BaseURL, "/") + "/responses"
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", "Bearer "+snapshot.APIKey)
	if snapshot.AccountID != "" {
		httpRequest.Header.Set("ChatGPT-Account-ID", snapshot.AccountID)
	}
	return httpRequest, nil
}

func localCodexResponsesInput(messages []ProviderMessage) []map[string]any {
	items := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		role := strings.TrimSpace(message.Role)
		if role != "assistant" {
			role = "user"
		}
		contentType := "input_text"
		if role == "assistant" {
			contentType = "output_text"
		}
		text := strings.TrimSpace(message.Content)
		if text == "" {
			continue
		}
		items = append(items, map[string]any{
			"type": "message",
			"role": role,
			"content": []map[string]string{{
				"type": contentType,
				"text": text,
			}},
		})
	}
	return items
}

func parseLocalCodexResponsesSSE(ctx context.Context, body io.Reader, ch chan<- ProviderEvent) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var dataLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if localCodexDispatchResponsesEvent(ctx, strings.Join(dataLines, "\n"), ch) {
				return
			}
			dataLines = nil
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if localCodexDispatchResponsesEvent(ctx, strings.Join(dataLines, "\n"), ch) {
		return
	}
	if err := scanner.Err(); err != nil {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventError, ErrorCode: "stream_error", Message: "Provider stream failed."})
		return
	}
	_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventEmptyResponse, Message: "Model returned an empty response."})
}

func localCodexDispatchResponsesEvent(ctx context.Context, data string, ch chan<- ProviderEvent) bool {
	data = strings.TrimSpace(data)
	if data == "" || data == "[DONE]" {
		return false
	}
	var event struct {
		Type     string `json:"type"`
		Delta    string `json:"delta"`
		Response struct {
			Error struct {
				Type string `json:"type"`
				Code any    `json:"code"`
			} `json:"error"`
		} `json:"response"`
	}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventError, ErrorCode: "parse_error", Message: "Provider stream could not be parsed."})
		return true
	}
	switch event.Type {
	case "response.output_text.delta":
		if event.Delta != "" {
			_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventTextDelta, Text: event.Delta})
		}
	case "response.completed":
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventCompleted})
		return true
	case "response.failed", "response.incomplete":
		metadata := map[string]any{}
		if errorType := sanitizeProviderErrorField(event.Response.Error.Type); errorType != "" {
			metadata["provider_error_type"] = errorType
		}
		if errorCode := sanitizeProviderErrorField(strings.TrimSpace(jsonScalarString(event.Response.Error.Code))); errorCode != "" {
			metadata["provider_error_code"] = errorCode
		}
		_ = sendProviderEvent(ctx, ch, ProviderEvent{Type: ProviderEventError, ErrorCode: "provider_error", Message: "Provider response failed.", Metadata: metadata})
		return true
	}
	return false
}

func jsonScalarString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(typed, 'f', -1, 64), "0"), ".")
	default:
		return ""
	}
}
