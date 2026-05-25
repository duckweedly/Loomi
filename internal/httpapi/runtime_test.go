package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestStartRunHandlerCreatesLocalSimulatedRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"script_name":"m4_smoke"}`)

	if res.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Run productdata.Run `json:"run"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Run.ThreadID != thread.ID || body.Run.Source != productdata.RunSourceLocalSimulated || body.Run.Status != productdata.RunStatusQueued {
		t.Fatalf("run = %+v", body.Run)
	}
}

func TestModelProviderPreflightAllowsBrowserReads(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	for _, origin := range []string{"http://127.0.0.1:5173", "http://127.0.0.1:5180"} {
		req := httptest.NewRequest(http.MethodOptions, "/v1/model-providers", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
		res := httptest.NewRecorder()

		srv.ServeHTTP(res, req)

		if res.Code != http.StatusNoContent {
			t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
		}
		if res.Header().Get("Access-Control-Allow-Origin") != origin {
			t.Fatalf("origin %s allow origin = %q", origin, res.Header().Get("Access-Control-Allow-Origin"))
		}
		if res.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PATCH, DELETE, OPTIONS" {
			t.Fatalf("allow methods = %q", res.Header().Get("Access-Control-Allow-Methods"))
		}
	}
}

func TestModelProviderHandlersExposeRedactedCapability(t *testing.T) {
	svc := productdata.NewMemoryService()
	cfg := config.Config{AppEnv: "local", ModelProviders: []config.ModelProvider{{ID: "custom", Family: "openai_compatible", BaseURL: "https://user:secret@example.test/v1?token=secret", APIKey: "key", Model: "gpt-5.5", Enabled: true}}}
	srv := NewServerWithProduct(cfg, fakeChecker{}, svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/model-providers", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	res := httptest.NewRecorder()
	srv.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if res.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:5173" {
		t.Fatalf("allow origin = %q", res.Header().Get("Access-Control-Allow-Origin"))
	}
	if strings.Contains(res.Body.String(), "secret") || !strings.Contains(res.Body.String(), "gpt-5.5") {
		t.Fatalf("body = %s", res.Body.String())
	}

	check := requestJSON(t, srv, http.MethodPost, "/v1/model-providers/check", `{"provider_id":"custom"}`)
	if check.Code != http.StatusOK {
		t.Fatalf("check status = %d body=%s", check.Code, check.Body.String())
	}
}

func TestModelProviderCheckAcceptsEnabledLocalProvider(t *testing.T) {
	svc := productdata.NewMemoryService()
	home := t.TempDir()
	writeHTTPRuntimeTestFile(t, filepath.Join(home, ".codex", "auth.json"), `{"tokens":{"access_token":"access-runtime-secret","refresh_token":"refresh-runtime-secret"},"base_url":"https://gateway.example.test/v1","model":"gpt-local-fixture"}`)
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, productruntime.NewGateway(svc, nil, nil))
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home}

	enable := requestJSON(t, srv, http.MethodPost, "/v1/local-provider-detections/local_codex/enable", "")
	if enable.Code != http.StatusOK {
		t.Fatalf("enable status = %d body=%s", enable.Code, enable.Body.String())
	}
	check := requestJSON(t, srv, http.MethodPost, "/v1/model-providers/check", `{"provider_id":"local_codex"}`)
	if check.Code != http.StatusOK {
		t.Fatalf("check status = %d body=%s", check.Code, check.Body.String())
	}
	if strings.Contains(check.Body.String(), "access-runtime-secret") || strings.Contains(check.Body.String(), "refresh-runtime-secret") || strings.Contains(check.Body.String(), home) {
		t.Fatalf("check leaked local secret/path: %s", check.Body.String())
	}
}

func TestModelProviderHandlerSavesLocalCustomProvider(t *testing.T) {
	svc := productdata.NewMemoryService()
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/model-providers", `{"base_url":"https://gateway.example.test/v1","model":"gpt-5.5","api_key":"secret-key"}`)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "secret-key") || !strings.Contains(res.Body.String(), `"id":"custom"`) || !strings.Contains(res.Body.String(), `"status":"available"`) {
		t.Fatalf("body = %s", res.Body.String())
	}
	listed := requestJSON(t, srv, http.MethodGet, "/v1/model-providers", "")
	if !strings.Contains(listed.Body.String(), "gpt-5.5") || strings.Contains(listed.Body.String(), "secret-key") {
		t.Fatalf("listed body = %s", listed.Body.String())
	}
}

func TestModelProviderHandlersExposeUnavailableAndMisconfigured(t *testing.T) {
	svc := productdata.NewMemoryService()
	cfg := config.Config{AppEnv: "local", ModelProviders: []config.ModelProvider{
		{ID: "disabled", Family: "openai", APIKey: "key", Model: "gpt-4.1", Enabled: false},
		{ID: "custom", Family: "openai_compatible", APIKey: "key", Model: "gpt-5.5", Enabled: true},
	}}
	srv := NewServerWithProduct(cfg, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/model-providers", "")
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	if !strings.Contains(body, `"status":"unavailable"`) || !strings.Contains(body, `"status":"misconfigured"`) {
		t.Fatalf("body = %s", body)
	}

	unavailable := requestJSON(t, srv, http.MethodPost, "/v1/model-providers/check", `{"provider_id":"disabled"}`)
	if unavailable.Code != http.StatusServiceUnavailable || !strings.Contains(unavailable.Body.String(), "provider_unavailable") {
		t.Fatalf("unavailable status = %d body=%s", unavailable.Code, unavailable.Body.String())
	}

	misconfigured := requestJSON(t, srv, http.MethodPost, "/v1/model-providers/check", `{"provider_id":"custom"}`)
	if misconfigured.Code != http.StatusBadRequest || !strings.Contains(misconfigured.Body.String(), "provider_misconfigured") {
		t.Fatalf("misconfigured status = %d body=%s", misconfigured.Code, misconfigured.Body.String())
	}
}

func TestLocalProviderDetectionHandlerReturnsSafeProviders(t *testing.T) {
	home := t.TempDir()
	writeHTTPRuntimeTestFile(t, home+"/.claude.json", `{"primaryApiKey":"sk-ant-http-secret"}`)
	writeHTTPRuntimeTestFile(t, home+"/.codex/auth.json", `{"auth_mode":"chatgpt","tokens":{"access_token":"access-http-secret","refresh_token":"refresh-http-secret"}}`)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}

	res := requestJSON(t, srv, http.MethodGet, "/v1/local-provider-detections", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, expected := range []string{"local_claude_code", "Local Claude Code", "local_codex", "Local Codex", `"redaction_applied":true`, "Explicit opt-in"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body missing %q: %s", expected, body)
		}
	}
}

func TestLocalProviderDetectionHandlerAcceptsTrailingSlash(t *testing.T) {
	home := t.TempDir()
	writeHTTPRuntimeTestFile(t, home+"/.codex/auth.json", `{"auth_mode":"chatgpt","tokens":{"access_token":"access-http-secret"}}`)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}

	res := requestJSON(t, srv, http.MethodGet, "/v1/local-provider-detections/", "")

	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), "local_codex") {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
}

func TestLocalProviderDetectionHandlerDoesNotExposeSecretsOrPrivatePaths(t *testing.T) {
	home := t.TempDir()
	writeHTTPRuntimeTestFile(t, home+"/.claude/settings.json", `{"env":{"ANTHROPIC_AUTH_TOKEN":"Bearer private-token","ANTHROPIC_BASE_URL":"https://example.test/private/path","ANTHROPIC_MODEL":"claude-test"}}`)
	writeHTTPRuntimeTestFile(t, home+"/.codex/auth.json", `{"OPENAI_API_KEY":"sk-private-codex"}`)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}

	res := requestJSON(t, srv, http.MethodGet, "/v1/local-provider-detections", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, forbidden := range []string{"sk-", "Bearer", "private-token", "access_token", "refresh_token", home, "/private/path"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("body leaked %q: %s", forbidden, body)
		}
	}
}

func TestLocalProviderDetectionHandlerReturnsDisabledAndUnsupportedStableStatuses(t *testing.T) {
	home := t.TempDir()
	writeHTTPRuntimeTestFile(t, home+"/.claude.json", `{"apiKeyHelper":"echo secret"}`)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}

	unsupported := requestJSON(t, srv, http.MethodGet, "/v1/local-provider-detections", "")
	if unsupported.Code != http.StatusOK || !strings.Contains(unsupported.Body.String(), `"status":"unsupported"`) {
		t.Fatalf("unsupported status = %d body=%s", unsupported.Code, unsupported.Body.String())
	}

	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}, Disabled: true}
	disabled := requestJSON(t, srv, http.MethodGet, "/v1/local-provider-detections", "")
	if disabled.Code != http.StatusOK || !strings.Contains(disabled.Body.String(), `"status":"disabled"`) {
		t.Fatalf("disabled status = %d body=%s", disabled.Code, disabled.Body.String())
	}
}

func TestLocalProviderEnablementRequiresExplicitOptInBeforeModelProviderList(t *testing.T) {
	home := t.TempDir()
	providerServer := newOpenAICompatibleRuntimeTestServer(t, "M20 local codex works.")
	writeHTTPRuntimeTestFile(t, home+"/.codex/auth.json", `{"auth_mode":"chatgpt","tokens":{"access_token":"access-enable-secret","refresh_token":"refresh-enable-secret"},"base_url":"`+providerServer.URL+`/v1","model":"gpt-local-fixture"}`)
	svc := productdata.NewMemoryService()
	gateway := productruntime.NewGateway(svc, nil, nil)
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, productruntime.NewBroadcaster(), nil, gateway)
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}

	before := requestJSON(t, srv, http.MethodGet, "/v1/model-providers", "")
	if strings.Contains(before.Body.String(), "local_codex") {
		t.Fatalf("detected-only provider leaked into model list: %s", before.Body.String())
	}

	enabled := requestJSON(t, srv, http.MethodPost, "/v1/local-provider-detections/local_codex/enable", "")
	if enabled.Code != http.StatusOK {
		t.Fatalf("enable status = %d body=%s", enabled.Code, enabled.Body.String())
	}
	body := enabled.Body.String()
	for _, expected := range []string{`"id":"local_codex"`, `"local_provider":true`, `"session_local":true`, `"credential_reference":"redacted"`, `"execution_state":"supported"`, `"status":"available"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("enabled body missing %q: %s", expected, body)
		}
	}
	for _, forbidden := range []string{"access-enable-secret", "refresh-enable-secret", "access_token", "refresh_token", home} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("enabled body leaked %q: %s", forbidden, body)
		}
	}

	listed := requestJSON(t, srv, http.MethodGet, "/v1/model-providers", "")
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), `"id":"local_codex"`) || strings.Contains(listed.Body.String(), "access-enable-secret") {
		t.Fatalf("listed status = %d body=%s", listed.Code, listed.Body.String())
	}

	disabled := requestJSON(t, srv, http.MethodDelete, "/v1/local-provider-detections/local_codex/enable", "")
	if disabled.Code != http.StatusOK {
		t.Fatalf("disable status = %d body=%s", disabled.Code, disabled.Body.String())
	}
	after := requestJSON(t, srv, http.MethodGet, "/v1/model-providers", "")
	if strings.Contains(after.Body.String(), "local_codex") {
		t.Fatalf("disabled provider remained in model list: %s", after.Body.String())
	}

	thread := createRuntimeTestThread(t, svc)
	message, _, err := svc.CreateMessage(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.CreateMessageInput{Content: "disabled local codex"})
	if err != nil {
		t.Fatal(err)
	}
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"message_id":"`+message.ID+`","source":"model_gateway","provider_id":"local_codex","model":"gpt-local-fixture"}`)
	if runRes.Code != http.StatusServiceUnavailable || !strings.Contains(runRes.Body.String(), "not enabled") {
		t.Fatalf("disabled local codex start status = %d body=%s", runRes.Code, runRes.Body.String())
	}
}

func TestLocalCodexEnabledProviderRunsThroughGatewayWorker(t *testing.T) {
	home := t.TempDir()
	providerServer := newOpenAICompatibleRuntimeTestServer(t, "Local Codex assistant reply.")
	writeHTTPRuntimeTestFile(t, home+"/.codex/auth.json", `{"auth_mode":"chatgpt","tokens":{"access_token":"access-chat-secret","refresh_token":"refresh-chat-secret"},"base_url":"`+providerServer.URL+`/v1","model":"gpt-local-fixture"}`)
	svc := productdata.NewMemoryService()
	broadcaster := productruntime.NewBroadcaster()
	gateway := productruntime.NewGateway(svc, broadcaster, nil)
	worker := productruntime.NewWorker(svc, broadcaster, productruntime.QueuedRunRouter{Gateway: gateway})
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, broadcaster, nil, gateway)
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Local Codex", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello local codex"})
	if err != nil {
		t.Fatal(err)
	}

	enabled := requestJSON(t, srv, http.MethodPost, "/v1/local-provider-detections/local_codex/enable", "")
	if enabled.Code != http.StatusOK || !strings.Contains(enabled.Body.String(), `"execution_state":"supported"`) || !strings.Contains(enabled.Body.String(), `"status":"available"`) {
		t.Fatalf("enable status = %d body=%s", enabled.Code, enabled.Body.String())
	}
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"message_id":"`+message.ID+`","source":"model_gateway","provider_id":"local_codex","model":"gpt-local-fixture"}`)
	if runRes.Code != http.StatusAccepted {
		t.Fatalf("run status = %d body=%s", runRes.Code, runRes.Body.String())
	}
	var runBody struct {
		Run productdata.Run `json:"run"`
	}
	if err := json.Unmarshal(runRes.Body.Bytes(), &runBody); err != nil {
		t.Fatal(err)
	}
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("worker ok=%v err=%v", ok, err)
	}

	events := waitForRuntimeTestEvents(t, svc, runBody.Run.ID, productdata.EventRunCompleted)
	for _, want := range []string{"model_request_started", "model_output_delta", productdata.EventRunCompleted} {
		if !runtimeTestEventsContain(events, want) {
			t.Fatalf("events missing %s: %+v", want, events)
		}
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "Local Codex assistant reply." {
		t.Fatalf("messages = %+v", messages)
	}
	assertRuntimeTestNoLeak(t, enabled.Body.String()+runRes.Body.String()+runtimeTestEventsJSON(t, events)+runtimeTestMessagesJSON(t, messages), "access-chat-secret", "refresh-chat-secret", "Authorization", home)
}

func TestLocalCodexProviderFailureDoesNotFabricateAssistantMessage(t *testing.T) {
	home := t.TempDir()
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":{"type":"authentication_error"}}`, http.StatusUnauthorized)
	}))
	t.Cleanup(providerServer.Close)
	writeHTTPRuntimeTestFile(t, home+"/.codex/auth.json", `{"OPENAI_API_KEY":"sk-failure-secret","base_url":"`+providerServer.URL+`/v1"}`)
	svc := productdata.NewMemoryService()
	gateway := productruntime.NewGateway(svc, nil, nil)
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, productruntime.NewBroadcaster(), nil, gateway)
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}
	thread := createRuntimeTestThread(t, svc)
	message, _, err := svc.CreateMessage(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if res := requestJSON(t, srv, http.MethodPost, "/v1/local-provider-detections/local_codex/enable", ""); res.Code != http.StatusOK {
		t.Fatalf("enable status=%d body=%s", res.Code, res.Body.String())
	}
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"message_id":"`+message.ID+`","source":"model_gateway","provider_id":"local_codex"}`)
	if runRes.Code != http.StatusAccepted {
		t.Fatalf("run status=%d body=%s", runRes.Code, runRes.Body.String())
	}
	var runBody struct {
		Run productdata.Run `json:"run"`
	}
	if err := json.Unmarshal(runRes.Body.Bytes(), &runBody); err != nil {
		t.Fatal(err)
	}
	if ok, err := worker.ProcessOne(context.Background()); !ok || err == nil {
		t.Fatalf("worker ok=%v err=%v", ok, err)
	}
	events := waitForRuntimeTestEvents(t, svc, runBody.Run.ID, "run_failed")
	messages, err := svc.ListMessages(context.Background(), identity.LocalDevIdentity(), thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("fabricated assistant message: %+v", messages)
	}
	assertRuntimeTestNoLeak(t, runtimeTestEventsJSON(t, events), "sk-failure-secret", home)
}

func TestLocalProviderConcurrentEnableDisableListSaveAndCheck(t *testing.T) {
	home := t.TempDir()
	providerServer := newOpenAICompatibleRuntimeTestServer(t, "ok")
	writeHTTPRuntimeTestFile(t, home+"/.codex/auth.json", `{"OPENAI_API_KEY":"sk-concurrent-secret","base_url":"`+providerServer.URL+`/v1"}`)
	svc := productdata.NewMemoryService()
	gateway := productruntime.NewGateway(svc, nil, nil)
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, productruntime.NewBroadcaster(), nil, gateway)
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}

	errs := make(chan string, 80)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 8; j++ {
				for _, res := range []httptest.ResponseRecorder{
					*requestJSON(t, srv, http.MethodPost, "/v1/local-provider-detections/local_codex/enable", ""),
					*requestJSON(t, srv, http.MethodGet, "/v1/model-providers", ""),
					*requestJSON(t, srv, http.MethodPost, "/v1/model-providers", `{"base_url":"https://gateway.example.test/v1","model":"gpt-5.5","api_key":"secret-key"}`),
					*requestJSON(t, srv, http.MethodPost, "/v1/model-providers/check", `{"provider_id":"custom"}`),
				} {
					if strings.Contains(res.Body.String(), "sk-concurrent-secret") || strings.Contains(res.Body.String(), home) {
						errs <- res.Body.String()
					}
				}
				_ = requestJSON(t, srv, http.MethodDelete, "/v1/local-provider-detections/local_codex/enable", "")
			}
			errs <- ""
		}()
	}
	for i := 0; i < 10; i++ {
		if err := <-errs; err != "" {
			t.Fatalf("leak during concurrent provider operations: %s", err)
		}
	}
}

func TestLocalProviderEnablementRejectsUnavailableUnsupportedAndClaudeCode(t *testing.T) {
	home := t.TempDir()
	writeHTTPRuntimeTestFile(t, home+"/.claude.json", `{"primaryApiKey":"sk-ant-enable-secret"}`)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	srv.localProviderDetectionInput = productruntime.LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}}

	claude := requestJSON(t, srv, http.MethodPost, "/v1/local-provider-detections/local_claude_code/enable", "")
	if claude.Code != http.StatusBadRequest || !strings.Contains(claude.Body.String(), "unsupported") {
		t.Fatalf("claude enable status = %d body=%s", claude.Code, claude.Body.String())
	}

	codex := requestJSON(t, srv, http.MethodPost, "/v1/local-provider-detections/local_codex/enable", "")
	if codex.Code != http.StatusServiceUnavailable || !strings.Contains(codex.Body.String(), "not available") {
		t.Fatalf("codex enable status = %d body=%s", codex.Code, codex.Body.String())
	}

	listed := requestJSON(t, srv, http.MethodGet, "/v1/model-providers", "")
	if strings.Contains(listed.Body.String(), "local_claude_code") || strings.Contains(listed.Body.String(), "local_codex") || strings.Contains(listed.Body.String(), "sk-ant-enable-secret") {
		t.Fatalf("rejected provider leaked into model list: %s", listed.Body.String())
	}
}

func TestStartRunRejectsUnsupportedEnabledLocalProvider(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	srv.localProviderEnablements["local_claude_code"] = productruntime.LocalProviderCapability{ProviderID: "local_claude_code", DisplayName: "Local Claude Code", ProviderKind: productruntime.LocalProviderKindClaudeCode, AuthMode: productruntime.LocalProviderAuthModeAPIKey, Status: productruntime.LocalProviderStatusAvailable, ModelCandidates: []string{"claude-sonnet-4-5"}, Source: productruntime.LocalProviderSourceLocalConfig, RedactionApplied: true}

	res := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"source":"model_gateway","provider_id":"local_claude_code","model":"claude-sonnet-4-5"}`)
	if res.Code != http.StatusServiceUnavailable || !strings.Contains(res.Body.String(), "execution is unsupported") {
		t.Fatalf("start run status = %d body=%s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "access-run-secret") {
		t.Fatalf("start run leaked local secret/path: %s", res.Body.String())
	}
}

func TestStartRunHandlerCreatesModelGatewayRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"message_id":"msg_1","source":"model_gateway","provider_id":"custom","model":"gpt-5.5"}`)

	if res.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Run productdata.Run `json:"run"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Run.Source != productdata.RunSourceModelGateway || body.Run.Title != "Model gateway run" {
		t.Fatalf("run = %+v", body.Run)
	}
}

func writeHTTPRuntimeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestStartRunHandlerAcceptsPersonaOverride(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default persona",
		SystemPrompt:     "secret prompt",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "persona-model"},
		AllowedToolNames: []string{productdata.ToolNameCurrentTime},
		ReasoningMode:    "balanced",
		BudgetSummary:    "budget",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	personas, err := svc.ListPersonas(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	thread := createRuntimeTestThread(t, svc)
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"message_id":"`+message.ID+`","source":"model_gateway","provider_id":"custom","model":"fallback","persona_id":"`+personas[0].ID+`"}`)

	if res.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Run productdata.Run `json:"run"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Run.PersonaID != personas[0].ID {
		t.Fatalf("run = %+v", body.Run)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_persona", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	context, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if context.Persona.ID != personas[0].ID || context.ProviderRoute.Model != "persona-model" {
		t.Fatalf("context = %+v", context)
	}
}

func TestStartRunHandlerQueuesModelGatewayRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	message, _, err := svc.CreateMessage(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs", `{"message_id":"`+message.ID+`","source":"model_gateway","provider_id":"custom"}`)
	if res.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Run productdata.Run `json:"run"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Run.Status != productdata.RunStatusQueued {
		t.Fatalf("run = %+v", body.Run)
	}
	events, err := svc.ListRunEvents(context.Background(), identity.LocalDevIdentity(), body.Run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[1].Type != productdata.EventRunQueued {
		t.Fatalf("events = %+v", events)
	}
}

func TestRunEventHistoryHandlerReturnsOrderedEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryProgress, "context_loaded")
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events?after_sequence=1", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Events []productdata.RunEvent `json:"events"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Events) != 3 || body.Events[0].Sequence != 2 || body.Events[1].Sequence != 3 || body.Events[2].Sequence != 4 {
		t.Fatalf("events = %+v", body.Events)
	}
}

func TestRunEventStreamDeliversHistoryBeforeCloseMarker(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryProgress, "context_loaded")
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events/stream", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	if !strings.Contains(res.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("content-type = %q", res.Header().Get("Content-Type"))
	}
	if strings.Index(body, "run_created") > strings.Index(body, "context_loaded") {
		t.Fatalf("history order body=%s", body)
	}
	if !strings.Contains(body, "event: stream_closed") {
		t.Fatalf("missing close marker body=%s", body)
	}
}

func TestRunEventStreamReconnectUsesAfterSequence(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryProgress, "context_loaded")
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events/stream?after_sequence=1", "")

	body := res.Body.String()
	if strings.Contains(body, "run_created") || !strings.Contains(body, "context_loaded") {
		t.Fatalf("reconnect body=%s", body)
	}
}

func TestRunEventStreamReplaysApprovalRequiredToolEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	if _, _, err := svc.RecordToolCallRequest(context.Background(), identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: productdata.ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithRuntime(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events/stream?after_sequence=2", "")

	body := res.Body.String()
	requested := strings.Index(body, productdata.EventToolCallRequested)
	required := strings.Index(body, productdata.EventToolCallApprovalRequired)
	if requested < 0 || required < 0 || requested > required {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, `"tool_call_id":"tc_1"`) || !strings.Contains(body, `"execution_status":"blocked"`) {
		t.Fatalf("body=%s", body)
	}
}

func TestScopedToolCallHandlerReturnsProjection(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	run, err := svc.StartRun(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: productdata.ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+thread.ID+"/runs/"+run.ID+"/tool-calls/tc_1", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"tool_call_id":"tc_1"`) || !strings.Contains(res.Body.String(), `"execution_status":"blocked"`) {
		t.Fatalf("body=%s", res.Body.String())
	}
}

func TestScopedToolCallHandlerReturnsMCPProjection(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	run, err := svc.StartRun(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{
		ToolCallID:          "tc_mcp",
		ToolName:            "mcp.local-search.search",
		CandidateSchemaHash: "sha256:test-local-search",
		ArgumentsSummary:    map[string]any{"query": "public", "api_key": "[redacted]"},
		ArgumentsHash:       "hash_mcp",
		ApprovalStatus:      productdata.ToolCallApprovalRequired,
		ExecutionStatus:     productdata.ToolCallExecutionBlocked,
	}); err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+thread.ID+"/runs/"+run.ID+"/tool-calls/tc_mcp", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	if !strings.Contains(body, `"tool_name":"mcp.local-search.search"`) || !strings.Contains(body, `"query":"public"`) || !strings.Contains(body, `"api_key":"[redacted]"`) {
		t.Fatalf("body=%s", body)
	}
	if strings.Contains(body, "secret") {
		t.Fatalf("unredacted MCP projection: %s", body)
	}
}

func TestToolCallApproveDenyHandlersAreIdempotentAndScoped(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	run, err := svc.StartRun(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_approve", ToolName: productdata.ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	path := "/v1/threads/" + thread.ID + "/runs/" + run.ID + "/tool-calls/tc_approve/approve"
	res := requestJSON(t, srv, http.MethodPost, path, "")
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"approval_status":"approved"`) || !strings.Contains(res.Body.String(), `"execution_status":"not_started"`) {
		t.Fatalf("approve status=%d body=%s", res.Code, res.Body.String())
	}
	again := requestJSON(t, srv, http.MethodPost, path, "")
	if again.Code != http.StatusOK || !strings.Contains(again.Body.String(), `"approval_status":"approved"`) {
		t.Fatalf("approve retry status=%d body=%s", again.Code, again.Body.String())
	}
	wrongScope := requestJSON(t, srv, http.MethodPost, "/v1/threads/wrong/runs/"+run.ID+"/tool-calls/tc_approve/approve", "")
	if wrongScope.Code != http.StatusNotFound {
		t.Fatalf("wrong scope status=%d body=%s", wrongScope.Code, wrongScope.Body.String())
	}
}

func TestToolCallDenyHandlerStopsRunWithoutExecution(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	run, err := svc.StartRun(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_deny", ToolName: productdata.ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	path := "/v1/threads/" + thread.ID + "/runs/" + run.ID + "/tool-calls/tc_deny/deny"
	res := requestJSON(t, srv, http.MethodPost, path, "")
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"approval_status":"denied"`) {
		t.Fatalf("deny status=%d body=%s", res.Code, res.Body.String())
	}
	again := requestJSON(t, srv, http.MethodPost, path, "")
	if again.Code != http.StatusOK || !strings.Contains(again.Body.String(), `"approval_status":"denied"`) {
		t.Fatalf("deny retry status=%d body=%s", again.Code, again.Body.String())
	}
	gotRun, err := svc.GetRun(context.Background(), identity.LocalDevIdentity(), run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotRun.Status != productdata.RunStatusStopped {
		t.Fatalf("run = %+v", gotRun)
	}
	events, err := svc.ListRunEvents(context.Background(), identity.LocalDevIdentity(), run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == productdata.EventToolCallExecuting {
			t.Fatalf("denied run executed: %+v", events)
		}
	}
}

func TestRunEventStreamReplaysToolApprovalExecutionAndResultEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	run, err := svc.StartRun(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: productdata.ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), identity.LocalDevIdentity(), thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.StartToolCallExecution(context.Background(), identity.LocalDevIdentity(), thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.CompleteToolCallSuccess(context.Background(), identity.LocalDevIdentity(), thread.ID, run.ID, "tc_1", map[string]any{"timezone": "UTC"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), identity.LocalDevIdentity(), run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: productdata.EventRunCompleted, Summary: "Run completed"}); err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/v1/runs/"+run.ID+"/events/stream", nil).WithContext(ctx)
	res := httptest.NewRecorder()
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	srv.ServeHTTP(res, req)

	body := res.Body.String()
	approved := strings.Index(body, productdata.EventToolCallApproved)
	executing := strings.Index(body, productdata.EventToolCallExecuting)
	succeeded := strings.Index(body, productdata.EventToolCallSucceeded)
	if approved < 0 || executing < 0 || succeeded < 0 || !(approved < executing && executing < succeeded) {
		t.Fatalf("body=%s", body)
	}
	if !strings.Contains(body, `"result_summary":{"timezone":"UTC"}`) {
		t.Fatalf("body=%s", body)
	}
}

func TestRunEventStreamSubscribesBeforeHistoryRead(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	broadcaster := productruntime.NewBroadcaster()
	srv := NewServerWithRuntime(config.Config{AppEnv: "local"}, fakeChecker{}, publishDuringListService{Service: svc, broadcaster: broadcaster}, broadcaster, nil)

	res := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+run.ID+"/events/stream", "")

	body := res.Body.String()
	if !strings.Contains(body, "context_loaded") || !strings.Contains(body, "stream_closed") {
		t.Fatalf("body = %s", body)
	}
}

func TestRunEventStreamFlushesHistoryAndCloseMarker(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	res := newFlushRecorder()
	srv.ServeHTTP(res, httptestRequest(http.MethodGet, "/v1/runs/"+run.ID+"/events/stream", ""))

	if res.flushes == 0 {
		t.Fatal("flushes = 0, want at least one flush for history/close marker")
	}
}

func TestStopRunHandlerStopsQueuedBackgroundRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread := createRuntimeTestThread(t, svc)
	run, err := svc.StartRun(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/runs/"+run.ID+"/stop", "")
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"result":"stopped"`) || !strings.Contains(res.Body.String(), `"status":"stopped"`) {
		t.Fatalf("body = %s", res.Body.String())
	}
	if _, _, ok, err := svc.ClaimBackgroundJob(context.Background(), identity.LocalDevIdentity(), productdata.ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 1}); err != nil || ok {
		t.Fatalf("claim after stop ok=%v err=%v", ok, err)
	}
}

func TestStopRunHandlerPublishesStopEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	broadcaster := productruntime.NewBroadcaster()
	srv := NewServerWithRuntime(config.Config{AppEnv: "local"}, fakeChecker{}, svc, broadcaster, nil)
	events := broadcaster.Subscribe(context.Background(), run.ID)

	res := requestJSON(t, srv, http.MethodPost, "/v1/runs/"+run.ID+"/stop", "")
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	first := <-events
	second := <-events
	if first.Type != productdata.EventStopRequested || second.Category != productdata.RunEventCategoryFinal {
		t.Fatalf("published events = %+v %+v", first, second)
	}
}

func TestStopRunHandlerReturnsAlreadyTerminalForCompletedRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	run := createRuntimeTestRun(t, svc)
	appendRuntimeTestEvent(t, svc, run.ID, productdata.RunEventCategoryFinal, "run_completed")
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodPost, "/v1/runs/"+run.ID+"/stop", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "already_terminal") {
		t.Fatalf("body = %s", res.Body.String())
	}
}

type publishDuringListService struct {
	productdata.Service
	broadcaster *productruntime.Broadcaster
	published   bool
}

func (s publishDuringListService) ListRunEvents(ctx context.Context, ident identity.LocalIdentity, runID string, afterSequence int) ([]productdata.RunEvent, error) {
	events, err := s.Service.ListRunEvents(ctx, ident, runID, afterSequence)
	if err != nil || s.published {
		return events, err
	}
	event, err := s.Service.AppendRunEvent(ctx, ident, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: "context_loaded", Summary: "Context loaded"})
	if err != nil {
		return events, err
	}
	final, err := s.Service.AppendRunEvent(ctx, ident, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"})
	if err != nil {
		return events, err
	}
	s.broadcaster.Publish(event)
	s.broadcaster.Publish(final)
	return events, nil
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushes int
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func (r *flushRecorder) Flush() {
	r.flushes++
}

func httptestRequest(method string, path string, body string) *http.Request {
	if body == "" {
		return httptest.NewRequest(method, path, nil)
	}
	return httptest.NewRequest(method, path, bytes.NewBufferString(body))
}

func createRuntimeTestThread(t *testing.T, svc *productdata.MemoryService) productdata.Thread {
	t.Helper()
	thread, err := svc.CreateThread(context.Background(), identity.LocalDevIdentity(), productdata.CreateThreadInput{Title: "Runtime", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	return thread
}

func createRuntimeTestRun(t *testing.T, svc *productdata.MemoryService) productdata.Run {
	t.Helper()
	thread := createRuntimeTestThread(t, svc)
	run, err := svc.StartRun(context.Background(), identity.LocalDevIdentity(), thread.ID, productdata.StartRunInput{ScriptName: "m4_smoke"})
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func appendRuntimeTestEvent(t *testing.T, svc *productdata.MemoryService, runID string, category productdata.RunEventCategory, eventType string) productdata.RunEvent {
	t.Helper()
	event, err := svc.AppendRunEvent(context.Background(), identity.LocalDevIdentity(), runID, productdata.AppendRunEventInput{Category: category, Type: eventType, Summary: eventType})
	if err != nil {
		t.Fatal(err)
	}
	return event
}

func waitForRuntimeTestEvents(t *testing.T, svc *productdata.MemoryService, runID string, eventType string) []productdata.RunEvent {
	t.Helper()
	var last []productdata.RunEvent
	for attempt := 0; attempt < 100; attempt++ {
		events, err := svc.ListRunEvents(context.Background(), identity.LocalDevIdentity(), runID, 0)
		if err != nil {
			t.Fatal(err)
		}
		last = events
		for _, event := range events {
			if event.Type == eventType {
				return events
			}
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("event %s not found; events = %+v", eventType, last)
	return nil
}

func newOpenAICompatibleRuntimeTestServer(t *testing.T, responseText string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" && r.URL.Path != "/v1/responses" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Fatalf("missing authorization header")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		if r.URL.Path == "/v1/responses" {
			_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"" + responseText + "\"}\n\n"))
			_, _ = w.Write([]byte("data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\"}}\n\n"))
			return
		}
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"" + responseText + "\"},\"finish_reason\":\"stop\"}]}\n\n"))
	}))
	t.Cleanup(server.Close)
	return server
}

func runtimeTestEventsContain(events []productdata.RunEvent, eventType string) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func runtimeTestEventsJSON(t *testing.T, events []productdata.RunEvent) string {
	t.Helper()
	raw, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func runtimeTestMessagesJSON(t *testing.T, messages []productdata.Message) string {
	t.Helper()
	raw, err := json.Marshal(messages)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func assertRuntimeTestNoLeak(t *testing.T, body string, forbidden ...string) {
	t.Helper()
	for _, value := range forbidden {
		if value != "" && strings.Contains(body, value) {
			t.Fatalf("body leaked %q: %s", value, body)
		}
	}
}
