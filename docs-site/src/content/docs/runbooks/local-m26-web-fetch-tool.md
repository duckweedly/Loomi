---
title: Local M26 Web Fetch Tool Validation
description: Commands for validating the M26 web.fetch runtime slice locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesWebFetchTool|TestWorkspaceLSPAndWebToolsOnlyEnabledForWorkModeRunContext|TestValidateWebFetchToolCallArguments'
go test ./internal/runtime -run 'TestWebFetch|TestToolBrokerExecutesWebFetchThroughOneEntrypoint|TestWorkerExecutesApprovedWebFetchAndContinuesModel'
go test ./internal/httpapi -run 'TestM26WebFetchApproveExecuteFinalSmoke'
bun test --cwd web SettingsView.tools RunRail.runtime runtimeScripts
```

## Full Closeout

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual Smoke

1. Start the API and web app.
2. Open Settings > Tools and confirm `web.fetch` renders as builtin, web-scoped, read-only, approval-required, medium risk, and public HTTP only.
3. Replay or trigger a Work mode `web.fetch` lifecycle and confirm RunRail labels the rows as a web fetch tool.
4. Confirm the UI shows URL/status/truncation metadata without raw response bodies, cookies, credentials, Set-Cookie values, or local paths.

M26 does not support browser automation, JavaScript rendering, cookies, authenticated fetch, crawling, search provider queries, artifact runtime, channels, desktop activity recording, or multi-agent orchestration.
