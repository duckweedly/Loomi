package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM21WorkspaceReadToolsSmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	for _, tc := range []struct {
		name      string
		toolName  string
		args      map[string]any
		wantKey   string
		wantValue string
	}{
		{name: "glob", toolName: productdata.ToolNameWorkspaceGlob, args: map[string]any{"pattern": "**/*.txt", "limit": 10}, wantKey: "match_count", wantValue: "2"},
		{name: "grep", toolName: productdata.ToolNameWorkspaceGrep, args: map[string]any{"query": "needle", "path": "src", "include": "*.txt", "limit": 10}, wantKey: "match_count", wantValue: "1"},
		{name: "read", toolName: productdata.ToolNameWorkspaceRead, args: map[string]any{"path": "src/notes.txt", "limit": 6}, wantKey: "content", wantValue: "needle"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, eventsBody := runM21WorkspaceTool(t, tc.toolName, tc.args, true)
			if got := strings.TrimSpace(stringValue(result, tc.wantKey)); got != tc.wantValue {
				t.Fatalf("%s result = %+v", tc.name, result)
			}
			assertBodyExcludes(t, eventsBody, tc.name+" events", root, "fixture-secret", ".env", "secrets/token")
		})
	}

	result, eventsBody := runM21WorkspaceTool(t, productdata.ToolNameWorkspaceRead, map[string]any{"path": ".env"}, true)
	if result["error_code"] == nil && !strings.Contains(eventsBody, "tool_call_failed") {
		t.Fatalf("sensitive read should fail: result=%+v events=%s", result, eventsBody)
	}
	assertBodyExcludes(t, eventsBody, "sensitive failure events", "fixture-secret")

	_, symlinkEvents := runM21WorkspaceTool(t, productdata.ToolNameWorkspaceRead, map[string]any{"path": "src/outside-link.txt"}, true)
	if !strings.Contains(symlinkEvents, "tool_call_failed") {
		t.Fatalf("symlink escape should fail: events=%s", symlinkEvents)
	}
	assertBodyExcludes(t, symlinkEvents, "symlink failure events", "outside-secret")

	_, autoEvents := runM21WorkspaceTool(t, productdata.ToolNameWorkspaceRead, map[string]any{"path": "src/notes.txt"}, false)
	if strings.Contains(autoEvents, productdata.EventToolCallApprovalRequired) {
		t.Fatalf("workspace read should use directory-scoped auto execution: %s", autoEvents)
	}
}

func TestM21WorkspaceReadToolsRejectChatModeProviderRequest(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)
	const toolCallID = "tc_m21_chat_workspace"

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Do not use workspace tools in chat mode.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameCurrentTime, productdata.ToolNameWorkspaceRead},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m21WorkspaceProvider{toolName: productdata.ToolNameWorkspaceRead, toolCallID: toolCallID, args: map[string]any{"path": "src/notes.txt"}}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m21_chat_workspace"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M21 chat workspace rejected","mode":"chat"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Read workspace","client_message_id":"m21-chat-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); !ok || err == nil || !strings.Contains(err.Error(), "Queued gateway run did not complete") {
		t.Fatalf("ProcessOne ok=%v err=%v", ok, err)
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusFailed || run.ErrorCode == nil || *run.ErrorCode != "tool_call_rejected" {
		t.Fatalf("chat workspace run = %+v", run)
	}
	if _, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID); err == nil {
		t.Fatal("chat workspace tool call was recorded; want rejected before approval")
	}
	eventsBody := fetchM21Events(t, srv, runID)
	if strings.Contains(eventsBody, productdata.EventToolCallApprovalRequired) || strings.Contains(eventsBody, productdata.EventToolCallSucceeded) {
		t.Fatalf("chat workspace events should not enter approval/execution: %s", eventsBody)
	}
	assertBodyExcludes(t, eventsBody, "chat workspace rejection events", root, "needle", "fixture-secret")
}

func runM21WorkspaceTool(t *testing.T, toolName string, args map[string]any, approve bool) (map[string]any, string) {
	t.Helper()
	const toolCallID = "tc_m21_workspace"
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved workspace read tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameCurrentTime, productdata.ToolNameWorkspaceGlob, productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceRead},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m21WorkspaceProvider{toolName: toolName, toolCallID: toolCallID, args: args}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m21_workspace"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M21 workspace smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Run workspace tool","client_message_id":"m21-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		if call, getErr := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID); getErr == nil {
			return call.ResultSummary, fetchM21Events(t, srv, runID)
		}
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ApprovalStatus == productdata.ToolCallApprovalRequired {
		if !approve {
			return map[string]any{}, fetchM21Events(t, srv, runID)
		}
		approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
		assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	} else if call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
		if call.ExecutionStatus == productdata.ToolCallExecutionSucceeded {
			return call.ResultSummary, fetchM21Events(t, srv, runID)
		}
		if call.ExecutionStatus == productdata.ToolCallExecutionFailed {
			return toolCallErrorSummary(call), fetchM21Events(t, srv, runID)
		}
		t.Fatalf("pre-execution call = %+v", call)
	}
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("second ProcessOne ok=%v err=%v", ok, err)
	}
	finalCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if finalCall.ExecutionStatus == productdata.ToolCallExecutionFailed {
		return toolCallErrorSummary(finalCall), fetchM21Events(t, srv, runID)
	}
	return finalCall.ResultSummary, fetchM21Events(t, srv, runID)
}

func toolCallErrorSummary(call productdata.ToolCall) map[string]any {
	result := map[string]any{"ok": false}
	if call.ErrorCode != nil {
		result["error_code"] = *call.ErrorCode
	}
	if call.ErrorMessage != nil {
		result["error_message"] = *call.ErrorMessage
	}
	return result
}

type m21WorkspaceProvider struct {
	toolName   string
	toolCallID string
	args       map[string]any
	calls      int
}

func (p *m21WorkspaceProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m21WorkspaceProvider) Stream(_ context.Context, _ productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: p.toolName, Metadata: map[string]any{"tool_call_id": p.toolCallID, "arguments_summary": p.args}}}
	if p.calls == 2 {
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M21 workspace "}, {Type: productruntime.ProviderEventCompleted, Text: "M21 workspace complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func createM21WorkspaceFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "secrets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "notes.txt"), []byte("needle\nsecond\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "other.txt"), []byte("haystack\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("fixture-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "secrets", "token.txt"), []byte("fixture-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "outside.txt"), []byte("outside-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "outside.txt"), filepath.Join(root, "src", "outside-link.txt")); err != nil {
		t.Fatal(err)
	}
	return root
}

func fetchM21Events(t *testing.T, srv http.Handler, runID string) string {
	t.Helper()
	eventsRes := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+runID+"/events", "")
	assertStatus(t, eventsRes.Code, http.StatusOK, eventsRes.Body.String())
	return eventsRes.Body.String()
}

func stringValue(values map[string]any, key string) string {
	switch value := values[key].(type) {
	case string:
		return value
	case float64:
		return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintf("%.0f", value), ".0"), ".")
	case int:
		return fmt.Sprint(value)
	default:
		return fmt.Sprint(value)
	}
}
