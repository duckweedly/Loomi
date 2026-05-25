package runtime

import "testing"

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
