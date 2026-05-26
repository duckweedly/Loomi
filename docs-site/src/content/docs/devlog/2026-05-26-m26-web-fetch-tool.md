---
title: 2026-05-26 M26 Web Fetch Tool Foundation
description: Implementation notes and validation for M26 web.fetch.
---

## Completed

- Added Spec Kit feature `specs/034-web-fetch-tool-foundation/`.
- Added builtin `web.fetch` catalog identity, default persona allowlist, Work-mode filtering, and safe tool-call metadata grouping.
- Added `WebToolExecutor` with HTTP(S)-only URL validation, credential rejection, private/local network denial, DNS checks, redirect validation, timeout/byte bounds, text-like content handling, title extraction, and bounded excerpts.
- Routed `web.fetch` through ToolBroker, worker approved-tool resume, provider continuation, and HTTP smoke coverage.
- Updated Settings > Tools, RunRail labels, mock catalog, and runtime scripts for visible web lifecycle metadata.

## Validation

Focused validation during implementation:

```bash
go test ./internal/productdata
go test ./internal/runtime
go test ./internal/httpapi -run 'TestM26WebFetch|TestM25LSPReadonlyApproveExecuteFinalSmoke'
bun test --cwd web SettingsView.tools RunRail.runtime runtimeScripts
```

Full closeout commands should also run before marking M26 complete:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Non-goals

No browser automation, JavaScript rendering, cookies, authenticated fetch, crawler, search provider, artifact runtime, activity recorder, plugin marketplace, channels, heartbeat, or multi-agent orchestration were added.
