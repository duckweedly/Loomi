# Contract: MCP Approval Gate

## Purpose

Define how a provider-requested MCP tool call becomes an M7 approval-blocked projection without executing the tool.

## Input

```json
{
  "thread_id": "thread_...",
  "run_id": "run_...",
  "provider_tool_call_id": "tc_1",
  "tool_name": "mcp.local-search.search",
  "arguments": {
    "query": "redacted before persistence"
  },
  "persona_snapshot": {
    "allowed_tools": ["mcp.local-search.search"]
  }
}
```

## Acceptance Rules

- `tool_name` must match a discovered M11 namespaced MCP candidate.
- The candidate server must be enabled and locally configured for stdio.
- The selected persona snapshot must include the same namespaced tool.
- Arguments must pass candidate schema validation and be redacted before persistence.
- Projection identity is `(run_id, provider_tool_call_id)`.

## Approval-Required Projection

```json
{
  "tool_call_id": "tc_1",
  "tool_name": "mcp.local-search.search",
  "tool_source": "mcp",
  "server_slug": "local-search",
  "candidate_schema_hash": "sha256:...",
  "arguments_summary": {
    "query": "[redacted-summary]"
  },
  "arguments_hash": "sha256:...",
  "approval_status": "required",
  "execution_status": "blocked"
}
```

## Rejection Output

```json
{
  "type": "run_failed",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_name": "mcp.unknown.search",
    "tool_source": "mcp",
    "error_code": "mcp_tool_not_allowed",
    "safe_message": "MCP tool is not available for this run."
  }
}
```

## Idempotency

Repeated provider requests for the same `(run_id, provider_tool_call_id)` return the existing projection and do not duplicate events.

## Forbidden Behavior

- No MCP stdio process startup before approval.
- No approval for undiscovered, disabled, unnamespaced, or persona-disallowed tools.
- No raw arguments, env, command path, stdout/stderr, tokens, credentials, secret-looking paths, file contents, shell output, browser state, or desktop captured data in persisted metadata.
