---
title: M12 MCP Approval-Gated Execution API and Events
description: Safe event metadata for executing one approved local stdio MCP tool.
---

M12 does not add public MCP server management endpoints. It reuses M7 scoped tool-call read, approve, and deny endpoints for MCP calls:

- `GET /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}`
- `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/approve`
- `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/deny`

## Approval-Required Event

```json
{
  "type": "tool_call_approval_required",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_mcp_1",
    "tool_name": "mcp.local-search.search",
    "tool_source": "mcp",
    "server_slug": "local-search",
    "candidate_schema_hash": "sha256:...",
    "arguments_summary": { "query": "status" },
    "approval_status": "required",
    "execution_status": "blocked"
  }
}
```

MCP approval is offered only when the candidate was discovered, the discovery event includes a candidate schema hash, and the persona allowed-tools snapshot references the same namespaced tool.

## Execution Events

Approved MCP execution reuses M7 event names:

| Backend type | Meaning |
| --- | --- |
| `tool_call_executing` | Worker marked the approved MCP call as started before stdio invocation. |
| `tool_call_succeeded` | Stdio call completed and produced a redacted result summary. |
| `tool_call_failed` | Timeout, exit, invalid response, missing config, or unsafe response failed safely. |
| `tool_call_denied` | User denied before execution. |

Success metadata is safe for replay and continuation:

```json
{
  "type": "tool_call_succeeded",
  "metadata": {
    "tool_call_id": "tc_mcp_1",
    "tool_name": "mcp.local-search.search",
    "tool_source": "mcp",
    "server_slug": "local-search",
    "candidate_schema_hash": "sha256:...",
    "execution_status": "succeeded",
    "result_summary": {
      "summary": "Redacted MCP result"
    }
  }
}
```

## Continuation

Provider continuation uses the matching `tool_call_requested` plus `tool_call_succeeded` events. Raw MCP output is never used. If `result_for_model_redacted` is present, it wins; otherwise the safe `result_summary` is serialized as the tool result.

If the continuation asks for another tool, the run fails with `unsupported_tool_loop` and no tool executes.

## Redaction

These fields must not appear in run events, API responses, Timeline/debug replay, docs examples, or provider continuation:

- raw command paths
- raw args
- env values
- raw stdout or stderr
- tokens, credentials, Authorization headers
- secret-looking paths
- file contents
- shell output
- browser or desktop captured state
- raw MCP result payloads
