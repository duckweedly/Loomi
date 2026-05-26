---
title: 2026-05-26 M12 Todo Write Planning Tool
description: runtime.todo_write implementation notes.
---

## Completed

- Added Spec Kit artifacts under `specs/014-todo-write-planning-tool/`.
- Added allowlisted `runtime.todo_write`.
- Kept todo planning approval-required through the existing tool lifecycle.
- Added bounded todo item validation: 1-20 items, trimmed titles, fixed status enum.
- Added runtime execution that returns counts and sanitized item summaries.
- Added worker execution coverage for approved todo_write calls.
- Added ToolCallCard coverage for todo_write summaries.
- Added mock tool catalog entry for Settings > Tools visibility.

## Validation

Focused validation:

```bash
go test ./internal/productdata ./internal/runtime
bun test --cwd web ./src/components/ToolCallCard.test.tsx
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

- Opened Settings > Tools.
- Confirmed `runtime.todo_write`, `plan`, and `no_side_effect_internal` render.
- Checked browser console errors: `0`.
- Captured `m12-todo-write-smoke.png`.

## Known Limits

- Todo state is stored only in tool-call projections for the current run.
- M12 does not add editable task management, MCP, LSP, spawn-agent, auto-approval, or multi-agent delegation.
