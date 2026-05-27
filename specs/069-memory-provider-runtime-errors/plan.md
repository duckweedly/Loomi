# Plan: M69 Memory Provider Runtime Errors

## Scope

Extend the existing memory recent-errors boundary to include safe runtime provider failure events.

## Runtime

- Add event type `memory_external_snapshot_failed`.
- Emit it from `Gateway.withExternalMemorySnapshot` when the active external provider handles recall but returns an error.
- Keep the original prepared context and continue the run.

## Product Data / API

- Add a safe `MemoryProviderErrorEvent` read model.
- Include current provider diagnostics and recent runtime provider failure events.
- Map `/v1/memory/errors` to the read model and include optional `run_id` / `event_type`.

## Frontend

- Preserve existing Settings > Memory recent errors UI.
- Map optional runtime run/event fields in the real API client.

## Validation

```bash
go test ./internal/runtime ./internal/httpapi -run 'TestGatewayRecordsExternalMemorySnapshotFailureForRecentErrors|TestMemoryErrorsReportsRuntimeProviderFailures|TestMemoryErrorsReportsProviderDiagnostic' -count=1
bun test --cwd web src/memory.test.ts
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```
