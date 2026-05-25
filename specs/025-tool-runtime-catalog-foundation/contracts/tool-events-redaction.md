# Contract: Tool Events and Redaction

M18 keeps existing event names:

- `tool_call_requested`
- `tool_call_approval_required`
- `tool_call_approved`
- `tool_call_denied`
- `tool_call_executing`
- `tool_call_succeeded`
- `tool_call_failed`

Tool event metadata may include:

- `tool_call_id`
- `tool_name`
- `tool_source`
- `tool_group`
- `risk_level`
- `approval_policy`
- `candidate_schema_hash`
- `arguments_summary`
- `result_summary`
- `error_code`

Metadata MUST NOT include secrets, raw tool args, raw tool results, MCP process command/env/stderr, provider trace, local credential paths, or continuation-only raw payloads.
