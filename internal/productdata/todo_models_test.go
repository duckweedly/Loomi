package productdata

import "testing"

func TestValidateTodoWriteToolArgumentsNormalizesSafeItems(t *testing.T) {
	input, err := ValidateToolCallRequestInput(RecordToolCallRequestInput{
		ToolCallID: "tc_todo",
		ToolName:   ToolNameTodoWrite,
		ArgumentsSummary: map[string]any{"items": []any{
			map[string]any{"id": "todo-1", "title": "Review patch", "status": "running", "summary": "Checking tests"},
			map[string]any{"id": "todo-2", "title": "Run bash /tmp/secret", "status": "wat", "extra": "drop"},
		}},
		ApprovalStatus:  ToolCallApprovalRequired,
		ExecutionStatus: ToolCallExecutionBlocked,
	})
	if err != nil {
		t.Fatal(err)
	}
	items, ok := input.ArgumentsSummary["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("items = %#v", input.ArgumentsSummary["items"])
	}
	first := items[0].(map[string]any)
	second := items[1].(map[string]any)
	if first["title"] != "Review patch" || first["status"] != "running" || first["summary"] != "Checking tests" {
		t.Fatalf("first = %+v", first)
	}
	if second["title"] != "[redacted]" || second["status"] != "pending" || second["redaction_applied"] != true {
		t.Fatalf("second = %+v", second)
	}
}

func TestValidateTodoWriteToolRejectsInvalidArguments(t *testing.T) {
	cases := []RecordToolCallRequestInput{
		{ToolCallID: "tc_todo", ToolName: ToolNameTodoWrite, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_todo", ToolName: ToolNameTodoWrite, ArgumentsSummary: map[string]any{"items": []any{}}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_todo", ToolName: ToolNameTodoWrite, ArgumentsSummary: map[string]any{"items": []any{map[string]any{"title": "ok"}}, "path": "notes.txt"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
	}
	for _, input := range cases {
		if _, err := ValidateToolCallRequestInput(input); err == nil {
			t.Fatalf("expected error for %+v", input)
		}
	}
}
