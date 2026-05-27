package productdata

import "testing"

func TestValidateWebFetchToolCallArguments(t *testing.T) {
	base := RecordToolCallRequestInput{ToolCallID: "tc_web", ToolName: ToolNameWebFetch, ArgumentsSummary: map[string]any{"url": " https://example.com/docs ", "max_bytes": 1024, "timeout_ms": 1000}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	input, err := ValidateToolCallRequestInput(base)
	if err != nil {
		t.Fatal(err)
	}
	if input.ArgumentsSummary["url"] != "https://example.com/docs" {
		t.Fatalf("url was not normalized: %+v", input.ArgumentsSummary)
	}
	for _, args := range []map[string]any{
		{},
		{"url": ""},
		{"url": "https://example.com", "api_key": "sk-secret"},
	} {
		input := base
		input.ArgumentsSummary = args
		if _, err := ValidateToolCallRequestInput(input); err == nil || ErrorCode(err) != CodeInvalidRequest {
			t.Fatalf("ValidateToolCallRequestInput(%+v) err = %v", input, err)
		}
	}
}

func TestValidateWebSearchToolCallArguments(t *testing.T) {
	base := RecordToolCallRequestInput{ToolCallID: "tc_search", ToolName: ToolNameWebSearch, ArgumentsSummary: map[string]any{"query": "  最新 AI 新闻  ", "provider": "tavily", "limit": 5, "timeout_ms": 1000}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	input, err := ValidateToolCallRequestInput(base)
	if err != nil {
		t.Fatal(err)
	}
	if input.ArgumentsSummary["query"] != "最新 AI 新闻" {
		t.Fatalf("query was not normalized: %+v", input.ArgumentsSummary)
	}
	for _, args := range []map[string]any{
		{},
		{"query": ""},
		{"query": "news", "provider": "bing"},
		{"query": "news", "api_key": "tvly-secret"},
	} {
		input := base
		input.ArgumentsSummary = args
		if _, err := ValidateToolCallRequestInput(input); err == nil || ErrorCode(err) != CodeInvalidRequest {
			t.Fatalf("ValidateToolCallRequestInput(%+v) err = %v", input, err)
		}
	}
}

func TestValidateWebSearchToolCallNormalizesProviderArgumentAliases(t *testing.T) {
	input, err := ValidateToolCallRequestInput(RecordToolCallRequestInput{
		ToolCallID:       "tc_search",
		ToolName:         ToolNameWebSearch,
		ArgumentsSummary: map[string]any{"q": "latest AI news", "count": 3, "provider": "brave"},
		ApprovalStatus:   ToolCallApprovalApproved,
		ExecutionStatus:  ToolCallExecutionNotStarted,
	})
	if err != nil {
		t.Fatalf("ValidateToolCallRequestInput() error = %v", err)
	}
	if input.ArgumentsSummary["query"] != "latest AI news" || input.ArgumentsSummary["limit"] != 3 {
		t.Fatalf("arguments = %+v", input.ArgumentsSummary)
	}
	if _, ok := input.ArgumentsSummary["q"]; ok {
		t.Fatalf("q alias leaked into normalized arguments: %+v", input.ArgumentsSummary)
	}
	if _, ok := input.ArgumentsSummary["count"]; ok {
		t.Fatalf("count alias leaked into normalized arguments: %+v", input.ArgumentsSummary)
	}
}

func TestValidateWebFetchAllowsAutoApprovedPublicRead(t *testing.T) {
	input, err := ValidateToolCallRequestInput(RecordToolCallRequestInput{
		ToolCallID:       "tc_fetch",
		ToolName:         ToolNameWebFetch,
		ArgumentsSummary: map[string]any{"url": " https://example.com/repo ", "max_bytes": 4096},
		ApprovalStatus:   ToolCallApprovalApproved,
		ExecutionStatus:  ToolCallExecutionNotStarted,
	})
	if err != nil {
		t.Fatalf("ValidateToolCallRequestInput() error = %v", err)
	}
	if input.ArgumentsSummary["url"] != "https://example.com/repo" {
		t.Fatalf("arguments = %+v", input.ArgumentsSummary)
	}
}
