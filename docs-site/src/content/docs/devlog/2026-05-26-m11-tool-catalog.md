---
title: 2026-05-26 M11 Tool Catalog Visibility
description: Read-only tool catalog API and Settings panel implementation notes.
---

## Completed

- Added Spec Kit artifacts under `specs/013-tool-catalog-visibility/`.
- Added backend `ToolCatalog()` metadata for all current runtime/workspace tools.
- Added `GET /v1/tools/catalog` with deterministic read-only entries.
- Added frontend domain/API mapping and mock catalog support.
- Replaced Settings > Tools placeholder with a read-only catalog panel.
- Updated Settings category copy/status for Tools.

## Validation

Focused validation:

```bash
go test ./internal/runtime ./internal/httpapi
bun test --cwd web ./src/realApiClient.test.ts ./src/components/SettingsView.runtime.test.tsx ./src/components/settingsCatalog.test.ts
```

Full validation:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```

Browser smoke:

- Open Settings > Tools.
- Confirm catalog cards render and placeholder copy is gone.
- Confirm `workspace.exec_command` renders with `exec`, `required`, `high`, `process`, and `workspace_exec`.
- Check browser console errors: `0`.
- Captured `m11-tool-catalog-smoke.png`.

## Known Limits

- Catalog is read-only metadata.
- Permission editing, auto-approval, tool execution controls, MCP, browser automation, sandbox sessions, and multi-agent delegation remain future slices.
