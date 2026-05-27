---
title: Tool Result Continuation
description: Minimal model continuation boundary after an approved tool result.
---

Tool result continuation is the M7 slice after approval-gated tool execution. It lets Loomi take a redacted `tool_call_succeeded` result, feed it back to the configured model provider, and persist follow-up assistant output in the same run.

The M22 loop foundation starts from the same continuation boundary. Continuation may request another enabled Work-mode tool after each terminal tool result, but every new tool call is persisted, blocks on approval, and re-enters the worker only after approval. Chat-mode continuation still uses the original narrower boundary.

## Context source

Continuation context is built from existing conversation messages plus the current run's persisted events:

1. Load thread messages through the user message that triggered the run.
2. Find the matching `tool_call_requested` event.
3. Find the matching `tool_call_succeeded` event.
4. Build one provider-neutral assistant tool-call item.
5. Build one provider-neutral tool-result item from the redacted result metadata.

No durable `messages.role = tool` row is written for this slice. Run events and the tool-call projection remain the audit source of truth.

On worker restart, the queued runner may re-enter with the same approved `tool_call_id`. If the tool projection already says `execution_status = succeeded`, Loomi does not execute the tool again. It scans the durable run events after that tool's `tool_call_succeeded` event:

- if no continuation `model_request_started`, later `tool_call_requested`, or final run event exists, it resumes provider continuation from the persisted tool result;
- if continuation already started, a later tool was requested, or the run is terminal, it treats the retry as already handled.

This keeps the provider input rebuild durable without mixing an unfinished `tool_call_requested` into continuation and without duplicating final assistant output after a retry.

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

## Large result compaction

Provider continuation uses the redacted result payload as its source, then compacts oversized string fields before serializing the provider `tool_result` message. Compaction is deterministic and signal-preserving: it keeps early context, path/status/error-like lines, tail context, and a `[tool output compacted]` marker. Small results are not modified.

M92 tightens the compaction boundary so ordinary summaries remain readable even when the source text contains terminal labels such as `stdout` or `tool output`. Sensitive lines are still replaced with `[redacted]`, but benign `summary`, `status`, `path`, and excerpt lines stay available to continuation. This prevents the model from seeing only `[redacted]` after a successful tool chain and then producing an empty or unreadable final answer.

This compaction affects only provider input. Persisted run events remain the audit source of truth and continue to store safe summaries only.

## Loop limit

Continuation is bounded to six accepted tool calls per run. If the provider asks for a seventh continuation tool, Loomi records `tool_loop_limit_reached` and fails the run without recording or executing the extra call.

Enabled workspace, bounded command, LSP, web, browser, artifact, and coordination tools can use the bounded continuation path. If the continuation provider asks for `runtime.get_current_time`, MCP, an unknown tool, or any tool outside the run's enabled tool snapshot, Loomi records `unsupported_tool_loop` and fails the run.

Provider tool schemas are generated from the run's enabled builtin tool snapshot. Provider-facing function names use safe identifiers such as `workspace_read`, `workspace_edit`, `sandbox_exec_command`, `sandbox_start_process`, and `lsp_symbols`; inbound provider tool calls are mapped back to Loomi's internal dotted tool names before approval and execution.

Provider tool call ids are single-use inside a run. If continuation repeats an already-requested `tool_call_id`, Loomi records `duplicate_tool_call_id` and fails the run without duplicating approval events or reusing a terminal projection.

If the user stops a run while it is blocked on approval or after approval but before the resume worker executes the tool, the queued runner treats `stopped` as terminal. It exits without preparing another continuation, executing the approved tool, calling the provider, or writing a worker failure over the stopped run.

## Draft behavior

Frontend replay treats a successful tool event as a pause point for an existing assistant draft. When continuation deltas arrive with `model_phase = continuation`, the visible draft starts from the continuation answer and finalizes once. This prevents pre-tool text from becoming a duplicate final assistant message.

## Safety boundary

Only redacted result metadata is eligible for provider continuation. Tool result context must not include raw provider payloads, raw executor internals, credentials, unredacted file contents, arbitrary network responses, or hidden local state. Oversized redacted text is compacted before it reaches the provider so the next assistant answer stays focused on the actionable signal.

Approved `runtime.get_current_time` worker execution now calls continuation immediately after `tool_call_succeeded`. Denied and `tool_call_failed` runs are terminal and never re-enter the model. If the continuation provider fails, Loomi records one redacted failed terminal state and does not persist a final assistant message.

M92 also makes terminal state a hard write boundary for the agent loop. Once a run is `completed`, `failed`, `stopped`, or `cancelled`, late model/tool events must be rejected by product data or ignored by the frontend replay adapter. A retry can resume only from a durable succeeded tool result that has no later continuation, later tool request, or final event.

## Tool choice guidance

The provider prompt now includes an explicit Work-mode strategy:

- Directory questions start with a broad workspace listing from the selected root.
- Content questions use grep/read after the relative path is known.
- Modification questions read first, then preview a patch, then apply only after approval.
- Shell/process tools are reserved for explicit shell requests or validation such as build, test, and lint.

`tool.load_tools` remains a query-only catalog lookup. The provider schema advertises `query`/`queries` and does not expose `names`; omitting the query lists a bounded safe catalog instead of failing validation.
