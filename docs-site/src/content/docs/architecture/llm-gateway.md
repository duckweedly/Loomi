---
title: M5 LLM Gateway Architecture
description: Backend gateway boundary for Anthropic, OpenAI, Gemini, and OpenAI-compatible custom providers.
---

M5 adds a backend model gateway on top of the existing thread/message and run/event/SSE foundations. The frontend still consumes Loomi events instead of provider-native streams.

## Boundary

The gateway owns provider configuration, request context construction, provider stream consumption, redaction, and conversion into Loomi run events. Provider API keys are never sent to the browser or persisted in user-visible events.

`cmd/loomi-api` builds HTTP providers from local configuration and injects them into the API server. `POST /v1/threads/{thread_id}/runs` routes `source: "model_gateway"` runs to the gateway runner; local simulated runs continue to use the M4 runner.

## Request context

A model-gateway run starts from an existing durable user `message_id`. The gateway loads messages from the selected thread in creation order, includes assistant history where present, and stops at the trigger message. If the trigger message is missing or outside the thread, the run fails with `invalid_request`.

The gateway background worker does not inherit the HTTP request context, so a normal response completion does not cancel the provider stream.

Before the provider request is serialized, the gateway applies a deterministic provider-context window. Small runs are left untouched. Large runs are compacted to a bounded recent window plus one user-role `<conversation_summary>` message, with per-message truncation and tool-call/tool-result pair preservation. Tool history compaction keeps paired tool calls/results together while preserving later normal assistant/user messages, including the latest user trigger. The durable thread messages and run events remain unchanged; only the provider input is compacted. A `context_compacted` progress event records counts, byte budgets, preserved tool pairs, and strategy metadata without raw message or tool-result payloads.

## Provider normalization

M5 uses Go stdlib HTTP streaming for Anthropic, OpenAI, Gemini, and OpenAI-compatible providers. Provider-specific chunks are converted into internal provider events first, then into Loomi run events:

- text deltas become `model_output_delta`
- final text becomes `model_output_completed` plus one assistant message and `run_completed`
- refusals, timeouts, rate limits, generic errors, misconfiguration, and empty responses become redacted failure states
- tool/function-call output becomes tool-call requests with stable ids and redacted arguments; OpenAI-compatible, Anthropic, and Gemini requests all include enabled tool schemas and their streams preserve multiple tool calls from one model turn, including Anthropic `input_json_delta` argument chunks

Assistant persistence is treated as part of completion. If the final assistant message cannot be written, the run fails with `assistant_message_persist_failed` instead of reporting a false completion.

Current gateway runs retry transient provider failures with bounded backoff only before durable output exists. Rate limits, HTTP 408/504 timeouts, retryable network errors, retryable stream errors, and empty attempts can be retried up to three attempts. Retry scheduling is recorded as redacted `model_request_retry_scheduled` telemetry. If retries exhaust, the final failure keeps the provider route, model phase, attempt, retryable flag, and safe provider error metadata for debugging without exposing keys or raw provider traces. Misconfiguration, refusal, validation errors, user stop, provider-incomplete output, and any failure after text/tool/final state has been written are not replayed, preventing duplicate assistant messages or duplicate tool calls. Provider streams also stop reading after terminal tool-call/completion frames; trailing transport noise after a complete tool-call turn is ignored. OpenAI-compatible `finish_reason=length`, Anthropic `max_tokens`, Gemini `MAX_TOKENS`, and Local Codex Responses `response.incomplete` are treated as incomplete output, not successful final answers.

Run context prompt construction can include safe registered context sources. URL, GitHub, and note-like sources are summarized in a bounded `<context_sources>` block after redaction; workspace-path sources are not injected into the prompt because they may represent local or sensitive paths.

Prepared run context is reused at the gateway boundary where it is fresh: initial provider tool schemas and initial scoped tool-call validation use `EnabledTools` directly instead of replaying the full event stream. Initial model-gateway jobs and approved-tool resume jobs can rebuild their `RunContext` from the materialized run-step projection when route metadata is present, including MCP tool schema hashes and safe MCP availability summaries captured during discovery. This keeps MCP visibility and diagnostics intact even when prepare-context does not replay `run_events`. Continuation carries the same system prompt policy back into the provider request so persona, thread mode, memory/notebook snapshots, workspace policy, and safety boundaries do not disappear after the first tool result. Continuation reads the same projection for completed tool results, enabled tools, loop counts, workspace guardrail facts, and MCP schema validation. If that projection cannot be read after the product-data boundary has had a chance to catch it up from `run_events`, continuation fails explicitly with context unavailable instead of silently replaying the whole event stream on the gateway hot path.

The direct asynchronous gateway entry remains request-context independent so HTTP cancellation does not kill a queued model run, but it now watches the durable run stop state and cancels the provider stream when the run is stopped.

## Safety boundary

M5 records tool-like provider output but does not execute tools, call external systems, or persist provider-supplied tool arguments. The event stream shows the boundary so future tool-use work can build on observable state without creating hidden side effects in this milestone.

## Deferred capabilities

Actual tool execution, worker/job queues, settings UI, RAG, memory, desktop runtime, and multi-agent behavior remain outside M5.
