# Contract: M7 Approval Execution HTTP and Event API

## Approve

```http
POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/approve
```

Response `200` returns the current redacted tool-call projection for a newly approved call or a repeated approve.

Response `404` is used for unknown, wrong-thread, wrong-run, or wrong-user scoped lookup failures.

Response `409` is used for incompatible current state such as denied, executing, succeeded, failed, cancelled, or any state that cannot safely transition to approved.

## Deny

```http
POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/deny
```

Response `200` returns the current redacted denied projection for a newly denied call or repeated deny.

Response `404` is used for unknown, wrong-thread, wrong-run, or wrong-user scoped lookup failures.

Response `409` is used for incompatible current state such as executing, succeeded, failed, cancelled, or already approved after execution became possible.

## Event Contract

Approval/execution events must be persisted and replayed in sequence:

```text
tool.call.requested
tool.call.approval_required
tool.call.approved | tool.call.denied
tool.call.executing
tool.call.succeeded | tool.call.failed
```

Denied path:

```text
tool.call.requested
tool.call.approval_required
tool.call.denied
run.stopped
```

Approved success path:

```text
tool.call.requested
tool.call.approval_required
tool.call.approved
tool.call.executing
tool.call.succeeded
run.completed | run.stopped
```

Approved failure path:

```text
tool.call.requested
tool.call.approval_required
tool.call.approved
tool.call.executing
tool.call.failed
run.failed
```

## Redaction

Events may include:

- `tool_call_id`
- `tool_name`
- `arguments_summary`
- `result_summary`
- `error_code`
- `error_message`

Events must not include raw provider payloads, secrets, shell output, file contents, arbitrary URL contents, authorization headers, or unvalidated arguments.
