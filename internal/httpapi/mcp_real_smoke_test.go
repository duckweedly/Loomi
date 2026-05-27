package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM12RealLocalMCPApprovalSmoke(t *testing.T) {
	if os.Getenv("LOOMI_M12_MCP_FIXTURE") == "1" {
		runM12MCPFixture(t)
		return
	}

	countFile := t.TempDir() + "/tools-call-count"
	rawConfig, err := json.Marshal([]productruntime.MCPServerConfig{{
		Slug:        "local-smoke",
		DisplayName: "Local Smoke",
		Enabled:     true,
		Transport:   productruntime.MCPTransportStdio,
		Command:     os.Args[0],
		Args:        []string{"-test.run=^TestM12RealLocalMCPApprovalSmoke$"},
		Env: map[string]string{
			"LOOMI_M12_MCP_FIXTURE":    "1",
			"LOOMI_M12_MCP_COUNT_FILE": countFile,
			"LOOMI_M12_MCP_SECRET":     "sk-secret-fixture",
		},
		TimeoutMS: 1000,
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOOMI_MCP_SERVERS_JSON", string(rawConfig))

	configs, err := productruntime.MCPServerConfigsFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	mcpConfig := configs["local-smoke"]
	discovery, err := productruntime.DiscoverMCPTools(nil, mcpConfig)
	if err != nil {
		t.Fatal(err)
	}
	if discovery.Status != productruntime.MCPDiscoverySucceeded || len(discovery.Candidates) != 1 {
		t.Fatalf("discovery = %+v", discovery)
	}
	candidate := discovery.Candidates[0]
	discoveryMetadata := productruntime.MCPDiscoveryEventMetadata(discovery)
	hashes, ok := discoveryMetadata["candidate_schema_hashes"].(map[string]any)
	if !ok || hashes[candidate.Name] == "" {
		t.Fatalf("discovery metadata missing schema hashes: %+v", discoveryMetadata)
	}

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	syncM12SmokePersona(t, svc, candidate.Name)
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "M12.5 MCP smoke", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Run local MCP smoke"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: "mcp_discovery_succeeded", Summary: "MCP discovery succeeded", Metadata: discoveryMetadata}); err != nil {
		t.Fatal(err)
	}

	provider := &m12SmokeProvider{toolName: candidate.Name}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway, MCPExecutor: productruntime.StdioMCPToolExecutor{Configs: configs}})
	worker.WorkerID = "worker_m12_real_smoke"

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	if _, err := os.Stat(countFile); err == nil {
		t.Fatal("tools/call executed before approval")
	}
	call, err := svc.GetToolCall(context.Background(), ident, thread.ID, run.ID, "tc_mcp_real")
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != candidate.Name || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked {
		t.Fatalf("approval call = %+v", call)
	}
	if call.CandidateSchemaHash == "" {
		t.Fatalf("candidate schema hash missing: %+v", call)
	}

	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)
	approve := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+thread.ID+"/runs/"+run.ID+"/tool-calls/tc_mcp_real/approve", "")
	if approve.Code != http.StatusOK || !strings.Contains(approve.Body.String(), `"approval_status":"approved"`) {
		t.Fatalf("approve status=%d body=%s", approve.Code, approve.Body.String())
	}

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("second ProcessOne ok=%v err=%v", ok, err)
	}
	countRaw, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(countRaw)) != "1" {
		t.Fatalf("tools/call count = %q", string(countRaw))
	}

	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertM12SmokeEventOrder(t, events)
	renderedEvents, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	renderedContinuation, err := json.Marshal(provider.continuationRequest.Messages)
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{"sk-secret-fixture", "/Users/xuean", ".ssh", "id_ed25519"} {
		if strings.Contains(string(renderedEvents), leaked) || strings.Contains(string(renderedContinuation), leaked) {
			t.Fatalf("smoke leaked %q\nevents=%s\ncontinuation=%s", leaked, string(renderedEvents), string(renderedContinuation))
		}
	}
	if provider.calls != 2 || len(provider.continuationRequest.Messages) != 3 {
		t.Fatalf("provider calls=%d continuation=%+v", provider.calls, provider.continuationRequest.Messages)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || !strings.Contains(messages[1].Content, "MCP smoke complete") {
		t.Fatalf("messages = %+v", messages)
	}
}

type m12SmokeProvider struct {
	toolName            string
	calls               int
	continuationRequest productruntime.ProviderRequest
}

func (p *m12SmokeProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m12SmokeProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{{Type: productruntime.ProviderEventToolCall, ToolName: p.toolName, Metadata: map[string]any{"tool_call_id": "tc_mcp_real", "arguments_summary": map[string]any{"text": "status", "token": "sk-secret-fixture"}}}}
	if p.calls == 2 {
		p.continuationRequest = request
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "MCP smoke "}, {Type: productruntime.ProviderEventCompleted, Text: "MCP smoke complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func syncM12SmokePersona(t *testing.T, svc *productdata.MemoryService, toolName string) {
	t.Helper()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), identity.LocalDevIdentity(), []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use local MCP only after approval.",
		ModelRoute:       productdata.PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{productdata.ToolNameCurrentTime, toolName},
		ReasoningMode:    "balanced",
		BudgetSummary:    "small",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
}

func assertM12SmokeEventOrder(t *testing.T, events []productdata.RunEvent) {
	t.Helper()
	positions := map[string]int{}
	for index, event := range events {
		switch event.Type {
		case "mcp_discovery_succeeded":
			positions[event.Type] = index
		case productdata.EventToolCallApprovalRequired:
			positions[event.Type] = index
			if event.Metadata["tool_source"] != "mcp" || event.Metadata["server_slug"] != "local-smoke" {
				t.Fatalf("approval metadata = %+v", event.Metadata)
			}
		case productdata.EventToolCallApproved, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, productdata.EventRunCompleted:
			positions[event.Type] = index
		case "model_output_delta":
			if event.Metadata["model_phase"] == "continuation" {
				positions["continuation_delta"] = index
			}
		}
	}
	ordered := []string{"mcp_discovery_succeeded", productdata.EventToolCallApprovalRequired, productdata.EventToolCallApproved, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, "continuation_delta", productdata.EventRunCompleted}
	for _, key := range ordered {
		if _, ok := positions[key]; !ok {
			t.Fatalf("missing %s in events: %+v", key, events)
		}
	}
	for index := 1; index < len(ordered); index++ {
		if positions[ordered[index-1]] >= positions[ordered[index]] {
			t.Fatalf("event order %s=%d before %s=%d; events=%+v", ordered[index-1], positions[ordered[index-1]], ordered[index], positions[ordered[index]], events)
		}
	}
}

func runM12MCPFixture(t *testing.T) {
	t.Helper()
	reader := bufio.NewReader(os.Stdin)
	for {
		frame, ok := readM12MCPFrame(t, reader)
		if !ok {
			return
		}
		method, _ := frame["method"].(string)
		switch method {
		case "initialize":
			_, _ = os.Stdout.Write(m12MCPFrame(`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{},"serverInfo":{"name":"loomi-m12-smoke","version":"test"}}}`))
		case "tools/list":
			_, _ = os.Stdout.Write(m12MCPFrame(`{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"echo","description":"Echo smoke fixture","inputSchema":{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}}]}}`))
			return
		case "tools/call":
			countFile := os.Getenv("LOOMI_M12_MCP_COUNT_FILE")
			if countFile == "" {
				t.Fatal("missing count file")
			}
			raw, _ := os.ReadFile(countFile)
			count, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
			count++
			if err := os.WriteFile(countFile, []byte(strconv.Itoa(count)), 0o600); err != nil {
				t.Fatal(err)
			}
			_, _ = os.Stderr.WriteString("fixture stderr sk-secret-fixture /Users/xuean/.ssh/id_ed25519\n")
			_, _ = os.Stdout.Write(m12MCPFrame(`{"jsonrpc":"2.0","id":2,"result":{"summary":"ok","token":"sk-secret-fixture","path":"/Users/xuean/.ssh/id_ed25519"}}`))
			return
		default:
			if frame["id"] != nil {
				t.Fatalf("unexpected MCP method %q", method)
			}
		}
	}
}

func readM12MCPFrame(t *testing.T, reader *bufio.Reader) (map[string]any, bool) {
	t.Helper()
	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, false
			}
			t.Fatal(err)
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			break
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "Content-Length") {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				t.Fatal(err)
			}
			contentLength = parsed
		}
	}
	if contentLength <= 0 {
		t.Fatal("missing content length")
	}
	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		t.Fatal(err)
	}
	var frame map[string]any
	if err := json.Unmarshal(payload, &frame); err != nil {
		t.Fatal(err)
	}
	return frame, true
}

func m12MCPFrame(payload string) []byte {
	return []byte("Content-Length: " + strconv.Itoa(len(payload)) + "\r\n\r\n" + payload)
}
