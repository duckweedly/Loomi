# Contract: Bounded Agent Loop + Todo Foundation

## Loop Event Contract

A second approved tool call in the same run follows the same event shape as the first:

```text
tool_call_requested -> tool_call_approval_required
tool_call_approved -> tool_call_executing -> tool_call_succeeded
model_request_started(model_phase=continuation)
```

Each tool call has a distinct `tool_call_id` and a 1-based `loop_index`.

## Approval Contract

Approval for one tool call authorizes only that tool call:

```json
{
  "tool_call_id": "tc_read_2",
  "tool_name": "workspace.read",
  "loop_index": 2,
  "approval_status": "required",
  "execution_status": "blocked"
}
```

No later tool call may inherit this approval.

## Loop Limit Contract

When the provider requests a tool after the configured limit:

```json
{
  "error_code": "tool_loop_limit_reached",
  "loop_count": 3,
  "max_tool_calls": 3
}
```

No approval-required event or tool execution is recorded for the over-limit tool request.

## Todo Contract

Todo state is safe run metadata:

```json
{
  "todo_items": [
    {"id": "todo_1", "title": "Find candidate files", "status": "completed"},
    {"id": "todo_2", "title": "Read the selected file", "status": "running"}
  ],
  "redaction_applied": false
}
```

Todo metadata must not include raw provider payloads, raw tool results, file contents, host absolute paths, shell commands, browser state, credentials, tokens, or secret-looking values.
