# Loomi API Service

M2 adds the first local API service boundary for Loomi.

## Scope

Included:

- `/healthz` process liveness
- `/readyz` dependency readiness
- Local PostgreSQL connectivity
- Explicit schema baseline migration workflow
- Structured diagnostics with request or operation identifiers

Deferred:

- Authentication
- Users, threads, messages
- Runs, events, SSE
- Workers, LLM gateway, tools
- Desktop runtime and production deployment

## Local command shape

See `specs/001-m2-api-db-base/quickstart.md` for the canonical validation flow.
