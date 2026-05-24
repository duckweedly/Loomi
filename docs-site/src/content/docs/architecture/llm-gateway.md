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

## Provider normalization

M5 uses Go stdlib HTTP streaming for Anthropic, OpenAI, Gemini, and OpenAI-compatible providers. Provider-specific chunks are converted into internal provider events first, then into Loomi run events:

- text deltas become `model_output_delta`
- final text becomes `model_output_completed` plus one assistant message and `run_completed`
- refusals, timeouts, rate limits, generic errors, misconfiguration, and empty responses become redacted failure states
- tool/function-call output becomes `tool_call_blocked`

Assistant persistence is treated as part of completion. If the final assistant message cannot be written, the run fails with `assistant_message_persist_failed` instead of reporting a false completion.

## Safety boundary

M5 records tool-like provider output but does not execute tools, call external systems, or persist provider-supplied tool arguments. The event stream shows the boundary so future tool-use work can build on observable state without creating hidden side effects in this milestone.

## Deferred capabilities

Actual tool execution, worker/job queues, settings UI, RAG, memory, desktop runtime, and multi-agent behavior remain outside M5.
