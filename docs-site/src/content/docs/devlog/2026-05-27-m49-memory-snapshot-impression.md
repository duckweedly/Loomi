---
title: M49 Memory Snapshot And Impression
description: Settings > Memory snapshot and impression endpoints, UI cards, tests, and validation notes.
---

## Changed

- Added safe local memory overview snapshot and memory impression models.
- Added `/v1/memory/snapshot`, `/v1/memory/snapshot/rebuild`, `/v1/memory/impression`, and `/v1/memory/impression/rebuild`.
- Added real and mock web API methods plus Settings > Memory cards for Memory Snapshot and Memory Impression.
- Added rebuild actions that refresh the safe local projections from approved memories.

## Safety

- Snapshot and impression outputs are built from approved memory search results.
- Raw memory content, proposal bodies, idempotency keys, tool output, provider traces, local paths, credentials, and secret-like values stay out of API responses and UI state.
- M49 does not execute external Nowledge/OpenViking adapters.

## Validation

- `go test ./internal/productdata ./internal/httpapi -run 'TestMemory.*Snapshot|TestMemorySnapshot|TestMemoryOverviewAndImpression|TestMemoryProvider|TestPrepareRunContextIncludesMemoryProviderReadiness|TestPrepareRunContextIncludesSafeMemorySnapshot' -count=1`
- `bun test --cwd web src/memory.test.ts src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts`
