# Contract: M7 Worker Approval Wait and Resume

M7 reuses the M6 background job and worker pipeline. It must not introduce a second queue for the first tool slice.

## Worker Responsibilities

1. Execute model-backed run work through the existing M6 job pipeline.
2. Detect provider-normalized tool requests from the M5 gateway boundary.
3. Validate tool name and arguments against the allowlisted Tool Definition.
4. Persist `tool_call_requested` and either:
   - `tool_call_approval_required` and block the run, or
   - `tool_call_failed` for unsupported/malformed requests, or
   - `tool_call_executing` for a future no-approval tool policy.
5. Release or pause active work while waiting for approval; do not keep a worker lease busy indefinitely.
6. Resume after idempotent approval by scheduling/waking the existing run job.
7. Observe stop requests before approval wait, before execution, and before result write.
8. Persist exactly one terminal tool event.

## Run/Job States

M7 may add a run status or job substatus such as `blocked_on_tool_approval`. Exact implementation can vary, but the contract is:

| State | Worker Behavior |
|-------|-----------------|
| queued | Claimable by normal M6 worker flow |
| running | Worker may process model/tool-safe steps |
| blocked_on_tool_approval | Not executing; visible to UI; not treated as failed/stale worker work |
| resumable_after_tool_approval | Existing job pipeline can continue tool execution |
| stopped/failed/completed | Terminal; no further tool writes except idempotent current-state reads |

## Approval Wait

When a tool requires approval:

1. Worker records `tool_call_requested`.
2. Worker records `tool_call_approval_required`.
3. Worker marks the run/job as blocked on approval.
4. Worker stops active execution for that run.
5. Browser can reconnect and see the approval state through history-first SSE.

The worker must not:

- Keep renewing a lease only to wait for the user.
- Auto-execute approval-required tools.
- Convert approval wait into worker failure/retry noise.

## Approval Resume

When approve is recorded:

1. Approval API validates scope by `thread_id`, `run_id`, `tool_call_id`.
2. Approval API atomically transitions approval state to approved if pending.
3. Approval API records `tool_call_approved` exactly once.
4. Approval API schedules or wakes existing M6 job processing exactly once.
5. Worker claims/resumes and records `tool_call_executing`.
6. Executor runs allowlisted internal tool and records `tool_call_succeeded` or `tool_call_failed`.

Repeated approve calls must return the same current state without duplicate wakeups or events.

## Denial

When deny is recorded:

1. Deny API validates scope.
2. Deny API atomically transitions approval state to denied if pending.
3. Deny API records `tool_call_denied` exactly once.
4. No execution job is scheduled for that tool call.
5. MVP finalizes the run as `run_stopped` after denial, with a visible user-safe summary and no tool execution.

Repeated deny calls must return the same current state without duplicate events.

## Cancellation

Stop/cancel behavior must win over approval/execution races:

- Pending approval -> `tool_call_cancelled`, `run_stopped`.
- Approved but not executing -> `tool_call_cancelled`, no executor start.
- Executing -> cooperative cancellation at safe boundary; if tool result returns after cancellation, it must not overwrite cancelled state.

## Recovery and Lease Semantics

- Blocked approval wait is not a stale lease.
- Recovery should resume only if the tool call is approved and non-terminal.
- Recovery must not duplicate `tool_call_executing` or terminal tool events.
- Two workers racing after approval must produce at most one execution attempt and one terminal event.

## Diagnostics

Worker diagnostics should include redacted counts for runs blocked on tool approval and resumable tool calls, but must not include raw tool arguments or results.
