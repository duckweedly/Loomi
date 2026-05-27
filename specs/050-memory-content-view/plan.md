# Implementation Plan: M50 Memory Content View

## Backend

- Add `GET /v1/memory/content?uri=memory://{entry_id}&layer=overview|read`.
- Reuse `GetMemoryEntry` authorization and safe projection.
- Return only safe `summary`, optionally prefixed by title for `read`.

## Frontend

- Add `getMemoryContent` to real and mock API clients.
- Expose the helper through workspace state.
- Add a Settings > Memory modal opened from snapshot hit chips.

## Validation

- HTTP contract test for safe content response.
- Frontend source/mock tests.
- Web build, docs build, diff check, browser smoke.
