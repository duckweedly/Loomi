---
title: M7 Tool Call Approval Core Architecture
description: Approval-gated internal tool-call foundation, audit events, and safety boundaries for Loomi M7.
---

M7 starts the tool-calling slice without opening broad tool execution. The foundation records model-requested internal tool calls, blocks runs on approval, keeps a redacted current-state projection, and replays the lifecycle through persisted run events.

The implemented slice includes the infrastructure foundation plus US1-US4. US1 converts allowlisted provider `runtime.get_current_time` requests into durable approval-required tool calls. US2 records idempotent approve/deny decisions through scoped HTTP actions and frontend client controls. US3 queues the existing M6 worker after approval and executes the approved current-time tool through the M7 internal tool executor. US4 keeps tool lifecycle rows distinct from model stream rows across history-first replay and live continuation.

## Boundary

The product data layer owns durable tool-call state and run events. Runtime code owns the allowlisted `runtime.get_current_time` tool definition and executor. Frontend runtime adapters replay tool events into a stable `ToolCall` view model.

M7 reuses the M6 worker/job and M4 run/event/SSE foundations:

- runs can enter `blocked_on_tool_approval`
- history-first SSE continues to replay persisted events by sequence
- worker diagnostics can count approval-blocked and resumable tool calls
- approval creates one resumable worker job for the existing run
- RunRail/runtime grouping separates tool-call events from model stream and worker/job rows

## Durable projection

M7 adds a minimal `tool_calls` projection in migration `000006_m7_tool_call_approval`:

- one row per `(run_id, tool_call_id)`
- scoped to thread and run
- `tool_name`, redacted `arguments_summary`, and `arguments_hash`
- `approval_status` and `execution_status`
- optional redacted result/error summaries for later phases

Run events remain the audit and replay contract. The projection exists for idempotency, scoped lookup, worker resume, and concurrency guards.

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

## Current limitations

The US3 MVP finalizes the run after the approved tool succeeds, fails, or is cancelled. It does not yet continue a multi-step model loop with tool results. Stop/cancel has precedence over pending, approved, and executing tool calls; later worker attempts see the cancelled projection and avoid duplicate terminal tool events.
