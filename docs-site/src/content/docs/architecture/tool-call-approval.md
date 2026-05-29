---
title: M7 Tool Call Approval Core Architecture
description: Approval-gated internal tool-call foundation, audit events, and safety boundaries for Loomi M7.
---

M7 starts the tool-calling slice without opening broad tool execution. The foundation records model-requested internal tool calls, blocks runs on approval, keeps a redacted current-state projection, and replays the lifecycle through persisted run events.

The implemented execution closure now converts allowlisted provider `runtime.get_current_time` requests into durable approval-required tool calls, lets the user approve or deny, resumes approved runs through the existing M6 worker/job path, executes the internal current-time tool, and replays approval/execution/result states through SSE.

## Boundary

The product data layer owns durable tool-call state and run events. Runtime code owns the allowlisted `runtime.get_current_time` tool definition and executor. Frontend runtime adapters replay tool events into a stable `ToolCall` view model.

M7 reuses the M6 worker/job and M4 run/event/SSE foundations:

- runs can enter `blocked_on_tool_approval`
- approve moves a blocked run back to a queued worker-resumable state
- deny finalizes the MVP run as stopped only while the call is still `required` and `blocked`
- history-first SSE continues to replay persisted events by sequence
- worker diagnostics can count approval-blocked and resumable tool calls
- RunRail/runtime grouping separates tool-call events from model stream and worker/job rows

## Durable projection

M7 adds a minimal `tool_calls` projection in migration `000006_m7_tool_call_approval`:

- one row per `(run_id, tool_call_id)`
- scoped to thread and run
- `tool_name`, redacted `arguments_summary`, and `arguments_hash`
- `approval_status` and `execution_status`
- optional redacted result/error summaries for later phases

Run events remain the audit and replay contract. The projection exists for idempotency, scoped lookup, worker resume, and concurrency guards.

M98 keeps approval-gated execution serialized while allowing a model turn to record more than one approval-required request. Multiple `blocked` tool calls can coexist for the same run, so the timeline can show the full batch the model asked for instead of failing on the second request. Once a user approves one of those calls, that approved/executing path remains guarded by the worker job and tool-call state; write-capable, sandbox, MCP, browser, artifact mutation, memory-write, and other gated tools are not batch-executed automatically. If the run is stopped or one pending approval is denied, every other unresolved tool call in that run is cancelled and emits `tool_call_cancelled`, so the run-step projection has no hidden pending or executing tools after the terminal event.

## Lifecycle events

Backend event types use the existing run-event categories and frontend maps them to dotted runtime names:

| Backend type | Frontend type | Purpose |
| --- | --- | --- |
| `tool_call_requested` | `tool.call.requested` | Model requested an allowlisted tool. |
| `tool_call_approval_required` | `tool.call.approval_required` | Run is blocked waiting for approval. |
| `tool_call_approved` | `tool.call.approved` | User approval was recorded. |
| `tool_call_denied` | `tool.call.denied` | User denial was recorded. |
| `tool_call_executing` | `tool.call.executing` | Executor started the tool. |
| `tool_call_succeeded` | `tool.call.succeeded` | Tool completed with a redacted result. |
| `tool_call_failed` | `tool.call.failed` | Validation or execution failed safely. |
| `tool_call_cancelled` | `tool.call.cancelled` | Tool call was cancelled. |

Tool lifecycle events use existing `progress` or `error` categories. No separate `tool` event category is introduced.

## Current tool boundary

The only Phase 2 executable tool definition is `runtime.get_current_time`.

Rules:

- approval policy is `always_required`
- safety class is `no_side_effect_internal`
- allowed arguments are omitted `timezone` or `timezone: "UTC"`
- omitted timezone normalizes to `UTC`
- unknown argument fields are rejected
- result shape contains `iso_time`, `timezone`, and `source`

The product data boundary repeats the same schema checks because it cannot import runtime without creating a cycle. Later M7 work may move the shared schema into a lower-level package if more tools are added.

## Safety boundaries

M7 Phase 2 does not add:

- shell or terminal tools
- filesystem read/write tools
- MCP integration
- browser automation
- arbitrary network tools
- multi-agent execution
- long-term memory or RAG
- approval bypass

Model-generated tool arguments are untrusted data. Persisted summaries and events must not contain raw provider payloads, API keys, Authorization headers, passwords, tokens, credentials, shell output, file contents, or arbitrary URL contents. Redaction checks both sensitive keys and sensitive-looking values before metadata is persisted.

## Approval execution closure

Approve and deny decisions are owned by the product data layer so projection updates and run events share one transaction boundary. Repeated same-decision requests return current state without duplicate decision events. Wrong-scope or incompatible states are rejected before state changes. Once a call is approved, even if execution has not started yet, a later deny request is treated as an incompatible decision and cannot flip the run into stopped. Denying one blocked call is a terminal run decision: the denied call records `tool_call_denied`, sibling unresolved calls record `tool_call_cancelled`, and then the run records `run_stopped`.

Approved calls create a normal queued run job with `tool_call_id` metadata. The queued run router recognizes that metadata and executes the approved tool directly instead of making another provider request. Execution writes `tool_call_executing` before invocation and then exactly one terminal tool event:

- `tool_call_succeeded` with redacted `result_summary`
- `tool_call_failed` with redacted `error_code` and `error_message`

The MVP completes the run after a successful tool result and fails the run after a tool execution failure. Denial writes `tool_call_denied` and `run_stopped`.

## Current limitations

M7 itself does not add shell, filesystem, arbitrary network, MCP, browser automation, memory/RAG, desktop runtime, or multi-agent capabilities. Later milestones layer those tools onto the same approval projection and worker resume boundary.
