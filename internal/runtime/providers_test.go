package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestProviderConfigsFromConfig(t *testing.T) {
	providers := ProviderConfigsFromConfig(config.Config{ModelProviders: []config.ModelProvider{{ID: "custom", Family: "openai_compatible", BaseURL: "https://example.test/v1?token=secret", APIKey: "key", Model: "model", Enabled: true}}})
	if len(providers) != 1 {
		t.Fatalf("providers = %+v", providers)
	}
	capability := providers[0].Capability()
	if capability.Status != ProviderStatusConfigured || capability.BaseURL != "https://example.test" {
		t.Fatalf("capability = %+v", capability)
	}
}

func TestProviderCapabilityReportsMisconfiguredCustomProvider(t *testing.T) {
	capability := ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, APIKey: "key", Model: "model", Enabled: true}.Capability()
	if capability.Status != ProviderStatusMisconfigured {
		t.Fatalf("capability = %+v", capability)
	}
}

func TestProviderCapabilityRedactsCustomBaseURLPath(t *testing.T) {
	capability := ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://gateway.example.test/key/secret-token/v1?token=secret", APIKey: "key", Model: "model", Enabled: true}.Capability()
	if capability.BaseURL != "https://gateway.example.test" {
		t.Fatalf("base url = %q", capability.BaseURL)
	}
}

func TestSelectProviderRejectsUnavailableProvider(t *testing.T) {
	_, err := SelectProvider([]ProviderConfig{{ID: "custom", Family: ProviderFamilyOpenAICompatible, Model: "model", Enabled: false}}, "custom")
	if err == nil {
		t.Fatal("SelectProvider() error = nil, want error")
	}
}

func TestHTTPProviderStreamsAnthropicTextAndToolEvents(t *testing.T) {
	var auth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("x-api-key")
		if r.URL.Path != "/v1/messages" || r.Header.Get("anthropic-version") == "" {
			t.Fatalf("request path=%s headers=%v", r.URL.Path, r.Header)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"hel\"}}\n\n"))
		_, _ = w.Write([]byte("event: content_block_start\ndata: {\"type\":\"content_block_start\",\"content_block\":{\"type\":\"tool_use\",\"name\":\"read_file\"}}\n\n"))
		_, _ = w.Write([]byte("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "anthropic", Family: ProviderFamilyAnthropic, BaseURL: server.URL, APIKey: "secret-key", Model: "claude-opus-4-7", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if auth != "secret-key" {
		t.Fatalf("x-api-key = %q", auth)
	}
	if len(events) != 3 || events[0].Type != ProviderEventTextDelta || events[0].Text != "hel" || events[1].Type != ProviderEventToolCall || events[1].ToolName != "read_file" || events[2].Type != ProviderEventCompleted {
		t.Fatalf("events = %+v", events)
	}
}

func TestHTTPProviderStreamsOpenAICompatibleTextAndToolEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer secret-key" {
			t.Fatalf("request path=%s headers=%v", r.URL.Path, r.Header)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"function\":{\"name\":\"search\"}}]}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if len(events) != 3 || events[0].Type != ProviderEventTextDelta || events[0].Text != "hi" || events[1].Type != ProviderEventToolCall || events[1].ToolName != productdata.ToolNameWebSearch || events[2].Type != ProviderEventCompleted {
		t.Fatalf("events = %+v", events)
	}
}

func TestHTTPProviderPreservesOpenAIToolArgumentsAsMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"id\":\"tc_1\",\"function\":{\"name\":\"runtime.get_current_time\",\"arguments\":\"{\\\"timezone\\\":\\\"Asia/Shanghai\\\"}\"}}]}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if len(events) != 1 || events[0].Type != ProviderEventToolCall || events[0].Metadata["tool_call_id"] != "tc_1" {
		t.Fatalf("events = %+v", events)
	}
	arguments, ok := events[0].Metadata["arguments_summary"].(map[string]any)
	if !ok || arguments["timezone"] != "Asia/Shanghai" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
}

func TestHTTPProviderAccumulatesOpenAIToolArgumentsAcrossChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"tc_search\",\"function\":{\"name\":\"web_search\",\"arguments\":\"{\\\"que\"}}]}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"ry\\\":\\\"今天最新 AI\\\"}\"}}]}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"tool_calls\"}]}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEventsForRequest(t, provider, ProviderRequest{ThreadID: "thr_1", MessageID: "msg_1", Model: "gpt-5.5", Messages: []ProviderMessage{{Role: "user", Content: "search latest ai"}}, Tools: []ProviderToolDefinition{WebSearchProviderToolDefinition()}})

	if len(events) != 1 || events[0].Type != ProviderEventToolCall || events[0].ToolName != "web.search" {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Metadata["tool_call_id"] != "tc_search" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
	arguments, ok := events[0].Metadata["arguments_summary"].(map[string]any)
	if !ok || arguments["query"] != "今天最新 AI" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
}

func TestHTTPProviderMapsCommonSearchFunctionAlias(t *testing.T) {
	if got := internalProviderToolName("search"); got != productdata.ToolNameWebSearch {
		t.Fatalf("internalProviderToolName(search) = %q", got)
	}
	if got := internalProviderToolName("web.search"); got != productdata.ToolNameWebSearch {
		t.Fatalf("internalProviderToolName(web.search) = %q", got)
	}
	if got := internalProviderToolName("fetch"); got != productdata.ToolNameWebFetch {
		t.Fatalf("internalProviderToolName(fetch) = %q", got)
	}
	if got := internalProviderToolName("web.fetch"); got != productdata.ToolNameWebFetch {
		t.Fatalf("internalProviderToolName(web.fetch) = %q", got)
	}
}

func TestHTTPProviderSerializesOpenAIToolResultContinuation(t *testing.T) {
	var body struct {
		Messages []map[string]any `json:"messages"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"done\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEventsForRequest(t, provider, ProviderRequest{ThreadID: "thr_1", MessageID: "msg_1", Model: "gpt-5.5", Messages: []ProviderMessage{
		{Role: "user", Content: "What time is it?"},
		{Role: ProviderMessageRoleAssistantToolCall, ToolCallID: "tc_1", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}},
		{Role: ProviderMessageRoleToolResult, ToolCallID: "tc_1", ToolName: "runtime.get_current_time", Content: `{"iso_time":"2026-05-25T10:00:00Z","timezone":"UTC","source":"runtime"}`},
	}})

	if len(events) != 2 || events[0].Type != ProviderEventTextDelta || events[1].Type != ProviderEventCompleted {
		t.Fatalf("events = %+v", events)
	}
	if len(body.Messages) != 3 {
		t.Fatalf("messages = %+v", body.Messages)
	}
	assistant := body.Messages[1]
	if assistant["role"] != "assistant" || assistant["content"] != nil {
		t.Fatalf("assistant tool call message = %+v", assistant)
	}
	toolCalls, ok := assistant["tool_calls"].([]any)
	if !ok || len(toolCalls) != 1 {
		t.Fatalf("tool_calls = %+v", assistant["tool_calls"])
	}
	tool := body.Messages[2]
	if tool["role"] != "tool" || tool["tool_call_id"] != "tc_1" || tool["content"] == "" {
		t.Fatalf("tool result message = %+v", tool)
	}
}

func TestHTTPProviderSendsSystemPromptAsOpenAISystemMessage(t *testing.T) {
	var body struct {
		Messages []map[string]any `json:"messages"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"done\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEventsForRequest(t, provider, ProviderRequest{ThreadID: "thr_1", MessageID: "msg_1", Model: "gpt-5.5", SystemPrompt: "Use tools only when needed.", Messages: []ProviderMessage{{Role: "user", Content: "hello"}}})

	if len(events) != 2 || events[1].Type != ProviderEventCompleted {
		t.Fatalf("events = %+v", events)
	}
	if len(body.Messages) != 2 || body.Messages[0]["role"] != "system" || body.Messages[0]["content"] != "Use tools only when needed." {
		t.Fatalf("messages = %+v", body.Messages)
	}
}

func TestHTTPProviderSendsEnabledWebSearchToolSchema(t *testing.T) {
	var body struct {
		Tools []map[string]any `json:"tools"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"done\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEventsForRequest(t, provider, ProviderRequest{ThreadID: "thr_1", MessageID: "msg_1", Model: "gpt-5.5", Messages: []ProviderMessage{{Role: "user", Content: "search latest news"}}, Tools: []ProviderToolDefinition{WebSearchProviderToolDefinition()}})

	if len(events) != 2 || events[1].Type != ProviderEventCompleted {
		t.Fatalf("events = %+v", events)
	}
	if len(body.Tools) != 1 || body.Tools[0]["type"] != "function" {
		t.Fatalf("tools = %+v", body.Tools)
	}
	function, ok := body.Tools[0]["function"].(map[string]any)
	if !ok || function["name"] != "web_search" {
		t.Fatalf("function tool = %+v", body.Tools[0])
	}
}

func TestHTTPProviderNormalizesStreamingErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"error\":{\"type\":\"rate_limit_error\",\"message\":\"raw secret body\"}}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if len(events) != 1 || events[0].Type != ProviderEventRateLimited {
		t.Fatalf("events = %+v", events)
	}
}

func TestHTTPProviderReportsRedactedHTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"type":"invalid_request_error","code":"unsupported_parameter","message":"raw token secret should not appear"}}`))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if len(events) != 1 || events[0].Type != ProviderEventError || events[0].Message != "Provider request failed with HTTP 400." {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Metadata["http_status"] != http.StatusBadRequest || events[0].Metadata["provider_error_type"] != "invalid_request_error" || events[0].Metadata["provider_error_code"] != "unsupported_parameter" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
}

func TestCheckProviderCompletionReportsHTTP503WithoutLeakingBody(t *testing.T) {
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer secret-key" {
			t.Fatalf("request path=%s headers=%v", r.URL.Path, r.Header)
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"message":"raw provider body sk-leak prompt trace"}}`))
	}))
	defer server.Close()

	capability := CheckProviderCompletion(context.Background(), ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	if capability.Status != ProviderStatusCompletionFailed || capability.CheckCode != "completion-failed-503" || capability.HTTPStatus != http.StatusServiceUnavailable {
		t.Fatalf("capability = %+v", capability)
	}
	if capability.Message != "Provider completion check failed with HTTP 503." {
		t.Fatalf("message = %q", capability.Message)
	}
	rendered, err := json.Marshal(capability)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rendered), "sk-leak") || strings.Contains(string(rendered), "prompt trace") || strings.Contains(string(rendered), "secret-key") {
		t.Fatalf("capability leaked provider data: %s", rendered)
	}
	if requestBody["stream"] != false {
		t.Fatalf("request body = %+v", requestBody)
	}
	if messages, ok := requestBody["messages"].([]any); !ok || len(messages) != 1 {
		t.Fatalf("request body messages = %+v", requestBody["messages"])
	}
}

func TestHTTPProviderStreamsGeminiTextAndFunctionEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/gemini-3.5-flash:streamGenerateContent" || r.URL.Query().Get("alt") != "sse" || r.Header.Get("x-goog-api-key") != "secret-key" {
			t.Fatalf("request url=%s headers=%v", r.URL.String(), r.Header)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hey\"}]}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"candidates\":[{\"content\":{\"parts\":[{\"functionCall\":{\"name\":\"lookup\"}}]}}]}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "gemini", Family: ProviderFamilyGemini, BaseURL: server.URL, APIKey: "secret-key", Model: "gemini-3.5-flash", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if len(events) != 3 || events[0].Type != ProviderEventTextDelta || events[0].Text != "hey" || events[1].Type != ProviderEventToolCall || events[1].ToolName != "lookup" || events[2].Type != ProviderEventCompleted {
		t.Fatalf("events = %+v", events)
	}
}

func collectProviderEvents(t *testing.T, provider Provider) []ProviderEvent {
	t.Helper()
	return collectProviderEventsForRequest(t, provider, ProviderRequest{ThreadID: "thr_1", MessageID: "msg_1", Model: provider.Config().Model, Messages: []ProviderMessage{{Role: "user", Content: "hello"}}})
}

func collectProviderEventsForRequest(t *testing.T, provider Provider, request ProviderRequest) []ProviderEvent {
	t.Helper()
	stream, err := provider.Stream(context.Background(), request)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	var events []ProviderEvent
	for event := range stream {
		events = append(events, event)
	}
	return events
}
