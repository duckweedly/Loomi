# Contract: MCP Worker Execution

## Purpose

Define how the worker executes one approved local stdio MCP tool call without duplicate execution under lease, cancellation, and recovery.

## Readiness Input

```json
{
  "run_id": "run_...",
  "job_id": "job_...",
  "worker_id": "worker_1",
  "tool_call_id": "tc_1",
  "approval_status": "approved",
  "execution_status": "not_started",
  "tool_name": "mcp.local-search.search"
}
```

## Readiness Rules

- Worker must currently own the run/job lease.
- Run must not be stopped, cancelled, terminal, or owned by another worker.
- Projection must be approved and `not_started`.
- Candidate must still resolve to the same discovered server/tool/schema hash approved by the user.
- If any rule fails, do not start the MCP process.

## Execution Start Event

Persist before process startup:

```json
{
  "type": "tool_call_executing",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_name": "mcp.local-search.search",
    "tool_source": "mcp",
    "server_slug": "local-search",
    "approval_status": "approved",
    "execution_status": "started"
  }
}
```

## Success Event

```json
{
  "type": "tool_call_succeeded",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_name": "mcp.local-search.search",
    "execution_status": "succeeded",
    "result_summary": {
      "kind": "json",
      "summary": "Redacted MCP result summary.",
      "result_hash": "sha256:..."
    },
    "result_for_model_redacted": {
      "summary": "Redacted MCP result for provider continuation."
    }
  }
}
```

## Failure Event

```json
{
  "type": "tool_call_failed",
  "category": "error",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_name": "mcp.local-search.search",
    "execution_status": "failed",
    "error_code": "mcp_stdio_timeout",
    "safe_message": "MCP tool execution timed out.",
    "retryable": false
  }
}
```

## Recovery Rules

- If status is `not_started`, a valid worker may execute after rechecking approval and ownership.
- If status is `started`, `succeeded`, `failed`, `denied`, or `cancelled`, M12 must not invoke the MCP tool again.
- A stale `started` state after worker crash is treated as unsafe to retry and must resolve through a redacted failed/cancelled recovery event, not re-execution.
- Stop/cancel racing with startup must be checked before process start and during timeout/cancellation handling.

## Stdio Lifecycle Rules

- Bound the call by per-tool timeout.
- Cleanup the process on timeout, cancellation, early exit, or terminal run state.
- Classify stderr without persisting raw stderr.
- Do not expose raw command, args, env, stdout, stderr, tokens, credentials, secret paths, file contents, shell output, browser state, or desktop captured data.
