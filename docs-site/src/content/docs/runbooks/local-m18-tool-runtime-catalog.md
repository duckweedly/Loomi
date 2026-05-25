---
title: Local M18 Tool Runtime Catalog Validation
description: Local validation commands and smoke expectations for M18.
---

## Commands

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Focused Smoke

```bash
go test ./internal/httpapi -run 'TestToolsCatalogHandlerReturnsSafeCatalog|TestM18BuiltinToolApprovalRunsThroughBrokerSmoke|TestM12RealLocalMCPApprovalSmoke'
```

Expected chain:

1. Deterministic provider requests `runtime.get_current_time`.
2. Existing tool-call API records approval required.
3. Approval queues worker resume.
4. Worker executes through broker.
5. Replay events include requested, approval required, approved, executing, succeeded, continuation, and completed.
6. Local stdio MCP smoke follows the same broker path after discovery and approval.
7. Sensitive canaries do not appear in API responses, run events, or continuation context.

## Settings Check

With the web shell connected to the local API, open Settings > Tools. The page should show tool name, source, group, risk, approval policy, enabled state, execution state, and schema hash. It should not show install/edit/enable/disable controls.

## Known Limits

M18 does not add workspace file tools, shell, sandbox, browser, web fetch/search, artifact runtime, remote MCP/OAuth, plugin marketplace, provider autodetect, multi-agent, or worker queue changes.
