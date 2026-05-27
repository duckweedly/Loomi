---
title: Local M27 Browser Automation Validation
description: Commands for validating the M27 browser automation runtime slice locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesBrowserAutomationTools|TestWorkspaceLSPWebAndBrowserToolsOnlyEnabledForWorkModeRunContext|TestValidateBrowserToolCallArguments'
go test ./internal/runtime -run 'TestBrowser|TestToolBrokerExecutesBrowserOpenThroughOneEntrypoint|TestWorkerExecutesApprovedBrowserOpenAndContinuesModel'
go test ./internal/httpapi -run 'TestM27Browser'
bun test --cwd web SettingsView.tools.test.tsx RunRail.runtime.test.ts runtimeScripts.test.ts mockExecutionAdapter.test.ts mockApiClient.test.ts
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
2. Open Settings > Tools and confirm `browser.open`, `browser.snapshot`, and `browser.click_link` render as builtin, browser-scoped, approval-required, medium risk, public HTTP only, and stateful.
3. Replay or trigger a Work mode browser lifecycle and confirm RunRail labels the rows as browser automation tool rows.
4. Confirm the UI shows URL/title/link/session metadata without raw HTML, cookies, credentials, Set-Cookie values, Authorization values, or local paths.

M27 does not support Chrome profile automation, JavaScript rendering, forms, screenshots, downloads, authenticated browsing, crawling, artifact runtime, channels, desktop activity recording, or multi-agent orchestration.
