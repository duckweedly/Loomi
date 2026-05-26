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
	tools := ToolResolutionsForPersona([]string{productdata.ToolNameCurrentTime, productdata.ToolNameWorkspaceRead, productdata.ToolNameSandboxExecCommand, "runtime.shell"})
	if len(tools) != 3 || tools[0].Name != productdata.ToolNameCurrentTime || tools[1].Name != productdata.ToolNameWorkspaceRead || tools[2].Name != productdata.ToolNameSandboxExecCommand {
		t.Fatalf("tools = %+v", tools)
	}
	if tools[1].Group != string(productdata.ToolCatalogGroupWorkspace) || tools[1].ExecutionState != string(productdata.ToolExecutionStateExecutable) {
		t.Fatalf("workspace resolution = %+v", tools[1])
	}
	if tools[2].Group != string(productdata.ToolCatalogGroupSandbox) || tools[2].RiskLevel != string(productdata.ToolRiskHigh) || tools[2].ExecutionState != string(productdata.ToolExecutionStateExecutable) {
		t.Fatalf("sandbox resolution = %+v", tools[2])
	}
}

func TestWorkspaceToolDefinitionsSeparateReadAndMutationRisk(t *testing.T) {
	defs := WorkspaceToolDefinitions()
	if len(defs) != 5 {
		t.Fatalf("defs = %+v", defs)
	}
	for _, def := range defs {
		if !productdata.IsWorkspaceToolName(def.Name) || def.ApprovalPolicy != ToolApprovalAlwaysRequired {
			t.Fatalf("workspace definition = %+v", def)
		}
		if def.Name == productdata.ToolNameWorkspaceWriteFile || def.Name == productdata.ToolNameWorkspaceEdit {
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
	if len(defs) != 1 || defs[0].Name != productdata.ToolNameSandboxExecCommand {
		t.Fatalf("defs = %+v", defs)
	}
	if defs[0].ApprovalPolicy != ToolApprovalAlwaysRequired || defs[0].SafetyClass != ToolSafetySandboxCommand || defs[0].ExecutionState != ToolExecutionAllowlisted {
		t.Fatalf("sandbox definition = %+v", defs[0])
	}
}
