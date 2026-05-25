---
title: Local M12 MCP Approval Execution
description: Local validation checklist for approval-gated MCP execution.
---

## Scope

This runbook validates the M12 minimal loop:

- provider-requested namespaced MCP tool is blocked behind M7 approval
- undiscovered or persona-disallowed MCP tools are rejected before approval
- approved local stdio MCP execution runs once through the worker
- retry/recovery does not duplicate execution after start
- redacted result is used for one provider continuation
- frontend replay preserves MCP tool-call metadata

## Local MCP Config

The real API worker loads local stdio MCP execution configs from `LOOMI_MCP_SERVERS_JSON`. The value is a JSON array using the same fields as M11 discovery:

```json
[
  {
    "slug": "local-search",
    "display_name": "Local Search",
    "enabled": true,
    "transport": "stdio",
    "command": "local-mcp-server",
    "args": ["--stdio"],
    "env": {},
    "timeout_ms": 5000
  }
]
```

Do not paste real tokens or private paths into validation logs or docs examples.

## Commands

```bash
go test ./internal/httpapi -run TestM12RealLocalMCPApprovalSmoke
go test ./internal/productdata ./internal/runtime ./internal/httpapi
bun test --cwd web src/runtime/realExecutionAdapter.test.ts src/runtime/runtimeEventGroups.test.ts
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Smoke Expectations

### M12.5 Real Local Smoke

`TestM12RealLocalMCPApprovalSmoke` is the closeout evidence test. It uses a real local stdio fixture process and MCP `Content-Length` frames for both discovery and execution:

1. The fixture responds to discovery `tools/list`.
2. `mcp_discovery_succeeded` records the namespaced candidate and `candidate_schema_hashes`.
3. The default persona snapshot allows the discovered MCP tool.
4. The provider requests `mcp.local-smoke.echo`.
5. Loomi records `tool_call_approval_required` and no `tools/call` has run yet.
6. The scoped HTTP approve endpoint records `tool_call_approved`.
7. The worker uses `StdioMCPToolExecutor` loaded from `LOOMI_MCP_SERVERS_JSON`.
8. The fixture receives exactly one `tools/call`.
9. Loomi records `tool_call_executing`, redacted `tool_call_succeeded`, continuation delta, final completion, and one assistant message.
10. The smoke fails if fixture secrets or private paths appear in persisted events or continuation content.

### General Expectations

1. A model request for `mcp.local-search.search` creates `tool_call_approval_required`.
2. The event metadata includes `tool_source: "mcp"`, `server_slug: "local-search"`, and `candidate_schema_hash`.
3. Deny records `tool_call_denied` and never starts stdio.
4. Approve queues the existing worker path.
5. Worker marks `tool_call_executing` before stdio invocation.
6. The stdio fixture uses MCP `Content-Length` frames for both discovery and execution.
7. Success records `tool_call_succeeded` with only redacted result data.
8. Provider continuation runs once with `model_phase = "continuation"`.
9. A second tool request from continuation fails with `unsupported_tool_loop`.

## Safety Checks

Stop and fix redaction if any of these appear in events, UI replay, provider continuation, or docs examples:

- env values
- raw args
- command paths
- stdout/stderr
- tokens or credentials
- private paths
- file contents
- shell output
- browser or desktop captured data

## Out of Scope

M12 does not validate remote MCP, MCP HTTP/SSE/OAuth, marketplace/plugin install, DB-managed MCP server admin, shell/filesystem/browser automation, automatic execution, complex sandboxing, admin UI, or multi-step tool loops.

## Browser Smoke Status

M12.5 backend/httpapi/runtime smoke covers the same state sequence required for UI verification. Browser smoke should be run only when a live local API, database, deterministic provider fixture, and web dev server are available together. In this closeout session, no long-running browser stack was started; the evidence is the in-process HTTP approve plus real worker/stdin/stdout smoke above.
