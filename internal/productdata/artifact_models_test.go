package productdata

import "testing"

func TestValidateArtifactToolCallArguments(t *testing.T) {
	create := RecordToolCallRequestInput{ToolCallID: "tc_artifact", ToolName: ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{"title": " Notes ", "filename": " notes.md ", "mime_type": " text/markdown ", "display": " inline ", "content": "hello", "max_bytes": 1024}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	input, err := ValidateToolCallRequestInput(create)
	if err != nil {
		t.Fatal(err)
	}
	if input.ArgumentsSummary["title"] != "Notes" || input.ArgumentsSummary["filename"] != "notes.md" || input.ArgumentsSummary["mime_type"] != "text/markdown" || input.ArgumentsSummary["display"] != "inline" {
		t.Fatalf("artifact fields were not normalized: %+v", input.ArgumentsSummary)
	}

	read := RecordToolCallRequestInput{ToolCallID: "tc_artifact_read", ToolName: ToolNameArtifactRead, ArgumentsSummary: map[string]any{"artifact_id": "art_123", "max_bytes": 512}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(read); err != nil {
		t.Fatal(err)
	}

	list := RecordToolCallRequestInput{ToolCallID: "tc_artifact_list", ToolName: ToolNameArtifactList, ArgumentsSummary: map[string]any{"limit": 10}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(list); err != nil {
		t.Fatal(err)
	}

	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_artifact", ToolName: ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_artifact", ToolName: ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{"title": "", "content": "hello"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_artifact", ToolName: ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{"title": "Notes", "content": "", "api_key": "sk-secret"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_artifact", ToolName: ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{"title": "Notes", "content": "hello", "display": "drawer"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_artifact_read", ToolName: ToolNameArtifactRead, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_artifact_list", ToolName: ToolNameArtifactList, ArgumentsSummary: map[string]any{"query": "secret"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
	} {
		if _, err := ValidateToolCallRequestInput(input); err == nil || ErrorCode(err) != CodeInvalidRequest {
			t.Fatalf("ValidateToolCallRequestInput(%+v) err = %v", input, err)
		}
	}
}
