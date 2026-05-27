package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM26WebFetchAutoExecuteFinalSmoke(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><title>M26 Fixture</title><body>Approved web fetch body.</body></html>"))
	}))
	defer target.Close()
	const toolCallID = "tc_m26_web_fetch"

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use public web fetch only.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameWebFetch},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	provider := &m21WorkspaceProvider{toolName: productdata.ToolNameWebFetch, toolCallID: toolCallID, args: map[string]any{"url": target.URL, "max_bytes": 4096, "timeout_ms": 1000}}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway, WebExecutor: productruntime.WebToolExecutor{AllowPrivateHosts: true}})
	worker.WorkerID = "worker_m26_web_fetch"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M26 web fetch smoke","mode":"work"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")
	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Fetch docs","client_message_id":"m26-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("ProcessOne ok=%v err=%v", ok, err)
	}
	call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != productdata.ToolNameWebFetch || call.ApprovalStatus != productdata.ToolCallApprovalApproved || call.ExecutionStatus != productdata.ToolCallExecutionSucceeded || call.ResultSummary["operation"] != "fetch" || call.ResultSummary["scope"] != "web" || call.ResultSummary["title"] != "M26 Fixture" {
		t.Fatalf("call = %+v", call)
	}
	run, err := svc.GetRun(context.Background(), ident, runID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != productdata.RunStatusCompleted || provider.calls != 2 {
		t.Fatalf("run=%+v provider calls=%d", run, provider.calls)
	}
	eventsBody := fetchM21Events(t, srv, runID)
	for _, expected := range []string{productdata.EventToolCallApproved, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, `"tool_name":"web.fetch"`, `"scope":"web"`, `"status_code":200`} {
		if !strings.Contains(eventsBody, expected) {
			t.Fatalf("events missing %s: %s", expected, eventsBody)
		}
	}
	if strings.Contains(eventsBody, productdata.EventToolCallApprovalRequired) {
		t.Fatalf("web.fetch should not require manual approval: %s", eventsBody)
	}
	assertBodyExcludes(t, eventsBody, "m26 web fetch events", "Set-Cookie", "Authorization", "sk-secret", "/Users/")
}
