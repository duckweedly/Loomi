# Implementation Plan: M47 Memory Provider Config Details

## Slice

This slice extends the existing M42 provider foundation instead of creating a new runtime subsystem. It changes the persisted config shape, backend status projection, HTTP update contract, frontend mapping, and Settings panel.

## Files

- `migrations/000015_memory_provider_config_details.*.sql`
- `internal/productdata/models.go`
- `internal/productdata/service.go`
- `internal/productdata/repository.go`
- `internal/httpapi/memory.go`
- `web/src/domain.ts`
- `web/src/realApiClient.ts`
- `web/src/components/SettingsView.tsx`
- `web/src/styles.css`
- docs-site memory architecture/API/runbook/devlog pages

## Validation

- `go test ./internal/productdata ./internal/httpapi -run 'TestMemoryProvider|TestPrepareRunContextIncludesMemoryProviderReadiness' -count=1`
- `bun test --cwd web src/memory.test.ts src/components/SettingsView.runtime.test.tsx`
- `bun run --cwd web build`
- `bun run --cwd docs-site build`
