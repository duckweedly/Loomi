package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
)

func TestProviderConfigsFromConfig(t *testing.T) {
	providers := ProviderConfigsFromConfig(config.Config{ModelProviders: []config.ModelProvider{{ID: "custom", Family: "openai_compatible", BaseURL: "https://example.test/v1?token=secret", APIKey: "key", Model: "model", Enabled: true}}})
	if len(providers) != 1 {
		t.Fatalf("providers = %+v", providers)
	}
	capability := providers[0].Capability()
	if capability.Status != ProviderStatusAvailable || capability.BaseURL != "https://example.test" {
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
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL+"/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if len(events) != 3 || events[0].Type != ProviderEventTextDelta || events[0].Text != "hi" || events[1].Type != ProviderEventToolCall || events[1].ToolName != "search" || events[2].Type != ProviderEventCompleted {
		t.Fatalf("events = %+v", events)
	}
}

func TestHTTPProviderPreservesOpenAIToolArgumentsAsMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"id\":\"tc_1\",\"function\":{\"name\":\"runtime.get_current_time\",\"arguments\":\"{\\\"timezone\\\":\\\"Asia/Shanghai\\\"}\"}}]}}]}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL+"/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if len(events) != 1 || events[0].Type != ProviderEventToolCall || events[0].Metadata["tool_call_id"] != "tc_1" {
		t.Fatalf("events = %+v", events)
	}
	arguments, ok := events[0].Metadata["arguments_summary"].(map[string]any)
	if !ok || arguments["timezone"] != "Asia/Shanghai" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
}

func TestHTTPProviderNormalizesStreamingErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"error\":{\"type\":\"rate_limit_error\",\"message\":\"raw secret body\"}}\n\n"))
	}))
	defer server.Close()
	provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL+"/v1", APIKey: "secret-key", Model: "gpt-5.5", Enabled: true}, server.Client())

	events := collectProviderEvents(t, provider)

	if len(events) != 1 || events[0].Type != ProviderEventRateLimited {
		t.Fatalf("events = %+v", events)
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
	stream, err := provider.Stream(context.Background(), ProviderRequest{ThreadID: "thr_1", MessageID: "msg_1", Model: provider.Config().Model, Messages: []ProviderMessage{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	var events []ProviderEvent
	for event := range stream {
		events = append(events, event)
	}
	return events
}
