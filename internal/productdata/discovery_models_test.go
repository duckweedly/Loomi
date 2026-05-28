package productdata

import "testing"

func TestValidateDiscoveryToolCalls(t *testing.T) {
	loadTools := RecordToolCallRequestInput{ToolCallID: "tc_load_tools", ToolName: ToolNameLoadTools, ArgumentsSummary: map[string]any{"queries": []any{"workspace"}, "names": []any{ToolNameWorkspaceRead}, "limit": 5}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}
	if _, err := ValidateToolCallRequestInput(loadTools); err != nil {
		t.Fatalf("load_tools err = %v", err)
	}
	stringArgs := RecordToolCallRequestInput{ToolCallID: "tc_load_tools_string", ToolName: ToolNameLoadTools, ArgumentsSummary: map[string]any{"queries": "workspace", "names": ToolNameWorkspaceRead, "limit": 5}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}
	validatedStringArgs, err := ValidateToolCallRequestInput(stringArgs)
	if err != nil {
		t.Fatalf("load_tools string args err = %v", err)
	}
	if got := validatedStringArgs.ArgumentsSummary["names"]; len(got.([]any)) != 1 || got.([]any)[0] != ToolNameWorkspaceRead {
		t.Fatalf("load_tools string names were not normalized: %+v", got)
	}
	queriesOnly := RecordToolCallRequestInput{ToolCallID: "tc_load_tools_queries_only", ToolName: ToolNameLoadTools, ArgumentsSummary: map[string]any{"queries": []any{"workspace list files directory glob ls"}, "names": []any{}, "limit": 10}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}
	validatedQueriesOnly, err := ValidateToolCallRequestInput(queriesOnly)
	if err != nil {
		t.Fatalf("queries-only load_tools err = %v", err)
	}
	if got := validatedQueriesOnly.ArgumentsSummary["names"]; len(got.([]any)) != 0 {
		t.Fatalf("empty load_tools names were not preserved: %+v", got)
	}
	queryOnly := RecordToolCallRequestInput{ToolCallID: "tc_load_tools_query", ToolName: ToolNameLoadTools, ArgumentsSummary: map[string]any{"query": "workspace list files directory glob ls", "names": []any{}, "limit": 10}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}
	validatedQueryOnly, err := ValidateToolCallRequestInput(queryOnly)
	if err != nil {
		t.Fatalf("query-only load_tools err = %v", err)
	}
	if got := validatedQueryOnly.ArgumentsSummary["queries"]; len(got.([]any)) != 1 || got.([]any)[0] != "workspace list files directory glob ls" {
		t.Fatalf("query was not normalized into queries: %+v", validatedQueryOnly.ArgumentsSummary)
	}
	emptyQuery := RecordToolCallRequestInput{ToolCallID: "tc_load_tools_empty", ToolName: ToolNameLoadTools, ArgumentsSummary: map[string]any{"names": []any{}, "limit": 10}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}
	if _, err := ValidateToolCallRequestInput(emptyQuery); err != nil {
		t.Fatalf("empty load_tools query should list safe catalog: %v", err)
	}
	loadSkill := RecordToolCallRequestInput{ToolCallID: "tc_load_skill", ToolName: ToolNameLoadSkill, ArgumentsSummary: map[string]any{"name": "speckit"}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}
	if _, err := ValidateToolCallRequestInput(loadSkill); err != nil {
		t.Fatalf("load_skill err = %v", err)
	}
	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_load_skill", ToolName: ToolNameLoadSkill, ArgumentsSummary: map[string]any{"name": ""}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted},
		{ToolCallID: "tc_load_skill", ToolName: ToolNameLoadSkill, ArgumentsSummary: map[string]any{"name": "skill", "path": "/tmp/SKILL.md"}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted},
	} {
		if _, err := ValidateToolCallRequestInput(input); err == nil {
			t.Fatalf("expected validation error for %+v", input)
		}
	}
}
