package runtime

import (
	"strconv"
	"strings"
	"testing"
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
