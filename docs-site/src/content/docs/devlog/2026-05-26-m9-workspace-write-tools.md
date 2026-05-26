---
title: 2026-05-26 M9 Workspace Write Tools
description: M9 approval-gated workspace write/edit implementation notes.
---

## Completed

- Added Spec Kit artifacts under `specs/011-safe-workspace-write-tools/`.
- Added allowlisted `workspace.write_file` and `workspace.edit` definitions.
- Kept both mutation tools approval-required.
- Added mutation path checks for traversal, sensitive paths, missing parents, directories, and symlink escape.
- Added bounded UTF-8 text write behavior.
- Added exact single replacement edit behavior with no-mutation failure cases.
- Added worker-level execution with injectable workspace root for tests and local runtime configuration.
- Added frontend replay and ToolCallCard coverage for write/edit result summaries.

## Validation

Run so far:

```bash
go test ./internal/runtime
go test ./internal/productdata
bun test --cwd web ./src/runtime/executionAdapter.test.ts ./src/components/ToolCallCard.test.tsx
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

- Started Vite on `http://127.0.0.1:5173/`.
- Loaded the Loomi app through Playwright.
- Captured `m9-workspace-write-smoke.png`.
- Checked browser console errors: `0`.

## Known Limits

- M9 writes text only.
- `workspace.edit` requires an exact single match.
- Directory creation, patch parsing, shell execution, MCP, browser automation, multi-agent delegation, and activity recording remain later milestones.
