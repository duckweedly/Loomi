package productdata

import "testing"

func TestValidateMemoryToolCallArguments(t *testing.T) {
	search := RecordToolCallRequestInput{ToolCallID: "tc_memory_search", ToolName: ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": " project context ", "limit": 5}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	input, err := ValidateToolCallRequestInput(search)
	if err != nil {
		t.Fatal(err)
	}
	if input.ArgumentsSummary["query"] != "project context" {
		t.Fatalf("query was not normalized: %+v", input.ArgumentsSummary)
	}

	list := RecordToolCallRequestInput{ToolCallID: "tc_memory_list", ToolName: ToolNameMemoryList, ArgumentsSummary: map[string]any{"limit": 5}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(list); err != nil {
		t.Fatal(err)
	}

	read := RecordToolCallRequestInput{ToolCallID: "tc_memory_read", ToolName: ToolNameMemoryRead, ArgumentsSummary: map[string]any{"entry_id": "mem_123"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(read); err != nil {
		t.Fatal(err)
	}

	write := RecordToolCallRequestInput{ToolCallID: "tc_memory_write", ToolName: ToolNameMemoryWrite, ArgumentsSummary: map[string]any{"title": "Decision", "content": "Keep memory writes approval gated."}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(write); err != nil {
		t.Fatal(err)
	}

	edit := RecordToolCallRequestInput{ToolCallID: "tc_memory_edit", ToolName: ToolNameMemoryEdit, ArgumentsSummary: map[string]any{"proposal_id": "memprop_123", "title": "Decision", "content": "Update the pending proposal."}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(edit); err != nil {
		t.Fatal(err)
	}

	forget := RecordToolCallRequestInput{ToolCallID: "tc_memory_forget", ToolName: ToolNameMemoryForget, ArgumentsSummary: map[string]any{"entry_id": "mem_123", "reason": "obsolete"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(forget); err != nil {
		t.Fatal(err)
	}

	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_memory_context", ToolName: ToolNameMemoryContext, ArgumentsSummary: map[string]any{"query": "project", "limit": 5}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_timeline", ToolName: ToolNameMemoryTimeline, ArgumentsSummary: map[string]any{"limit": 5}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_connections", ToolName: ToolNameMemoryConnections, ArgumentsSummary: map[string]any{"entry_id": "mem_123", "limit": 5}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_thread_search", ToolName: ToolNameMemoryThreadSearch, ArgumentsSummary: map[string]any{"query": "project", "limit": 5}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_thread_fetch", ToolName: ToolNameMemoryThreadFetch, ArgumentsSummary: map[string]any{"thread_id": "thr_123", "limit": 5}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
	} {
		if _, err := ValidateToolCallRequestInput(input); err != nil {
			t.Fatalf("valid memory tool rejected: %+v err=%v", input, err)
		}
	}

	status := RecordToolCallRequestInput{ToolCallID: "tc_memory_status", ToolName: ToolNameMemoryStatus, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}
	if _, err := ValidateToolCallRequestInput(status); err != nil {
		t.Fatal(err)
	}

	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_notebook_read", ToolName: ToolNameNotebookRead, ArgumentsSummary: map[string]any{"entry_id": "mem_123"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_notebook_write", ToolName: ToolNameNotebookWrite, ArgumentsSummary: map[string]any{"title": "Notebook", "content": "Keep durable notebook notes structured."}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_notebook_edit", ToolName: ToolNameNotebookEdit, ArgumentsSummary: map[string]any{"entry_id": "mem_123", "title": "Notebook", "content": "Replace notebook notes through audit."}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_notebook_forget", ToolName: ToolNameNotebookForget, ArgumentsSummary: map[string]any{"entry_id": "mem_123", "reason": "obsolete"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
	} {
		if _, err := ValidateToolCallRequestInput(input); err != nil {
			t.Fatalf("valid notebook tool rejected: %+v err=%v", input, err)
		}
	}

	for _, input := range []RecordToolCallRequestInput{
		{ToolCallID: "tc_memory_search", ToolName: ToolNameMemorySearch, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_search", ToolName: ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": "x", "api_key": "sk-secret"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_read", ToolName: ToolNameMemoryRead, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_write", ToolName: ToolNameMemoryWrite, ArgumentsSummary: map[string]any{"title": "", "content": "hello"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_write", ToolName: ToolNameMemoryWrite, ArgumentsSummary: map[string]any{"title": "Decision", "content": ""}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_edit", ToolName: ToolNameMemoryEdit, ArgumentsSummary: map[string]any{"title": "Decision", "content": "hello"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_forget", ToolName: ToolNameMemoryForget, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_connections", ToolName: ToolNameMemoryConnections, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_thread_search", ToolName: ToolNameMemoryThreadSearch, ArgumentsSummary: map[string]any{"query": ""}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_thread_fetch", ToolName: ToolNameMemoryThreadFetch, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_status", ToolName: ToolNameMemoryStatus, ArgumentsSummary: map[string]any{"query": "secret"}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_memory_status", ToolName: ToolNameMemoryStatus, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted},
		{ToolCallID: "tc_notebook_read", ToolName: ToolNameNotebookRead, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_notebook_write", ToolName: ToolNameNotebookWrite, ArgumentsSummary: map[string]any{"title": "Notebook", "content": ""}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_notebook_edit", ToolName: ToolNameNotebookEdit, ArgumentsSummary: map[string]any{"entry_id": "mem_123", "title": "Notebook", "content": ""}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
		{ToolCallID: "tc_notebook_forget", ToolName: ToolNameNotebookForget, ArgumentsSummary: map[string]any{}, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked},
	} {
		if _, err := ValidateToolCallRequestInput(input); err == nil || ErrorCode(err) != CodeInvalidRequest {
			t.Fatalf("ValidateToolCallRequestInput(%+v) err = %v", input, err)
		}
	}
}
