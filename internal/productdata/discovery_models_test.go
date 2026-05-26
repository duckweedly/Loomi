package productdata

import "testing"

func TestValidateDiscoveryToolCalls(t *testing.T) {
	loadTools := RecordToolCallRequestInput{ToolCallID: "tc_load_tools", ToolName: ToolNameLoadTools, ArgumentsSummary: map[string]any{"queries": []any{"workspace"}, "names": []any{ToolNameWorkspaceRead}, "limit": 5}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}
	if _, err := ValidateToolCallRequestInput(loadTools); err != nil {
		t.Fatalf("load_tools err = %v", err)
	}
	loadSkill := RecordToolCallRequestInput{ToolCallID: "tc_load_skill", ToolName: ToolNameLoadSkill, ArgumentsSummary: map[string]any{"name": "speckit"}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}
	if _, err := ValidateToolCallRequestInput(loadSkill); err != nil {
		t.Fatalf("load_skill err = %v", err)
	}
	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_load_tools", ToolName: ToolNameLoadTools, ArgumentsSummary: map[string]any{"query": "workspace"}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted},
		{ToolCallID: "tc_load_skill", ToolName: ToolNameLoadSkill, ArgumentsSummary: map[string]any{"name": ""}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted},
		{ToolCallID: "tc_load_skill", ToolName: ToolNameLoadSkill, ArgumentsSummary: map[string]any{"name": "skill", "path": "/tmp/SKILL.md"}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted},
	} {
		if _, err := ValidateToolCallRequestInput(input); err == nil {
			t.Fatalf("expected validation error for %+v", input)
		}
	}
}
