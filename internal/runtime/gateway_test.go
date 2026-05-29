package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func TestGatewayPersistsPartialAssistantMessageWhenStoppedAfterVisibleOutput(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "long answer"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := newBlockingDeltaProvider("partial answer")
	done := make(chan struct{})
	go func() {
		defer close(done)
		NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})
	}()
	<-provider.firstSent
	waitForRunEventType(t, svc, run.ID, "model_output_delta")
	if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
		t.Fatal(err)
	}
	close(provider.release)
	<-done

	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "partial answer" {
		t.Fatalf("messages = %+v", messages)
	}
	if messages[1].Metadata["run_id"] != run.ID || messages[1].Metadata["interrupted"] != true {
		t.Fatalf("assistant metadata = %+v", messages[1].Metadata)
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

func TestLocalCodexResponsesRequestAllowsParallelToolCalls(t *testing.T) {
	request, err := buildLocalCodexResponsesRequest(context.Background(), LocalCodexCredentialSnapshot{BaseURL: "https://example.test/v1", APIKey: "token", Model: "gpt-local"}, ProviderRequest{
		ThreadID: "thread_1",
		Model:    "gpt-local",
		Tools: []ProviderToolDefinition{
			providerTool(productdata.ToolNameWorkspaceRead, "Read file", map[string]any{"path": map[string]any{"type": "string"}}, []string{"path"}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer request.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["parallel_tool_calls"] != true {
		t.Fatalf("parallel_tool_calls = %v", body["parallel_tool_calls"])
	}
}

func TestLocalCodexResponsesParserEmitsMultipleToolCalls(t *testing.T) {
	stream := strings.Join([]string{
		`data: {"type":"response.output_item.done","item":{"type":"function_call","name":"workspace_read","call_id":"tc_read_a","arguments":"{\"path\":\"a.txt\",\"limit\":128}"}}`,
		``,
		`data: {"type":"response.output_item.done","item":{"type":"function_call","name":"workspace_read","call_id":"tc_read_b","arguments":"{\"path\":\"b.txt\",\"limit\":128}"}}`,
		``,
		`data: {"type":"response.completed","response":{"id":"resp_1"}}`,
		``,
	}, "\n")
	ch := make(chan ProviderEvent, 8)
	parseLocalCodexResponsesSSE(context.Background(), strings.NewReader(stream), ch)
	close(ch)

	events := []ProviderEvent{}
	for event := range ch {
		events = append(events, event)
	}
	if len(events) != 3 {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Type != ProviderEventToolCall || events[0].ToolName != productdata.ToolNameWorkspaceRead || events[0].Metadata["tool_call_id"] != "tc_read_a" {
		t.Fatalf("first event = %+v", events[0])
	}
	if events[1].Type != ProviderEventToolCall || events[1].ToolName != productdata.ToolNameWorkspaceRead || events[1].Metadata["tool_call_id"] != "tc_read_b" {
		t.Fatalf("second event = %+v", events[1])
	}
	if events[2].Type != ProviderEventCompleted {
		t.Fatalf("third event = %+v", events[2])
	}
}

func TestLocalCodexResponsesParserUsesCompletedOutputTextWhenNoDeltas(t *testing.T) {
	stream := strings.Join([]string{
		`data: {"type":"response.completed","response":{"output":[{"type":"message","content":[{"type":"output_text","text":"Done from completed output."}]}]}}`,
		``,
	}, "\n")
	ch := make(chan ProviderEvent, 4)
	parseLocalCodexResponsesSSE(context.Background(), strings.NewReader(stream), ch)
	close(ch)

	events := []ProviderEvent{}
	for event := range ch {
		events = append(events, event)
	}
	if len(events) != 1 || events[0].Type != ProviderEventCompleted || events[0].Text != "Done from completed output." {
		t.Fatalf("events = %+v", events)
	}
}

func TestLocalCodexResponsesParserCompletesOnDoneAfterTextDelta(t *testing.T) {
	stream := strings.Join([]string{
		`data: {"type":"response.output_text.delta","delta":"Done from delta."}`,
		``,
		`data: [DONE]`,
		``,
	}, "\n")
	ch := make(chan ProviderEvent, 4)
	parseLocalCodexResponsesSSE(context.Background(), strings.NewReader(stream), ch)
	close(ch)

	events := []ProviderEvent{}
	for event := range ch {
		events = append(events, event)
	}
	if len(events) != 2 || events[0].Type != ProviderEventTextDelta || events[1].Type != ProviderEventCompleted {
		t.Fatalf("events = %+v", events)
	}
}

func TestLocalCodexResponsesParserMapsIncompleteToProviderIncomplete(t *testing.T) {
	stream := strings.Join([]string{
		`data: {"type":"response.incomplete","response":{"incomplete_details":{"reason":"max_output_tokens"}}}`,
		``,
	}, "\n")
	ch := make(chan ProviderEvent, 4)
	parseLocalCodexResponsesSSE(context.Background(), strings.NewReader(stream), ch)
	close(ch)

	events := []ProviderEvent{}
	for event := range ch {
		events = append(events, event)
	}
	if len(events) != 1 || events[0].Type != ProviderEventError || events[0].ErrorCode != "provider_incomplete" || events[0].Metadata["incomplete_reason"] != "max_output_tokens" {
		t.Fatalf("events = %+v", events)
	}
}

func TestRunSystemPromptGuidesWorkModeToWorkspaceTools(t *testing.T) {
	chat := runSystemPrompt(&productdata.RunContext{Thread: productdata.Thread{Mode: productdata.ThreadModeChat}})
	if strings.Contains(chat, "File and folder tasks are tool-first") {
		t.Fatalf("chat prompt contains work-only policy: %s", chat)
	}
	for _, expected := range []string{"Answer first", "No preface", "Do not repeat the user's request", "For code changes, report what changed and what was verified"} {
		if !strings.Contains(chat, expected) {
			t.Fatalf("chat prompt missing concise output rule %q: %s", expected, chat)
		}
	}
	work := runSystemPrompt(&productdata.RunContext{Thread: productdata.Thread{Mode: productdata.ThreadModeWork}})
	for _, expected := range []string{
		"Directory questions: use workspace_tree_summary or workspace_list_directory first",
		"Use workspace_glob only for file-name pattern matching or a narrow follow-up",
		"Content questions: use workspace_grep or workspace_read after you have a relative path",
		"Modification questions: use workspace_read first, then workspace_patch_preview, then workspace_patch_apply only after approval",
		"Use sandbox commands only when the user explicitly asks for a shell/process action or when verifying",
		"workspace_glob",
		"workspace_grep",
		"workspace_read",
		"Do not tell the user to run shell commands",
		"use \".\" for it and do not repeat the root folder name",
		"Do not stop with a final answer that says you still need to continue reading files",
		"request the next workspace tool call in the same run",
	} {
		if !strings.Contains(work, expected) {
			t.Fatalf("work prompt missing %q: %s", expected, work)
		}
	}
	for _, expected := range []string{
		"Reports, articles, Markdown, and saveable documents should use artifact.create_text",
		"Diagrams, SVG drawings, HTML mockups, charts, and visual explanations should use artifact.create_visual",
		"Do not place raw <svg>, <html>, or fenced SVG/HTML blocks directly in the final reply",
		"Reference saved artifacts as [title](artifact:<key>)",
		"Do not invent artifact keys",
	} {
		if !strings.Contains(work, expected) {
			t.Fatalf("work prompt missing artifact contract %q: %s", expected, work)
		}
	}
}

func TestRunSystemPromptUsesSelectedWorkspaceLabelForDirectoryReferences(t *testing.T) {
	prompt := runSystemPrompt(&productdata.RunContext{
		Thread:        productdata.Thread{Mode: productdata.ThreadModeWork},
		WorkspaceRoot: productdata.WorkspaceRootConfig{Path: "/Users/xuean/Downloads", DisplayName: "Downloads"},
	})

	for _, expected := range []string{
		"Selected workspace: Downloads",
		"current directory, this directory, selected directory, just selected directory, 当前目录, 这个目录, 刚选目录",
		"use the selected workspace root",
		"download directory or 下载目录",
		"only treat it as Downloads when the selected workspace label is Downloads",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt missing workspace reference rule %q: %s", expected, prompt)
		}
	}
	if strings.Contains(prompt, "/Users/xuean/Downloads") {
		t.Fatalf("prompt leaked absolute workspace path: %s", prompt)
	}
}

func TestLoadToolsProviderSchemaIsQueryOnlyAndOptional(t *testing.T) {
	tool, ok := builtinProviderToolDefinition(productdata.ToolNameLoadTools)
	if !ok {
		t.Fatal("load_tools provider definition missing")
	}
	if !strings.Contains(tool.Description, "query-only") {
		t.Fatalf("load_tools description should tell the model it is query-only: %+v", tool)
	}
	properties, ok := tool.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties = %+v", tool.Parameters)
	}
	if _, ok := properties["names"]; ok {
		t.Fatalf("load_tools provider schema should not expose names: %+v", properties)
	}
	required, ok := tool.Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required = %+v", tool.Parameters["required"])
	}
	if len(required) != 0 {
		t.Fatalf("load_tools should accept empty query args, required=%+v", required)
	}
}

func TestArtifactCreateTextProviderSchemaIncludesDocumentMetadata(t *testing.T) {
	tool, ok := builtinProviderToolDefinition(productdata.ToolNameArtifactCreateText)
	if !ok {
		t.Fatal("artifact.create_text provider definition missing")
	}
	properties, ok := tool.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties = %+v", tool.Parameters)
	}
	for _, want := range []string{"title", "filename", "mime_type", "display", "content", "max_bytes"} {
		if _, ok := properties[want]; !ok {
			t.Fatalf("artifact.create_text schema missing %s: %+v", want, properties)
		}
	}
	display, _ := properties["display"].(map[string]any)
	if fmt.Sprint(display["enum"]) != "[inline panel]" {
		t.Fatalf("display schema = %+v", display)
	}
	required, ok := tool.Parameters["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "content" {
		t.Fatalf("artifact.create_text required = %+v", tool.Parameters["required"])
	}
}

func TestArtifactCreateVisualProviderSchemaIncludesRenderableMetadata(t *testing.T) {
	tool, ok := builtinProviderToolDefinition(productdata.ToolNameArtifactCreateVisual)
	if !ok {
		t.Fatal("artifact.create_visual provider definition missing")
	}
	properties, ok := tool.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties = %+v", tool.Parameters)
	}
	for _, want := range []string{"title", "filename", "mime_type", "display", "content", "max_bytes"} {
		if _, ok := properties[want]; !ok {
			t.Fatalf("artifact.create_visual schema missing %s: %+v", want, properties)
		}
	}
	mimeType, _ := properties["mime_type"].(map[string]any)
	if fmt.Sprint(mimeType["enum"]) != "[image/svg+xml text/html]" {
		t.Fatalf("mime_type schema = %+v", mimeType)
	}
	required, ok := tool.Parameters["required"].([]string)
	if !ok || strings.Join(required, ",") != "title,mime_type,content" {
		t.Fatalf("artifact.create_visual required = %+v", tool.Parameters["required"])
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
		t.Fatalf("run = %+v error=%s", got, runErrorCodeForTest(got))
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

func TestGatewayRecordsMultipleApprovalRequiredToolCalls(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "check two clocks"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_time_a", "arguments_summary": map[string]any{"timezone": "UTC"}}},
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_time_b", "arguments_summary": map[string]any{"timezone": "UTC"}}},
	}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	for _, toolCallID := range []string{"tc_time_a", "tc_time_b"} {
		call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, toolCallID)
		if err != nil {
			t.Fatal(err)
		}
		if call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
			t.Fatalf("call %s = %+v", toolCallID, call)
		}
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusBlockedOnToolApproval {
		t.Fatalf("run = %+v", got)
	}
	state, err := svc.GetRunStepState(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.PendingToolCalls) != 2 {
		t.Fatalf("pending = %+v", state.PendingToolCalls)
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

func TestGatewayRejectsMemoryToolOutsideEnabledSnapshot(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search memory"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameCurrentTime}}}); err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameMemorySearch, Metadata: map[string]any{"tool_call_id": "tc_memory_search", "arguments_summary": map[string]any{"query": "plans", "limit": 5}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_call_rejected" {
		t.Fatalf("run = %+v", got)
	}
	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_memory_search"); err == nil {
		t.Fatal("memory tool call was recorded outside enabled snapshot")
	}
}

func TestGatewayRecordsMultipleProviderToolCallsInOneTurn(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search and fetch"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWebSearch, productdata.ToolNameWebFetch}}}); err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWebSearch, Metadata: map[string]any{"tool_call_id": "tc_search", "arguments_summary": map[string]any{"query": "latest ai news"}}},
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWebFetch, Metadata: map[string]any{"tool_call_id": "tc_fetch", "arguments_summary": map[string]any{"url": "https://example.com/repo"}}},
		{Type: ProviderEventCompleted, Text: "done"},
	}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	for _, want := range []struct {
		id   string
		name string
	}{
		{id: "tc_search", name: productdata.ToolNameWebSearch},
		{id: "tc_fetch", name: productdata.ToolNameWebFetch},
	} {
		call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, want.id)
		if err != nil {
			t.Fatalf("GetToolCall(%s) error = %v", want.id, err)
		}
		if call.ToolName != want.name || call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
			t.Fatalf("call %s = %+v", want.id, call)
		}
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusQueued {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRetriesTransientProviderFailureBeforeOutput(t *testing.T) {
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
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{
		{Type: ProviderEventRateLimited},
	}, {
		{Type: ProviderEventCompleted, Text: "Recovered."},
	}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if len(provider.requests) != 2 {
		t.Fatalf("provider requests = %d, want 2", len(provider.requests))
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
	if len(messages) != 2 || messages[1].Content != "Recovered." {
		t.Fatalf("messages = %+v", messages)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRunEventType(events, "model_request_retry_scheduled") {
		t.Fatalf("events = %+v", events)
	}
}

func TestGatewayRetriesTimeoutEmptyAndStreamErrorsBeforeOutput(t *testing.T) {
	cases := []struct {
		name     string
		provider Provider
	}{
		{
			name: "timeout",
			provider: &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{
				{Type: ProviderEventTimeout},
			}, {
				{Type: ProviderEventCompleted, Text: "Recovered."},
			}}},
		},
		{
			name: "empty",
			provider: &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{
				{Type: ProviderEventEmptyResponse},
			}, {
				{Type: ProviderEventCompleted, Text: "Recovered."},
			}}},
		},
		{
			name:     "stream error",
			provider: &streamErrorThenSuccessProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, err: errors.New("temporary connection reset")},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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

			NewGateway(svc, nil, []Provider{tc.provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

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
			if len(messages) != 2 || messages[1].Content != "Recovered." {
				t.Fatalf("messages = %+v", messages)
			}
			events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
			if err != nil {
				t.Fatal(err)
			}
			if !hasRunEventType(events, "model_request_retry_scheduled") {
				t.Fatalf("events = %+v", events)
			}
		})
	}
}

func TestGatewayRetriesHTTPProviderTransientFailuresBeforeOutput(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		networkErr bool
	}{
		{name: "rate limited", status: http.StatusTooManyRequests},
		{name: "request timeout", status: http.StatusRequestTimeout},
		{name: "gateway timeout", status: http.StatusGatewayTimeout},
		{name: "network error", networkErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := productdata.NewMemoryService()
			ident := identity.LocalDevIdentity()
			thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway HTTP retry", Mode: productdata.ThreadModeChat})
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

			attempts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				if r.URL.Path != "/v1/chat/completions" {
					t.Fatalf("request path = %s", r.URL.Path)
				}
				if attempts == 1 && !tc.networkErr {
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(`{"error":{"type":"rate_limit_error","code":"retryable","message":"retry"}}`))
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Recovered.\"}}]}\n\n"))
				_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"))
			}))
			defer server.Close()
			client := server.Client()
			var transport *singleFailureRoundTripper
			if tc.networkErr {
				transport = &singleFailureRoundTripper{base: http.DefaultTransport, err: errors.New("temporary connection reset")}
				client.Transport = transport
			}
			provider := NewHTTPProvider(ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: server.URL + "/v1", APIKey: "secret-key", Model: "model", Enabled: true}, client)

			NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

			wantAttempts := 2
			if tc.networkErr {
				wantAttempts = 1
				if transport == nil || !transport.used {
					t.Fatalf("network failure transport was not used")
				}
			}
			if attempts != wantAttempts {
				t.Fatalf("server attempts = %d, want %d", attempts, wantAttempts)
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
			if len(messages) != 2 || messages[1].Content != "Recovered." {
				t.Fatalf("messages = %+v", messages)
			}
			events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
			if err != nil {
				t.Fatal(err)
			}
			if !hasRunEventType(events, "model_request_retry_scheduled") {
				t.Fatalf("events = %+v", events)
			}
		})
	}
}

type singleFailureRoundTripper struct {
	base http.RoundTripper
	err  error
	used bool
}

func (t *singleFailureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.used {
		t.used = true
		return nil, t.err
	}
	return t.base.RoundTrip(req)
}

func TestGatewayFailsAfterProviderRetryExhaustionWithoutAssistantMessage(t *testing.T) {
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
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{
		{{Type: ProviderEventRateLimited}},
		{{Type: ProviderEventTimeout}},
		{{Type: ProviderEventEmptyResponse}},
	}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if len(provider.requests) != 3 {
		t.Fatalf("provider requests = %d, want 3", len(provider.requests))
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "empty_response" {
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
	var scheduled int
	for _, event := range events {
		if event.Type == "model_request_retry_scheduled" {
			scheduled++
		}
	}
	if scheduled != maxProviderAttempts-1 {
		t.Fatalf("retry scheduled events = %d, want %d; events = %+v", scheduled, maxProviderAttempts-1, events)
	}
}

func TestGatewayRedactsRetryTelemetryAndExhaustionFailure(t *testing.T) {
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
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{
		{{Type: ProviderEventError, ErrorCode: "provider_error", Message: "provider failed", Metadata: map[string]any{"api_key": "sk-retry-leak", "provider_trace": "secret provider trace", "safe": "visible"}}},
		{{Type: ProviderEventError, ErrorCode: "provider_error", Message: "provider failed", Metadata: map[string]any{"authorization": "Bearer sk-retry-leak", "safe": "visible"}}},
		{{Type: ProviderEventError, ErrorCode: "provider_error", Message: "provider failed", Metadata: map[string]any{"workspace_root_path": "/Users/xuean/secret", "safe": "visible"}}},
	}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	body := string(rendered)
	for _, forbidden := range []string{"sk-retry-leak", "Bearer", "secret provider trace", "/Users/xuean/secret"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("retry telemetry leaked %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "visible") {
		t.Fatalf("retry telemetry dropped safe metadata: %s", body)
	}
	var finalFailure productdata.RunEvent
	for _, event := range events {
		if event.Type == productdata.EventRunFailed {
			finalFailure = event
		}
	}
	if finalFailure.Metadata["provider_id"] != "custom" || finalFailure.Metadata["model"] != "model" || fmt.Sprint(finalFailure.Metadata["attempt"]) != fmt.Sprint(maxProviderAttempts) {
		t.Fatalf("final failure metadata = %+v", finalFailure.Metadata)
	}
}

func TestGatewayDoesNotRetryProviderFailureAfterVisibleOutput(t *testing.T) {
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
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{
		{Type: ProviderEventTextDelta, Text: "Partial "},
		{Type: ProviderEventRateLimited},
	}, {
		{Type: ProviderEventCompleted, Text: "Should not happen."},
	}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if len(provider.requests) != 1 {
		t.Fatalf("provider requests = %d, want 1", len(provider.requests))
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "provider_rate_limited" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayAutoApprovesChatWebFetchToolCall(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Analyze https://example.com/repo"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWebFetch}}}); err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWebFetch, Metadata: map[string]any{"tool_call_id": "tc_fetch_1", "arguments_summary": map[string]any{"url": "https://example.com/repo", "max_bytes": 4096}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusQueued {
		t.Fatalf("run = %+v", got)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_fetch_1")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWebFetch || call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
		t.Fatalf("call = %+v", call)
	}
}

func TestGatewayRejectsDirectoryInventoryStartingWithGrep(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Directory guard", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "请盘点当前目录有哪些东西，并按源码/文档/配置分类。"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceListDirectory, productdata.ToolNameWorkspaceTreeSummary}}}); err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceGrep, Metadata: map[string]any{"tool_call_id": "tc_bad_inventory_grep", "arguments_summary": map[string]any{"query": ".", "path": ".", "limit": 200}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_bad_inventory_grep"); err == nil {
		t.Fatal("directory inventory grep was recorded")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_planner_guardrail" {
		t.Fatalf("run = %+v", got)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Type != productdata.EventRunFailed || last.Metadata["recommended_tool"] != productdata.ToolNameWorkspaceTreeSummary {
		t.Fatalf("events = %+v", events)
	}
}

func TestGatewayAllowsDirectoryInventoryStartingWithTreeSummary(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Directory guard", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "盘点这个目录都有哪些东西。"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWorkspaceTreeSummary}}}); err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceTreeSummary, Metadata: map[string]any{"tool_call_id": "tc_inventory_tree", "arguments_summary": map[string]any{"path": ".", "depth": 2, "max_entries": 200}}}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_inventory_tree")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWorkspaceTreeSummary || call.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
		t.Fatalf("call = %+v", call)
	}
}

func TestGatewayRejectsRepeatedWorkspaceReadArguments(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Repeat guard", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Read the notes and summarize."})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWorkspaceRead}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_read_1", "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": "notes.txt", "limit": 128}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_read_1", "tool_name": productdata.ToolNameWorkspaceRead, "result_summary": map[string]any{"path": "notes.txt", "content": "hello"}}}); err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_2", "arguments_summary": map[string]any{"path": "notes.txt", "limit": 128}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_read_1"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_2"); err == nil {
		t.Fatal("repeated read tool call was recorded")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_planner_guardrail" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRejectsRepeatedWorkspaceReadArgumentsInSameContinuationTurn(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Repeat guard", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Read notes.txt and summarize it."})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameCurrentTime, productdata.ToolNameWorkspaceRead}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_time", "tool_name": productdata.ToolNameCurrentTime, "arguments_summary": map[string]any{"timezone": "UTC"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_time", "tool_name": productdata.ToolNameCurrentTime, "result_summary": map[string]any{"iso_time": "2026-05-29T00:00:00Z"}}}); err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_a", "arguments_summary": map[string]any{"path": "notes.txt", "limit": 128}}},
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_b", "arguments_summary": map[string]any{"path": "notes.txt", "limit": 128}}},
	}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_time"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_a"); err == nil {
		t.Fatal("first duplicate read tool call was recorded")
	}
	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_b"); err == nil {
		t.Fatal("second duplicate read tool call was recorded")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_planner_guardrail" {
		t.Fatalf("run = %+v", got)
	}
}

func TestRepeatedWorkspaceToolRequestThisTurnKeysByToolAndCoversGlob(t *testing.T) {
	seen := map[string]bool{}
	if repeatedWorkspaceToolRequestThisTurn(productdata.ToolNameWorkspaceListDirectory, map[string]any{"path": "."}, seen) {
		t.Fatal("first list_directory was treated as repeated")
	}
	if repeatedWorkspaceToolRequestThisTurn(productdata.ToolNameWorkspaceRead, map[string]any{"path": "."}, seen) {
		t.Fatal("read with same argument shape should not collide with list_directory")
	}
	if repeatedWorkspaceToolRequestThisTurn(productdata.ToolNameWorkspaceGlob, map[string]any{"pattern": "*.go", "limit": 10}, seen) {
		t.Fatal("first glob was treated as repeated")
	}
	if !repeatedWorkspaceToolRequestThisTurn(productdata.ToolNameWorkspaceGlob, map[string]any{"pattern": "*.go", "limit": 10}, seen) {
		t.Fatal("second identical glob was not treated as repeated")
	}
}

func TestGatewayRejectsRepeatedWorkspaceReadArgumentsAfterDifferentTool(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Repeat guard", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Read the notes and summarize."})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range []productdata.AppendRunEventInput{
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspaceGrep}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_read_1", "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": "notes.txt", "limit": 128}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_read_1", "tool_name": productdata.ToolNameWorkspaceRead, "result_summary": map[string]any{"path": "notes.txt", "content": "hello"}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_grep_1", "tool_name": productdata.ToolNameWorkspaceGrep, "arguments_summary": map[string]any{"query": "hello", "path": ".", "limit": 20}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_grep_1", "tool_name": productdata.ToolNameWorkspaceGrep, "result_summary": map[string]any{"match_count": 1}}},
	} {
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, event); err != nil {
			t.Fatal(err)
		}
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_2", "arguments_summary": map[string]any{"path": "notes.txt", "limit": 128}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_grep_1"})

	if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_2"); err == nil {
		t.Fatal("repeated read after grep was recorded")
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_planner_guardrail" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayAllowsRepeatedWorkspaceReadAfterMutationAndExec(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Repeat guard", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Patch then verify the file."})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range []productdata.AppendRunEventInput{
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspacePatchApply, productdata.ToolNameSandboxExecCommand}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_read_1", "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 64}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_read_1", "tool_name": productdata.ToolNameWorkspaceRead, "result_summary": map[string]any{"path": "src/notes.txt", "content": "needle"}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_apply_1", "tool_name": productdata.ToolNameWorkspacePatchApply, "arguments_summary": map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "m76 loop\n"}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_apply_1", "tool_name": productdata.ToolNameWorkspacePatchApply, "result_summary": map[string]any{"path": "src/notes.txt", "changed": true}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_exec_1", "tool_name": productdata.ToolNameSandboxExecCommand, "arguments_summary": map[string]any{"argv": []any{"cat", "src/notes.txt"}, "cwd": "."}}},
		{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_exec_1", "tool_name": productdata.ToolNameSandboxExecCommand, "result_summary": map[string]any{"exit_code": 0}}},
	} {
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, event); err != nil {
			t.Fatal(err)
		}
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_2", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 64}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_exec_1"})

	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_2")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWorkspaceRead {
		t.Fatalf("call = %+v", call)
	}
}

func TestGatewayFinalizesStructuredProviderPayloadAsNaturalLanguage(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Final guard", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Summarize Project Tokenizer at https://example.com/tokenizer"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	raw := `{"answer":"Project Tokenizer lives at https://example.com/tokenizer and is safe to mention.","tool_calls":[{"name":"workspace.read"}]}`
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventCompleted, Text: raw}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatalf("messages = %+v", messages)
	}
	content := messages[1].Content
	if strings.HasPrefix(content, "{") || strings.Contains(content, `"tool_calls"`) || !strings.Contains(content, "https://example.com/tokenizer") || content == "[redacted]" {
		t.Fatalf("final content = %q", content)
	}
}

func TestNaturalLanguageFinalContentExtractsNestedStructuredAnswers(t *testing.T) {
	for name, tc := range map[string]struct {
		input string
		want  string
	}{
		"nested result summary": {
			input: `{"result":{"summary":"Nested summary from provider."},"tool_calls":[]}`,
			want:  "Nested summary from provider.",
		},
		"output text": {
			input: `{"output_text":"Provider output text."}`,
			want:  "Provider output text.",
		},
		"array text segments": {
			input: `{"content":[{"type":"output_text","text":"First paragraph."},{"type":"output_text","text":"Second paragraph."}]}`,
			want:  "First paragraph.\n\nSecond paragraph.",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := naturalLanguageFinalContent(tc.input); got != tc.want {
				t.Fatalf("naturalLanguageFinalContent() = %q, want %q", got, tc.want)
			}
		})
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

func TestGatewayCompactsLargeProviderContextBeforeRequest(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 12; index++ {
		if _, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: fmt.Sprintf("older user %02d %s", index, strings.Repeat("u", 3500))}); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, productdata.AppendAssistantMessageInput{Content: fmt.Sprintf("older assistant %02d %s", index, strings.Repeat("a", 3500))}); err != nil {
			t.Fatal(err)
		}
	}
	current, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "final trigger must stay"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: current.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: current.ID, ProviderID: "custom"})

	if providerMessagesBytes(provider.request.Messages) > maxProviderContextBytes {
		t.Fatalf("provider context bytes = %d", providerMessagesBytes(provider.request.Messages))
	}
	if len(provider.request.Messages) == 0 || provider.request.Messages[0].Role != "user" || !strings.Contains(provider.request.Messages[0].Content, "<conversation_summary>") {
		t.Fatalf("messages = %+v", provider.request.Messages)
	}
	last := provider.request.Messages[len(provider.request.Messages)-1]
	if last.Role != "user" || last.Content != "final trigger must stay" {
		t.Fatalf("last message = %+v", last)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	for _, event := range events {
		if event.Type == "context_compacted" {
			metadata = event.Metadata
			break
		}
	}
	if metadata == nil || metadata["strategy"] != "recent_window_with_tool_pair_preservation" || metadata["redaction_applied"] != true {
		t.Fatalf("context_compacted metadata = %+v", metadata)
	}
	for key := range metadata {
		if strings.Contains(key, "content") || strings.Contains(key, "payload") || strings.Contains(key, "arguments") {
			t.Fatalf("unsafe metadata key %q in %+v", key, metadata)
		}
	}
}

func TestProviderContextCompactionPreservesToolCallResultPairs(t *testing.T) {
	messages := []ProviderMessage{}
	for index := 0; index < 12; index++ {
		messages = append(messages, ProviderMessage{Role: "user", Content: fmt.Sprintf("older user %02d %s", index, strings.Repeat("u", 3500))})
	}
	for index := 0; index < 12; index++ {
		toolCallID := fmt.Sprintf("tc_%02d", index)
		messages = append(messages, ProviderMessage{Role: ProviderMessageRoleAssistantToolCall, ToolCallID: toolCallID, ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": fmt.Sprintf("file-%02d.go", index)}})
		messages = append(messages, ProviderMessage{Role: ProviderMessageRoleToolResult, ToolCallID: toolCallID, ToolName: productdata.ToolNameWorkspaceRead, Content: strings.Repeat("result", 1200)})
	}

	result := compactProviderMessages(messages, maxProviderContextBytes)

	if !result.Compacted || result.PreservedPairs == 0 || result.CompactedBytes > maxProviderContextBytes {
		t.Fatalf("result = %+v", result)
	}
	for index, message := range result.Messages {
		if message.Role != ProviderMessageRoleToolResult {
			continue
		}
		if index == 0 || result.Messages[index-1].Role != ProviderMessageRoleAssistantToolCall || result.Messages[index-1].ToolCallID != message.ToolCallID {
			t.Fatalf("orphan tool result at %d: %+v", index, result.Messages)
		}
	}
}

func TestProviderContextCompactionKeepsLatestUserAfterToolHistory(t *testing.T) {
	messages := []ProviderMessage{
		{Role: "user", Content: strings.Repeat("old context ", 5000)},
		{Role: ProviderMessageRoleAssistantToolCall, ToolCallID: "tc_read", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "notes.txt"}},
		{Role: ProviderMessageRoleToolResult, ToolCallID: "tc_read", ToolName: productdata.ToolNameWorkspaceRead, Content: strings.Repeat("tool result ", 3000)},
		{Role: "assistant", Content: "I read the notes."},
		{Role: "user", Content: "final trigger must stay"},
	}

	result := compactProviderMessages(messages, maxProviderContextBytes)

	if !result.Compacted || len(result.Messages) == 0 || result.CompactedBytes > maxProviderContextBytes {
		t.Fatalf("result = %+v", result)
	}
	last := result.Messages[len(result.Messages)-1]
	if last.Role != "user" || last.Content != "final trigger must stay" {
		t.Fatalf("last message = %+v, messages = %+v", last, result.Messages)
	}
	foundPair := false
	for index, message := range result.Messages {
		if message.Role != ProviderMessageRoleToolResult {
			continue
		}
		if index == 0 || result.Messages[index-1].Role != ProviderMessageRoleAssistantToolCall || result.Messages[index-1].ToolCallID != message.ToolCallID {
			t.Fatalf("orphan tool result at %d: %+v", index, result.Messages)
		}
		foundPair = true
	}
	if !foundPair {
		t.Fatalf("tool pair was dropped: %+v", result.Messages)
	}
}

func TestProviderContextCompactionBoundsLargeToolArguments(t *testing.T) {
	messages := []ProviderMessage{
		{Role: "user", Content: strings.Repeat("old context ", 5000)},
		{Role: ProviderMessageRoleAssistantToolCall, ToolCallID: "tc_edit", ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "notes.txt", "content": strings.Repeat("large patch ", 8000)}},
		{Role: ProviderMessageRoleToolResult, ToolCallID: "tc_edit", ToolName: productdata.ToolNameWorkspaceEdit, Content: "ok"},
		{Role: "user", Content: "final trigger must stay"},
	}

	result := compactProviderMessages(messages, maxProviderContextBytes)

	if !result.Compacted || result.CompactedBytes > maxProviderContextBytes {
		t.Fatalf("result = %+v", result)
	}
	last := result.Messages[len(result.Messages)-1]
	if last.Role != "user" || last.Content != "final trigger must stay" {
		t.Fatalf("last message = %+v, messages = %+v", last, result.Messages)
	}
	foundToolPair := false
	for _, message := range result.Messages {
		if message.Role != ProviderMessageRoleAssistantToolCall || message.ToolCallID != "tc_edit" {
			continue
		}
		foundToolPair = true
		if providerMessagesBytes([]ProviderMessage{message}) > maxProviderContextMessageBytes {
			t.Fatalf("tool call arguments were not compacted: %+v", message.ArgumentsSummary)
		}
		if _, ok := message.ArgumentsSummary["content"]; ok {
			t.Fatalf("large argument content was preserved: %+v", message.ArgumentsSummary)
		}
	}
	if !foundToolPair {
		t.Fatalf("latest tool pair was dropped: %+v", result.Messages)
	}
}

func TestMCPToolAllowedForStateUsesProjectedSchemaHash(t *testing.T) {
	state := productdata.RunStepState{
		EnabledToolNames:    []string{"mcp.local-search.search"},
		MCPToolSchemaHashes: map[string]string{"mcp.local-search.search": "sha256:search"},
	}

	allowed, hash := mcpToolAllowedForState(state, "mcp.local-search.search")
	if !allowed || hash != "sha256:search" {
		t.Fatalf("allowed=%v hash=%q", allowed, hash)
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
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameCurrentTime}}}); err != nil {
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

func TestGatewayCompactsLargeContinuationToolResult(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
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
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": "web/src/App.tsx"}}}); err != nil {
		t.Fatal(err)
	}
	largeOutput := strings.Repeat("repeated progress line\n", 300) + "path: web/src/App.tsx\nstatus: 503\nerror: provider unavailable\n"
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": productdata.ToolNameWorkspaceRead, "result_summary": map[string]any{"output": largeOutput}}}); err != nil {
		t.Fatal(err)
	}

	messages, err := NewGateway(svc, nil, nil).loadContinuationMessages(context.Background(), thread.ID, message.ID, run.ID, "tc_1")
	if err != nil {
		t.Fatal(err)
	}

	content := messages[2].Content
	if len(content) > maxProviderToolResultContentBytes {
		t.Fatalf("content length = %d", len(content))
	}
	for _, want := range []string{"web/src/App.tsx", "503", "provider unavailable", "tool output compacted"} {
		if !strings.Contains(content, want) {
			t.Fatalf("content missing %q: %s", want, content)
		}
	}
}

func TestGatewayContinuationSkipsPendingToolCalls(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Pending continuation", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Use tools"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_pending", "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": "pending.txt"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_done", "tool_name": productdata.ToolNameWorkspaceGrep, "arguments_summary": map[string]any{"query": "needle"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_done", "tool_name": productdata.ToolNameWorkspaceGrep, "result_summary": map[string]any{"matches": []any{"notes.txt"}}}}); err != nil {
		t.Fatal(err)
	}
	gateway := NewGateway(svc, nil, nil)

	messages, err := gateway.loadContinuationMessages(context.Background(), thread.ID, message.ID, run.ID, "tc_done")
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 3 {
		t.Fatalf("messages = %+v", messages)
	}
	for _, message := range messages {
		if message.ToolCallID == "tc_pending" {
			t.Fatalf("pending tool call entered continuation: %+v", messages)
		}
	}
	if messages[1].ToolCallID != "tc_done" || messages[2].ToolCallID != "tc_done" {
		t.Fatalf("continuation messages = %+v", messages)
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

func TestQueuedRunRouterResumesContinuationAfterSucceededToolRestart(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Restart resume", Mode: productdata.ThreadModeChat})
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
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_resume", ToolName: productdata.ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_resume", ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_resume"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_resume"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.CompleteToolCallSuccess(context.Background(), ident, thread.ID, run.ID, "tc_resume", map[string]any{"iso_time": "2026-05-25T10:00:00Z", "timezone": "UTC", "source": "runtime"}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: "Recovered "}, {Type: ProviderEventCompleted, Text: "Recovered after restart."}}}
	router := QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}

	if err := router.runApprovedTool(context.Background(), run, productdata.BackgroundJob{}, "tc_resume", nil, true); err != nil {
		t.Fatal(err)
	}

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertToolEventOrder(t, events, "tc_resume", []string{productdata.EventToolCallApprovalRequired, productdata.EventToolCallApproved, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded})
	if len(provider.request.Messages) != 3 || provider.request.Messages[1].ToolCallID != "tc_resume" || provider.request.Messages[2].ToolCallID != "tc_resume" {
		t.Fatalf("provider request = %+v", provider.request.Messages)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Content != "Recovered after restart." {
		t.Fatalf("messages = %+v", messages)
	}
}

func assertToolEventOrder(t *testing.T, events []productdata.RunEvent, toolCallID string, expected []string) {
	t.Helper()
	position := 0
	for _, event := range events {
		if metadataString(event.Metadata, "tool_call_id") != toolCallID {
			continue
		}
		if position < len(expected) && event.Type == expected[position] {
			position++
		}
	}
	if position != len(expected) {
		t.Fatalf("tool event order for %s missing %v in events %+v", toolCallID, expected[position:], events)
	}
}

func runErrorCodeForTest(run productdata.Run) string {
	if run.ErrorCode == nil {
		return ""
	}
	return *run.ErrorCode
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
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameCurrentTime}}}); err != nil {
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

func TestGatewayContinuationRequestsSecondCurrentTimeToolForApproval(t *testing.T) {
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
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameCurrentTime}}}); err != nil {
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
	if got.Status != productdata.RunStatusBlockedOnToolApproval {
		t.Fatalf("run = %+v error=%s", got, runErrorCodeForTest(got))
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_2")
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
	if got.Status != productdata.RunStatusQueued {
		t.Fatalf("run = %+v", got)
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read_2")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWorkspaceRead || call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
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

func TestGatewayContinuationRequestsMemoryToolForApproval(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Memory loop", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Use memory"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	enabled := []string{productdata.ToolNameMemorySearch, productdata.ToolNameMemoryRead}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": enabled}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_memory_search_1", ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": "plans", "limit": 5}, ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_memory_search_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_memory_search_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.CompleteToolCallSuccess(context.Background(), ident, thread.ID, run.ID, "tc_memory_search_1", map[string]any{"items": []any{map[string]any{"id": "mem_123", "title": "Plan"}}}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameMemoryRead, Metadata: map[string]any{"tool_call_id": "tc_memory_read_2", "arguments_summary": map[string]any{"entry_id": "mem_123"}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_memory_search_1"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusBlockedOnToolApproval {
		t.Fatalf("run = %+v error=%s", got, runErrorCodeForTest(got))
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_memory_read_2")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameMemoryRead || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("call = %+v", call)
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

func TestGatewayContinuationRejectsOverBudgetToolTurnAtomically(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace over budget", Mode: productdata.ThreadModeWork})
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
	for index := 0; index < productdata.DefaultMaxBoundedToolCallsPerRun-1; index++ {
		toolCallID := fmt.Sprintf("tc_%d", index+1)
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": toolCallID, "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": fmt.Sprintf("safe-%d.txt", index), "limit": 128}, "loop_index": index + 1}}); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": toolCallID, "tool_name": productdata.ToolNameWorkspaceRead, "result_summary": map[string]any{"path": fmt.Sprintf("safe-%d.txt", index), "truncated": false}}}); err != nil {
			t.Fatal(err)
		}
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": fmt.Sprintf("tc_%d", productdata.DefaultMaxBoundedToolCallsPerRun), "arguments_summary": map[string]any{"path": "within-budget.txt", "limit": 128}}},
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": fmt.Sprintf("tc_%d", productdata.DefaultMaxBoundedToolCallsPerRun+1), "arguments_summary": map[string]any{"path": "over-budget.txt", "limit": 128}}},
	}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: fmt.Sprintf("tc_%d", productdata.DefaultMaxBoundedToolCallsPerRun-1)})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "tool_loop_limit_reached" {
		t.Fatalf("run = %+v", got)
	}
	for _, toolCallID := range []string{fmt.Sprintf("tc_%d", productdata.DefaultMaxBoundedToolCallsPerRun), fmt.Sprintf("tc_%d", productdata.DefaultMaxBoundedToolCallsPerRun+1)} {
		if _, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, toolCallID); err == nil {
			t.Fatalf("over-budget tool call %s was recorded", toolCallID)
		}
	}
}

func TestGatewayContinuationAllowsMoreThanSixProjectSurveyTools(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Project survey", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "梳理页面和源码实现"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	enabled := []string{productdata.ToolNameWorkspaceListDirectory, productdata.ToolNameWorkspaceTreeSummary, productdata.ToolNameWorkspaceRead}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": enabled}}); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 6; index++ {
		toolCallID := fmt.Sprintf("tc_%d", index+1)
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": toolCallID, "tool_name": productdata.ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": fmt.Sprintf("page-%d.tsx", index+1), "limit": 128}, "loop_index": index + 1}}); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": toolCallID, "tool_name": productdata.ToolNameWorkspaceRead, "result_summary": map[string]any{"path": fmt.Sprintf("page-%d.tsx", index+1), "truncated": false}}}); err != nil {
			t.Fatal(err)
		}
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_7", "arguments_summary": map[string]any{"path": "page-7.tsx", "limit": 128}}}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_6"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status == productdata.RunStatusFailed {
		t.Fatalf("run = %+v error=%s", got, runErrorCodeForTest(got))
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_7")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWorkspaceRead {
		t.Fatalf("call = %+v", call)
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

func TestGatewayContinuationOmitsWorkspaceGlobAfterSuccessfulListing(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Workspace listing", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "List files"})
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
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_glob", "tool_name": productdata.ToolNameWorkspaceGlob, "arguments_summary": map[string]any{"pattern": "**/*", "limit": 500}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_glob", "tool_name": productdata.ToolNameWorkspaceGlob, "result_summary": map[string]any{"count": 3, "truncated": false}}}); err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "Final answer."}}}

	NewGateway(svc, nil, []Provider{provider}).ContinueAfterToolResult(context.Background(), run, GatewayContinuationInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model", ToolCallID: "tc_glob"})

	for _, tool := range provider.request.Tools {
		if tool.Name == productdata.ToolNameWorkspaceGlob {
			t.Fatalf("workspace.glob should be omitted after successful listing: %+v", provider.request.Tools)
		}
	}
	if !providerHasTool(provider.request.Tools, productdata.ToolNameWorkspaceRead) || !providerHasTool(provider.request.Tools, productdata.ToolNameWorkspaceGrep) {
		t.Fatalf("continuation should still expose read/grep tools: %+v", provider.request.Tools)
	}
}

func TestGatewayLoadToolsProviderSchemaIsQueriesOnly(t *testing.T) {
	tool, ok := builtinProviderToolDefinition(productdata.ToolNameLoadTools)
	if !ok {
		t.Fatal("load_tools provider definition missing")
	}
	properties, ok := tool.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties = %+v", tool.Parameters["properties"])
	}
	if _, ok := properties["names"]; ok {
		t.Fatalf("load_tools provider schema should not expose names: %+v", properties)
	}
	if _, ok := properties["queries"]; !ok {
		t.Fatalf("load_tools provider schema should expose queries: %+v", properties)
	}
	required, ok := tool.Parameters["required"].([]string)
	if !ok || len(required) != 0 {
		t.Fatalf("load_tools required = %+v", tool.Parameters["required"])
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
	enabled := []string{productdata.ToolNameLoadTools, productdata.ToolNameLoadSkill, productdata.ToolNameWorkspaceListDirectory, productdata.ToolNameWorkspaceTreeSummary, productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspaceEdit, productdata.ToolNameWorkspacePatchPreview, productdata.ToolNameWorkspacePatchApply, productdata.ToolNameSandboxExecCommand, productdata.ToolNameSandboxStartProcess, productdata.ToolNameSandboxContinueProcess, productdata.ToolNameSandboxTerminateProcess, productdata.ToolNameLSPSymbols, productdata.ToolNameLSPDefinition, productdata.ToolNameLSPHover, productdata.ToolNameWebSearch, productdata.ToolNameBrowserOpen, productdata.ToolNameBrowserScreenshot, productdata.ToolNameBrowserType, productdata.ToolNameBrowserPress, productdata.ToolNameMemorySearch, productdata.ToolNameMemoryList, productdata.ToolNameMemoryRead, productdata.ToolNameMemoryWrite, productdata.ToolNameMemoryEdit, productdata.ToolNameMemoryForget, productdata.ToolNameMemoryContext, productdata.ToolNameMemoryTimeline, productdata.ToolNameMemoryConnections, productdata.ToolNameMemoryThreadSearch, productdata.ToolNameMemoryThreadFetch, productdata.ToolNameMemoryStatus, productdata.ToolNameNotebookRead, productdata.ToolNameNotebookWrite, productdata.ToolNameNotebookEdit, productdata.ToolNameNotebookForget, productdata.ToolNameTodoWrite}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": enabled}}); err != nil {
		t.Fatal(err)
	}

	tools := NewGateway(svc, nil, nil).providerToolsForRun(context.Background(), run.ID, nil)
	names := []string{}
	for _, tool := range tools {
		names = append(names, tool.ProviderName)
		if tool.ProviderName == "" || len(tool.Parameters) == 0 {
			t.Fatalf("tool missing provider schema: %+v", tool)
		}
	}
	want := []string{"tool_load_tools", "skill_load_skill", "workspace_list_directory", "workspace_tree_summary", "workspace_read", "workspace_edit", "workspace_patch_preview", "workspace_patch_apply", "sandbox_exec_command", "sandbox_start_process", "sandbox_continue_process", "sandbox_terminate_process", "lsp_symbols", "lsp_definition", "lsp_hover", "web_search", "browser_open", "browser_screenshot", "browser_type", "browser_press", "memory_search", "memory_list", "memory_read", "memory_write", "memory_edit", "memory_forget", "memory_context", "memory_timeline", "memory_connections", "memory_thread_search", "memory_thread_fetch", "memory_status", "notebook_read", "notebook_write", "notebook_edit", "notebook_forget", "todo_write"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("provider tools = %+v", names)
	}
	listTool, ok := builtinProviderToolDefinition(productdata.ToolNameWorkspaceListDirectory)
	if !ok || !strings.Contains(listTool.Description, "Use this before grep for folder listing") {
		t.Fatalf("list directory provider schema = %+v", listTool)
	}
	if internalProviderToolName("tool_load_tools") != productdata.ToolNameLoadTools || providerToolName(productdata.ToolNameLoadSkill) != "skill_load_skill" || internalProviderToolName("workspace_list_directory") != productdata.ToolNameWorkspaceListDirectory || providerToolName(productdata.ToolNameWorkspaceTreeSummary) != "workspace_tree_summary" || internalProviderToolName("workspace_edit") != productdata.ToolNameWorkspaceEdit || internalProviderToolName("workspace_patch_preview") != productdata.ToolNameWorkspacePatchPreview || providerToolName(productdata.ToolNameWorkspacePatchApply) != "workspace_patch_apply" || providerToolName(productdata.ToolNameSandboxExecCommand) != "sandbox_exec_command" || internalProviderToolName("sandbox_start_process") != productdata.ToolNameSandboxStartProcess || internalProviderToolName("browser_type") != productdata.ToolNameBrowserType || providerToolName(productdata.ToolNameLSPDefinition) != "lsp_definition" || internalProviderToolName("memory_search") != productdata.ToolNameMemorySearch || internalProviderToolName("memory_thread_fetch") != productdata.ToolNameMemoryThreadFetch || providerToolName(productdata.ToolNameMemoryStatus) != "memory_status" || internalProviderToolName("notebook_write") != productdata.ToolNameNotebookWrite || providerToolName(productdata.ToolNameNotebookForget) != "notebook_forget" || internalProviderToolName("todo_write") != productdata.ToolNameTodoWrite {
		t.Fatalf("provider tool name mapping failed")
	}
}

func TestGatewayRunWithPreparedContextUsesPreparedToolDefinitions(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Prepared tools", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Read a file"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}
	prepared := &productdata.RunContext{
		Run:          run,
		Thread:       thread,
		EnabledTools: []productdata.ToolResolution{toolResolutionForTest(productdata.ToolNameWorkspaceRead)},
	}

	NewGateway(svc, nil, []Provider{provider}).runWithContext(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model"}, prepared)

	if !providerHasTool(provider.request.Tools, productdata.ToolNameWorkspaceRead) {
		t.Fatalf("tools = %+v", provider.request.Tools)
	}
}

func TestGatewayRecordToolCallRequestUsesPreparedScopedToolSnapshot(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Prepared scoped tool", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Read a file"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{
		config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true},
		events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read", "arguments_summary": map[string]any{"path": "safe.txt", "limit": 128}}}},
	}
	prepared := &productdata.RunContext{
		Run:          run,
		Thread:       thread,
		EnabledTools: []productdata.ToolResolution{toolResolutionForTest(productdata.ToolNameWorkspaceRead)},
	}

	NewGateway(svc, nil, []Provider{provider}).runWithContext(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom", Model: "model"}, prepared)

	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWorkspaceRead {
		t.Fatalf("call = %+v", call)
	}
}

func TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots(t *testing.T) {
	prompt := runSystemPrompt(&productdata.RunContext{
		Persona: productdata.PersonaSnapshot{SystemPrompt: "Use safe context."},
		Thread:  productdata.Thread{Mode: productdata.ThreadModeWork},
		ContextSources: []productdata.ContextSource{{
			Kind:    productdata.ContextSourceKindURL,
			Title:   "Docs",
			Locator: "https://example.com/docs",
			Summary: "Public docs",
		}, {
			Kind:    productdata.ContextSourceKindWorkspacePath,
			Title:   "Secrets",
			Locator: ".env",
			Summary: "sk-leak",
		}},
		MemorySnapshot: productdata.MemorySnapshot{Entries: []productdata.MemorySearchResult{{
			Title:      "Preference",
			Summary:    "Keep memory concise.",
			SourceType: "manual",
		}, {
			Title:      "Notebook duplicate",
			Summary:    "Should only appear in notebook block.",
			SourceType: "notebook",
		}}},
		NotebookSnapshot: productdata.MemorySnapshot{Entries: []productdata.MemorySearchResult{{
			Title:      "Project notebook",
			Summary:    "Use structured notes for durable facts.",
			SourceType: "notebook",
		}}},
	})
	if !strings.Contains(prompt, "<memory>\n- Preference: Keep memory concise.\n</memory>") {
		t.Fatalf("prompt missing memory block: %s", prompt)
	}
	if strings.Contains(prompt, "Notebook duplicate") {
		t.Fatalf("memory block should not include notebook entries: %s", prompt)
	}
	if !strings.Contains(prompt, "<notebook>\n- Project notebook: Use structured notes for durable facts.\n</notebook>") {
		t.Fatalf("prompt missing notebook block: %s", prompt)
	}
	if !strings.Contains(prompt, "<context_sources>\n- url Docs: https://example.com/docs — Public docs\n</context_sources>") {
		t.Fatalf("prompt missing context source block: %s", prompt)
	}
	if strings.Contains(prompt, "sk-") || strings.Contains(prompt, "/Users/") {
		t.Fatalf("prompt leaked unsafe content: %s", prompt)
	}
}

func TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "External memory", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/find" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "root-secret" {
			t.Fatalf("missing openviking root key")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"memories": []map[string]any{{"uri": "viking://user/user_local_dev/memories/pref", "abstract": "Preference\nUse provider memory before answering.", "score": 0.9, "match_reason": "query"}}}})
	}))
	defer server.Close()
	if _, err := svc.SaveMemoryProviderConfig(context.Background(), ident, productdata.MemoryProviderConfig{Enabled: true, Provider: productdata.MemoryProviderOpenViking, OpenViking: productdata.OpenVikingMemoryConfig{BaseURL: server.URL, RootAPIKey: "root-secret", EmbeddingModel: "embed", VLMModel: "vlm"}}); err != nil {
		t.Fatal(err)
	}
	prepared := &productdata.RunContext{Run: run, Thread: thread}
	enriched := NewGateway(svc, nil, nil).withExternalMemorySnapshot(context.Background(), prepared, []ProviderMessage{{Role: "user", Content: "What should I remember?"}})
	prompt := runSystemPrompt(enriched)
	if !strings.Contains(prompt, "<memory>\n- Preference: Use provider memory before answering.\n</memory>") {
		t.Fatalf("prompt missing external memory: %s", prompt)
	}
	if len(prepared.MemorySnapshot.Entries) != 0 {
		t.Fatalf("original prepared context should not be mutated: %+v", prepared.MemorySnapshot)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, event := range events {
		if event.Type == productdata.EventMemoryExternalSnapshotLoaded {
			found = true
			if event.Metadata["provider"] != string(productdata.MemoryProviderOpenViking) || fmt.Sprint(event.Metadata["entry_count"]) != "1" {
				t.Fatalf("external snapshot metadata = %+v", event.Metadata)
			}
		}
	}
	if !found {
		t.Fatalf("external snapshot event not recorded: %+v", events)
	}
}

func TestGatewayEnrichesPromptMemorySnapshotFromNowledgeProvider(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Nowledge external memory", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/memories/search" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("query") != "Recall project preference" || r.URL.Query().Get("limit") != "5" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") != "Bearer nowledge-secret" || r.Header.Get("x-nmem-api-key") != "nowledge-secret" {
			t.Fatalf("missing nowledge auth headers")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"memories": []map[string]any{{"id": "pref", "title": "Preference", "content": "Use Nowledge memory before answering.", "score": 0.88, "relevance_reason": "query"}}})
	}))
	defer server.Close()
	if _, err := svc.SaveMemoryProviderConfig(context.Background(), ident, productdata.MemoryProviderConfig{Enabled: true, Provider: productdata.MemoryProviderNowledge, Nowledge: productdata.NowledgeMemoryConfig{BaseURL: server.URL, APIKey: "nowledge-secret", RequestTimeoutMS: 30000}}); err != nil {
		t.Fatal(err)
	}
	prepared := &productdata.RunContext{Run: run, Thread: thread}
	enriched := NewGateway(svc, nil, nil).withExternalMemorySnapshot(context.Background(), prepared, []ProviderMessage{{Role: "user", Content: "Recall project preference"}})
	prompt := runSystemPrompt(enriched)
	if !strings.Contains(prompt, "<memory>\n- Preference: Use Nowledge memory before answering.\n</memory>") {
		t.Fatalf("prompt missing nowledge memory: %s", prompt)
	}
	if strings.Contains(prompt, "nowledge-secret") {
		t.Fatalf("prompt leaked nowledge secret: %s", prompt)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, event := range events {
		if event.Type == productdata.EventMemoryExternalSnapshotLoaded {
			found = true
			if event.Metadata["provider"] != string(productdata.MemoryProviderNowledge) || fmt.Sprint(event.Metadata["entry_count"]) != "1" {
				t.Fatalf("nowledge external snapshot metadata = %+v", event.Metadata)
			}
			if _, ok := event.Metadata["query"]; ok {
				t.Fatalf("nowledge external snapshot leaked unsafe metadata = %+v", event.Metadata)
			}
			if _, ok := event.Metadata["content"]; ok {
				t.Fatalf("nowledge external snapshot leaked unsafe metadata = %+v", event.Metadata)
			}
		}
	}
	if !found {
		t.Fatalf("nowledge external snapshot event not recorded: %+v", events)
	}
}

func TestGatewayRecordsExternalMemorySnapshotFailureForRecentErrors(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "External memory failure", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "sk-secret upstream trace", http.StatusBadGateway)
	}))
	defer server.Close()
	if _, err := svc.SaveMemoryProviderConfig(context.Background(), ident, productdata.MemoryProviderConfig{Enabled: true, Provider: productdata.MemoryProviderNowledge, Nowledge: productdata.NowledgeMemoryConfig{BaseURL: server.URL, APIKey: "nowledge-secret", RequestTimeoutMS: 30000}}); err != nil {
		t.Fatal(err)
	}
	prepared := &productdata.RunContext{Run: run, Thread: thread}
	enriched := NewGateway(svc, nil, nil).withExternalMemorySnapshot(context.Background(), prepared, []ProviderMessage{{Role: "user", Content: "Recall failure path"}})
	if enriched != prepared {
		t.Fatalf("failure should keep original prepared context")
	}
	errors, err := svc.ListMemoryProviderErrors(context.Background(), ident, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(errors) != 1 || errors[0].Code != productdata.EventMemoryExternalSnapshotFailed || errors[0].Provider != productdata.MemoryProviderNowledge || errors[0].RunID != run.ID {
		t.Fatalf("recent provider errors = %+v", errors)
	}
	encoded, _ := json.Marshal(errors)
	if strings.Contains(string(encoded), "nowledge-secret") || strings.Contains(string(encoded), "sk-secret") || strings.Contains(string(encoded), "Recall failure path") {
		t.Fatalf("recent provider error leaked unsafe data: %s", string(encoded))
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

type sequencedProvider struct {
	config    ProviderConfig
	requests  []ProviderRequest
	eventSets [][]ProviderEvent
}

type streamErrorThenSuccessProvider struct {
	config   ProviderConfig
	requests []ProviderRequest
	err      error
}

type scriptedToolExecutor struct {
	results map[string]map[string]any
}

type concurrentSleepToolExecutor struct {
	mu          sync.Mutex
	inFlight    int
	maxInFlight int
	delay       time.Duration
}

func providerHasTool(tools []ProviderToolDefinition, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func hasRunEventType(events []productdata.RunEvent, eventType string) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
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

func (p *sequencedProvider) Config() ProviderConfig { return p.config }

func (p *sequencedProvider) Stream(_ context.Context, request ProviderRequest) (<-chan ProviderEvent, error) {
	p.requests = append(p.requests, request)
	index := len(p.requests) - 1
	events := []ProviderEvent{{Type: ProviderEventCompleted, Text: "ok"}}
	if index < len(p.eventSets) {
		events = p.eventSets[index]
	}
	ch := make(chan ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func (p *streamErrorThenSuccessProvider) Config() ProviderConfig { return p.config }

func (p *streamErrorThenSuccessProvider) Stream(_ context.Context, request ProviderRequest) (<-chan ProviderEvent, error) {
	p.requests = append(p.requests, request)
	if len(p.requests) == 1 {
		return nil, p.err
	}
	ch := make(chan ProviderEvent, 1)
	ch <- ProviderEvent{Type: ProviderEventCompleted, Text: "Recovered."}
	close(ch)
	return ch, nil
}

func (e scriptedToolExecutor) ExecuteTool(_ context.Context, invocation ToolInvocation) (ToolResult, error) {
	result := e.results[invocation.ToolCallID]
	if result == nil {
		result = map[string]any{"ok": true}
	}
	return ToolResult{ResultSummary: result}, nil
}

func (e *concurrentSleepToolExecutor) ExecuteTool(ctx context.Context, invocation ToolInvocation) (ToolResult, error) {
	e.mu.Lock()
	e.inFlight++
	if e.inFlight > e.maxInFlight {
		e.maxInFlight = e.inFlight
	}
	e.mu.Unlock()

	timer := time.NewTimer(e.delay)
	select {
	case <-timer.C:
	case <-ctx.Done():
		timer.Stop()
		e.mu.Lock()
		e.inFlight--
		e.mu.Unlock()
		return ToolResult{}, ctx.Err()
	}
	e.mu.Lock()
	e.inFlight--
	e.mu.Unlock()
	return ToolResult{ResultSummary: map[string]any{"tool_call_id": invocation.ToolCallID, "ok": true}}, nil
}

func toolResolutionForTest(name string) productdata.ToolResolution {
	return productdata.ToolResolution{
		Name:           name,
		ApprovalPolicy: string(productdata.ToolApprovalReadOnly),
		ExecutionState: string(productdata.ToolExecutionStateExecutable),
		Source:         string(productdata.ToolCatalogSourceBuiltin),
		Group:          string(productdata.ToolCatalogGroupRuntime),
		RiskLevel:      string(productdata.ToolRiskLow),
	}
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

func TestQueuedRunRouterHydratesContinuationInputFromRunStepProjection(t *testing.T) {
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
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "projected-model"})
	if err != nil {
		t.Fatal(err)
	}

	input := (QueuedRunRouter{Gateway: NewGateway(svc, nil, nil)}).gatewayContinuationInput(context.Background(), run, "tc_1", nil)

	if input.ThreadID != thread.ID || input.ToolCallID != "tc_1" || input.MessageID != message.ID || input.ProviderID != "custom" || input.Model != "projected-model" {
		t.Fatalf("input = %+v", input)
	}
}

func TestQueuedRunRouterExecutesParallelAutoApprovedToolsBeforeContinuation(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "search and fetch"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameLoadTools, productdata.ToolNameWorkspaceRead}}}); err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_gateway", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameLoadTools, Metadata: map[string]any{"tool_call_id": "tc_load_tools", "arguments_summary": map[string]any{"names": []any{productdata.ToolNameWorkspaceRead}}}},
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read", "arguments_summary": map[string]any{"path": "notes.txt", "limit": 128}}},
	}, {
		{Type: ProviderEventCompleted, Text: "Both tools finished."},
	}}}
	enabledTools := []productdata.ToolResolution{
		toolResolutionForTest(productdata.ToolNameLoadTools),
		toolResolutionForTest(productdata.ToolNameWorkspaceRead),
	}
	prepared := &productdata.RunContext{EnabledTools: enabledTools}
	prepared.Thread = productdata.Thread{Mode: productdata.ThreadModeWork}
	prepared.WorkspaceRoot = productdata.WorkspaceRootConfig{Path: "/Users/xuean/Downloads", DisplayName: "Downloads"}
	executor := scriptedToolExecutor{results: map[string]map[string]any{
		"tc_load_tools": {"tools": []any{productdata.ToolNameWorkspaceRead}},
		"tc_read":       {"path": "notes.txt", "content": "hello"},
	}}
	broadcaster := NewBroadcaster()
	liveCtx, liveCancel := context.WithCancel(context.Background())
	defer liveCancel()
	live := broadcaster.Subscribe(liveCtx, run.ID)
	liveKeys := make(chan string, 32)
	go func() {
		for event := range live {
			toolCallID, _ := event.Metadata["tool_call_id"].(string)
			liveKeys <- event.Type + ":" + toolCallID
		}
	}()

	if err := (QueuedRunRouter{Gateway: NewGateway(svc, broadcaster, []Provider{provider}), Local: &LocalRunner{Service: svc, Broadcaster: broadcaster}, WebExecutor: WebToolExecutor{}}).dispatchWithExecutorForTest(context.Background(), claimedRun, job, prepared, executor); err != nil {
		t.Fatal(err)
	}
	seenToolEvents := map[string]bool{}
	deadline := time.After(time.Second)
	for !(seenToolEvents[productdata.EventToolCallExecuting+":tc_load_tools"] &&
		seenToolEvents[productdata.EventToolCallSucceeded+":tc_load_tools"] &&
		seenToolEvents[productdata.EventToolCallExecuting+":tc_read"] &&
		seenToolEvents[productdata.EventToolCallSucceeded+":tc_read"]) {
		select {
		case key := <-liveKeys:
			seenToolEvents[key] = true
		case <-deadline:
			t.Fatalf("missing live tool execution events: %+v", seenToolEvents)
		}
	}

	if len(provider.requests) != 2 {
		t.Fatalf("provider requests = %d, want 2", len(provider.requests))
	}
	if !strings.Contains(provider.requests[1].SystemPrompt, "Selected workspace: Downloads") {
		t.Fatalf("continuation system prompt = %q", provider.requests[1].SystemPrompt)
	}
	continuationMessages := provider.requests[1].Messages
	if len(continuationMessages) < 5 {
		t.Fatalf("continuation messages = %+v", continuationMessages)
	}
	if continuationMessages[len(continuationMessages)-4].ToolCallID != "tc_load_tools" ||
		continuationMessages[len(continuationMessages)-3].ToolCallID != "tc_load_tools" ||
		continuationMessages[len(continuationMessages)-2].ToolCallID != "tc_read" ||
		continuationMessages[len(continuationMessages)-1].ToolCallID != "tc_read" {
		t.Fatalf("continuation messages = %+v", continuationMessages)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
}

func TestQueuedRunRouterRunsReadyAutoApprovedToolsConcurrently(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "read two files"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWorkspaceRead}}}); err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_gateway", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_a", "arguments_summary": map[string]any{"path": "a.txt", "limit": 128}}},
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_b", "arguments_summary": map[string]any{"path": "b.txt", "limit": 128}}},
	}, {
		{Type: ProviderEventCompleted, Text: "Both reads finished."},
	}}}
	enabledTools := []productdata.ToolResolution{toolResolutionForTest(productdata.ToolNameWorkspaceRead)}
	prepared := &productdata.RunContext{EnabledTools: enabledTools}
	executor := &concurrentSleepToolExecutor{delay: 150 * time.Millisecond}

	done := make(chan error, 1)
	go func() {
		done <- (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider}), Local: &LocalRunner{Service: svc}}).dispatchWithExecutorForTest(context.Background(), claimedRun, job, prepared, executor)
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("dispatch did not return while waiting for unresolved parallel tool call")
	}

	if executor.maxInFlight < 2 {
		t.Fatalf("tools did not run concurrently, max in-flight = %d", executor.maxInFlight)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("provider requests = %d, want 2", len(provider.requests))
	}
	for _, toolCallID := range []string{"tc_read_a", "tc_read_b"} {
		call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, toolCallID)
		if err != nil {
			t.Fatal(err)
		}
		if call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
			t.Fatalf("call %s = %+v", toolCallID, call)
		}
	}
}

func TestQueuedRunRouterDoesNotContinueUntilAllParallelToolCallsResolved(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "read a file and check time"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameWorkspaceRead, productdata.ToolNameCurrentTime}}}); err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_gateway", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read", "arguments_summary": map[string]any{"path": "notes.txt", "limit": 128}}},
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_time", "arguments_summary": map[string]any{"timezone": "UTC"}}},
	}, {
		{Type: ProviderEventCompleted, Text: "This should wait."},
	}}}
	prepared := &productdata.RunContext{
		Thread:       productdata.Thread{Mode: productdata.ThreadModeWork},
		EnabledTools: []productdata.ToolResolution{toolResolutionForTest(productdata.ToolNameWorkspaceRead), toolResolutionForTest(productdata.ToolNameCurrentTime)},
	}
	executor := scriptedToolExecutor{results: map[string]map[string]any{"tc_read": {"path": "notes.txt", "content": "hello"}}}

	done := make(chan error, 1)
	go func() {
		done <- (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider}), Local: &LocalRunner{Service: svc}}).dispatchWithExecutorForTest(context.Background(), claimedRun, job, prepared, executor)
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("dispatch did not return while waiting for unresolved parallel tool call")
	}

	if len(provider.requests) != 1 {
		t.Fatalf("provider requests = %d, want initial request only", len(provider.requests))
	}
	readCall, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_read")
	if err != nil {
		t.Fatal(err)
	}
	if readCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("read call = %+v", readCall)
	}
	timeCall, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_time")
	if err != nil {
		t.Fatal(err)
	}
	if timeCall.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("time call = %+v", timeCall)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusBlockedOnToolApproval {
		t.Fatalf("run = %+v", got)
	}
}

func TestQueuedRunRouterDrainsParallelAutoApprovedToolsAfterContinuation(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "load then read two files"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{"step": string(productdata.PipelineStepResolveTools), "enabled_tools": []string{productdata.ToolNameLoadTools, productdata.ToolNameWorkspaceRead}}}); err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_gateway", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	provider := &sequencedProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, eventSets: [][]ProviderEvent{{
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameLoadTools, Metadata: map[string]any{"tool_call_id": "tc_load_tools", "arguments_summary": map[string]any{"names": []any{productdata.ToolNameWorkspaceRead}}}},
	}, {
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_a", "arguments_summary": map[string]any{"path": "a.txt", "limit": 128}}},
		{Type: ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_b", "arguments_summary": map[string]any{"path": "b.txt", "limit": 128}}},
	}, {
		{Type: ProviderEventCompleted, Text: "Both continuation tools finished."},
	}}}
	enabledTools := []productdata.ToolResolution{
		toolResolutionForTest(productdata.ToolNameLoadTools),
		toolResolutionForTest(productdata.ToolNameWorkspaceRead),
	}
	prepared := &productdata.RunContext{EnabledTools: enabledTools}
	executor := scriptedToolExecutor{results: map[string]map[string]any{
		"tc_load_tools": {"tools": []any{productdata.ToolNameWorkspaceRead}},
		"tc_read_a":     {"path": "a.txt", "content": "alpha"},
		"tc_read_b":     {"path": "b.txt", "content": "beta"},
	}}

	if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider}), Local: &LocalRunner{Service: svc}}).dispatchWithExecutorForTest(context.Background(), claimedRun, job, prepared, executor); err != nil {
		t.Fatal(err)
	}

	if len(provider.requests) != 3 {
		t.Fatalf("provider requests = %d, want 3", len(provider.requests))
	}
	finalMessages := provider.requests[2].Messages
	if len(finalMessages) < 7 {
		t.Fatalf("final continuation messages = %+v", finalMessages)
	}
	if finalMessages[len(finalMessages)-4].ToolCallID != "tc_read_a" ||
		finalMessages[len(finalMessages)-3].ToolCallID != "tc_read_a" ||
		finalMessages[len(finalMessages)-2].ToolCallID != "tc_read_b" ||
		finalMessages[len(finalMessages)-1].ToolCallID != "tc_read_b" {
		t.Fatalf("final continuation messages = %+v", finalMessages)
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

func TestGatewayRunAsyncCancelsProviderAfterStopRun(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway stop", Mode: productdata.ThreadModeChat})
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
	provider := newBlockingDeltaProvider("partial")

	NewGateway(svc, nil, []Provider{provider}).RunAsync(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	select {
	case <-provider.firstSent:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	if _, err := svc.StopRun(context.Background(), ident, run.ID); err != nil {
		t.Fatal(err)
	}
	got := waitForTerminalRun(t, svc, run.ID)
	if got.Status != productdata.RunStatusStopped {
		t.Fatalf("run = %+v", got)
	}
	select {
	case provider.release <- struct{}{}:
	default:
	}
	time.Sleep(150 * time.Millisecond)
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
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

type blockingDeltaProvider struct {
	config    ProviderConfig
	text      string
	firstSent chan struct{}
	release   chan struct{}
}

func newBlockingDeltaProvider(text string) *blockingDeltaProvider {
	return &blockingDeltaProvider{
		config:    ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true},
		text:      text,
		firstSent: make(chan struct{}),
		release:   make(chan struct{}),
	}
}

func (p *blockingDeltaProvider) Config() ProviderConfig { return p.config }

func (p *blockingDeltaProvider) Stream(ctx context.Context, _ ProviderRequest) (<-chan ProviderEvent, error) {
	ch := make(chan ProviderEvent)
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
			return
		case ch <- ProviderEvent{Type: ProviderEventTextDelta, Text: p.text}:
			close(p.firstSent)
		}
		select {
		case <-ctx.Done():
			return
		case <-p.release:
		}
		select {
		case <-ctx.Done():
		case ch <- ProviderEvent{Type: ProviderEventCompleted, Text: p.text}:
		}
	}()
	return ch, nil
}

func waitForRunEventType(t *testing.T, svc productdata.Service, runID string, eventType string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		events, err := svc.ListRunEvents(context.Background(), identity.LocalDevIdentity(), runID, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			if event.Type == eventType {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("run event %q not found", eventType)
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
