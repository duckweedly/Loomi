package productdata

import "testing"

func TestValidateSandboxToolCallArgumentsUsesTinyAllowlist(t *testing.T) {
	base := RecordToolCallRequestInput{ToolCallID: "tc_exec", ToolName: ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"pwd"}}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(base); err != nil {
		t.Fatalf("valid pwd rejected: %v", err)
	}
	base.ArgumentsSummary = map[string]any{"argv": []any{"ls", "."}}
	if _, err := ValidateToolCallRequestInput(base); err != nil {
		t.Fatalf("valid ls rejected: %v", err)
	}
	base.ArgumentsSummary = map[string]any{"argv": []any{"git", "status"}}
	if _, err := ValidateToolCallRequestInput(base); err != nil {
		t.Fatalf("valid git status rejected: %v", err)
	}

	for _, args := range []map[string]any{
		{"argv": []any{"cat", ".env"}},
		{"argv": []any{"head", "notes.txt"}},
		{"argv": []any{"tail", "notes.txt"}},
		{"argv": []any{"sed", "-n", "1p", "notes.txt"}},
		{"argv": []any{"wc", "notes.txt"}},
		{"argv": []any{"rg", "token"}},
		{"argv": []any{"ls", "secrets"}},
		{"argv": []any{"ls", "-la"}},
		{"argv": []any{"git", "diff"}},
		{"argv": []any{"git", "log"}},
		{"argv": []any{"python", "-c", "print(1)"}},
		{"argv": []any{"sh", "-c", "pwd"}},
	} {
		input := base
		input.ArgumentsSummary = args
		if _, err := ValidateToolCallRequestInput(input); err == nil {
			t.Fatalf("unsafe sandbox args accepted: %+v", args)
		}
	}
}
