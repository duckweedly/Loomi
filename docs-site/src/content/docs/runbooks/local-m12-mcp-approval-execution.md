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
go test ./internal/productdata ./internal/runtime ./internal/httpapi
bun test --cwd web src/runtime/realExecutionAdapter.test.ts src/runtime/runtimeEventGroups.test.ts
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Smoke Expectations

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
