---
title: M69 Memory Provider Runtime Errors
description: Safe runtime provider failures in Settings > Memory recent errors.
---

## Completed

- Added `memory_external_snapshot_failed` for external prompt recall failures.
- Kept external prompt recall failures non-fatal: Gateway returns the original prepared context and continues the run.
- Added a safe memory provider error read model that combines configuration diagnostics and runtime provider failure events.
- Extended `/v1/memory/errors` with optional safe `run_id` and `event_type` fields.
- Updated frontend mapping for runtime memory error fields while preserving the existing recent errors UI.

## Validation

- `go test ./internal/runtime ./internal/httpapi -run 'TestGatewayRecordsExternalMemorySnapshotFailureForRecentErrors|TestMemoryErrorsReportsRuntimeProviderFailures|TestMemoryErrorsReportsProviderDiagnostic' -count=1`
- `bun test --cwd web src/memory.test.ts`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
- `git diff --check`

## Notes

- Error items do not include prompt query text, raw memory content, upstream response bodies, API keys, Authorization headers, local paths, or provider traces.
- This does not add retries, provider restart controls, or raw log inspection.
