# Implementation Plan: M49 Memory Snapshot And Impression

## Slice

M49 adds a thin local snapshot/impression read layer over approved Loomi memories. It mirrors the target settings concepts without adding a new storage engine.

## Backend

- Add `MemoryOverviewSnapshot`, `MemorySnapshotHit`, and `MemoryImpressionSnapshot` models.
- Add `MemorySnapshotService` methods for get/rebuild snapshot and get/rebuild impression.
- Build both outputs from `SearchMemory` with bounded approved safe summaries.
- Add `/v1/memory/snapshot`, `/v1/memory/snapshot/rebuild`, `/v1/memory/impression`, and `/v1/memory/impression/rebuild`.

## Frontend

- Add real/mock API client methods and mapping tests.
- Load snapshot/impression state with memory provider status.
- Render Settings > Memory cards with updated time, hit count, safe preview text, and rebuild actions.

## Validation

- Focused productdata/httpapi tests for safe snapshot/impression output and rebuild flags.
- Focused web tests for API mapping, mock client rebuilds, and settings source contract.
- Web build, docs build, diff check, and browser smoke for Settings > Memory.
