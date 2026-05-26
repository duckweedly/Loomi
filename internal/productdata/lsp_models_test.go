package productdata

import "testing"

func TestValidateLSPToolCallArguments(t *testing.T) {
	base := RecordToolCallRequestInput{ToolCallID: "tc_lsp", ToolName: ToolNameLSPSymbols, ArgumentsSummary: map[string]any{"path": "src/main.go"}, ArgumentsHash: "hash_lsp", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(base); err != nil {
		t.Fatalf("valid symbols err = %v", err)
	}
	cases := []RecordToolCallRequestInput{
		{ToolCallID: "tc_lsp", ToolName: ToolNameLSPSymbols, ArgumentsSummary: map[string]any{}, ArgumentsHash: "hash_lsp", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_lsp", ToolName: ToolNameLSPSymbols, ArgumentsSummary: map[string]any{"path": "src/main.go", "api_key": "sk-secret"}, ArgumentsHash: "hash_lsp", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_lsp", ToolName: ToolNameLSPReferences, ArgumentsSummary: map[string]any{"path": "src/main.go", "line": 3}, ArgumentsHash: "hash_lsp", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_lsp", ToolName: ToolNameLSPReferences, ArgumentsSummary: map[string]any{"path": "src/main.go", "line": 0, "column": 6}, ArgumentsHash: "hash_lsp", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
	}
	for _, input := range cases {
		if _, err := ValidateToolCallRequestInput(input); err == nil || ErrorCode(err) != CodeInvalidRequest {
			t.Fatalf("ValidateToolCallRequestInput(%+v) err = %v", input, err)
		}
	}
}
