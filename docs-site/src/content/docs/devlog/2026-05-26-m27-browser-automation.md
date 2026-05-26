---
title: 2026-05-26 M27 Browser Automation Foundation
description: Implementation notes and validation for M27 browser automation tools.
---

## Completed

- Added Spec Kit feature `specs/035-browser-automation-foundation/`.
- Added builtin `browser.open`, `browser.snapshot`, `browser.click_link`, `browser.screenshot`, `browser.type`, and `browser.press` catalog identity, default persona allowlist, Work-mode filtering, and safe tool-call metadata grouping.
- Added `BrowserToolExecutor` with run-scoped sessions, HTTP(S)-only URL validation, credential rejection, private/local network denial, redirect validation, timeout/byte bounds, title/text/link/input extraction, text screenshot summaries, bounded type/press state, and bounded safe result summaries.
- Routed browser tools through ToolBroker, worker approved-tool resume, provider continuation, and bounded HTTP smoke coverage.
- Broadened Gateway continuation support from workspace-only to workspace plus browser tools while keeping duplicate tool-call and max-loop limits.
- Updated Settings > Tools, RunRail labels, mock catalog, seeded run data, and runtime scripts for visible browser lifecycle metadata.

## Validation

Focused validation during implementation:

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesBrowserAutomationTools|TestWorkspaceLSPWebAndBrowserToolsOnlyEnabledForWorkModeRunContext|TestValidateBrowserToolCallArguments'
go test ./internal/runtime -run 'TestBrowser|TestToolBrokerExecutesBrowserOpenThroughOneEntrypoint|TestWorkerExecutesApprovedBrowserOpenAndContinuesModel'
go test ./internal/httpapi -run 'TestM27Browser'
bun test --cwd web runtimeScripts.test.ts mockExecutionAdapter.test.ts mockApiClient.test.ts SettingsView.tools.test.tsx RunRail.runtime.test.ts
```

Full closeout commands should also run before marking M27 complete:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Non-goals

No Chrome profile control, cookies, JavaScript rendering, binary screenshots, downloads, authenticated browsing, crawler, search provider, artifact runtime, activity recorder, plugin marketplace, channels, heartbeat, or multi-agent orchestration were added. `browser.type` and `browser.press` update only the run-scoped HTTP session state.
