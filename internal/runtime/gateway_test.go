package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestGatewayPersistsProviderDeltasAndCompletion(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: "hel"}, {Type: ProviderEventTextDelta, Text: "lo"}, {Type: ProviderEventCompleted}}}
	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if events[len(events)-1].Type != "run_completed" {
		t.Fatalf("events = %+v", events)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "hello" {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestLocalCodexProviderProducesAssistantMessageThroughGateway(t *testing.T) {
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access-runtime-secret" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("ChatGPT-Account-ID") != "account_runtime" {
			t.Fatalf("account header = %q", r.Header.Get("ChatGPT-Account-ID"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"Local \"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"Codex\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\"}}\n\n"))
	}))
	t.Cleanup(providerServer.Close)
	home := t.TempDir()
	writeRuntimeGatewayTestFile(t, filepath.Join(home, ".codex", "auth.json"), `{"auth_mode":"chatgpt","tokens":{"access_token":"access-runtime-secret","refresh_token":"refresh-runtime-secret","account_id":"account_runtime"},"base_url":"`+providerServer.URL+`/v1","model":"gpt-local-fixture"}`)
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "local_codex", Model: "gpt-local-fixture"})
	if err != nil {
		t.Fatal(err)
	}
	provider := NewLocalCodexProvider(LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}})

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "local_codex"})

	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Content != "Local Codex" {
		t.Fatalf("messages = %+v", messages)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	body := runtimeGatewayEventsJSON(t, events) + runtimeGatewayMessagesJSON(t, messages)
	for _, forbidden := range []string{"access-runtime-secret", "refresh-runtime-secret", home, "Authorization"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("leaked %q: %s", forbidden, body)
		}
	}
}

func TestGatewayRemoveProviderPreventsSelection(t *testing.T) {
	gateway := NewGateway(productdata.NewMemoryService(), nil, []Provider{
		StaticProvider{ProviderConfig: ProviderConfig{ID: "local_codex", Family: ProviderFamilyOpenAICompatible, APIKey: "redacted", Model: "gpt-local-fixture", Enabled: true}},
	})

	gateway.RemoveProvider("local_codex")

	if _, err := gateway.selectProvider("local_codex"); err == nil {
		t.Fatal("expected removed provider to be unavailable")
	}
}

func TestGatewayRecordsApprovalRequiredCurrentTimeToolCall(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_1"}}, {Type: ProviderEventCompleted, Text: "done"}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	wantTypes := []string{productdata.EventToolCallRequested, productdata.EventToolCallApprovalRequired}
	for i, want := range wantTypes {
		if events[len(events)-2+i].Type != want {
			t.Fatalf("events = %+v", events)
		}
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusBlockedOnToolApproval {
		t.Fatalf("run = %+v", got)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameCurrentTime || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("call = %+v", call)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestGatewayAutoApprovesWebSearchToolCall(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search ai news"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWebSearch}}}); err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWebSearch, Metadata: map[string]any{"tool_call_id": "tc_search_1", "arguments_summary": map[string]any{"query": "latest ai news", "provider": "brave", "limit": 5}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusQueued {
		t.Fatalf("run = %+v", got)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_search_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWebSearch || call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
		t.Fatalf("call = %+v", call)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if events[len(events)-1].Type != productdata.EventToolCallApproved {
		t.Fatalf("events = %+v", events)
	}
}

func TestGatewayRejectsInvalidToolArgumentsWithoutFabricatingDefaults(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_bad", "arguments_summary": map[string]any{"timezone": "Asia/Shanghai"}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_bad"); err == nil {
		t.Fatal("GetToolCall() error = nil, want invalid tool argument rejection")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRejectsWorkspaceMutationToolInChatMode(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceWriteFile, Metadata: map[string]any{"tool_call_id": "tc_write_chat", "arguments_summary": map[string]any{"path": "chat.txt", "content": "chat\n"}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_write_chat"); err == nil {
		t.Fatal("workspace mutation tool call was recorded in chat mode")
	}
	if _, err := os.Stat(filepath.Join(root, "chat.txt")); err == nil {
		t.Fatal("workspace mutation wrote file in chat mode")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_call_rejected" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRejectsSandboxExecCommandInChatMode(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameSandboxExecCommand, Metadata: map[string]any{"tool_call_id": "tc_exec_chat", "arguments_summary": map[string]any{"argv": []any{"touch", "chat-exec.txt"}}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_exec_chat"); err == nil {
		t.Fatal("sandbox exec tool call was recorded in chat mode")
	}
	if _, err := os.Stat(filepath.Join(root, "chat-exec.txt")); err == nil {
		t.Fatal("sandbox exec command ran in chat mode")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_call_rejected" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRejectsLSPToolInChatMode(t *testing.T) {
	root := createLSPFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameLSPSymbols, Metadata: map[string]any{"tool_call_id": "tc_lsp_chat", "arguments_summary": map[string]any{"path": "src/main.go", "query": "Tool"}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_lsp_chat"); err == nil {
		t.Fatal("lsp tool call was recorded in chat mode")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_call_rejected" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRejectsAgentToolInChatMode(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, productdata.BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameAgentSpawn, Metadata: map[string]any{"tool_call_id": "tc_agent_chat", "arguments_summary": map[string]any{"role": "reviewer", "goal": "Review chat"}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_agent_chat"); err == nil {
		t.Fatal("agent tool call was recorded in chat mode")
	}
	tasks, err := svc.ListAgentTasks(context.Background(), ident, productdata.ListAgentTasksInput{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Fatalf("agent tasks created in chat mode: %+v", tasks)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_call_rejected" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayMapsProviderFailureToRedactedRunFailure(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventRateLimited}}}
	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "provider_rate_limited" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayLoadsCurrentThreadContextThroughTriggerMessage(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	otherThread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Other", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.CreateMessage(context.Background(), ident, otherThread.ID, productdata.CreateMessageInput{Content: "do not include"}); err != nil {
		t.Fatal(err)
	}
	first, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, productdata.AppendAssistantMessageInput{Content: "hi there"}); err != nil {
		t.Fatal(err)
	}
	current, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "continue"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: current.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: current.ID, ProviderID: "custom"})

	if provider.request.ThreadID != thread.ID || provider.request.MessageID != current.ID {
		t.Fatalf("request = %+v", provider.request)
	}
	want := []ProviderMessage{{Role: "user", Content: first.Content}, {Role: "assistant", Content: "hi there"}, {Role: "user", Content: current.Content}}
	if len(provider.request.Messages) != len(want) {
		t.Fatalf("messages = %+v", provider.request.Messages)
	}
	for i := range want {
		if provider.request.Messages[i].Role != want[i].Role || provider.request.Messages[i].Content != want[i].Content {
			t.Fatalf("messages = %+v", provider.request.Messages)
		}
	}
}

func TestGatewayBuildsContinuationContextFromToolResultEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "What time is it?"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameCurrentTime, "arguments_summary": map[string]any{"timezone": "UTC"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameCurrentTime, "result_summary": map[string]any{"iso_time": "2026-05-25T10:00:00Z", "timezone": "UTC", "source": "runtime", "api_key": "sk-secret"}}}); err != nil {
		t.Fatal(err)
	}
	gateway := NewGateway(svc, nil, nil)

	messages, err := gateway.loadContinuationMessages(context.Background(), thread.ID, message.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 3 {
		t.Fatalf("messages = %+v", messages)
	}
	if messages[1].Role != ProviderMessageRoleAssistantToolCall || messages[1].ToolCallID != "tc_1" || messages[1].ToolName != productdata.ToolNameCurrentTime {
		t.Fatalf("tool call message = %+v", messages[1])
	}
	if messages[2].Role != ProviderMessageRoleToolResult || messages[2].ToolCallID != "tc_1" {
		t.Fatalf("tool result message = %+v", messages[2])
	}
	if messages[2].Content == "" || messages[2].Content == "sk-secret" {
		t.Fatalf("tool result content = %q", messages[2].Content)
	}
}

func TestGatewayBuildsMCPContinuationContextFromRedactedToolResult(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "MCP continuation", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Search"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_mcp_1", "tool_name": "mcp.local-search.search", "tool_source": "mcp", "arguments_summary": map[string]any{"query": "status"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_mcp_1", "tool_name": "mcp.local-search.search", "tool_source": "mcp", "result_summary": map[string]any{"raw": "sk-secret"}, "result_for_model_redacted": map[string]any{"summary": "safe"}}}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", ToolCallID: "tc_mcp_1"})

	if len(provider.request.Messages) != 3 {
		t.Fatalf("messages = %+v", provider.request.Messages)
	}
	toolResult := provider.request.Messages[2]
	if toolResult.Role != ProviderMessageRoleToolResult || toolResult.ToolName != "mcp.local-search.search" || !strings.Contains(toolResult.Content, "safe") || strings.Contains(toolResult.Content, "sk-secret") {
		t.Fatalf("tool result = %+v", toolResult)
	}
}

func TestGatewayContinuesAfterToolResultAndPersistsFinalAssistant(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "What time is it?"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameCurrentTime, "arguments_summary": map[string]any{"timezone": "UTC"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameCurrentTime, "result_summary": map[string]any{"iso_time": "2026-05-25T10:00:00Z", "timezone": "UTC", "source": "runtime"}}}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: "It is "}, {Type: ProviderEventCompleted, Text: "It is 2026-05-25T10:00:00Z."}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_1"})

	if provider.request.Messages[1].Role != ProviderMessageRoleAssistantToolCall || provider.request.Messages[2].Role != ProviderMessageRoleToolResult {
		t.Fatalf("request = %+v", provider.request.Messages)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var succeeded, continuationDelta, completed bool
	for _, event := range events {
		if event.Type == productdata.EventToolCallSucceeded {
			succeeded = true
		}
		if succeeded && event.Type == "model_output_delta" && event.Metadata["model_phase"] == "continuation" {
			continuationDelta = true
		}
		if continuationDelta && event.Type == productdata.EventRunCompleted {
			completed = true
		}
	}
	if !completed {
		t.Fatalf("events = %+v", events)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "It is 2026-05-25T10:00:00Z." {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestGatewayFailsWhenContinuationRequestsAnotherTool(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "What time is it?"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameCurrentTime, "arguments_summary": map[string]any{"timezone": "UTC"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameCurrentTime, "result_summary": map[string]any{"iso_time": "2026-05-25T10:00:00Z", "timezone": "UTC", "source": "runtime"}}}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_2"}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_1"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "unsupported_tool_loop" {
		t.Fatalf("run = %+v", got)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestGatewayContinuationRequestsSecondWorkspaceToolForApproval(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace loop", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Inspect workspace"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	enabled := []string{productdata.ToolNameCurrentTime, productdata.ToolNameWorkspaceGlob, productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceRead}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": enabled}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_glob_1", "tool_name": productdata.ToolNameWorkspaceGlob, "arguments_summary": map[string]any{"pattern": "*.go", "limit": 10}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_glob_1", "tool_name": productdata.ToolNameWorkspaceGlob, "result_summary": map[string]any{"matches": []any{"internal/runtime/gateway.go"}}}}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_2", "arguments_summary": map[string]any{"path": "internal/runtime/gateway.go", "limit": 512}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_glob_1"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusBlockedOnToolApproval {
		t.Fatalf("run = %+v", got)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_2")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWorkspaceRead || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("call = %+v", call)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestGatewayContinuationStopsAtWorkspaceToolLoopLimit(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace limit", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Inspect workspace"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	enabled := []string{productdata.ToolNameWorkspaceGlob, productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceRead}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": enabled}}); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < productdata.DefaultMaxBoundedToolCallsPerRun; index++ {
		toolCallID := fmt.Sprintf("tc_%d", index+1)
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": toolCallID, "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": "safe.txt", "limit": 128}, "loop_index": index + 1}}); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": toolCallID, "tool_name": productdata.ToolNameWorkspaceRead, "result_summary": map[string]any{"path": "safe.txt", "truncated": false}}}); err != nil {
			t.Fatal(err)
		}
	}
	overLimitToolCallID := fmt.Sprintf("tc_%d", productdata.DefaultMaxBoundedToolCallsPerRun+1)
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": overLimitToolCallID, "arguments_summary": map[string]any{"path": "safe.txt", "limit": 128}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: fmt.Sprintf("tc_%d", productdata.DefaultMaxBoundedToolCallsPerRun)})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_loop_limit_reached" {
		t.Fatalf("run = %+v", got)
	}
	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, overLimitToolCallID); err == nil {
		t.Fatalf("over-limit tool call was recorded")
	}
}

func TestGatewayContinuationOmitsToolsAtLoopLimitSoProviderCanFinalize(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace limit final", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Inspect workspace"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	enabled := []string{productdata.ToolNameWorkspaceRead}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": enabled}}); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < productdata.DefaultMaxBoundedToolCallsPerRun; index++ {
		toolCallID := fmt.Sprintf("tc_%d", index+1)
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": toolCallID, "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": "safe.txt", "limit": 128}, "loop_index": index + 1}}); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": toolCallID, "tool_name": productdata.ToolNameWorkspaceRead, "result_summary": map[string]any{"path": "safe.txt", "truncated": false}}}); err != nil {
			t.Fatal(err)
		}
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Final answer from gathered tool results."}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: fmt.Sprintf("tc_%d", productdata.DefaultMaxBoundedToolCallsPerRun)})

	if len(provider.request.Tools) != 0 {
		t.Fatalf("continuation tools = %+v", provider.request.Tools)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Content != "Final answer from gathered tool results." {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestGatewayExposesCodeAgentToolsToProvider(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Code tools", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Patch and verify"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	enabled := []string{productdata.ToolNameLoadTools, productdata.ToolNameLoadSkill, productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspaceEdit, productdata.ToolNameWorkspacePatchPreview, productdata.ToolNameWorkspacePatchApply, productdata.ToolNameSandboxExecCommand, productdata.ToolNameSandboxStartProcess, productdata.ToolNameSandboxContinueProcess, productdata.ToolNameSandboxTerminateProcess, productdata.ToolNameLSPSymbols, productdata.ToolNameLSPDefinition, productdata.ToolNameLSPHover, productdata.ToolNameWebSearch, productdata.ToolNameBrowserOpen, productdata.ToolNameBrowserScreenshot, productdata.ToolNameBrowserType, productdata.ToolNameBrowserPress, productdata.ToolNameTodoWrite}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": enabled}}); err != nil {
		t.Fatal(err)
	}

	tools := NewGateway(svc, nil, nil).providerToolsForRun(context.Background(), run.ID)
	names := []string{}
	for _, tool := range tools {
		names = append(names, tool.ProviderName)
		if tool.ProviderName == "" || len(tool.Parameters) == 0 {
			t.Fatalf("tool missing provider schema: %+v", tool)
		}
	}
	want := []string{"tool_load_tools", "skill_load_skill", "workspace_read", "workspace_edit", "workspace_patch_preview", "workspace_patch_apply", "sandbox_exec_command", "sandbox_start_process", "sandbox_continue_process", "sandbox_terminate_process", "lsp_symbols", "lsp_definition", "lsp_hover", "web_search", "browser_open", "browser_screenshot", "browser_type", "browser_press", "todo_write"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("provider tools = %+v", names)
	}
	if internalProviderToolName("tool_load_tools") != productdata.ToolNameLoadTools || providerToolName(productdata.ToolNameLoadSkill) != "skill_load_skill" || internalProviderToolName("workspace_edit") != productdata.ToolNameWorkspaceEdit || internalProviderToolName("workspace_patch_preview") != productdata.ToolNameWorkspacePatchPreview || providerToolName(productdata.ToolNameWorkspacePatchApply) != "workspace_patch_apply" || providerToolName(productdata.ToolNameSandboxExecCommand) != "sandbox_exec_command" || internalProviderToolName("sandbox_start_process") != productdata.ToolNameSandboxStartProcess || internalProviderToolName("browser_type") != productdata.ToolNameBrowserType || providerToolName(productdata.ToolNameLSPDefinition) != "lsp_definition" || internalProviderToolName("todo_write") != productdata.ToolNameTodoWrite {
		t.Fatalf("provider tool name mapping failed")
	}
}

func TestGatewayContinuationRejectsRepeatedWorkspaceToolCallID(t *testing.T) {
	svc, ident, thread, message, run := setupWorkspaceContinuationRun(t)
	completeWorkspaceReadToolCall(t, svc, ident, thread.ID, run.ID, "tc_1")
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_1", "arguments_summary": map[string]any{"path": "safe.txt", "limit": 128}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_1"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "duplicate_tool_call_id" {
		t.Fatalf("run = %+v", got)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var approvalRequired int
	for _, event := range events {
		if event.Type == productdata.EventToolCallApprovalRequired {
			approvalRequired++
		}
	}
	if approvalRequired != 1 {
		t.Fatalf("approvalRequired = %d, events = %+v", approvalRequired, events)
	}
}

func TestGatewayContinuationRejectsSecondPendingWorkspaceToolCall(t *testing.T) {
	svc, ident, thread, message, run := setupWorkspaceContinuationRun(t)
	completeWorkspaceReadToolCall(t, svc, ident, thread.ID, run.ID, "tc_1")
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_pending", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "safe.txt", "limit": 128}, ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_3", "arguments_summary": map[string]any{"path": "safe.txt", "limit": 128}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_1"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_call_rejected" {
		t.Fatalf("run = %+v", got)
	}
	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_3"); err == nil {
		t.Fatalf("second pending tool call was recorded")
	}
}

func TestGatewayContinuationRejectsWorkspaceToolOutsideEnabledSnapshot(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace loop", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Inspect workspace"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWorkspaceGlob}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_glob_1", "tool_name": productdata.ToolNameWorkspaceGlob, "arguments_summary": map[string]any{"pattern": "*.go", "limit": 10}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_glob_1", "tool_name": productdata.ToolNameWorkspaceGlob, "result_summary": map[string]any{"matches": []any{"internal/runtime/gateway.go"}}}}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_2", "arguments_summary": map[string]any{"path": "internal/runtime/gateway.go", "limit": 512}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_glob_1"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "unsupported_tool_loop" {
		t.Fatalf("run = %+v", got)
	}
	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_2"); err == nil {
		t.Fatalf("disabled workspace tool call was recorded")
	}
}

func TestGatewayFailsContinuationProviderErrorWithoutFinalAssistant(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "What time is it?"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameCurrentTime, "arguments_summary": map[string]any{"timezone": "UTC"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameCurrentTime, "result_summary": map[string]any{"iso_time": "2026-05-25T10:00:00Z", "timezone": "UTC", "source": "runtime"}}}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: "Partial "}, {Type: ProviderEventError, ErrorCode: "provider_error", Message: "secret provider token leaked"}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_1"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorMessage == nil || *got.ErrorMessage != "[redacted]" {
		t.Fatalf("run = %+v", got)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var finalCount int
	for _, event := range events {
		if event.Category == productdata.RunEventCategoryFinal {
			finalCount++
		}
		if event.Summary == "secret provider token leaked" {
			t.Fatalf("unredacted event = %+v", event)
		}
	}
	if finalCount != 1 {
		t.Fatalf("final events = %d, events = %+v", finalCount, events)
	}
}

type capturingProvider struct {
	config  ProviderConfig
	request ProviderRequest
	events  []ProviderEvent
}

func (p *capturingProvider) Config() ProviderConfig { return p.config }

func (p *capturingProvider) Stream(_ context.Context, request ProviderRequest) (<-chan ProviderEvent, error) {
	p.request = request
	events := p.events
	if len(events) == 0 {
		events = []ProviderEvent{{Type: ProviderEventCompleted, Text: "ok"}}
	}
	ch := make(chan ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func TestQueuedRunRouterHydratesGatewayInputFromJobMetadata(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "override"})
	if err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_gateway", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "base", Enabled: true}}

	if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err != nil {
		t.Fatal(err)
	}

	if provider.request.ThreadID != thread.ID || provider.request.MessageID != message.ID || provider.request.Model != "override" {
		t.Fatalf("request = %+v", provider.request)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
}

func TestQueuedRunRouterReturnsErrorWhenGatewayRunFails(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_gateway", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventRateLimited}}}

	if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err == nil {
		t.Fatal("Run() error = nil")
	}
	if _, changed, err := svc.CompleteBackgroundJob(context.Background(), ident, productdata.CompleteBackgroundJobInput{JobID: job.ID, WorkerID: "worker_gateway", OwnershipVersion: job.OwnershipVersion}); err != nil || !changed {
		t.Fatalf("CompleteBackgroundJob() changed=%v err=%v", changed, err)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRunAsyncOutlivesRequestContext(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	provider := contextAwareProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}

	NewGateway(svc, nil, []Provider{provider}).RunAsync(ctx, run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got := waitForTerminalRun(t, svc, run.ID)
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayFailsWhenAssistantMessageCannotBePersisted(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "hello"}}}
	wrapped := assistantPersistFailingService{Service: svc}

	NewGateway(wrapped, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "assistant_message_persist_failed" {
		t.Fatalf("run = %+v", got)
	}
}

type contextAwareProvider struct {
	config ProviderConfig
}

func (p contextAwareProvider) Config() ProviderConfig { return p.config }

func (p contextAwareProvider) Stream(ctx context.Context, _ ProviderRequest) (<-chan ProviderEvent, error) {
	ch := make(chan ProviderEvent, 1)
	if ctx.Err() != nil {
		ch <- ProviderEvent{Type: ProviderEventError, ErrorCode: "request_context_canceled", Message: "request context was canceled"}
	} else {
		ch <- ProviderEvent{Type: ProviderEventCompleted, Text: "ok"}
	}
	close(ch)
	return ch, nil
}

type assistantPersistFailingService struct {
	productdata.Service
}

func (s assistantPersistFailingService) AppendAssistantMessage(context.Context, identity.LocalIdentity, string, productdata.AppendAssistantMessageInput) (productdata.Message, error) {
	return productdata.Message{}, productdata.NewError(productdata.CodeInternalError, "assistant message persistence failed")
}

func waitForTerminalRun(t *testing.T, svc productdata.Service, runID string) productdata.Run {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		run, err := svc.GetRun(context.Background(), identity.LocalDevIdentity(), runID)
		if err != nil {
			t.Fatal(err)
		}
		if productdata.IsRunTerminal(run.Status) {
			return run
		}
		time.Sleep(10 * time.Millisecond)
	}
	run, err := svc.GetRun(context.Background(), identity.LocalDevIdentity(), runID)
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func setupWorkspaceContinuationRun(t *testing.T) (*productdata.MemoryService, identity.LocalIdentity, productdata.Thread, productdata.Message, productdata.Run) {
	t.Helper()
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace loop", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Inspect workspace"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	enabled := []string{productdata.ToolNameWorkspaceGlob, productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceRead}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": enabled}}); err != nil {
		t.Fatal(err)
	}
	return svc, ident, thread, message, run
}

func completeWorkspaceReadToolCall(t *testing.T, svc productdata.Service, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) {
	t.Helper()
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, runID, productdata.RecordToolCallRequestInput{ToolCallID: toolCallID, ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "safe.txt", "limit": 128}, ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, threadID, runID, toolCallID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, threadID, runID, toolCallID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.CompleteToolCallSuccess(context.Background(), ident, threadID, runID, toolCallID, map[string]any{"path": "safe.txt", "truncated": false}); err != nil {
		t.Fatal(err)
	}
}

func writeRuntimeGatewayTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func runtimeGatewayEventsJSON(t *testing.T, events []productdata.RunEvent) string {
	t.Helper()
	raw, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func runtimeGatewayMessagesJSON(t *testing.T, messages []productdata.Message) string {
	t.Helper()
	raw, err := json.Marshal(messages)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
