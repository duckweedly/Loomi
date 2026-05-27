# Contract: MCP Tool Result Continuation

## Purpose

Define the single provider continuation after one approved MCP tool execution succeeds.

## Continuation Input

```json
{
  "run_id": "run_...",
  "thread_id": "thread_...",
  "model_phase": "continuation",
  "assistant_tool_call": {
    "tool_call_id": "tc_1",
    "tool_name": "mcp.local-search.search",
    "arguments_summary": {
      "query": "[redacted-summary]"
    }
  },
  "tool_result": {
    "tool_call_id": "tc_1",
    "tool_name": "mcp.local-search.search",
    "result_for_model_redacted": {
      "summary": "Redacted MCP result for provider continuation."
    }
  }
}
```

## Rules

- Continuation happens only after the matching MCP projection succeeds.
- Denied, failed, cancelled, stopped, or ownership-lost executions do not continue.
- Continuation input must use `result_for_model_redacted` when present; otherwise it may use the safe result summary.
- Raw MCP output is never eligible for continuation.
- Only one continuation attempt is allowed in M12.

## Continuation Events

```json
{
  "type": "model_output_delta",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_source": "mcp",
    "model_phase": "continuation"
  }
}
```

```json
{
  "type": "run_completed",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_source": "mcp",
    "model_phase": "continuation",
    "final_message_created": true
  }
}
```

## Additional Tool Request Rejection

If the continuation provider requests another tool:

```json
{
  "type": "run_failed",
  "category": "error",
  "metadata": {
    "tool_call_id": "tc_2",
    "tool_source": "mcp",
    "error_code": "unsupported_tool_loop",
    "safe_message": "M12 supports one approved MCP tool execution and one continuation per run."
  }
}
```

No additional MCP, runtime, shell, filesystem, browser, or desktop tool executes after this event.
