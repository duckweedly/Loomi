# Contract: MCP Redaction and Events

## Purpose

Define safe metadata for persisted events, API responses, Timeline/debug, worker diagnostics, and provider continuation.

## Safe Event Fields

Allowed fields:

- `tool_call_id`
- `tool_name`
- `tool_source`
- `server_slug`
- `candidate_schema_hash`
- `approval_status`
- `execution_status`
- `arguments_summary`
- `arguments_hash`
- `result_summary`
- `result_hash`
- `result_for_model_redacted`
- `error_code`
- `safe_message`
- `retryable`
- `model_phase`
- timestamps

## Forbidden Fields

Never persist, replay, document as normal examples, or send to provider continuation:

- raw command paths
- raw args
- env values
- raw stdout
- raw stderr
- tokens
- credentials
- Authorization headers
- secret-looking paths
- private absolute paths
- file contents
- shell output
- browser state
- desktop captured data
- raw MCP result payloads

## Event Type Mapping

M12 reuses existing M7 backend event names and distinguishes MCP calls with safe metadata such as `tool_source = mcp`, `server_slug`, and the namespaced `tool_name`.

| Backend event | Frontend runtime type | Group |
| --- | --- | --- |
| `tool_call_approval_required` with `tool_source=mcp` | `tool.call.approval_required` plus MCP metadata | tool-call |
| `tool_call_approved` with `tool_source=mcp` | `tool.call.approved` plus MCP metadata | tool-call |
| `tool_call_denied` with `tool_source=mcp` | `tool.call.denied` plus MCP metadata | tool-call |
| `tool_call_executing` with `tool_source=mcp` | `tool.call.executing` plus MCP metadata | worker-job |
| `tool_call_succeeded` with `tool_source=mcp` | `tool.call.succeeded` plus MCP metadata | tool-call |
| `tool_call_failed` with `tool_source=mcp` | `tool.call.failed` plus MCP metadata | error |
| `tool_call_cancelled` with `tool_source=mcp` | `tool.call.cancelled` plus MCP metadata | error |
| `model_output_delta` with `model_phase=continuation` | existing continuation model-output mapping | assistant |
| `run_failed` with `error_code=unsupported_tool_loop` | existing safe failure mapping | error |

## Replay Rules

- Live SSE and history replay must produce the same ToolCall/Timeline state.
- Older M7/M11 events without MCP metadata must not crash replay.
- Redacted metadata must be sufficient to diagnose approval, worker, process, and continuation state without raw process output.
