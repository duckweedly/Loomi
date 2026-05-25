package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM15ChatRealIntegratedSmoke(t *testing.T) {
	if os.Getenv("LOOMI_M15_MCP_FIXTURE") == "1" {
		runM15MCPFixture(t)
		return
	}
	if os.Getenv("LOOMI_M15_REAL_CHAT_SMOKE") != "1" {
		t.Skip("set LOOMI_M15_REAL_CHAT_SMOKE=1 to run the M15 real chat integrated smoke")
	}

	const (
		toolCallID         = "tc_m15_real"
		providerSecret     = "sk-m15-provider-secret"
		mcpSecret          = "sk-m15-mcp-secret"
		secretPath         = "/Users/xuean/.ssh/id_ed25519"
		authorizationValue = "Authorization: Bearer m15-secret-token"
	)

	countFile := t.TempDir() + "/m15-tools-call-count"
	rawConfig, err := json.Marshal([]productruntime.MCPServerConfig{{
		Slug:        "m15-smoke",
		DisplayName: "M15 Smoke",
		Enabled:     true,
		Transport:   productruntime.MCPTransportStdio,
		Command:     os.Args[0],
		Args:        []string{"-test.run=^TestM15ChatRealIntegratedSmoke$"},
		Env: map[string]string{
			"LOOMI_M15_MCP_FIXTURE":    "1",
			"LOOMI_M15_MCP_COUNT_FILE": countFile,
			"LOOMI_M15_MCP_SECRET":     mcpSecret,
			"LOOMI_M15_MCP_PATH":       secretPath,
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
	discovery, err := productruntime.DiscoverMCPTools(nil, configs["m15-smoke"])
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
		t.Fatalf("discovery metadata missing candidate hash: %+v", discoveryMetadata)
	}

	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	syncM15SmokePersona(t, svc, candidate.Name)
	provider := &m15SmokeProvider{toolName: candidate.Name, toolCallID: toolCallID, providerSecret: providerSecret, authorizationValue: authorizationValue}
	gateway := productruntime.NewGateway(svc, nil, []productruntime.Provider{provider})
	worker := productruntime.NewWorker(svc, nil, productruntime.QueuedRunRouter{Gateway: gateway, MCPExecutor: productruntime.StdioMCPToolExecutor{Configs: configs}})
	worker.WorkerID = "worker_m15_real_smoke"
	srv := NewServerWithRuntimes(config.Config{AppEnv: "local"}, fakeChecker{}, svc, nil, nil, gateway)

	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"M15 chat real integrated smoke","mode":"chat"}`)
	assertStatus(t, threadRes.Code, http.StatusCreated, threadRes.Body.String())
	threadID := decodeStringField(t, threadRes.Body.Bytes(), "thread", "id")

	memoryRes := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals", `{"scope_type":"thread","scope_id":"`+threadID+`","title":"M15 smoke preference","content":"Use the approved safe M15 chat smoke memory snapshot","source_thread_id":"`+threadID+`","idempotency_key":"m15-memory-proposal"}`)
	assertStatus(t, memoryRes.Code, http.StatusCreated, memoryRes.Body.String())
	proposalID := decodeStringField(t, memoryRes.Body.Bytes(), "proposal", "id")
	memoryApprove := requestJSON(t, srv, http.MethodPost, "/v1/memory/write-proposals/"+proposalID+"/approve", `{"reason":"m15 smoke setup","idempotency_key":"m15-memory-approve"}`)
	assertStatus(t, memoryApprove.Code, http.StatusOK, memoryApprove.Body.String())
	memoryEntryID := decodeStringField(t, memoryApprove.Body.Bytes(), "entry", "id")

	messageRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/messages", `{"content":"Run the M15 integrated chat smoke","client_message_id":"m15-user-message"}`)
	assertStatus(t, messageRes.Code, http.StatusCreated, messageRes.Body.String())
	messageID := decodeStringField(t, messageRes.Body.Bytes(), "message", "id")
	runRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs", `{"message_id":"`+messageID+`","source":"model_gateway","provider_id":"custom","model":"model"}`)
	assertStatus(t, runRes.Code, http.StatusAccepted, runRes.Body.String())
	runID := decodeStringField(t, runRes.Body.Bytes(), "run", "id")
	if _, err := svc.AppendRunEvent(context.Background(), ident, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: "mcp_discovery_succeeded", Summary: "MCP discovery succeeded", Metadata: discoveryMetadata}); err != nil {
		t.Fatal(err)
	}

	if ok, err := worker.ProcessOne(context.Background()); err != nil || !ok {
		t.Fatalf("first ProcessOne ok=%v err=%v", ok, err)
	}
	if _, err := os.Stat(countFile); err == nil {
		t.Fatal("MCP tools/call executed before HTTP approval")
	}
	call, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if call.ToolName != candidate.Name || call.ApprovalStatus != productdata.ToolCallApprovalRequired || call.ExecutionStatus != productdata.ToolCallExecutionBlocked || call.CandidateSchemaHash == "" {
		t.Fatalf("approval call = %+v", call)
	}

	approvalRes := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/runs/"+runID+"/tool-calls/"+toolCallID+"/approve", "")
	assertStatus(t, approvalRes.Code, http.StatusOK, approvalRes.Body.String())
	assertBodyExcludes(t, approvalRes.Body.String(), "approval response", providerSecret, mcpSecret, secretPath, authorizationValue)

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

	finalCall, err := svc.GetToolCall(context.Background(), ident, threadID, runID, toolCallID)
	if err != nil {
		t.Fatal(err)
	}
	if finalCall.ExecutionStatus != productdata.ToolCallExecutionSucceeded {
		t.Fatalf("final call = %+v", finalCall)
	}
	messagesRes := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+threadID+"/messages", "")
	assertStatus(t, messagesRes.Code, http.StatusOK, messagesRes.Body.String())
	messages := decodeM15Messages(t, messagesRes.Body.Bytes())
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "M15 chat smoke complete." {
		t.Fatalf("messages = %+v", messages)
	}
	runGet := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+runID, "")
	assertStatus(t, runGet.Code, http.StatusOK, runGet.Body.String())
	if !strings.Contains(runGet.Body.String(), `"status":"completed"`) {
		t.Fatalf("run response = %s", runGet.Body.String())
	}
	eventsRes := requestJSON(t, srv, http.MethodGet, "/v1/runs/"+runID+"/events", "")
	assertStatus(t, eventsRes.Code, http.StatusOK, eventsRes.Body.String())
	events := decodeM15Events(t, eventsRes.Body.Bytes())
	assertM15SmokeEvidence(t, events, memoryEntryID, candidate.Name, toolCallID)

	renderedProviderRequest, err := json.Marshal(provider.continuationRequest.Messages)
	if err != nil {
		t.Fatal(err)
	}
	renderedRunContextSummary := m15PrepareContextSummary(events)
	renderedToolResult, err := json.Marshal(finalCall.ResultSummary)
	if err != nil {
		t.Fatal(err)
	}
	for label, surface := range map[string]string{
		"run create API response":       runRes.Body.String(),
		"run fetch API response":        runGet.Body.String(),
		"events API response":           eventsRes.Body.String(),
		"messages API response":         messagesRes.Body.String(),
		"RunContext safe summary":       renderedRunContextSummary,
		"tool result summary":           string(renderedToolResult),
		"provider continuation request": string(renderedProviderRequest),
		"M15 docs examples":             readM15DocsExamples(t),
	} {
		assertBodyExcludes(t, surface, label, providerSecret, mcpSecret, secretPath, authorizationValue, "id_ed25519", ".ssh")
	}
	if provider.calls != 2 || len(provider.continuationRequest.Messages) != 3 || provider.continuationRequest.Messages[2].Role != productruntime.ProviderMessageRoleToolResult {
		t.Fatalf("provider calls=%d continuation=%+v", provider.calls, provider.continuationRequest.Messages)
	}
	t.Logf("M15 smoke run_id=%s tool_call_id=%s final=%q events=%v", runID, toolCallID, messages[1].Content, m15EventTypes(events))
}

type m15SmokeProvider struct {
	toolName            string
	toolCallID          string
	providerSecret      string
	authorizationValue  string
	calls               int
	continuationRequest productruntime.ProviderRequest
}

func (p *m15SmokeProvider) Config() productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}
}

func (p *m15SmokeProvider) Stream(_ context.Context, request productruntime.ProviderRequest) (<-chan productruntime.ProviderEvent, error) {
	p.calls++
	events := []productruntime.ProviderEvent{{
		Type:     productruntime.ProviderEventToolCall,
		ToolName: p.toolName,
		Metadata: map[string]any{
			"tool_call_id": p.toolCallID,
			"arguments_summary": map[string]any{
				"query":         "m15 status",
				"token":         p.providerSecret,
				"authorization": p.authorizationValue,
			},
		},
	}}
	if p.calls == 2 {
		p.continuationRequest = request
		events = []productruntime.ProviderEvent{{Type: productruntime.ProviderEventTextDelta, Text: "M15 chat smoke "}, {Type: productruntime.ProviderEventCompleted, Text: "M15 chat smoke complete."}}
	}
	ch := make(chan productruntime.ProviderEvent, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func syncM15SmokePersona(t *testing.T, svc *productdata.MemoryService, toolName string) {
	t.Helper()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), identity.LocalDevIdentity(), []productdata.BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default",
		SystemPrompt:     "Use approved local MCP only for deterministic closeout smoke.",
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

func runM15MCPFixture(t *testing.T) {
	t.Helper()
	reader := bufio.NewReader(os.Stdin)
	for {
		frame, ok := readM15MCPFrame(t, reader)
		if !ok {
			return
		}
		method, _ := frame["method"].(string)
		switch method {
		case "initialize":
			_, _ = os.Stdout.Write(m15MCPFrame(`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{},"serverInfo":{"name":"loomi-m15-smoke","version":"test"}}}`))
		case "tools/list":
			_, _ = os.Stdout.Write(m15MCPFrame(`{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"echo","description":"Echo M15 smoke fixture","inputSchema":{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}}]}}`))
			return
		case "tools/call":
			countFile := os.Getenv("LOOMI_M15_MCP_COUNT_FILE")
			if countFile == "" {
				t.Fatal("missing M15 count file")
			}
			raw, _ := os.ReadFile(countFile)
			count, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
			count++
			if err := os.WriteFile(countFile, []byte(strconv.Itoa(count)), 0o600); err != nil {
				t.Fatal(err)
			}
			_, _ = os.Stderr.WriteString("m15 fixture stderr " + os.Getenv("LOOMI_M15_MCP_SECRET") + " " + os.Getenv("LOOMI_M15_MCP_PATH") + "\n")
			_, _ = os.Stdout.Write(m15MCPFrame(`{"jsonrpc":"2.0","id":2,"result":{"summary":"M15 safe MCP result","token":"` + os.Getenv("LOOMI_M15_MCP_SECRET") + `","path":"` + os.Getenv("LOOMI_M15_MCP_PATH") + `"}}`))
			return
		default:
			if frame["id"] != nil {
				t.Fatalf("unexpected M15 MCP method %q", method)
			}
		}
	}
}

func readM15MCPFrame(t *testing.T, reader *bufio.Reader) (map[string]any, bool) {
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

func m15MCPFrame(payload string) []byte {
	return []byte("Content-Length: " + strconv.Itoa(len(payload)) + "\r\n\r\n" + payload)
}

func decodeM15Messages(t *testing.T, raw []byte) []productdata.Message {
	t.Helper()
	var body struct {
		Messages []productdata.Message `json:"messages"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body.Messages
}

func decodeM15Events(t *testing.T, raw []byte) []productdata.RunEvent {
	t.Helper()
	var body struct {
		Events []productdata.RunEvent `json:"events"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body.Events
}

func assertM15SmokeEvidence(t *testing.T, events []productdata.RunEvent, memoryEntryID string, toolName string, toolCallID string) {
	t.Helper()
	positions := map[string]int{}
	for index, event := range events {
		switch event.Type {
		case productdata.EventRunQueued, productdata.EventJobClaimed, productdata.EventMemorySnapshotLoaded, productdata.EventToolCallRequested, productdata.EventToolCallApprovalRequired, productdata.EventToolCallApproved, productdata.EventToolCallExecuting, productdata.EventToolCallSucceeded, productdata.EventRunCompleted:
			positions[event.Type] = index
		case productdata.EventPipelineStepStarted, productdata.EventPipelineStepCompleted:
			step, _ := event.Metadata["step"].(string)
			positions[event.Type+":"+step] = index
		case "mcp_discovery_succeeded":
			positions[event.Type] = index
			if event.Metadata["server_slug"] != "m15-smoke" {
				t.Fatalf("discovery metadata = %+v", event.Metadata)
			}
			hashes, ok := event.Metadata["candidate_schema_hashes"].(map[string]any)
			if !ok || hashes[toolName] == "" {
				t.Fatalf("discovery hash missing: %+v", event.Metadata)
			}
		case "model_request_started":
			if event.Metadata["model_phase"] == "continuation" {
				positions["continuation_started"] = index
			}
		case "model_output_completed":
			if event.Metadata["model_phase"] == "continuation" {
				positions["continuation_completed"] = index
			}
		}
		if event.Type == productdata.EventMemorySnapshotLoaded && event.Metadata["entry_count"] != float64(1) && event.Metadata["entry_count"] != 1 {
			t.Fatalf("memory event metadata = %+v", event.Metadata)
		}
		if event.Type == productdata.EventPipelineStepCompleted && event.Metadata["step"] == string(productdata.PipelineStepPrepareContext) {
			if event.Metadata["memory_entry_count"] != float64(1) && event.Metadata["memory_entry_count"] != 1 {
				t.Fatalf("prepare context metadata = %+v", event.Metadata)
			}
		}
		if event.Type == productdata.EventToolCallApprovalRequired && (event.Metadata["tool_call_id"] != toolCallID || event.Metadata["tool_name"] != toolName || event.Metadata["tool_source"] != "mcp") {
			t.Fatalf("approval metadata = %+v", event.Metadata)
		}
	}
	for _, key := range []string{
		productdata.EventRunQueued,
		productdata.EventJobClaimed,
		productdata.EventPipelineStepStarted + ":" + string(productdata.PipelineStepPrepareContext),
		productdata.EventPipelineStepCompleted + ":" + string(productdata.PipelineStepPrepareContext),
		productdata.EventPipelineStepStarted + ":" + string(productdata.PipelineStepInvokeRuntime),
		productdata.EventMemorySnapshotLoaded,
		"mcp_discovery_succeeded",
		productdata.EventToolCallRequested,
		productdata.EventToolCallApprovalRequired,
		productdata.EventToolCallApproved,
		productdata.EventToolCallExecuting,
		productdata.EventToolCallSucceeded,
		"continuation_started",
		"continuation_completed",
		productdata.EventRunCompleted,
	} {
		if _, ok := positions[key]; !ok {
			t.Fatalf("missing %s in event types %v; memory_entry_id=%s", key, m15EventTypes(events), memoryEntryID)
		}
	}
	if positions[productdata.EventToolCallRequested] >= positions[productdata.EventToolCallApprovalRequired] ||
		positions[productdata.EventToolCallApprovalRequired] >= positions[productdata.EventToolCallApproved] ||
		positions[productdata.EventToolCallApproved] >= positions[productdata.EventToolCallExecuting] ||
		positions[productdata.EventToolCallExecuting] >= positions[productdata.EventToolCallSucceeded] ||
		positions[productdata.EventToolCallSucceeded] >= positions["continuation_started"] ||
		positions["continuation_completed"] >= positions[productdata.EventRunCompleted] {
		t.Fatalf("unexpected event order: %+v", positions)
	}
}

func m15PrepareContextSummary(events []productdata.RunEvent) string {
	for _, event := range events {
		if event.Type == productdata.EventPipelineStepCompleted && event.Metadata["step"] == string(productdata.PipelineStepPrepareContext) {
			raw, _ := json.Marshal(event.Metadata)
			return string(raw)
		}
	}
	return ""
}

func m15EventTypes(events []productdata.RunEvent) []string {
	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.Type)
	}
	return types
}

func readM15DocsExamples(t *testing.T) string {
	t.Helper()
	var rendered strings.Builder
	for _, path := range []string{
		filepath.Join("..", "..", "docs-site", "src", "content", "docs", "devlog", "2026-05-25-m15-chat-real-integrated-smoke-closeout.md"),
		filepath.Join("..", "..", "docs-site", "src", "content", "docs", "runbooks", "local-m15-chat-smoke.md"),
		filepath.Join("..", "..", "docs-site", "src", "content", "docs", "roadmap", "current-status.md"),
		filepath.Join("..", "..", "docs-site", "src", "content", "docs", "spec-kit", "workflow.md"),
	} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read M15 docs example %s: %v", path, err)
		}
		rendered.Write(raw)
	}
	return rendered.String()
}
