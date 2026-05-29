---
title: M5 LLM Gateway API
description: Local API surface for model provider capability, model-gateway runs, and provider-normalized run events.
---

M5 extends the M4 run/event/SSE API with `model_gateway` runs and local provider capability checks. Provider credentials stay in backend configuration; frontend clients only receive redacted capability and Loomi-normalized run events.

## Provider capability

### `GET /v1/model-providers`

Returns configured provider capabilities:

```json
{
  "providers": [
    {
      "id": "custom",
      "family": "openai_compatible",
      "base_url": "https://gateway.example.test",
      "model": "gpt-5.5",
      "status": "available"
    }
  ],
  "request_id": "req_..."
}
```

`status` is one of `available`, `unavailable`, or `misconfigured`. Custom provider URLs are redacted to scheme and host; userinfo, path, query, and fragment are not exposed.

Local Claude Code and Local Codex autodetect results are not returned here because they are not configured model-gateway providers. Use `GET /v1/local-provider-detections` for detection-only local provider candidates.

### `POST /v1/model-providers/check`

Request:

```json
{ "provider_id": "custom" }
```

Returns the selected provider capability when available. Disabled providers return `provider_unavailable` with HTTP 503. Incomplete or invalid providers return `provider_misconfigured` with HTTP 400.

## Model-gateway run creation

`POST /v1/threads/{thread_id}/runs` accepts the M5 fields below:

```json
{
  "message_id": "msg_...",
  "source": "model_gateway",
  "provider_id": "custom",
  "model": "optional-model-override"
}
```

The gateway loads current-thread messages through `message_id`, starts a backend provider stream, and persists Loomi run events. Existing local simulation remains available by omitting `source` or using the M4 `script_name` path.

## Normalized event types

M5 provider output is converted into the existing run-event envelope:

| Category | Type | Meaning |
| --- | --- | --- |
| `progress` | `model_request_started` | Gateway selected a provider and started request context construction. |
| `progress` | `context_compacted` | Gateway compacted oversized provider input before serialization; metadata contains counts, byte budgets, preserved/dropped tool-pair counts, and strategy only. |
| `progress` | `model_request_retry_scheduled` | A retryable provider failure happened before visible output or tool requests; another attempt is scheduled with bounded backoff. |
| `message` | `model_output_delta` | Incremental assistant text from provider streaming. |
| `message` | `model_output_completed` | Final assistant text is ready and has been persisted as conversation history. |
| `progress` | `tool_call_blocked` | Provider requested tool/function use; M5 records the boundary and does not execute it. |
| `error` | `model_refusal` | Provider refused or blocked the response. |
| `error` | `provider_timeout` | Provider request timed out. |
| `error` | `provider_rate_limited` | Provider rate limit was reached. |
| `error` | `provider_error` | Provider request or stream failed with a redacted generic error. |
| `error` | `empty_response` | Provider completed without usable assistant text. |
| `final` | `run_completed`, `run_failed`, `run_stopped` | Terminal run outcome. |

SSE continues to use `run_event` frames and `after_sequence` history replay from M4.

Gateway provider retry is intentionally narrow. `provider_rate_limited`, `provider_timeout`, retryable `provider_error`, and empty attempts are retried up to three total attempts only when the attempt has not emitted visible text, a tool request, or a final assistant message. The HTTP provider maps 429 to `provider_rate_limited`, 408/504 to `provider_timeout`, and pre-output transport errors such as temporary connection resets to retryable provider failures. Once output or tool state is durable, the failure is recorded instead of replaying the provider request. OpenAI-compatible `finish_reason=length`, Anthropic `max_tokens`, Gemini `MAX_TOKENS`, and Local Codex Responses `response.incomplete` are recorded as `provider_incomplete`; Loomi does not persist truncated output as a final assistant answer.

Provider context compaction is an input-shaping event, not a mutation of conversation history. When emitted, `context_compacted.metadata` may include `model_phase`, `budget_bytes`, `message_budget_bytes`, `recent_message_window`, `original_bytes`, `compacted_bytes`, `original_count`, `compacted_count`, `omitted_messages`, `truncated_fields`, `preserved_tool_pairs`, `dropped_tool_pairs`, `strategy`, and `redaction_applied`. It must not include raw message content, tool arguments, tool output, provider payloads, file contents, or secret-bearing paths.

Frontend clients treat live token deltas as an interim draft only. `model_output_completed` promotes the draft immediately, and the terminal refresh reconciles the run against the persisted assistant message whose `run_id` matches the run. If those two sources disagree, the persisted assistant message wins because it is the conversation history source of truth. This keeps Markdown rendering stable even when replayed events contain partial token fragments or redacted debug summaries.

## Redaction rules

Run events and API responses must not include provider API keys, Authorization headers, raw provider request payloads, raw provider error bodies, or sensitive custom URL path/query fragments. Safe metadata can include provider id, provider family, selected model, and non-sensitive tool names.
