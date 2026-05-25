---
title: 2026-05-25 M11 MCP Stdio Foundation
description: Implementation notes and validation for the M11 local MCP stdio discovery slice.
---

## Completed

- Added local stdio MCP config validation.
- Added sensitive MCP redaction for config and process failure output.
- Added MCP list-tools response parsing.
- Added namespaced read-only MCP ToolSpec candidates.
- Prevented MCP candidates from overriding internal runtime tools.
- Allowed persona allowed-tools to reference MCP candidates as `discovered_non_executable`.
- Added RunContext safe MCP availability summary fields.
- Mapped MCP discovery/tool availability frontend events into Timeline/debug groups.
- Documented future MCP execution as M7 approval + audit only.

## Validation

Planned commands:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
bun test ./web/src/realApiClient.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/RunTimeline.runtime.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

Record final evidence in `specs/016-mcp-stdio-foundation/quickstart.md`.

## Boundaries

M11 does not execute MCP tools. It does not add HTTP/SSE/OAuth, remote MCP, marketplace install, shell/filesystem/browser automation, sandboxing, or approval bypass.
