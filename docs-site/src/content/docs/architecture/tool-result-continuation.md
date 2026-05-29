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

On worker restart, the queued runner may re-enter with the same approved `tool_call_id`. If the tool projection already says `execution_status = succeeded`, Loomi does not execute the tool again. It checks the durable run-step state after that tool's `tool_call_succeeded` event:

- if no continuation output, later `tool_call_requested`, or final run event exists, it may resume provider continuation from the persisted tool result;
- if continuation output exists, a later tool was requested, or the run is terminal, it treats the retry as already handled.

This keeps the provider input rebuild durable without mixing an unfinished `tool_call_requested` into continuation and without duplicating final assistant output after a retry.

M97 adds a Loomi-native run step ledger projection over the same durable event stream. It does not introduce a new queue or rename run events. Instead, `BuildRunStepLedger` and `RebuildRunStepState` classify persisted events into safe steps:

- `model_request`
- `tool_requested`
- `approval`
- `tool_execution`
- `continuation`
- `terminal`

The rebuilt state separates completed tool results from pending tool calls and derives the next action as `start_model`, `wait_for_tool_approval`, `execute_tool`, `continue_model`, `terminal`, or `none`. The queued runner uses this projection when deciding whether a retry should resume missing provider continuation after a succeeded tool result.

M98 materializes that projection as a checkpoint. `run_step_state_projections` stores `run_id`, `thread_id`, `user_id`, `last_sequence`, and a redacted `RunStepState` JSON snapshot. `run_events` remains the audit source of truth: missing, stale, or unreadable projections are rebuilt or caught up from run events. Hot queued-runner and gateway decisions such as the next approved tool, approved-tool resume context, post-success continuation resume, continuation route hydration, continuation provider messages, and continuation tool schemas read the projection instead of replaying the whole event stream every loop.

Projection writes are monotonic by `last_sequence`. A catch-up or rebuild path may advance the checkpoint, but it must not overwrite a newer checkpoint with an older event sequence if another writer already moved the run forward.

The projection tracks continuation request start separately from continuation output. A recovered worker may resume from a completed tool result if only `model_request_started` was recorded for the continuation; once continuation output, a later tool request, or a terminal run event exists, the retry is treated as already handled.

M98 adds an atomic continuation claim before the queued runner calls the provider. The claim writes `model_request_started` with `model_phase=continuation`, the completed `tool_call_id`, the current `job_id`, and a short `claim_expires_at` lease. The same job cannot claim the same completed tool result twice, and a different job cannot claim the same frontier while an earlier claim lease is still active. If the worker dies after claim but before output, a later recovery job with a different `job_id` can claim after the lease expires and continue from the same persisted tool result.

The PostgreSQL claim path reads the materialized run-step projection under the run transaction, catches up only events after the projection cursor, and materializes the caught-up state before appending the claim event. Missing or unreadable projections can still rebuild from `run_events`, but the hot continuation path no longer scans the full event stream for every claim.

Tool lifecycle events also derive loop metadata from the run-step projection. `tool_call_requested`, approval, execution, success/failure, recovery, and delegated-agent completion rows use `RunStepState.SeenToolCallIDs` to keep `loop_index` stable instead of rescanning the full event stream for each lifecycle write.

M98 closes the parallel tool-call gap while keeping the event stream as source of truth. If a provider emits multiple auto-approved read-only tool calls in one model turn, Loomi records all of them, executes the ready batch concurrently, and waits to call provider continuation until the batch has completed. The same drain rule applies after a continuation response asks for another parallel read-only batch. The continuation context then includes each assistant tool-call item followed by its matching tool-result item in original request order, even when concurrent tool executions finish in a different order.

Mixed batches do not produce partial continuation. If the same model turn contains a completed auto-approved tool and another pending approval-required call, the run remains blocked on approval; the queued runner only resumes the model after all pending tool calls from that turn are terminal or resolved. Manual-approval, write-capable, MCP, memory-write, sandbox, artifact mutation, and other gated tools remain serialized by approval state; they are not batch-executed.

Repeated workspace read/list/grep/glob/tree-summary arguments are rejected before the buffered continuation batch is recorded. The repeat key is `tool_name + arguments_hash`, so different workspace tools with similar argument shapes do not collide, while exact same-tool repeats in one provider turn are rejected. This keeps a single provider turn from queueing duplicate same-argument workspace reads that the historical repeat guard would only see after the first buffered call had already been persisted.

The worker also avoids one redundant full replay while publishing run history around lease renewal. `publishRunEvents` now returns the highest sequence it actually published, so the worker can request only events after that cursor instead of immediately calling `ListRunEvents(..., 0)` again.

## Provider boundary

Runtime code uses provider-neutral message roles:

| Role | Meaning |
| --- | --- |
| `assistant_tool_call` | The model's prior request for the allowlisted tool. |
| `tool_result` | The redacted result returned to the model. |

OpenAI-compatible and Local Codex providers serialize those items as an assistant message with `tool_calls` followed by `tool` messages with matching `tool_call_id`s. Anthropic serializes the same neutral items as `tool_use` assistant blocks followed by `tool_result` user blocks. Gemini serializes them as `functionCall` model parts followed by `functionResponse` user parts. Product data keeps one provider-neutral continuation shape; adapters are responsible for preserving the original tool-call/result order for their upstream API.

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

When the first model phase requests a parallel auto-approved batch, the event stream contains multiple `tool_call_requested` / `tool_call_approved` pairs before execution. Continuation starts only after those calls have terminal tool results, so the model does not see a partial batch.

Step metadata is also written into known lifecycle/tool/model events as safe metadata:

```json
{
  "run_step_kind": "tool_execution",
  "run_step_status": "succeeded",
  "run_step_summary": "Tool call succeeded"
}
```

Timeline and debug surfaces can read these fields without exposing raw arguments, tool output, provider internals, local paths, or secrets.

## Large result compaction

Provider continuation uses the redacted result payload as its source, then compacts oversized string fields before serializing the provider `tool_result` message. Compaction is deterministic and signal-preserving: it keeps early context, path/status/error-like lines, tail context, and a `[tool output compacted]` marker. Small results are not modified.

M92 tightens the compaction boundary so ordinary summaries remain readable even when the source text contains terminal labels such as `stdout` or `tool output`. Sensitive lines are still replaced with `[redacted]`, but benign `summary`, `status`, `path`, and excerpt lines stay available to continuation. This prevents the model from seeing only `[redacted]` after a successful tool chain and then producing an empty or unreadable final answer.

M98 also enforces the budget after the full redacted payload is serialized. Large aggregate structures such as long grep match arrays first receive normal per-field compaction; if the encoded JSON still exceeds the provider budget, Loomi replaces the payload with a compact summary that includes `truncated`, `omitted_bytes`, and a bounded excerpt. This keeps continuation context from ballooning even when every individual match entry is small.

This compaction affects only provider input. Persisted run events remain the audit source of truth and continue to store safe summaries only.

Provider context compaction also bounds assistant tool-call argument summaries, not only message content and tool-result payloads. If a recent tool call contains large arguments such as a patch body or generated content, Loomi replaces the oversized argument map with compact metadata and omitted-key names while preserving the matching assistant tool-call/tool-result pair. This avoids the worse fallback of dropping the latest tool pair to satisfy the context budget.

## Loop limit

Continuation is bounded to six accepted tool calls per run. If the provider asks for a seventh continuation tool, Loomi records `tool_loop_limit_reached` and fails the run without recording or executing the extra call.

Continuation tool requests are turn-atomic against that bound. If a single continuation response contains more tool calls than the remaining run budget, Loomi rejects the turn before recording any of those new tool calls, so no approved-but-unexecutable dangling tool call is left behind.

Enabled runtime, workspace, bounded command, LSP, web, browser, artifact, coordination, memory, notebook, and todo tools can use the bounded continuation path when they are present in the run's enabled tool snapshot. If the continuation provider asks for MCP, an unknown tool, or any tool outside that snapshot, Loomi records `unsupported_tool_loop` and fails the run.

Provider tool schemas are generated from the run's enabled builtin tool snapshot. Provider-facing function names use safe identifiers such as `workspace_read`, `workspace_edit`, `sandbox_exec_command`, `sandbox_start_process`, and `lsp_symbols`; inbound provider tool calls are mapped back to Loomi's internal dotted tool names before approval and execution. Generated Gemini stream ids are unique across the full stream, not just within one SSE frame, so multi-frame function calls do not collide.

Provider tool call ids are single-use inside a run. If continuation repeats an already-requested `tool_call_id`, Loomi records `duplicate_tool_call_id` and fails the run without duplicating approval events or reusing a terminal projection.

## Runtime tool guard

M95 adds a lightweight runtime guard before a provider-requested workspace tool is recorded. This is not an AI planner and does not reorder tools. It only checks the current run events, the latest user message, and the requested tool arguments:

- Work-mode directory inventory/classification intent must start with `workspace.tree_summary` or `workspace.list_directory`.
- `workspace.grep` is rejected as the first workspace tool for directory inventory because grep is for content search.
- `workspace.glob` remains available for file-name patterns or narrow follow-up after directory inventory, but it is not accepted as the first directory inventory tool.
- Repeating the same `workspace.read`, `workspace.list_directory`, or `workspace.grep` arguments in one run records `tool_planner_guardrail` and fails safely instead of looping.

The guard emits a terminal failure with safe metadata such as `guardrail`, `tool_name`, `arguments_summary`, and `recommended_tool`. It does not create a tool-call projection for the rejected request.

If the user stops a run while it is blocked on approval or after approval but before the resume worker executes the tool, the queued runner treats `stopped` as terminal. It exits without preparing another continuation, executing the approved tool, calling the provider, or writing a worker failure over the stopped run.

## Draft behavior

Frontend replay treats a successful tool event as a pause point for an existing assistant draft. When continuation deltas arrive with `model_phase = continuation`, the visible draft starts from the continuation answer and finalizes once. This prevents pre-tool text from becoming a duplicate final assistant message.

Provider completion text is normalized before the final assistant message is persisted. If the provider returns a raw structured payload such as JSON with an `answer`, `final`, `message`, `summary`, `content`, or `text` field, Loomi persists that field as the natural-language final answer. Raw `tool_calls` protocol payloads fall back to a readable failure-style summary instead of becoming visible assistant JSON.

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
