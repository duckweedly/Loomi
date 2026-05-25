package runtime

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestParseMCPListToolsResponseMapsTools(t *testing.T) {
	payload := []byte(`{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"search","description":"Search local index","inputSchema":{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}}]}}`)

	result, err := ParseMCPListToolsResponse("local-search", payload)
	if err != nil {
		t.Fatalf("ParseMCPListToolsResponse() error = %v", err)
	}
	if result.Status != MCPDiscoverySucceeded || len(result.Candidates) != 1 {
		t.Fatalf("result = %+v", result)
	}
	candidate := result.Candidates[0]
	if candidate.Name != "mcp.local-search.search" || candidate.MCPName != "search" || candidate.ExecutionEnabled {
		t.Fatalf("candidate = %+v", candidate)
	}
	if candidate.SchemaHash == "" {
		t.Fatalf("schema hash empty")
	}
	metadata := MCPDiscoveryEventMetadata(result)
	hashes, ok := metadata["candidate_schema_hashes"].(map[string]any)
	if !ok || hashes[candidate.Name] != candidate.SchemaHash {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestParseMCPListToolsResponseRejectsInvalidAndDuplicateSchemas(t *testing.T) {
	for _, payload := range [][]byte{
		[]byte(`{"result":{"tools":[{"description":"missing name","inputSchema":{"type":"object"}}]}}`),
		[]byte(`{"result":{"tools":[{"name":"search","inputSchema":{"type":"object"}},{"name":"search","inputSchema":{"type":"object"}}]}}`),
		[]byte(`{"result":{"tools":[{"name":"search","inputSchema":{"type":"array"}}]}}`),
		[]byte(`{"result":{"tools":[{"name":"search","inputSchema":{"type":"object","properties":[]}}]}}`),
		[]byte(`not-json`),
	} {
		if _, err := ParseMCPListToolsResponse("local-search", payload); err == nil {
			t.Fatalf("ParseMCPListToolsResponse(%s) error = nil", string(payload))
		}
	}
}

func TestParseMCPListToolsResponseRejectsOversizedSchemasAndToolLists(t *testing.T) {
	oversized := `{"result":{"tools":[{"name":"search","inputSchema":{"type":"object","description":"` + strings.Repeat("x", MaxMCPSchemaBytes) + `"}}]}}`
	if _, err := ParseMCPListToolsResponse("local-search", []byte(oversized)); err == nil {
		t.Fatal("oversized schema error = nil")
	}

	tools := make([]string, 0, MaxMCPDiscoveryTools+1)
	for i := 0; i < MaxMCPDiscoveryTools+1; i++ {
		tools = append(tools, `{"name":"tool`+strconv.Itoa(i)+`","inputSchema":{"type":"object"}}`)
	}
	payload := []byte(`{"result":{"tools":[` + strings.Join(tools, ",") + `]}}`)
	if _, err := ParseMCPListToolsResponse("local-search", payload); err == nil {
		t.Fatal("oversized tool list error = nil")
	}
}

func TestDiscoverMCPToolsSkipsDisabledConfig(t *testing.T) {
	result, err := DiscoverMCPTools(nil, MCPServerConfig{Slug: "local-search", DisplayName: "Local Search", Enabled: false, Transport: MCPTransportStdio, Command: "unused", TimeoutMS: 5000})
	if err != nil {
		t.Fatalf("DiscoverMCPTools() error = %v", err)
	}
	if result.Status != MCPDiscoveryDisabled || len(result.Candidates) != 0 {
		t.Fatalf("result = %+v", result)
	}
}

func TestDiscoverMCPToolsRunsLocalStdioListToolsSmokeWithoutLeaks(t *testing.T) {
	if os.Getenv("LOOMI_MCP_STDIO_FIXTURE") == "1" {
		runMCPStdioSmokeFixture(t)
		return
	}

	result, err := DiscoverMCPTools(nil, MCPServerConfig{
		Slug:        "local-smoke",
		DisplayName: "Local Smoke",
		Enabled:     true,
		Transport:   MCPTransportStdio,
		Command:     os.Args[0],
		Args:        []string{"-test.run", "^TestDiscoverMCPToolsRunsLocalStdioListToolsSmokeWithoutLeaks$", "--", "--token=secret-arg", "/Users/xuean/.ssh/id_ed25519"},
		Env:         map[string]string{"LOOMI_MCP_STDIO_FIXTURE": "1", "LOOMI_MCP_FIXTURE_TOKEN": "fixture-secret-token"},
		TimeoutMS:   1000,
	})
	if err != nil {
		t.Fatalf("DiscoverMCPTools() error = %v", err)
	}
	if result.Status != MCPDiscoverySucceeded || len(result.Candidates) != 1 {
		t.Fatalf("result = %+v", result)
	}
	candidate := result.Candidates[0]
	if candidate.Name != "mcp.local-smoke.echo" || candidate.ExecutionEnabled {
		t.Fatalf("candidate = %+v", candidate)
	}

	summary := productdata.RunContext{
		EnabledTools: []productdata.ToolResolution{{Name: candidate.Name, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: "discovered_non_executable"}},
		MCPAvailability: productdata.MCPToolAvailabilitySummary{
			ServersConfigured:           1,
			ServersEnabled:              1,
			ServersSucceeded:            1,
			ServerSummaries:             []productdata.MCPServerAvailabilitySummary{{ServerSafeID: "mcp:local-smoke", ServerSlug: "local-smoke", Enabled: true, DiscoveryStatus: string(result.Status), CandidateCount: 1, CandidateNames: []string{candidate.Name}}},
			CandidateNames:              []string{candidate.Name},
			NonExecutableCandidateNames: []string{candidate.Name},
			ExecutionEnabled:            false,
		},
	}.SafeSummary()
	rawSummary, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	safe := string(rawSummary)
	for _, leaked := range []string{"secret-arg", "fixture-secret-token", "id_ed25519", "/Users/xuean", ".ssh", "stderr", "tools/call"} {
		if strings.Contains(safe, leaked) {
			t.Fatalf("safe summary leaked %q: %s", leaked, safe)
		}
	}
	for _, expected := range []string{"mcp.local-smoke.echo", `"mcp_execution_enabled":false`, `"mcp_servers_succeeded":1`} {
		if !strings.Contains(safe, expected) {
			t.Fatalf("safe summary missing %q: %s", expected, safe)
		}
	}
}

func runMCPStdioSmokeFixture(t *testing.T) {
	t.Helper()
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(input), "tools/call") {
		t.Fatal("fixture received tools/call")
	}
	if !strings.Contains(string(input), `"method":"tools/list"`) {
		t.Fatalf("fixture did not receive tools/list: %s", string(input))
	}
	_, _ = os.Stderr.WriteString("stderr token fixture-secret-token /Users/xuean/.ssh/id_ed25519\n")
	os.Stdout.Write(mcpFrame(`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{},"serverInfo":{"name":"loomi-local-smoke","version":"test"}}}`))
	os.Stdout.Write(mcpFrame(`{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"echo","description":"Echo fixture token fixture-secret-token","inputSchema":{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}}]}}`))
	os.Exit(0)
}
