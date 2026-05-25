# Contract: Frontend Tool Approval Actions

## ToolCallCard Actions

When `approval_status=required`, ToolCallCard shows Approve and Deny controls.

Controls must:

- Call the real approve/deny API.
- Disable both controls while a request is in flight.
- Show an inline error if the request fails.
- Disable controls after approved, denied, executing, succeeded, failed, or cancelled.
- Accept SSE replay as the source of truth after the request returns.

## Adapter Mapping

The real execution adapter must map:

| Event | View Model |
|-------|------------|
| `tool.call.approved` | `approvalStatus=approved`, controls disabled |
| `tool.call.denied` | `approvalStatus=denied`, terminal denied |
| `tool.call.executing` | `executionStatus=executing` |
| `tool.call.succeeded` | `executionStatus=succeeded`, `resultSummary` set |
| `tool.call.failed` | `executionStatus=failed`, `errorCode` and `errorMessage` set |

History-first replay and live SSE must produce equivalent view models.

## Timeline and RunRail

Timeline and RunRail must label tool approval and execution events as tool lifecycle, not model output.
