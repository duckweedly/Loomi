---
title: Local M11 MCP Validation
description: Local validation checklist for MCP stdio foundation.
---

## Scope

This runbook validates the M11 minimal slice:

- local explicit stdio config validation
- discovery/list-tools parsing
- namespaced read-only ToolSpec candidates
- persona references to non-executable MCP tools
- RunContext MCP availability summary
- Timeline/debug labels for MCP discovery status
- no MCP tool execution

## Commands

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/RunTimeline.runtime.test.ts ./web/src/components/RunRail.runtime.test.ts
bun run --cwd docs-site build
git diff --check
```

## Real Local Stdio Smoke

`internal/runtime/mcp_discovery_test.go` includes `TestDiscoverMCPToolsRunsLocalStdioListToolsSmokeWithoutLeaks`, a repeatable local stdio fixture smoke.

The smoke launches the Go test binary itself as an enabled local stdio MCP server fixture. The fixture reads Loomi's MCP stdin frames, fails if it receives `tools/call`, and responds only to `tools/list` with one `echo` tool. The config intentionally includes sensitive-looking args, env, and fixture stderr; the assertion verifies the RunContext safe summary contains only safe MCP availability fields such as `mcp.local-smoke.echo`, disabled execution state, and discovery counts.

Expected evidence:

- enabled local stdio config can discover/list-tools through the real stdio runner
- no MCP tool execution request is sent
- discovered tool is projected as `mcp.local-smoke.echo`
- execution remains disabled/non-executable
- safe summary does not include env values, args, stderr text, token-like values, or private paths

## Backend Checks

Expected:

- `stdio` local configs validate.
- HTTP/SSE/OAuth/remote config is rejected.
- Discovery parser maps valid list-tools output into namespaced candidates.
- Duplicate, invalid, or unsupported tool schema fails safely.
- Env, args, raw stderr, tokens, credentials, and secret-looking paths are redacted.
- `mcp.<server_slug>.<tool_name>` persona references resolve as non-executable.
- RunContext safe summary includes MCP candidate counts and disabled execution state.

## Browser/Debug Smoke

1. Start local API/worker and web in real API mode.
2. Use a local explicit MCP fixture or mocked discovery result.
3. Trigger discovery/list-tools.
4. Create a run with a persona allowed-tool reference such as `mcp.local-search.search`.
5. Open Timeline/debug.
6. Confirm MCP discovery status and non-executable candidate state are visible.
7. Confirm no MCP tool executes.
8. Refresh and confirm history replay keeps the same safe labels when events were persisted.

## Safety Checks

Do not treat this slice as proof of:

- MCP HTTP/SSE/OAuth
- remote MCP
- marketplace/plugin install
- shell/filesystem/browser automation
- MCP tool execution
- approval bypass
- sandboxing

If any raw stderr, env value, token, credential, or private path appears in Timeline/debug, stop and fix redaction before continuing.
