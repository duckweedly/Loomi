package productdata

import "testing"

func TestValidateSandboxToolCallArgumentsUsesBoundedAllowlist(t *testing.T) {
	base := RecordToolCallRequestInput{ToolCallID: "tc_exec", ToolName: ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"pwd"}}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(base); err != nil {
		t.Fatalf("valid pwd rejected: %v", err)
	}
	start := base
	start.ToolName = ToolNameSandboxStartProcess
	if _, err := ValidateToolCallRequestInput(start); err != nil {
		t.Fatalf("valid start process rejected: %v", err)
	}
	start.ArgumentsSummary = map[string]any{"argv": []any{"cat"}, "stdin": true}
	if _, err := ValidateToolCallRequestInput(start); err != nil {
		t.Fatalf("valid stdin start process rejected: %v", err)
	}
	for _, name := range []string{ToolNameSandboxContinueProcess, ToolNameSandboxTerminateProcess} {
		input := base
		input.ToolName = name
		input.ArgumentsSummary = map[string]any{"process_id": "sp_test"}
		if _, err := ValidateToolCallRequestInput(input); err != nil {
			t.Fatalf("valid %s rejected: %v", name, err)
		}
		input.ArgumentsSummary = map[string]any{}
		if _, err := ValidateToolCallRequestInput(input); err == nil {
			t.Fatalf("%s without process_id accepted", name)
		}
	}
	continued := base
	continued.ToolName = ToolNameSandboxContinueProcess
	continued.ArgumentsSummary = map[string]any{"process_id": "sp_test", "cursor": 0, "stdin_text": "hello\n", "input_seq": 1, "close_stdin": true}
	if _, err := ValidateToolCallRequestInput(continued); err != nil {
		t.Fatalf("valid stdin continue rejected: %v", err)
	}
	continued.ArgumentsSummary = map[string]any{"process_id": "sp_test", "stdin_text": "hello\n"}
	if _, err := ValidateToolCallRequestInput(continued); err == nil {
		t.Fatal("stdin continue without input_seq accepted")
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
		{"argv": []any{"cat", "src/notes.txt"}},
		{"argv": []any{"head", "-n", "20", "src/notes.txt"}},
		{"argv": []any{"sed", "-n", "1,20p", "src/notes.txt"}},
		{"argv": []any{"wc", "src/notes.txt"}},
		{"argv": []any{"rg", "needle", "src"}},
		{"argv": []any{"git", "diff"}},
		{"argv": []any{"go", "test", "./internal/runtime"}},
		{"argv": []any{"bun", "run", "build", "web"}},
	} {
		input := base
		input.ArgumentsSummary = args
		if _, err := ValidateToolCallRequestInput(input); err != nil {
			t.Fatalf("safe sandbox args rejected: %+v err=%v", args, err)
		}
	}

	for _, args := range []map[string]any{
		{"argv": []any{"cat", ".env"}},
		{"argv": []any{"ls", "secrets"}},
		{"argv": []any{"ls", "-la"}},
		{"argv": []any{"rg", "token", "secrets"}},
		{"argv": []any{"git", "checkout", "--", "x"}},
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
