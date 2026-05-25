# Quickstart: MCP Stdio Foundation

## Purpose

Validate the M11 foundation slice after implementation: local explicit MCP stdio config validates, discovery/list-tools runs without executing tools, schemas map to namespaced read-only ToolSpec candidates, Persona allowed-tools may reference candidates as disabled, and RunContext plus Timeline/debug show safe MCP availability.

## Backend Validation

Run:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
```

Expected coverage:

- MCP config validation accepts only explicit local stdio config.
- MCP HTTP/SSE/OAuth/remote/marketplace/plugin install inputs are rejected.
- Env values, args, raw stderr, tokens, credentials, and secret-looking paths are redacted from events and summaries.
- Discovery parser accepts valid list-tools output and rejects invalid/oversized/conflicting schemas.
- Tool mapper creates `mcp.<server_slug>.<tool_name>` read-only ToolSpec candidates.
- Persona allowed tools can reference discovered MCP candidates while execution remains disabled by default.
- RunContext includes safe MCP availability summary without invoking MCP tools.
- Future execution boundary is represented as approval-required design state, not an executor.

## Web Validation

If Timeline/debug mapping or UI labels are touched, run related tests such as:

```text
web/src/realApiClient.test.ts
web/src/runtime/realExecutionAdapter.test.ts
web/src/runtime/runtimeEventGroups.test.ts
web/src/components/RunTimeline.runtime.test.ts
web/src/components/RunRail.runtime.test.ts
```

Expected coverage:

- Discovery success/failure/safety labels render from live or replayed metadata.
- MCP candidates show as non-executable.
- Raw stderr, env, args, tokens, credentials, and secret paths are absent from UI state and snapshots.

## Docs Validation

Run:

```bash
bun run --cwd docs-site build
```

Expected docs updates:

- `docs-site/src/content/docs/architecture/mcp-stdio-foundation.md`
- `docs-site/src/content/docs/api/mcp-stdio-foundation.md` or an extension to `docs-site/src/content/docs/api/tool-call-approval.md`
- `docs-site/src/content/docs/runbooks/local-m11-mcp.md`
- `docs-site/src/content/docs/roadmap/current-status.md`
- `docs-site/src/content/docs/spec-kit/workflow.md`
- `docs-site/src/content/docs/devlog/2026-05-25-m11-mcp-stdio-foundation.md`

## Browser/Debug Smoke

1. Start the local API/worker and web app in real API mode.
2. Enable one explicit local stdio MCP server fixture.
3. Run or trigger MCP discovery/list-tools.
4. Create a run whose persona allowed-tools references a discovered MCP candidate, or use a debug surface that displays current RunContext availability.
5. Open Timeline/debug details.
6. Confirm discovery status, candidate count, namespaced tool names, and non-executable state are visible.
7. Confirm no MCP tool executes.
8. Confirm raw stderr, env, args, tokens, credentials, command paths, and secret-looking paths are absent.
9. Refresh/reconnect and confirm history replay shows the same safe status when events were persisted.

## Non-Goals to Verify

- No MCP HTTP/SSE/OAuth.
- No remote network MCP.
- No marketplace or plugin install.
- No shell/filesystem/browser automation.
- No bypass of M7 approval.
- No automatic MCP tool execution.
- No raw stderr/env/tokens/secret paths in Timeline/debug.
- No complex sandbox.
- No rewrite of Persona/Skill, RunContext, or Worker queue.

## Validation Results

Completed on 2026-05-25:

- `go test ./...` passed.
- `bun test ./web/src/realApiClient.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/RunTimeline.runtime.test.ts` passed with 41 tests.
- `bun run --cwd web build` passed.
- `bun run --cwd docs-site build` passed.

Implemented scope:

- Local explicit stdio MCP config validation and sensitive config/process redaction.
- Bounded discovery/list-tools runner and parser.
- Namespaced read-only MCP ToolSpec candidates.
- Persona allowed-tools references for MCP candidates as non-executable.
- RunContext safe MCP availability summary fields.
- Timeline/debug grouping for MCP discovery success/failure/rejected/available/non-executable labels.

Known limitations:

- No MCP tool execution.
- No MCP HTTP/SSE/OAuth.
- No remote MCP.
- No marketplace/plugin install.
- No admin UI or DB-managed MCP server configuration.
- No sandbox or shell/filesystem/browser automation tool support.
