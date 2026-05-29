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

func TestM22BoundedAgentLoopWorkspaceSmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved workspace read tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWorkspaceGlob, productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceRead},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m22WorkspaceLoopProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m22_workspace_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M22 workspace loop","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Inspect project files","client_message_id":"m22-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("initial ProcessOne ok=%v err=%v", ok, err)
	}
	drainM22Worker(t, worker, svc, ident, runID)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_glob_1", productdata.ToolNameWorkspaceGlob)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_read_2", productdata.ToolNameWorkspaceRead)
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	messages, err := svc.ListMessages(context.Background(), ident, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "M22 workspace loop complete." {
		t.Fatalf("messages = %+v", messages)
	}
	readCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_read_2")
	if err != nil {
		t.Fatal(err)
	}
	if readCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded || stringValue(readCall.ResultSummary, "content") != "needle" {
		t.Fatalf("read call = %+v", readCall)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{productdata.EventToolCallSucceeded, `"tool_call_id":"tc_read_2"`, `"loop_index":2`, fmt.Sprintf(`"loop_max":%d`, productdata.DefaultMaxBoundedToolCallsPerRun), `"model_phase":"continuation"`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	if strings.Contains(eventsBody, productdata.EventToolCallApprovalRequired) {
		t.Fatalf("read-only workspace loop should not require approval: %s", eventsBody)
	}
	assertBodyExcludes(t, eventsBody, "m22 bounded loop events", root, "fixture-secret", ".env", "secrets/token")
}

func TestM22CodeAgentReadEditExecReadLoopSmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved code-agent tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspaceEdit, productdata.ToolNameSandboxExecCommand},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m22CodeAgentLoopProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m22_code_agent_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M22 code-agent loop","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Patch and verify workspace file","client_message_id":"m22-code-agent-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("read-before ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_read_before_1", productdata.ToolNameWorkspaceRead)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_edit_2", productdata.ToolNameWorkspaceEdit)
	approveEdit := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/tc_edit_2/approve", "")
	assertStatus(t, approveEdit.Code, http.StatusOK, approveEdit.Body.String())

	for _, step := range []struct {
		id   string
		name string
	}{
		{"tc_exec_3", productdata.ToolNameSandboxExecCommand},
	} {
		if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
			run, _ := svc.GetRun(context.Background(), ident, runID)
			errorCode := ""
			errorMessage := ""
			if run.ErrorCode != nil {
				errorCode = *run.ErrorCode
			}
			if run.ErrorMessage != nil {
				errorMessage = *run.ErrorMessage
			}
			t.Fatalf("%s ProcessOne ok=%v err=%v status=%s error=%s message=%s", step.id, ok, err, run.Status, errorCode, errorMessage)
		}
		assertM22ToolBlocked(t, svc, threadID, runID, step.id, step.name)
		approve := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+step.id+"/approve", "")
		assertStatus(t, approve.Code, http.StatusOK, approve.Body.String())
	}

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("read-after ProcessOne ok=%v err=%v", ok, err)
	}
	drainM22Worker(t, worker, svc, ident, runID)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_read_after_4", productdata.ToolNameWorkspaceRead)

	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	messages, err := svc.ListMessages(context.Background(), ident, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "Code-agent loop complete." {
		t.Fatalf("messages = %+v", messages)
	}
	readAfter, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_read_after_4")
	if err != nil {
		t.Fatal(err)
	}
	if readAfter.ExecutionStatus != productdata.ToolCallExecutionSucceeded || !strings.Contains(stringValue(readAfter.ResultSummary, "content"), "thread") {
		t.Fatalf("read after call = %+v", readAfter)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{`"tool_call_id":"tc_read_after_4"`, `"loop_index":4`, fmt.Sprintf(`"loop_max":%d`, productdata.DefaultMaxBoundedToolCallsPerRun), `"tool_name":"sandbox.exec_command"`, productdata.EventWorkTodoUpdated, "Run validation command", productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m22 code-agent loop events", root, "fixture-secret", ".env", "secrets/token")
}

func TestM75CodeAgentDailyLoopSmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use the daily code-agent loop: grep, read, preview patch, apply patch, run tests, summarize.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspacePatchPreview, productdata.ToolNameWorkspacePatchApply, productdata.ToolNameSandboxExecCommand},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m75CodeAgentDailyLoopProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m75_code_agent_daily_loop"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M75 code-agent daily loop","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Patch code and run go test validation","client_message_id":"m75-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("preview ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_grep_1", productdata.ToolNameWorkspaceGrep)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_read_2", productdata.ToolNameWorkspaceRead)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_preview_3", productdata.ToolNameWorkspacePatchPreview)
	beforePreviewApprove, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(beforePreviewApprove) != "needle\nsecond\n" {
		t.Fatalf("patch preview changed file before approval: %q", string(beforePreviewApprove))
	}
	approveM75Tool(t, srv, threadID, runID, "tc_preview_3")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("apply ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_preview_3", productdata.ToolNameWorkspacePatchPreview)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_apply_4", productdata.ToolNameWorkspacePatchApply)
	beforeApplyApprove, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(beforeApplyApprove) != "needle\nsecond\n" {
		t.Fatalf("patch apply changed file before approval: %q", string(beforeApplyApprove))
	}
	approveM75Tool(t, srv, threadID, runID, "tc_apply_4")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("exec ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_apply_4", productdata.ToolNameWorkspacePatchApply)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_exec_5", productdata.ToolNameSandboxExecCommand)
	approveM75Tool(t, srv, threadID, runID, "tc_exec_5")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("final ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_exec_5", productdata.ToolNameSandboxExecCommand)

	edited, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(edited) != "daily loop\nsecond\n" {
		t.Fatalf("edited notes = %q", string(edited))
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	messages, err := svc.ListMessages(context.Background(), ident, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "Daily loop complete: patch applied and tests passed." {
		t.Fatalf("messages = %+v", messages)
	}
	previewCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_preview_3")
	if err != nil {
		t.Fatal(err)
	}
	if previewCall.ResultSummary["operation"] != "patch_preview" || previewCall.ResultSummary["changed"] != false || !strings.Contains(stringValue(previewCall.ResultSummary, "diff"), "+daily loop") {
		t.Fatalf("preview call = %+v", previewCall)
	}
	applyCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_apply_4")
	if err != nil {
		t.Fatal(err)
	}
	if applyCall.ResultSummary["operation"] != "patch_apply" || applyCall.ResultSummary["changed"] != true || !strings.Contains(stringValue(applyCall.ResultSummary, "diff"), "+daily loop") {
		t.Fatalf("apply call = %+v", applyCall)
	}
	execCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_exec_5")
	if err != nil {
		t.Fatal(err)
	}
	if execCall.ResultSummary["operation"] != "exec_command" || execCall.ResultSummary["exit_code"] != 0 || !strings.Contains(stringValue(execCall.ResultSummary, "stdout"), "daily loop") {
		t.Fatalf("exec call = %+v", execCall)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{
		`"tool_call_id":"tc_grep_1"`,
		`"tool_call_id":"tc_read_2"`,
		`"tool_call_id":"tc_preview_3"`,
		`"tool_call_id":"tc_apply_4"`,
		`"tool_call_id":"tc_exec_5"`,
		`"operation":"patch_preview"`,
		`"operation":"patch_apply"`,
		`"operation":"exec_command"`,
		`"loop_index":5`,
		fmt.Sprintf(`"loop_max":%d`, productdata.DefaultMaxBoundedToolCallsPerRun),
		`"model_phase":"continuation"`,
		productdata.EventToolCallApprovalRequired,
		productdata.EventToolCallApproved,
		productdata.EventToolCallExecuting,
		productdata.EventToolCallSucceeded,
		productdata.EventRunCompleted,
	} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	assertBodyExcludes(t, eventsBody, "m75 daily loop events", root, "fixture-secret", ".env", "secrets/token")
}

func TestM91DirectoryClassificationSmoke(t *testing.T) {
	root := t.TempDir()
	for _, file := range []string{
		"src/main.go",
		"docs/spec.md",
		"config/app.json",
		"dist/generated.js",
		"tmp/cache.tmp",
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, file)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, file), []byte("fixture\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Classify selected directories with directory tools before reading files.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWorkspaceListDirectory, productdata.ToolNameWorkspaceTreeSummary, productdata.ToolNameWorkspaceRead},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m91DirectoryClassificationProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m91_directory_classification"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M91 directory classification","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"请看一下当前选择目录都有哪些东西，按源码/文档/配置/构建产物/临时文件分类列出。","client_message_id":"m91-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("directory classification ProcessOne ok=%v err=%v", ok, err)
	}
	drainM22Worker(t, worker, svc, ident, runID)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m91_list_1", productdata.ToolNameWorkspaceListDirectory)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m91_summary_2", productdata.ToolNameWorkspaceTreeSummary)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m91_read_3", productdata.ToolNameWorkspaceRead)
	if _, err := svc.GetToolCall(context.Background(), ident, threadID, runID, "tc_m91_grep"); err == nil {
		t.Fatal("directory smoke should not use grep")
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	messages, err := svc.ListMessages(context.Background(), ident, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || !strings.Contains(messages[1].Content, "源码") || !strings.Contains(messages[1].Content, "文档") || strings.Contains(messages[1].Content, "```json") {
		t.Fatalf("messages = %+v", messages)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{`"tool_name":"workspace.list_directory"`, `"tool_name":"workspace.tree_summary"`, `"tool_name":"workspace.read"`, `"operation":"tree_summary"`, productdata.EventRunCompleted} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	if strings.Contains(eventsBody, `"tool_name":"workspace.grep"`) {
		t.Fatalf("directory classification should not grep: %s", eventsBody)
	}
	assertBodyExcludes(t, eventsBody, "m91 directory classification events", root, "/Users/", "secret-token")
}

func TestM75CodeAgentDailyLoopSafetyBoundaries(t *testing.T) {
	t.Run("workspace tool is rejected outside work mode", func(t *testing.T) {
		root := createM21WorkspaceFixture(t)
		t.Setenv("LOOMI_WORKSPACE_ROOT", root)
		const toolCallID = "tc_m75_unauthorized_workspace"

		svc := productdata.NewMemoryService()
		ident := identity.LocalDevIdentity()
		if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
			Slug:             "default",
			Name:             "Default",
			Description:      "Default",
			SystemPrompt:     "Workspace tools must be Work-mode only.",
			ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
			AllowedToolNames: []string{productdata.ToolNameWorkspacePatchPreview},
			ReasoningMode:    "balanced",
			BudgetSummary:    "small",
			Version:          "1",
			IsDefault:        true,
		}}); err != nil {
			t.Fatal(err)
		}
		provider := &m21WorkspaceProvider{toolName: productdata.ToolNameWorkspacePatchPreview, toolCallID: toolCallID, args: map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "daily loop\n"}}
		gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
		worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
		worker.WorkerID = "worker_m75_unauthorized_workspace"
		srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

		threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M75 unauthorized workspace","mode":"chat"}`)
		assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
		threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
		messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Preview patch","client_message_id":"m75-unauthorized-user-message"}`)
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
			t.Fatalf("unauthorized workspace run = %+v", run)
		}
		if _, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID); err == nil {
			t.Fatal("unauthorized workspace tool call was recorded")
		}
		eventsBody := fetchM21Events(t, srv, runID)
		if strings.Contains(eventsBody, productdata.EventToolCallApprovalRequired) || strings.Contains(eventsBody, productdata.EventToolCallExecuting) || strings.Contains(eventsBody, productdata.EventToolCallSucceeded) {
			t.Fatalf("unauthorized workspace events should not execute: %s", eventsBody)
		}
	})

	t.Run("stale patch preview fails without writing", func(t *testing.T) {
		root := createM21WorkspaceFixture(t)
		t.Setenv("LOOMI_WORKSPACE_ROOT", root)
		result, eventsBody := runM75StalePatchApply(t, root)
		edited, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(edited) != "external change\nsecond\n" {
			t.Fatalf("stale patch apply wrote file: result=%+v edited=%q", result, string(edited))
		}
		if result["error_code"] == nil || !strings.Contains(stringValue(result, "error_message"), "read before editing") {
			t.Fatalf("stale patch result = %+v", result)
		}
		if !strings.Contains(eventsBody, productdata.EventToolCallFailed) || strings.Contains(eventsBody, productdata.EventRunCompleted) {
			t.Fatalf("stale patch events = %s", eventsBody)
		}
	})

	t.Run("terminal run cannot approve pending mutation", func(t *testing.T) {
		svc := productdata.NewMemoryService()
		ident := identity.LocalDevIdentity()
		thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "M75 terminal boundary", Mode: productdata.ThreadModeWork})
		if err != nil {
			t.Fatal(err)
		}
		message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Patch code and run go test"})
		if err != nil {
			t.Fatal(err)
		}
		run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{
			ToolCallID:       "tc_terminal_apply",
			ToolName:         productdata.ToolNameWorkspacePatchApply,
			ArgumentsSummary: map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "daily loop\n"},
			ArgumentsHash:    "hash_terminal_apply",
			ApprovalStatus:   productdata.ToolCallApprovalRequired,
			ExecutionStatus:  productdata.ToolCallExecutionBlocked,
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: productdata.EventRunCompleted, Summary: "Run completed"}); err != nil {
			t.Fatal(err)
		}
		srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, nil)
		approve := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs/"+run.ID+"/tool-calls/tc_terminal_apply/approve", "")
		assertStatus(t, approve.Code, http.StatusBadRequest, approve.Body.String())
		call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_terminal_apply")
		if err != nil {
			t.Fatal(err)
		}
		if call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
			t.Fatalf("terminal mutation call changed = %+v", call)
		}
	})
}

func TestM76ToolContinuationReliabilitySmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use six-step code-agent continuation without duplicating tool results.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWorkspaceGrep, productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspacePatchPreview, productdata.ToolNameWorkspacePatchApply, productdata.ToolNameSandboxExecCommand},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m76ContinuationReliabilityProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})
	worker.WorkerID = "worker_m76_tool_continuation"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M76 continuation reliability","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Run six tool continuations then summarize","client_message_id":"m76-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("initial ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m76_grep_1", productdata.ToolNameWorkspaceGrep)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m76_read_2", productdata.ToolNameWorkspaceRead)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_m76_preview_3", productdata.ToolNameWorkspacePatchPreview)
	assertNoPendingM76Approvals(t, svc, threadID, runID, "tc_m76_preview_3")
	approveM75Tool(t, srv, threadID, runID, "tc_m76_preview_3")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("apply ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m76_preview_3", productdata.ToolNameWorkspacePatchPreview)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_m76_apply_4", productdata.ToolNameWorkspacePatchApply)
	assertNoPendingM76Approvals(t, svc, threadID, runID, "tc_m76_apply_4")
	approveM75Tool(t, srv, threadID, runID, "tc_m76_apply_4")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("exec ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m76_apply_4", productdata.ToolNameWorkspacePatchApply)
	assertM22ToolBlocked(t, svc, threadID, runID, "tc_m76_exec_5", productdata.ToolNameSandboxExecCommand)
	assertNoPendingM76Approvals(t, svc, threadID, runID, "tc_m76_exec_5")
	approveM75Tool(t, srv, threadID, runID, "tc_m76_exec_5")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("read-after ProcessOne ok=%v err=%v", ok, err)
	}
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m76_exec_5", productdata.ToolNameSandboxExecCommand)
	assertM22ToolSucceeded(t, svc, threadID, runID, "tc_m76_read_6", productdata.ToolNameWorkspaceRead)
	drainM22Worker(t, worker, svc, ident, runID)

	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	messages, err := svc.ListMessages(context.Background(), ident, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "M76 continuation stable." {
		t.Fatalf("messages = %+v", messages)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, runID, 0)
	if err != nil {
		t.Fatal(err)
	}
	expectedToolCallIDs := []string{"tc_m76_grep_1", "tc_m76_read_2", "tc_m76_preview_3", "tc_m76_apply_4", "tc_m76_exec_5", "tc_m76_read_6"}
	assertM76TerminalToolEventsOnce(t, events, expectedToolCallIDs)
	if got := provider.toolResultIDsByRequest(); len(got) != 7 ||
		!stringSlicesEqual(got[1], expectedToolCallIDs[:1]) ||
		!stringSlicesEqual(got[2], expectedToolCallIDs[:2]) ||
		!stringSlicesEqual(got[3], expectedToolCallIDs[:3]) ||
		!stringSlicesEqual(got[4], expectedToolCallIDs[:4]) ||
		!stringSlicesEqual(got[5], expectedToolCallIDs[:5]) ||
		!stringSlicesEqual(got[6], expectedToolCallIDs) {
		t.Fatalf("continuation tool result order = %#v", got)
	}
	finalCount := 0
	for _, event := range events {
		if event.Type == productdata.EventRunCompleted {
			finalCount++
		}
	}
	if finalCount != 1 {
		t.Fatalf("run_completed count = %d events=%+v", finalCount, events)
	}
	lateText := "late"
	if _, err := svc.AppendRunEvent(context.Background(), ident, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryMessage, Type: "model_output_delta", Summary: "Late model output", Content: &lateText}); err == nil {
		t.Fatal("terminal run accepted late model output")
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, runID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_m76_late", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt"}, ArgumentsHash: "hash_late", ApprovalStatus: productdata.ToolCallApprovalApproved, ExecutionStatus: productdata.ToolCallExecutionNotStarted}); err == nil {
		t.Fatal("terminal run accepted late tool result continuation")
	}
}

func assertM22ToolBlocked(t *testing.T, svc productdata.Service, threadID string, runID string, toolCallID string, toolName string) {
	t.Helper()
	call, err := svc.GetToolCall(context.Background(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != toolName || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("call = %+v", call)
	}
}

func assertNoPendingM76Approvals(t *testing.T, svc productdata.Service, threadID string, runID string, pendingToolCallID string) {
	t.Helper()
	events, err := svc.ListRunEvents(context.Background(), identity.LocalDevIdentity(), runID, 0)
	if err != nil {
		t.Fatal(err)
	}
	pending := map[string]bool{}
	for _, event := range events {
		toolCallID := m76MetadataString(event.Metadata, "tool_call_id")
		if toolCallID == "" {
			continue
		}
		switch event.Type {
		case productdata.EventToolCallApprovalRequired:
			pending[toolCallID] = true
		case productdata.EventToolCallApproved, productdata.EventToolCallDenied, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, productdata.EventToolCallFailed, productdata.EventToolCallCancelled:
			delete(pending, toolCallID)
		}
	}
	if len(pending) != 1 || !pending[pendingToolCallID] {
		t.Fatalf("pending approvals = %+v, want only %s", pending, pendingToolCallID)
	}
	for toolCallID := range pending {
		call, err := svc.GetToolCall(context.Background(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
		if err != nil {
			t.Fatal(err)
		}
		if call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
			t.Fatalf("pending call = %+v", call)
		}
	}
}

func assertM76TerminalToolEventsOnce(t *testing.T, events []productdata.RunEvent, toolCallIDs []string) {
	t.Helper()
	terminalCounts := map[string]int{}
	for _, event := range events {
		switch event.Type {
		case productdata.EventToolCallSucceeded, productdata.EventToolCallFailed, productdata.EventToolCallDenied, productdata.EventToolCallCancelled:
			terminalCounts[m76MetadataString(event.Metadata, "tool_call_id")]++
		}
	}
	for _, toolCallID := range toolCallIDs {
		if terminalCounts[toolCallID] != 1 {
			t.Fatalf("terminal count for %s = %d counts=%+v", toolCallID, terminalCounts[toolCallID], terminalCounts)
		}
	}
}

func stringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func m76MetadataString(metadata map[string]any, key string) string {
	value, ok := metadata[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func approveM75Tool(t *testing.T, srv http.Handler, threadID string, runID string, toolCallID string) {
	t.Helper()
	approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
	assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
}

func assertM22ToolSucceeded(t *testing.T, svc productdata.Service, threadID string, runID string, toolCallID string, toolName string) {
	t.Helper()
	call, err := svc.GetToolCall(context.Background(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != toolName || call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("call = %+v", call)
	}
}

func drainM22Worker(t *testing.T, worker *productruntime.Worker, svc productdata.Service, ident identity.LocalIdentity, runID string) {
	t.Helper()
	for i := 0; i < 6; i++ {
		run, err := svc.GetRun(context.Background(), ident, runID)
		if err != nil {
			t.Fatal(err)
		}
		if productdata.IsRunTerminal(run.Status) || run.Status == productdata.RunStatusBlockedOnToolApproval {
			return
		}
		ok, err := worker.ProcessOne(context.Background())
		if err != nil {
			t.Fatalf("drain ProcessOne ok=%v err=%v", ok, err)
		}
		if !ok {
			return
		}
	}
}

type m22WorkspaceLoopProvider struct {
	calls int
}

func (p *m22WorkspaceLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m22WorkspaceLoopProvider) Stream(_ context.Context, _ productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceGlob, Metadata: map[string]any{"tool_call_id": "tc_glob_1", "arguments_summary": map[string]any{"pattern": "src/*.txt", "limit": 10}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_2", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 6}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M22 workspace "}, {Type: productruntime.ProviderEventCompleted, Text: "M22 workspace loop complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

type m22CodeAgentLoopProvider struct {
	calls int
}

type m75CodeAgentDailyLoopProvider struct {
	calls int
}

type m91DirectoryClassificationProvider struct {
	calls int
}

type m76ContinuationReliabilityProvider struct {
	calls    int
	requests []productruntime.ProviderRequest
}

func (p *m75CodeAgentDailyLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m75CodeAgentDailyLoopProvider) Stream(_ context.Context, _ productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceGrep, Metadata: map[string]any{"tool_call_id": "tc_grep_1", "arguments_summary": map[string]any{"query": "needle", "path": "src", "include": "*.txt", "limit": 10}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_2", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 64}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspacePatchPreview, Metadata: map[string]any{"tool_call_id": "tc_preview_3", "arguments_summary": map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "daily loop\n"}}}}
	case 4:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspacePatchApply, Metadata: map[string]any{"tool_call_id": "tc_apply_4", "arguments_summary": map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "daily loop\n"}}}}
	case 5:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameSandboxExecCommand, Metadata: map[string]any{"tool_call_id": "tc_exec_5", "arguments_summary": map[string]any{"argv": []any{"cat", "src/notes.txt"}, "cwd": ".", "timeout_ms": 1000, "max_output_bytes": 4096}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "Daily loop "}, {Type: productruntime.ProviderEventCompleted, Text: "Daily loop complete: patch applied and tests passed."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func (p *m91DirectoryClassificationProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m91DirectoryClassificationProvider) Stream(_ context.Context, _ productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceListDirectory, Metadata: map[string]any{"tool_call_id": "tc_m91_list_1", "arguments_summary": map[string]any{"path": ".", "max_entries": 200, "depth": 2, "sort": "name"}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceTreeSummary, Metadata: map[string]any{"tool_call_id": "tc_m91_summary_2", "arguments_summary": map[string]any{"path": ".", "max_entries": 200, "depth": 3}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_m91_read_3", "arguments_summary": map[string]any{"path": "docs/spec.md", "limit": 64}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "目录分类："}, {Type: productruntime.ProviderEventCompleted, Text: "源码：src/main.go；文档：docs/spec.md；配置：config/app.json；构建产物：已跳过 dist；临时文件：tmp/cache.tmp。"}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func (p *m76ContinuationReliabilityProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m76ContinuationReliabilityProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	p.requests = append(p.requests, request)
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceGrep, Metadata: map[string]any{"tool_call_id": "tc_m76_grep_1", "arguments_summary": map[string]any{"query": "needle", "path": "src", "include": "*.txt", "limit": 10}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_m76_read_2", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 64}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspacePatchPreview, Metadata: map[string]any{"tool_call_id": "tc_m76_preview_3", "arguments_summary": map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "m76 loop\n"}}}}
	case 4:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspacePatchApply, Metadata: map[string]any{"tool_call_id": "tc_m76_apply_4", "arguments_summary": map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "m76 loop\n"}}}}
	case 5:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameSandboxExecCommand, Metadata: map[string]any{"tool_call_id": "tc_m76_exec_5", "arguments_summary": map[string]any{"argv": []any{"cat", "src/notes.txt"}, "cwd": ".", "timeout_ms": 1000, "max_output_bytes": 4096}}}}
	case 6:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_m76_read_6", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 64}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M76 "}, {Type: productruntime.ProviderEventCompleted, Text: "M76 continuation stable."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func (p *m76ContinuationReliabilityProvider) toolResultIDsByRequest() [][]string {
	result := make([][]string, 0, len(p.requests))
	for _, request := range p.requests {
		ids := []string{}
		for _, message := range request.Messages {
			if message.Role == productruntime.ProviderMessageRoleToolResult {
				ids = append(ids, message.ToolCallID)
			}
		}
		result = append(result, ids)
	}
	return result
}

func runM75StalePatchApply(t *testing.T, root string) (map[string]any, string) {
	t.Helper()
	const runID = "run_m75_stale"
	executor := productruntime.WorkspaceToolExecutor{Root: root}
	if _, err := executor.Execute(context.Background(), productruntime.ToolInvocation{RunID: runID, ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := executor.Execute(context.Background(), productruntime.ToolInvocation{RunID: runID, ToolName: productdata.ToolNameWorkspacePatchPreview, ArgumentsSummary: map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "daily loop\n"}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "notes.txt"), []byte("external change\nsecond\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := executor.Execute(context.Background(), productruntime.ToolInvocation{RunID: runID, ToolName: productdata.ToolNameWorkspacePatchApply, ArgumentsSummary: map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "daily loop\n"}})
	events := productdata.EventToolCallFailed
	if err != nil {
		result = map[string]any{"error_code": "tool_execution_failed", "error_message": err.Error()}
	}
	return result, events
}

func (p *m22CodeAgentLoopProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m22CodeAgentLoopProvider) Stream(_ context.Context, _ productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{}
	switch p.calls {
	case 1:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_before_1", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 32}}}}
	case 2:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceEdit, Metadata: map[string]any{"tool_call_id": "tc_edit_2", "arguments_summary": map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "thread\n"}}}}
	case 3:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameSandboxExecCommand, Metadata: map[string]any{"tool_call_id": "tc_exec_3", "arguments_summary": map[string]any{"argv": []any{"cat", "src/notes.txt"}, "timeout_ms": 1000, "max_output_bytes": 4096}}}}
	case 4:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameWorkspaceRead, Metadata: map[string]any{"tool_call_id": "tc_read_after_4", "arguments_summary": map[string]any{"path": "src/notes.txt", "limit": 64}}}}
	default:
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "Code-agent "}, {Type: productruntime.ProviderEventCompleted, Text: "Code-agent loop complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}
