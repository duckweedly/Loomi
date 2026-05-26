package productdata

import "testing"

func TestValidateBrowserToolCallArguments(t *testing.T) {
	open := RecordToolCallRequestInput{ToolCallID: "tc_browser", ToolName: ToolNameBrowserOpen, ArgumentsSummary: map[string]any{"url": " https://example.com/docs ", "max_bytes": 1024, "timeout_ms": 1000}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	input, err := ValidateToolCallRequestInput(open)
	if err != nil {
		t.Fatal(err)
	}
	if input.ArgumentsSummary["url"] != "https://example.com/docs" {
		t.Fatalf("url was not normalized: %+v", input.ArgumentsSummary)
	}

	snapshot := RecordToolCallRequestInput{ToolCallID: "tc_browser_snapshot", ToolName: ToolNameBrowserSnapshot, ArgumentsSummary: map[string]any{"session_id": "br_run_tc"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(snapshot); err != nil {
		t.Fatal(err)
	}

	click := RecordToolCallRequestInput{ToolCallID: "tc_browser_click", ToolName: ToolNameBrowserClickLink, ArgumentsSummary: map[string]any{"session_id": "br_run_tc", "link_index": 0}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(click); err != nil {
		t.Fatal(err)
	}

	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_browser", ToolName: ToolNameBrowserOpen, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_browser", ToolName: ToolNameBrowserOpen, ArgumentsSummary: map[string]any{"url": "", "api_key": "sk-secret"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_browser_snapshot", ToolName: ToolNameBrowserSnapshot, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_browser_click", ToolName: ToolNameBrowserClickLink, ArgumentsSummary: map[string]any{"session_id": "br_run_tc"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
	} {
		if _, err := ValidateToolCallRequestInput(input); err == nil || ErrorCode(err) != CodeInvalidRequest {
			t.Fatalf("ValidateToolCallRequestInput(%+v) err = %v", input, err)
		}
	}
}
