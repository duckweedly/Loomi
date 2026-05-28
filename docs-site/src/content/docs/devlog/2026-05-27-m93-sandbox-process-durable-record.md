---
title: 2026-05-27 M93 Sandbox Process Durable Record
description: productdata/Postgres-backed sandbox process summaries for API restart recovery.
---

M93 now persists Loomi sandbox process records through productdata/Postgres without expanding the sandbox boundary into Docker, Firecracker, a guest agent, PTY, shell service, artifact sync, or warm pool.

Implemented:

- Added `sandbox_process_records` migration with rollback.
- Added `productdata.SandboxProcessRecord` and repository methods for save, list, and stale-row cleanup.
- Wired the queued-run API worker path to `PostgresRepository` when Postgres is available.
- Persisted only safe process data: `run_id`, `process_id`, argv summary, cwd alias, status, cursor, byte counters, bounded tails, timestamps, stdin state, and terminal summary.
- Surfaced durable save failures as safe tool failures instead of returning successful `start_process`, `continue_process`, `terminate_process`, or completion polling results with missing process records.
- Redacted secret-looking text and host absolute paths before durable storage.
- Reconciled restored `running` records with no live `exec.Cmd` as `lost`; `continue_process` returns the terminal summary and never restarts or reattaches the OS process.
- Kept run ownership, Work-mode availability, approval required, argv-only validation, allowlist, no shell string, no PTY, and no arbitrary local terminal.

API restart recovery:

- Terminal, expired, and lost summaries can be restored from productdata/Postgres.
- Running process handles are not reattached after API restart. They are marked `lost` on reconcile.
- `continue_process` on a terminal durable record is read-only: it returns stored safe summary and rejects stdin/close mutations.

Focused validation:

```bash
go test ./internal/runtime ./internal/productdata ./internal/httpapi -run 'Test.*Sandbox|Test.*Process' -count=1
```

Additional regression coverage verifies that initial start save failure, continue-state save failure, and asynchronous completion save failure no longer get silently swallowed.
