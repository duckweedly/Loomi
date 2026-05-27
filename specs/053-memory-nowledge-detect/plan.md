# Implementation Plan: M53 Nowledge Local Detect

## Backend

- Add a GET endpoint under the memory provider route.
- Use a bounded localhost health check.
- Keep response redacted and deterministic on miss.

## Frontend

- Add real/mock client method and state passthrough.
- Add detect button in the Nowledge config section.
- On success, fill Nowledge base URL through the same provider update path.

## Validation

- HTTP safe miss test.
- Settings source test and web build.
- Docs build and browser smoke.
