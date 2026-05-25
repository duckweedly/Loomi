---
title: M12 MCP Approval-Gated Execution
description: Approval-gated execution boundary for already-discovered local stdio MCP tools.
---

M12 opens the smallest MCP execution loop after M11 discovery. Loomi can accept a provider request for an already-discovered local stdio MCP candidate, block it behind the existing M7 approval projection, execute one approved call through the worker, persist redacted tool events, and continue the provider once with a redacted result.

This is still not remote MCP, OAuth, marketplace/plugin install, shell/filesystem/browser automation, admin UI, DB-managed MCP servers, automatic execution, complex sandboxing, or a multi-tool loop.

## Entry Gate

MCP execution starts only when all of these are true:

- the tool name is namespaced as `mcp.<server_slug>.<tool_name>`
- a prior `mcp_discovery_succeeded` event lists that exact candidate via the production `MCPDiscoveryEventMetadata` projection
- the discovery metadata carries a schema hash for that candidate
- the active run pipeline resolved the selected persona allowed-tools and includes that exact name
- the provider request can create or reuse one M7 tool-call projection

If any check fails, the run fails safely before an approval action is offered.

## Approval Projection

M12 reuses the M7 `tool_calls` projection instead of adding a separate MCP permission table. MCP tool events add safe metadata:

- `tool_source: "mcp"`
- `server_slug`
- `candidate_schema_hash`
- namespaced `tool_name`
- redacted `arguments_summary`
- approval and execution statuses

Repeated provider requests for the same `(run_id, tool_call_id)` reuse the existing projection and do not duplicate approval events.

## Worker Execution

Approve queues the existing worker resume job. The queued run router checks the projection before execution:

- only `approved` + `not_started` can start
- `executing`, `succeeded`, `failed`, `denied`, or `cancelled` never re-execute
- execution is marked before invoking stdio
- denied or stopped runs never start a process

The API worker wires a `StdioMCPToolExecutor` into the real `QueuedRunRouter`; tests that inject a fake executor are not the only execution path. Local stdio server configs are loaded from `LOOMI_MCP_SERVERS_JSON` into the same config shape used by discovery.

The local stdio executor is bounded by config timeout, uses the same MCP `Content-Length` framing as M11 discovery, sends one `tools/call`, classifies timeout/exit/invalid response safely, and redacts result data before persistence.

## Continuation

A successful MCP result uses the existing M7 provider-neutral continuation. The continuation receives only the redacted result summary for the matching `tool_call_id`. If the continuation provider asks for another tool, Loomi records `unsupported_tool_loop` and executes no further tools.

## Replay

Frontend replay keeps M12 on existing event groups:

- `tool.call.*` events appear in Tool Call
- continuation `model_phase = continuation` events update the assistant draft
- failures and unsupported loops appear in Error

Older M7/M11 events without MCP metadata still replay without crashing.
