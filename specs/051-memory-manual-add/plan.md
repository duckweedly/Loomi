# Implementation Plan: M51 Manual Memory Add

## Backend

- Add `POST /v1/memory/entries`.
- Reuse `CreateMemoryEntry` and safe `MemorySearchResult` projection.
- Preserve existing `GET /v1/memory/entries` list behavior.

## Frontend

- Add `createMemoryEntry` to real/mock clients.
- Add workspace state action that refreshes entries, audit, and snapshots.
- Add a compact Settings > Memory manual-add form.

## Validation

- HTTP create/list test.
- Mock client and Settings source tests.
- Web build, docs build, diff check, browser smoke.
