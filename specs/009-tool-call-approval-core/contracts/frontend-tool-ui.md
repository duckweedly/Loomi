# Contract: M7 Frontend Tool UI

M7 extends existing ToolCallCard, RunRail, Timeline, and runtime adapter behavior. It must not mix tool lifecycle events into model text streaming.

## ToolCallCard

ToolCallCard must show:

| State | Required Display |
|-------|------------------|
| requested | Tool name and redacted argument summary |
| approval_required | Approval required label plus approve and deny controls |
| approved | Approved state; waiting/resuming indicator if execution has not begun |
| denied | Denied terminal state; no approve/deny controls |
| executing | Executing state |
| succeeded | Redacted result summary |
| failed | Redacted error code/message |
| cancelled | Cancelled state |

Rules:

- Approve/deny controls are visible only while approval is required and pending.
- Retry-clicks and network retries must be safe because backend decisions are idempotent.
- The card must never render raw provider payloads, API keys, Authorization headers, shell output, file contents, arbitrary URL contents, or unvalidated arguments.

## RunRail

RunRail should summarize tool states separately from model stream state.

Examples:

- `Waiting for tool approval`
- `Tool approved`
- `Running tool`
- `Tool succeeded`
- `Tool failed`
- `Tool cancelled`

RunRail must preserve existing queued/running/recovering/stopped/completed behavior from M6.

## Timeline

Timeline must group or label tool events distinctly.

Minimum grouping:

```text
Model
  model_request_started
  model_output_delta
Tool: runtime.get_current_time
  requested
  approval required
  approved
  executing
  succeeded
Run
  run_completed
```

Alternative UI grouping is acceptable if users can clearly distinguish tool lifecycle from model stream output.

## Runtime Adapter Mapping

The adapter must map run events to a stable tool-call view model:

| Event | View Model Update |
|-------|-------------------|
| `tool_call_requested` | Create/update tool card with requested state |
| `tool_call_approval_required` | Mark pending approval and enable controls |
| `tool_call_approved` | Mark approved and disable controls |
| `tool_call_denied` | Mark denied terminal and disable controls |
| `tool_call_executing` | Mark executing |
| `tool_call_succeeded` | Mark succeeded with result summary |
| `tool_call_failed` | Mark failed with redacted error |
| `tool_call_cancelled` | Mark cancelled |

History replay and live SSE must produce the same final view model.

## Browser Smoke Expectations

A local browser smoke should verify:

1. A model/fake-provider tool request appears as a ToolCallCard.
2. Arguments are summarized and safe.
3. Approval-required card exposes approve/deny controls.
4. Deny creates a denied terminal card and no execution state.
5. Approve moves through executing to result.
6. Timeline labels tool events separately from model events.
7. RunRail shows waiting/executing/result states clearly.
