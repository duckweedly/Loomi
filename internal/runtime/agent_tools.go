package runtime

import (
	"context"
	"errors"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type AgentToolExecutor struct {
	Tasks productdata.AgentTaskService
}

func AgentToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameAgentSpawn, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameAgentList, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameAgentStart, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameAgentDelegate, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameAgentComplete, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameAgentFail, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

func (e AgentToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if e.Tasks == nil {
		return nil, errors.New("agent task service is unavailable")
	}
	switch invocation.ToolName {
	case productdata.ToolNameAgentSpawn:
		return e.spawn(ctx, invocation)
	case productdata.ToolNameAgentList:
		return e.list(ctx, invocation)
	case productdata.ToolNameAgentStart:
		return e.start(ctx, invocation)
	case productdata.ToolNameAgentDelegate:
		return e.delegate(ctx, invocation)
	case productdata.ToolNameAgentComplete:
		return e.complete(ctx, invocation)
	case productdata.ToolNameAgentFail:
		return e.fail(ctx, invocation)
	default:
		return nil, errors.New("agent tool is not supported")
	}
}

func (e AgentToolExecutor) spawn(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	role, _ := invocation.ArgumentsSummary["role"].(string)
	goal, _ := invocation.ArgumentsSummary["goal"].(string)
	task, err := e.Tasks.SpawnAgentTask(ctx, identity.LocalDevIdentity(), productdata.SpawnAgentTaskInput{
		ThreadID: invocation.ThreadID,
		RunID:    invocation.RunID,
		Role:     role,
		Goal:     goal,
	})
	if err != nil {
		return nil, err
	}
	return agentTaskSummary(productdata.ToolNameAgentSpawn, "spawn", task), nil
}

func (e AgentToolExecutor) list(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	tasks, err := e.Tasks.ListAgentTasks(ctx, identity.LocalDevIdentity(), productdata.ListAgentTasksInput{
		ThreadID: invocation.ThreadID,
		Limit:    boundedInt(invocation.ArgumentsSummary, "limit", 20, 50),
	})
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, agentTaskSummary("", "list_item", task))
	}
	return map[string]any{
		"tool":                 productdata.ToolNameAgentList,
		"scope":                "agent",
		"operation":            "list",
		"tasks":                items,
		"count":                len(items),
		"autonomous_execution": false,
		"redaction_applied":    false,
	}, nil
}

func (e AgentToolExecutor) start(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	taskID, _ := invocation.ArgumentsSummary["task_id"].(string)
	task, err := e.Tasks.StartAgentTask(ctx, identity.LocalDevIdentity(), productdata.StartAgentTaskInput{
		ThreadID: invocation.ThreadID,
		TaskID:   taskID,
	})
	if err != nil {
		return nil, err
	}
	return agentTaskSummary(productdata.ToolNameAgentStart, "start", task), nil
}

func (e AgentToolExecutor) delegate(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	taskID, _ := invocation.ArgumentsSummary["task_id"].(string)
	task, err := e.Tasks.DelegateAgentTask(ctx, identity.LocalDevIdentity(), productdata.DelegateAgentTaskInput{
		ThreadID:         invocation.ThreadID,
		TaskID:           taskID,
		ParentToolCallID: invocation.ToolCallID,
	})
	if err != nil {
		return nil, err
	}
	return agentTaskSummary(productdata.ToolNameAgentDelegate, "delegate", task), nil
}

func (e AgentToolExecutor) complete(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	taskID, _ := invocation.ArgumentsSummary["task_id"].(string)
	resultSummary, _ := invocation.ArgumentsSummary["result_summary"].(string)
	task, err := e.Tasks.CompleteAgentTask(ctx, identity.LocalDevIdentity(), productdata.CompleteAgentTaskInput{
		ThreadID:      invocation.ThreadID,
		TaskID:        taskID,
		ResultSummary: resultSummary,
	})
	if err != nil {
		return nil, err
	}
	return agentTaskSummary(productdata.ToolNameAgentComplete, "complete", task), nil
}

func (e AgentToolExecutor) fail(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	taskID, _ := invocation.ArgumentsSummary["task_id"].(string)
	resultSummary, _ := invocation.ArgumentsSummary["result_summary"].(string)
	task, err := e.Tasks.FailAgentTask(ctx, identity.LocalDevIdentity(), productdata.FailAgentTaskInput{
		ThreadID:      invocation.ThreadID,
		TaskID:        taskID,
		ResultSummary: resultSummary,
	})
	if err != nil {
		return nil, err
	}
	return agentTaskSummary(productdata.ToolNameAgentFail, "fail", task), nil
}

func agentTaskSummary(tool string, operation string, task productdata.AgentTask) map[string]any {
	summary := map[string]any{
		"scope":                "agent",
		"operation":            operation,
		"task_id":              task.ID,
		"role":                 task.Role,
		"goal":                 task.Goal,
		"status":               string(task.Status),
		"result_summary":       task.ResultSummary,
		"autonomous_execution": false,
		"redaction_applied":    false,
	}
	if tool != "" {
		summary["tool"] = tool
	}
	if task.RunID != "" {
		summary["run_id"] = task.RunID
	}
	if task.ThreadID != "" {
		summary["thread_id"] = task.ThreadID
	}
	if task.ChildThreadID != "" {
		summary["child_thread_id"] = task.ChildThreadID
	}
	if task.ChildRunID != "" {
		summary["child_run_id"] = task.ChildRunID
	}
	if task.ParentToolCallID != "" {
		summary["parent_tool_call_id"] = task.ParentToolCallID
	}
	if task.DelegatedAt != nil {
		summary["delegated_at"] = task.DelegatedAt.Format(time.RFC3339Nano)
	}
	if tool == productdata.ToolNameAgentDelegate {
		summary["autonomous_execution"] = true
	}
	return summary
}
