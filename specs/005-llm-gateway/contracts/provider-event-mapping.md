# Contract: Provider Event Mapping

M5 maps provider-native streaming output into Loomi run events before events are persisted or streamed to the frontend.

## Provider families

| Family | Native stream shape | M5 normalization |
|--------|---------------------|------------------|
| Anthropic | Message stream events with text deltas, message completion, errors, aborts, and tool-use content blocks | Text deltas become `model_output_delta`; final message becomes `model_output_completed`; errors/aborts become redacted provider failures; tool-use blocks become `tool_call_blocked`. |
| OpenAI | Responses or chat streaming events with output text deltas, done/completed/failed events, refusals, and function-call argument deltas | Text deltas become `model_output_delta`; completion becomes `model_output_completed`; failed/error/refusal states become redacted events; function calls become `tool_call_blocked`. |
| Gemini | `streamGenerateContent` chunks or OpenAI-compatible streaming endpoint chunks, including partial text and safety/blocked outcomes | Text chunks become `model_output_delta`; complete candidate text becomes `model_output_completed`; blocked/refused outcomes become `model_refusal`; failures become redacted provider events. |
| OpenAI-compatible custom | OpenAI-style chat streaming over configurable base URL, API key, and model | Parsed as the OpenAI-compatible subset needed for text delta, completion, refusal/error, and function-call boundary events. |

## Normalized event sequence

```text
run_created
run_started
model_request_started
model_output_delta...
model_output_completed
run_completed
```

Failure sequence:

```text
run_created
run_started
model_request_started
provider_error | provider_timeout | provider_rate_limited | model_refusal
run_failed
```

Stop sequence:

```text
run_created
...
run_stopped
```

After `run_stopped`, later provider output for that run is ignored.

## Normalized events

| Loomi Event Type | Category | Content | Metadata |
|------------------|----------|---------|----------|
| `model_request_started` | `progress` | null | provider family, model label |
| `model_output_delta` | `message` | delta text | provider family, model label, provider event id when safe |
| `model_output_completed` | `message` | final assistant text | provider family, model label, finish reason, safe token usage if available |
| `model_refusal` | `progress` | user-safe refusal summary | provider family, model label, safe finish reason |
| `tool_call_blocked` | `progress` | user-safe non-execution summary | provider family, model label, tool name when safe |
| `provider_error` | `error` | user-safe failure summary | stable provider family and redacted error code |
| `provider_timeout` | `error` | user-safe timeout summary | stable provider family and timeout class |
| `provider_rate_limited` | `error` | user-safe throttling summary | stable provider family and retry hint when safe |

## Safety rules

- Provider API keys, Authorization headers, raw request payloads, and raw provider error bodies are never written to run events.
- Provider output is treated as data, not instructions.
- Tool/function-call arguments are not executed by M5.
- Custom provider base URL may be displayed only as a redacted or credential-free value.
- Event ordering remains per-run monotonic and uses the M4 SSE history-first delivery contract.
