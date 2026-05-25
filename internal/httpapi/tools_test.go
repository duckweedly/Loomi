package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestToolsCatalogHandlerReturnsSafeCatalog(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	run := createToolsAPIRun(t, svc, ident)
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{
		Category: productdata.RunEventCategoryProgress,
		Type:     "mcp_discovery_succeeded",
		Summary:  "MCP discovery succeeded",
		Metadata: map[string]any{
			"server_slug":             "local-smoke",
			"status":                  "succeeded",
			"candidate_names":         []string{"mcp.local-smoke.echo"},
			"candidate_schema_hashes": map[string]any{"mcp.local-smoke.echo": "sha256:test-schema"},
			"raw_result":              "SECRET_CANARY_TOOL_RESULT",
			"env":                     "LOOMI_TOKEN=SECRET_CANARY_TOKEN",
		},
	}); err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/tools/catalog", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, expected := range []string{"runtime.get_current_time", `"source":"builtin"`, `"group":"runtime"`, "mcp.local-smoke.echo", `"source":"mcp"`, `"approval_policy":"always_required"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body missing %q: %s", expected, body)
		}
	}
	for _, forbidden := range []string{"SECRET_CANARY", "raw_result", "LOOMI_TOKEN", "env"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("body leaked %q: %s", forbidden, body)
		}
	}
}

func TestM18BuiltinToolApprovalRunsThroughBrokerSmoke(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved tools.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameCurrentTime},
		ReasoningMode:    "balanced",
		BudgetSummary:    "test",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "M18 builtin smoke", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "time"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := &m18BuiltinSmokeProvider{}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway})

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	approve := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs/"+run.ID+"/tool-calls/tc_builtin_m18/approve", "")
	if approve.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", approve.Code, approve.Body.String())
	}
	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("second ProcessOne ok=%v err=%v", ok, err)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	renderedEvents, _ := json.Marshal(events)
	for _, expected := range []string{productdata.EventToolCallRequested, productdata.EventToolCallApprovalRequired, productdata.EventToolCallApproved, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, productdata.EventRunCompleted} {
		if !strings.Contains(string(renderedEvents), expected) {
			t.Fatalf("events missing %s: %s", expected, string(renderedEvents))
		}
	}
	if strings.Contains(string(renderedEvents), "SECRET_CANARY") || provider.calls != 2 {
		t.Fatalf("events/provider leak or wrong calls events=%s calls=%d", string(renderedEvents), provider.calls)
	}
}

type m18BuiltinSmokeProvider struct {
	calls int
}

func (p *m18BuiltinSmokeProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m18BuiltinSmokeProvider) Stream(_ context.Context, _ productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: productdata.ToolNameCurrentTime, Metadata: map[string]any{"tool_call_id": "tc_builtin_m18", "arguments_summary": map[string]any{"timezone": "UTC"}, "provider_trace": "SECRET_CANARY_BUILTIN"}}}
	if p.calls == 2 {
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "Time ready "}, {Type: productruntime.ProviderEventCompleted, Text: "Time ready."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func createToolsAPIRun(t *testing.T, svc *productdata.MemoryService, ident identity.LocalIdentity) productdata.Run {
	t.Helper()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Tools API", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "tools"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	return run
}
