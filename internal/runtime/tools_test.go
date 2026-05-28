package runtime

import (
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestCurrentTimeToolDefinitionValidatesTimezone(t *testing.T) {
	tool := CurrentTimeToolDefinition()
	if tool.Name != "runtime.get_current_time" {
		t.Fatalf("tool.Name = %q", tool.Name)
	}
	if tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("ApprovalPolicy = %q", tool.ApprovalPolicy)
	}
	if tool.SafetyClass != ToolSafetyNoSideEffectInternal {
		t.Fatalf("SafetyClass = %q", tool.SafetyClass)
	}
	if got, err := tool.NormalizeArguments(map[string]any{}); err != nil || got.Timezone != "UTC" {
		t.Fatalf("NormalizeArguments(empty) = %+v, %v", got, err)
	}
	if got, err := tool.NormalizeArguments(map[string]any{"timezone": "UTC"}); err != nil || got.Timezone != "UTC" {
		t.Fatalf("NormalizeArguments(UTC) = %+v, %v", got, err)
	}
	if _, err := tool.NormalizeArguments(map[string]any{"timezone": "Asia/Shanghai"}); err == nil {
		t.Fatal("NormalizeArguments(Asia/Shanghai) error = nil, want error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"shell": "pwd"}); err == nil {
		t.Fatal("NormalizeArguments(shell) error = nil, want error")
	}
}

func TestCurrentTimeToolExecutesSafeResult(t *testing.T) {
	tool := CurrentTimeToolDefinition()
	result, err := tool.Execute(ToolArguments{Timezone: "UTC"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result["timezone"] != "UTC" || result["source"] != "runtime" || result["iso_time"] == "" {
		t.Fatalf("result = %+v", result)
	}
}

func TestToolDefinitionsForPersonaIntersectAllowlist(t *testing.T) {
	tools := ToolResolutionsForPersona([]string{productdata.ToolNameLoadTools, productdata.ToolNameCurrentTime, productdata.ToolNameWorkspaceRead, productdata.ToolNameSandboxExecCommand, "runtime.shell"})
	if len(tools) != 4 || tools[0].Name != productdata.ToolNameLoadTools || tools[1].Name != productdata.ToolNameCurrentTime || tools[2].Name != productdata.ToolNameWorkspaceRead || tools[3].Name != productdata.ToolNameSandboxExecCommand {
		t.Fatalf("tools = %+v", tools)
	}
	if tools[0].Group != string(productdata.ToolCatalogGroupDiscovery) || tools[0].ApprovalPolicy != string(ToolApprovalNotRequired) {
		t.Fatalf("discovery resolution = %+v", tools[0])
	}
	if tools[2].Group != string(productdata.ToolCatalogGroupWorkspace) || tools[2].ExecutionState != string(productdata.ToolExecutionStateExecutable) {
		t.Fatalf("workspace resolution = %+v", tools[2])
	}
	if tools[3].Group != string(productdata.ToolCatalogGroupSandbox) || tools[3].RiskLevel != string(productdata.ToolRiskHigh) || tools[3].ExecutionState != string(productdata.ToolExecutionStateExecutable) {
		t.Fatalf("sandbox resolution = %+v", tools[3])
	}
}

func TestToolResolutionsMarkWebToolsAsNoApproval(t *testing.T) {
	tools := ToolResolutionsForPersona([]string{productdata.ToolNameWebSearch, productdata.ToolNameWebFetch})
	if len(tools) != 2 {
		t.Fatalf("tools = %+v", tools)
	}
	if tools[0].Name != productdata.ToolNameWebSearch || tools[0].ApprovalPolicy != string(ToolApprovalNotRequired) {
		t.Fatalf("web search resolution = %+v", tools[0])
	}
	if tools[1].Name != productdata.ToolNameWebFetch || tools[1].ApprovalPolicy != string(ToolApprovalNotRequired) {
		t.Fatalf("web fetch resolution = %+v", tools[1])
	}
}

func TestWorkspaceToolDefinitionsSeparateReadAndMutationRisk(t *testing.T) {
	defs := WorkspaceToolDefinitions()
	if len(defs) != 9 {
		t.Fatalf("defs = %+v", defs)
	}
	for _, def := range defs {
		if !productdata.IsWorkspaceToolName(def.Name) {
			t.Fatalf("workspace definition = %+v", def)
		}
		if productdata.IsWorkspaceReadOnlyToolName(def.Name) && def.ApprovalPolicy != ToolApprovalNotRequired {
			t.Fatalf("workspace read definition = %+v", def)
		}
		if !productdata.IsWorkspaceReadOnlyToolName(def.Name) && def.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("workspace mutation definition = %+v", def)
		}
		if def.Name == productdata.ToolNameWorkspaceWriteFile || def.Name == productdata.ToolNameWorkspaceEdit || def.Name == productdata.ToolNameWorkspacePatchApply {
			if def.SafetyClass != ToolSafetyWorkspaceMutation {
				t.Fatalf("workspace mutation definition = %+v", def)
			}
			continue
		}
		if def.SafetyClass != ToolSafetyNoSideEffectInternal {
			t.Fatalf("workspace read definition = %+v", def)
		}
	}
}

func TestSandboxToolDefinitionsAreHighRisk(t *testing.T) {
	defs := SandboxToolDefinitions()
	if len(defs) != 4 || defs[0].Name != productdata.ToolNameSandboxExecCommand || defs[1].Name != productdata.ToolNameSandboxStartProcess || defs[2].Name != productdata.ToolNameSandboxContinueProcess || defs[3].Name != productdata.ToolNameSandboxTerminateProcess {
		t.Fatalf("defs = %+v", defs)
	}
	for _, def := range defs {
		if def.ApprovalPolicy != ToolApprovalAlwaysRequired || def.SafetyClass != ToolSafetySandboxCommand || def.ExecutionState != ToolExecutionAllowlisted {
			t.Fatalf("sandbox definition = %+v", def)
		}
	}
}

func TestMemoryToolDefinitionsAreApprovalGated(t *testing.T) {
	defs := MemoryToolDefinitions()
	if len(defs) != 16 || defs[0].Name != productdata.ToolNameMemorySearch || defs[len(defs)-1].Name != productdata.ToolNameNotebookForget {
		t.Fatalf("defs = %+v", defs)
	}
	for _, def := range defs {
		if def.ApprovalPolicy != ToolApprovalAlwaysRequired || def.ExecutionState != ToolExecutionAllowlisted {
			t.Fatalf("memory definition = %+v", def)
		}
		if def.Name == productdata.ToolNameMemoryWrite || def.Name == productdata.ToolNameMemoryEdit || def.Name == productdata.ToolNameMemoryForget || def.Name == productdata.ToolNameNotebookWrite || def.Name == productdata.ToolNameNotebookEdit || def.Name == productdata.ToolNameNotebookForget {
			if def.SafetyClass != ToolSafetyWorkspaceMutation {
				t.Fatalf("memory mutation definition = %+v", def)
			}
			continue
		}
		if def.SafetyClass != ToolSafetyNoSideEffectInternal {
			t.Fatalf("memory read definition = %+v", def)
		}
	}
}
