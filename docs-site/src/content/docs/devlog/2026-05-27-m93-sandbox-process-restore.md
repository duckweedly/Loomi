---
title: 2026-05-27 M93 Sandbox Process In-Process Restore
description: In-process local process handle restore, reconciliation, and TTL cleanup before a sandbox service.
---

M93 closes the next Arkloop comparison gap without introducing Docker, Firecracker, a guest agent, a browser tier, or a sandbox warm pool.

Implemented:

- Added a minimal memory-backed sandbox process repository and in-process records with `run_id`, `process_id`, safe argv summary, cwd alias, status, cursor, timestamps, exit code, stdin state, and bounded stdout/stderr tails.
- Restored memory-backed process records into a rebuilt `SandboxProcessStore`.
- Kept terminal records readable after rebuild while rejecting stdin/close mutations.
- Marked restored `running` records as `failed` when the local command is missing, with a safe terminal summary instead of starting a replacement process.
- Added max-lifetime and idle-timeout reconciliation that marks registry-owned processes `expired` without killing unrelated host processes.
- Preserved absolute stdout cursor semantics and bounded latest-tail output.
- Kept run ownership checks for continue and terminate.
- Added terminal-run guards for start and continue mutation.
- Extended ToolCallCard safe previews for process resume metadata while continuing to redact raw stdout/stderr, absolute paths, cwd, and secret-looking output.
- Added HTTP smoke coverage for `start_process -> continue_process(close stdin) -> registry rebuild -> continue_process(read terminal summary)`.

Arkloop comparison:

- Arkloop has a sandbox service, Docker/Firecracker backends, session templates, warm pools, restore TTL, output flush, guest agent protocol, and browser tier.
- Loomi M93 now has the pre-service abstraction needed to support those later: in-process handle records, explicit lifecycle states, output tails, ownership, and cleanup semantics.
- Loomi M93 still runs only local argv-form allowlisted host commands and does not provide container/microVM isolation.
- Loomi M93 does not survive an API process restart yet; durable productdata/Postgres storage remains future work.

Focused validation:

```bash
go test ./internal/runtime ./internal/httpapi -run 'Test.*Sandbox|Test.*Process|Test.*Resume|Test.*TTL|Test.*Cursor' -count=1
bun test --cwd web ./src/components/ToolCallCard.test.tsx
```
