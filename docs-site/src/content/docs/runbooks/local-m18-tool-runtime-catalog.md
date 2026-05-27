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
go test ./internal/productdata ./internal/runtime -run 'TestValidateDiscovery|TestToolCatalogIncludesDiscovery|TestDiscovery|TestToolDefinitionsForPersona|TestGatewayExposesCodeAgentToolsToProvider' -count=1
go test ./internal/httpapi -run TestDiscoveryLoadToolsAutoApprovedSmoke -count=1
```

Expected chain:

1. Deterministic provider requests `runtime.get_current_time`.
2. Existing tool-call API records approval required.
3. Approval queues worker resume.
4. Worker executes through broker.
5. Replay events include requested, approval required, approved, executing, succeeded, continuation, and completed.
6. Local stdio MCP smoke follows the same broker path after discovery and approval.
7. Sensitive canaries do not appear in API responses, run events, or continuation context.
8. `tool.load_tools` is auto-approved, executes through the worker/broker path, and returns only enabled-tool safe catalog data.
9. `skill.load_skill` returns installed skill manifest summaries only; full instruction bodies stay out of tool results.

## Settings Check

With the web shell connected to the local API, open Settings > Tools. The page should show tool name, source, group, risk, approval policy, enabled state, execution state, and schema hash. It should not show install/edit/enable/disable controls.

## M21 Follow-Up

Workspace read tools now have their own validation runbook: [Local M21 Workspace Read Tools Validation](./local-m21-workspace-read-tools/).

## Known Limits

The current tool runtime still does not add true dynamic tool-schema injection, full skill instruction-body loading, shell, remote MCP/OAuth, plugin marketplace, autonomous multi-agent execution, or worker queue changes.
