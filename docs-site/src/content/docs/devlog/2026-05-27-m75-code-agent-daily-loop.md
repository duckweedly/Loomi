---
title: 2026-05-27 M75 Code-Agent Daily Loop
description: Dogfoodable Work-mode code task loop over existing tool, approval, worker, event, and continuation paths.
---

Implemented:

- Added an HTTP smoke for the daily code-agent loop: grep, read, patch preview, patch apply, sandbox validation command, provider continuation, and final assistant message.
- Kept the tool surface unchanged. The loop reuses `workspace.grep`, `workspace.read`, `workspace.patch_preview`, `workspace.patch_apply`, and `sandbox.exec_command`.
- Verified approval gates for patch preview, patch apply, and sandbox exec before execution.
- Covered safety boundaries for workspace tools outside Work mode, unapproved mutation, terminal-run approval, and stale patch application without writing.
- Updated RunRail and ToolCallCard readable summaries so patch and command steps expose safe operation, changed status, diff redaction, exit code, and loop count without raw paths/content.

Focused validation:

```bash
go test ./internal/httpapi -run 'TestM22|TestM75' -count=1
bun test --cwd web ./src/components/RunRail.runtime.test.ts ./src/components/ToolCallCard.test.tsx
```

Full validation:

```bash
go test ./internal/httpapi ./internal/runtime -run 'Test.*Workspace|Test.*Sandbox|Test.*Bounded|Test.*CodeAgent' -count=1
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
```
