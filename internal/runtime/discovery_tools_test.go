package runtime

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestDiscoveryLoadToolsReturnsEnabledToolDescriptions(t *testing.T) {
	invocation := ToolInvocation{
		RunID:    "run_discovery",
		ToolName: productdata.ToolNameLoadTools,
		ArgumentsSummary: map[string]any{
			"queries": []any{"workspace"},
			"limit":   2,
		},
		Catalog: []productdata.ToolCatalogEntry{
			{Name: productdata.ToolNameWorkspaceRead, DisplayName: "Workspace read", Description: "Read file", Group: productdata.ToolCatalogGroupWorkspace, RiskLevel: productdata.ToolRiskLow, ApprovalPolicy: productdata.ToolApprovalAlwaysRequired, SafeMetadata: map[string]any{"arguments": []string{"path"}}},
			{Name: productdata.ToolNameSandboxExecCommand, DisplayName: "Sandbox exec", Description: "Run command", Group: productdata.ToolCatalogGroupSandbox, RiskLevel: productdata.ToolRiskHigh, ApprovalPolicy: productdata.ToolApprovalAlwaysRequired},
		},
		EnabledTools: ToolResolutionsForPersona([]string{productdata.ToolNameWorkspaceRead}),
	}

	result, err := (DiscoveryToolExecutor{}).Execute(context.Background(), invocation)
	if err != nil {
		t.Fatal(err)
	}
	tools, ok := result["tools"].([]map[string]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("tools = %+v", result["tools"])
	}
	if tools[0]["name"] != productdata.ToolNameWorkspaceRead || result["scope"] != "runtime_catalog" || result["dynamic_schema_loader"] != false {
		t.Fatalf("result = %+v", result)
	}
}

func TestDiscoveryLoadSkillReturnsSafeManifestOnly(t *testing.T) {
	home := t.TempDir()
	writeSkill(t, filepath.Join(home, ".codex", "skills", "review", "SKILL.md"), "---\nname: review\ndescription: Review code.\n---\nSECRET_BODY")
	executor := DiscoveryToolExecutor{SkillInput: SkillDiscoveryInput{HomeDir: home, MaxFiles: 10}}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameLoadSkill, ArgumentsSummary: map[string]any{"name": "review"}})
	if err != nil {
		t.Fatal(err)
	}
	skills, ok := result["skills"].([]map[string]any)
	if !ok || len(skills) != 1 {
		t.Fatalf("skills = %+v", result["skills"])
	}
	if skills[0]["name"] != "review" || skills[0]["instruction_loaded"] != false || result["returns_skill_body"] != false {
		t.Fatalf("result = %+v", result)
	}
	for _, value := range skills[0] {
		if value == "SECRET_BODY" {
			t.Fatalf("skill body leaked: %+v", result)
		}
	}
}
