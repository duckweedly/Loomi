package runtime

import (
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestMCPToolCandidatesUseNamespacedReadOnlyToolSpecs(t *testing.T) {
	result, err := ParseMCPListToolsResponse("local-search", []byte(`{"result":{"tools":[{"name":"search","description":"Search local index","inputSchema":{"type":"object"}}]}}`))
	if err != nil {
		t.Fatal(err)
	}

	specs := MCPToolSpecs(result.Candidates)
	if len(specs) != 1 {
		t.Fatalf("specs = %+v", specs)
	}
	spec := specs[0]
	if spec.Name != "mcp.local-search.search" || spec.Source != ToolSourceMCP || spec.ExecutionState != ToolExecutionDisabled {
		t.Fatalf("spec = %+v", spec)
	}
}

func TestMCPToolCandidateDoesNotOverrideInternalTool(t *testing.T) {
	if _, err := MapMCPToolCandidate("runtime", "runtime.get_current_time", "spoof", map[string]any{"type": "object"}); err == nil {
		t.Fatal("MapMCPToolCandidate() error = nil")
	}
}

func TestToolResolutionsForPersonaIncludesMCPAsNonExecutable(t *testing.T) {
	resolutions := ToolResolutionsForPersona([]string{productdata.ToolNameCurrentTime, "mcp.local-search.search"})
	if len(resolutions) != 2 {
		t.Fatalf("resolutions = %+v", resolutions)
	}
	if resolutions[1].Name != "mcp.local-search.search" || resolutions[1].ExecutionState != "discovered_non_executable" {
		t.Fatalf("mcp resolution = %+v", resolutions[1])
	}
}
