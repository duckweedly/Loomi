---
title: 2026-05-26 M10 Workspace Exec Command
description: M10 approval-gated exec command implementation notes.
---

## Completed

- Added Spec Kit artifacts under `specs/012-safe-workspace-exec-command/`.
- Added allowlisted `workspace.exec_command`.
- Kept command execution approval-required.
- Added argv-only command validation, cwd root containment, timeout bounds, stdout/stderr truncation, and shell/destructive command rejection.
- Reused worker workspace root injection for approved command execution.
- Added frontend replay and ToolCallCard coverage for exec command summaries.

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
- Captured `m10-workspace-exec-smoke.png`.
- Checked browser console errors: `0`.

## Known Limits

- M10 runs one bounded argv command per tool call.
- No shell strings, PTY, persistent terminal session, process management, MCP, browser automation, multi-agent delegation, or external upload.
