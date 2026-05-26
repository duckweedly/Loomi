---
title: 2026-05-26 M8 Workspace Read Tools
description: M8 approval-gated workspace read tools implementation notes.
---

## Completed

- Added Spec Kit artifacts under `specs/010-safe-workspace-read-tools/`.
- Added allowlisted `workspace.glob`, `workspace.grep`, and `workspace.read_file` definitions.
- Kept all workspace read tools approval-required.
- Added root containment, traversal rejection, sensitive path rejection, binary rejection, bounded output, and relative-path result payloads.
- Reused the existing M7 worker execution path for approved workspace read calls.
- Added frontend replay coverage for workspace tool result summaries.
- Updated ToolCallCard formatting so grep match rows render as readable `path:line preview` text instead of `[object Object]`.

## Validation

Run so far:

```bash
go test ./internal/runtime ./internal/productdata
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
- Captured `m8-workspace-read-smoke.png`.
- Checked browser console errors: `0`.

## Known Limits

- M8 uses bounded read summaries only; it is not a full file browser.
- The default worker execution root is the API process working directory.
- Write/edit tools, shell execution, MCP, browser automation, multi-agent delegation, and activity recording remain later milestones.
