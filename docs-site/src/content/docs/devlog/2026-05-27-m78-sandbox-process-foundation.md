---
title: 2026-05-27 M78 Sandbox Process Foundation
description: Minimal local process lifecycle for approval-gated sandbox tools.
---

M78 turns the one-shot sandbox command slice into a minimal controlled process lifecycle for long tests, builds, and local dev-service foundations.

Implemented:

- Kept `sandbox.start_process`, `sandbox.continue_process`, and `sandbox.terminate_process` on the existing Work-mode approval-gated tool path.
- Used a local in-memory run-scoped process registry. Process handles cannot be continued or terminated from another run.
- Kept start argv-only, workspace-root scoped, approval-required, and allowlisted to the existing read/validation command set. No shell strings or model-supplied env are accepted.
- Added prompt continue behavior: `sandbox.continue_process` returns current status/output immediately when no new output is available instead of blocking indefinitely.
- Added terminate lifecycle summary through `terminal_summary`.
- Applied bounded output preview redaction for one-shot and process results: workspace root scrub, host absolute path redaction, invalid UTF-8 drop, and secret-looking content redaction.
- Updated RunRail and ToolCallCard labels for process lifecycle tools so they show safe `process_id`, cursor, status, and terminal summary details.

Arkloop comparison:

- Arkloop has a dedicated sandbox service, session manager, warm pool, Docker/Firecracker-capable isolation, guest agent process controller, bounded process buffers, and continue/terminate actions.
- Loomi M78 borrows the process lifecycle mechanics only: process id, ownership check, output snapshot, continuation, termination, and summary.
- Loomi M78 does not copy Arkloop config format, naming, API shape, product copy, tier expression, or full isolation architecture.

Explicit non-goals:

- No Firecracker, Docker, guest filesystem sync, network namespace, shell session, PTY, resize, artifact extraction, Redis, or persistent process store.
- No arbitrary background process manager and no unrestricted local terminal.

Focused validation:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi -run 'Test.*Sandbox|Test.*Process|Test.*ToolCatalog' -count=1
bun test --cwd web ./src/components/RunRail.runtime.test.ts ./src/components/ToolCallCard.test.tsx
```
