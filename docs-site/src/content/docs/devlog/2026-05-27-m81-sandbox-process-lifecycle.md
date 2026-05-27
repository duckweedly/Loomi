---
title: 2026-05-27 M81 Sandbox Process Lifecycle Recovery
description: Stable bounded process output cursors and terminal continue semantics.
---

M81 continues the M78 sandbox process foundation and closes the next smallest Arkloop comparison gap: process output/lifecycle recovery semantics.

Implemented:

- Changed process stdout/stderr capture to a bounded latest-tail buffer with an absolute captured-byte cursor.
- Kept `next_cursor` advancing even after the retained preview window is full, so a caller can poll long output without replaying already-read bytes.
- Preserved bounded memory: old output is omitted from the retained preview once the byte limit is exceeded.
- Made `sandbox.continue_process` state-only after process exit or termination. It returns status, cursor, byte counts, and `terminal_summary` without attempting stdin writes or close actions.
- Added focused regressions for long output cursor reads, exited terminal summaries, terminated continue safety, cross-run ownership, and output redaction.

Arkloop comparison:

- Arkloop has a sandbox service/agent process controller with bounded process buffers, continue cursors, terminate lifecycle, and session ownership.
- Loomi now covers those mechanism-level basics inside the current local process foundation: stable cursor, bounded retained output, terminal state reads, and run-scoped process ids.
- Loomi still does not implement Arkloop's Docker/Firecracker isolation, guest agent protocol, session pool, PTY/shell/resize, artifact sync, Redis-backed process store, or sandbox template/runtime tier system.

Focused validation:

```bash
go test ./internal/runtime -run 'TestSandboxProcessContinueCursorReadsBoundedLongOutput|TestSandboxProcessContinueAfterExitReturnsTerminalSummary|TestSandboxProcessContinueAfterTerminateOnlyReturnsSafeState|TestSandboxProcessOutputRedactionCoversPathsAndSecrets' -count=1
go test ./internal/runtime -run 'Test.*Sandbox|Test.*Process' -count=1
```
