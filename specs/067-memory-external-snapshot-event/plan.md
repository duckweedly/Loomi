# Plan: M67 Memory External Snapshot Event

## Scope

Add one run event from `Gateway.withExternalMemorySnapshot` after external provider recall succeeds.

## Runtime

- Add productdata event constant `memory_external_snapshot_loaded`.
- Append a progress event with safe provider/count metadata.
- Keep append errors non-fatal, matching the prompt enrichment fallback.

## Validation

```bash
go test ./internal/runtime -run TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider -count=1
```
