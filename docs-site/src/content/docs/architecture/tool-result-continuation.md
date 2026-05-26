---
title: Tool Result Continuation
description: Minimal model continuation boundary after an approved tool result.
---

Tool result continuation is the M7 slice after approval-gated tool execution. It lets Loomi take a redacted `tool_call_succeeded` result, feed it back to the configured model provider, and persist follow-up assistant output in the same run.

The M22 loop foundation starts from the same continuation boundary. Continuation may request another enabled `workspace.*` read tool in Work mode, but every new tool call is persisted, blocks on approval, and re-enters the worker only after approval. Non-workspace tools and Chat-mode continuation still use the original single-tool boundary.

## Context source

Continuation context is built from existing conversation messages plus the current run's persisted events:

1. Load thread messages through the user message that triggered the run.
2. Find the matching `tool_call_requested` event.
3. Find the matching `tool_call_succeeded` event.
4. Build one provider-neutral assistant tool-call item.
5. Build one provider-neutral tool-result item from the redacted result metadata.

No durable `messages.role = tool` row is written for this slice. Run events and the tool-call projection remain the audit source of truth.

## Provider boundary

Runtime code uses provider-neutral message roles:

| Role | Meaning |
| --- | --- |
| `assistant_tool_call` | The model's prior request for the allowlisted tool. |
| `tool_result` | The redacted result returned to the model. |

OpenAI-compatible providers serialize those items as an assistant message with `tool_calls` followed by a `tool` message with the same `tool_call_id`. Other provider families can adapt the same neutral shape later without changing product data.

## Event flow

The success path reuses the existing run-event stream:

```text
model_request_started       model_phase=initial
model_output_delta*         model_phase=initial
tool_call_requested
tool_call_approval_required
tool_call_approved
tool_call_executing
tool_call_succeeded
model_request_started       model_phase=continuation
model_output_delta*         model_phase=continuation
model_output_completed      model_phase=continuation
run_completed
```

History-first SSE remains unchanged. The second model phase is just another ordered set of persisted run events.

## Loop limit

Continuation is bounded to three accepted tool calls per run for the M22 foundation. If the provider asks for a fourth continuation tool, Loomi records `tool_loop_limit_reached` and fails the run without recording or executing the extra call.

Only enabled `workspace.glob`, `workspace.grep`, and `workspace.read` calls can use the bounded continuation path. If the continuation provider asks for `runtime.get_current_time`, MCP, an unknown tool, or any tool outside the run's enabled tool snapshot, Loomi records `unsupported_tool_loop` and fails the run. This keeps the first loop slice read-only, approval-gated, and limited to Work-mode workspace context.

Provider tool call ids are single-use inside a run. If continuation repeats an already-requested `tool_call_id`, Loomi records `duplicate_tool_call_id` and fails the run without duplicating approval events or reusing a terminal projection.

If the user stops a run while it is blocked on approval or after approval but before the resume worker executes the tool, the queued runner treats `stopped` as terminal. It exits without preparing another continuation, executing the approved tool, calling the provider, or writing a worker failure over the stopped run.

## Draft behavior

Frontend replay treats a successful tool event as a pause point for an existing assistant draft. When continuation deltas arrive with `model_phase = continuation`, the visible draft starts from the continuation answer and finalizes once. This prevents pre-tool text from becoming a duplicate final assistant message.

## Safety boundary

Only redacted result metadata is eligible for provider continuation. Tool result context must not include raw provider payloads, raw executor internals, credentials, file contents, shell output, arbitrary network responses, or hidden local state.

Approved `runtime.get_current_time` worker execution now calls continuation immediately after `tool_call_succeeded`. Denied and `tool_call_failed` runs are terminal and never re-enter the model. If the continuation provider fails, Loomi records one redacted failed terminal state and does not persist a final assistant message.
