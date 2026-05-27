# Implementation Plan: M52 Memory Recent Errors

## Backend

- Add `GET /v1/memory/errors`.
- Reuse `GetMemoryProviderStatus` and emit one safe diagnostic item when state code is not `ok`.

## Frontend

- Add domain/API/client/state support for memory error events.
- Render a compact recent-errors section in the memory provider panel.

## Validation

- HTTP diagnostic test.
- Frontend mapper and source tests.
- Web/docs build and browser smoke.
