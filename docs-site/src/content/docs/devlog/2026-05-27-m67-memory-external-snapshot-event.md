---
title: M67 Memory External Snapshot Event
description: Timeline visibility for external provider prompt recall.
---

## Completed

- Added `memory_external_snapshot_loaded`.
- Emitted a safe progress event after external provider prompt recall succeeds.
- Extended runtime coverage to assert provider and count metadata.

## Safety

- Event metadata includes provider, status, entry count, limit, and redaction flag only.
- Query text, raw hit content, credentials, provider traces, and local paths are not recorded.

## Validation

- `go test ./internal/runtime -run TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider -count=1`
- `go test ./internal/productdata ./internal/runtime -run 'TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
