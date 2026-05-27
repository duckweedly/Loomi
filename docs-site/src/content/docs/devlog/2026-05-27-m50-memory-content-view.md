---
title: M50 Memory Content View
description: Safe content endpoint and Settings > Memory snapshot-hit modal.
---

## Changed

- Added `GET /v1/memory/content` for `memory://{entry_id}` snapshot hit URIs.
- Added real/mock web API methods and state passthrough for memory content reads.
- Made Settings > Memory snapshot hit chips open a read-only safe content modal.

## Safety

- Content view returns title plus safe summary only.
- Raw memory content, content hash, proposal bodies, provider traces, local paths, credentials, and secret-like values stay out of the response and modal.
- The endpoint reuses memory entry scope authorization.

## Validation

- `go test ./internal/httpapi -run TestMemorySnapshotAndImpressionHandlers -count=1`
- `bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts`
