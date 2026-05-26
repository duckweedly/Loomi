package httpapi

import (
	"context"
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

func TestM23WorkspaceMutationToolsSmoke(t *testing.T) {
	root := createM21WorkspaceFixture(t)
	t.Setenv("LOOMI_WORKSPACE_ROOT", root)

	writeResult, writeEvents := runM23WorkspaceMutationTool(t, root, productdata.ToolNameWorkspaceWriteFile, "tc_m23_write", map[string]any{"path": "src/generated.txt", "content": "created\n"})
	written, err := os.ReadFile(filepath.Join(root, "src", "generated.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != "created\n" || writeResult["operation"] != "write_file" || writeResult["path"] != "src/generated.txt" {
		t.Fatalf("write result=%+v written=%q", writeResult, string(written))
	}
	assertBodyExcludes(t, writeEvents, "m23 write events", root, "fixture-secret", ".env", "secrets/token")
	assertBodyExcludes(t, writeEvents, "m23 write event arguments", `"content":"created`, `"content":"created\n"`)

	editResult, editEvents := runM23WorkspaceMutationTool(t, root, productdata.ToolNameWorkspaceEdit, "tc_m23_edit", map[string]any{"path": "src/notes.txt", "old_text": "second\n", "new_text": "changed\n"})
	edited, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(edited) != "needle\nchanged\n" || editResult["operation"] != "edit" || editResult["path"] != "src/notes.txt" {
		t.Fatalf("edit result=%+v edited=%q", editResult, string(edited))
	}
	assertBodyExcludes(t, editEvents, "m23 edit events", root, "fixture-secret", ".env", "secrets/token")
	assertBodyExcludes(t, editEvents, "m23 edit event arguments", "second")
}

func runM23WorkspaceMutationTool(t *testing.T, root string, toolName string, toolCallID string, args map[string]any) (map[string]any, string) {
	t.Helper()
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved workspace mutation tools only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWorkspaceWriteFile, productdata.ToolNameWorkspaceEdit},
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
	worker.WorkerID = "worker_m23_workspace_mutation"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M23 workspace mutation smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Run workspace mutation","client_message_id":"m23-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("pre-approval call = %+v", call)
	}
	if _, err := os.Stat(filepath.Join(root, strings.TrimSpace(stringValue(args, "path")))); toolName == productdata.ToolNameWorkspaceWriteFile && err == nil {
		t.Fatal("write_file created target before approval")
	}

	approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
	assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("second ProcessOne ok=%v err=%v", ok, err)
	}
	finalCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", run)
	}
	return finalCall.ResultSummary, fetchM21Events(t, srv, runID)
}
