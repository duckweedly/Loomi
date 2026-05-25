# Contract: Provider Tool Result Continuation

## Runtime-to-Gateway Request

The worker asks the gateway for one continuation after a successful approved tool result.

```json
{
  "thread_id": "thread_...",
  "run_id": "run_...",
  "model_phase": "continuation",
  "loop_count": 1,
  "messages": [
    { "role": "user", "content": "What time is it?" },
    {
      "role": "assistant_tool_call",
      "tool_call_id": "tc_1",
      "tool_name": "runtime.get_current_time",
      "arguments_summary": { "timezone": "UTC" }
    },
    {
      "role": "tool_result",
      "tool_call_id": "tc_1",
      "tool_name": "runtime.get_current_time",
      "content": {
        "iso_time": "2026-05-25T10:00:00Z",
        "timezone": "UTC",
        "source": "runtime"
      }
    }
  ]
}
```

Rules:

- `role` names are gateway-neutral and not necessarily durable message roles.
- `tool_result` must be derived from `tool_call_succeeded` and redacted before this request is built.
- The gateway must preserve the association between `assistant_tool_call.tool_call_id` and `tool_result.tool_call_id`.
- Provider adapters may serialize this as native provider messages, such as OpenAI-compatible assistant tool call plus tool role message.
- The request must not include raw provider payloads, raw executor internals, unvalidated arguments, secrets, file contents, shell output, arbitrary URL contents, or hidden local state.

Implementation note: `runtime.ProviderMessage` now uses the gateway-neutral roles `assistant_tool_call` and `tool_result`; the OpenAI-compatible adapter serializes them as native `tool_calls` and `tool` messages. The persistent Loomi message schema remains user/assistant only.

## Gateway Stream Output

Continuation output reuses existing model stream semantics:

```text
model_started    metadata.model_phase=continuation
model_delta      content="The current UTC time is ..."
model_final      content="The current UTC time is 2026-05-25T10:00:00Z."
model_usage?     metadata.model_phase=continuation
```

Rules:

- Continuation provider calls must be counted separately from the initial provider call for diagnostics.
- If the continuation provider emits another tool request, gateway/runtime must return an unsupported-loop result instead of executing another tool.
- A continuation provider error must be redacted and surfaced as model/provider failure.

## Unsupported Loop Contract

If the continuation response asks for any tool:

```json
{
  "type": "unsupported_tool_loop",
  "tool_call_id": "tc_2",
  "tool_name": "runtime.get_current_time",
  "message": "Additional tool calls are not supported in this run."
}
```

Rules:

- The second tool must not be recorded as approval-required for MVP.
- The second tool must not execute.
- The run must fail with a redacted, user-visible error event.
