# Research: M2 API and Database Base

## Decision: Use Go standard library HTTP for the M2 API service

**Rationale**: M2 only needs liveness and readiness endpoints plus process wiring. The standard library is enough, keeps the service small, avoids framework decisions before real API surface area exists, and aligns with Loomi's staged principle of avoiding premature platform complexity.

**Alternatives considered**:

- Gin/Echo/Fiber: rejected because M2 has only two endpoints and no routing complexity.
- gRPC: rejected because browser and smoke checks need simple HTTP health/readiness behavior.
- Reusing the web dev server: rejected because M2 must establish a backend service boundary independent of the M1 UI shell.

## Decision: Use `github.com/jackc/pgx/v5/pgxpool` for PostgreSQL connectivity

**Rationale**: `pgxpool` is the Go-native pgx connection pool for PostgreSQL. It can create a pool from `DATABASE_URL` and expose ping-style readiness behavior. The pool keeps M2 close to future Postgres-specific needs without adding ORM or query generation before business tables exist.

**Alternatives considered**:

- `database/sql` with `lib/pq`: rejected because `lib/pq` is older and pgx is the stronger Go Postgres default for new work.
- ORM: rejected because M2 creates no business tables and does not need entity mapping.
- Raw `pgx.Conn` only: rejected because future API requests will need pooled connections, and a pool is still minimal.

## Decision: Use `golang-migrate` SQL migration files for schema baseline

**Rationale**: M2 needs a reversible local schema workflow. `golang-migrate` supports sequential SQL files, `up`, `down`, and `version` commands against PostgreSQL. Keeping migration execution as a CLI contract avoids embedding migration control into the API process too early.

**Alternatives considered**:

- Custom SQL script runner: rejected because dirty/version state and rollback semantics would need to be reimplemented.
- Goose: acceptable alternative, but `golang-migrate` has a simple CLI contract and common sequential file pattern.
- App-start automatic migrations: rejected because M2 should keep startup, readiness, and schema control explicit for learning and safety.

## Decision: M2 schema baseline records migration version only and creates no business tables

**Rationale**: Clarification confirmed that users, threads, messages, runs, events, workers, tools, and related business models are out of scope. The first migration pair should be intentionally minimal so M3/M4 own their tables and tests.

**Alternatives considered**:

- Create empty users/threads/messages tables now: rejected because it pulls M3 into M2.
- Create all future run/event tables now: rejected because it pulls M4 and observability model decisions forward.

## Decision: Service starts when PostgreSQL is unavailable; readiness reports not ready

**Rationale**: Clarification confirmed alive and ready must be separate. This improves local diagnosis and matches standard service health behavior: the process can answer liveness while dependent behavior remains unavailable.

**Alternatives considered**:

- Fail startup when PostgreSQL is unavailable: rejected because it makes failure diagnosis less visible and weakens liveness/readiness separation.
- Degraded mode: rejected because M2 does not need a broader policy for serving dependency-free routes beyond health checks.

## Decision: Use structured JSON diagnostics with request/operation identifiers

**Rationale**: M2 needs enough observability to debug startup, health/readiness, config, and schema workflow failures. Go's `log/slog` supports structured JSON logs without a third-party logging dependency.

**Alternatives considered**:

- Plain console output: rejected because it does not meet the clarification requirement for structured diagnostics.
- Full OpenTelemetry metrics/tracing: rejected because M2 does not yet have run/event execution or production operations.

## Decision: Local development PostgreSQL via `compose.yaml`

**Rationale**: The spec limits M2 persistence support to local development. A local compose service gives repeatable developer setup without introducing deployment readiness, hosted environments, or release packaging.

**Alternatives considered**:

- Production-ready deployment: rejected by clarification and milestone scope.
- SQLite first: rejected because current roadmap specifies PostgreSQL first and SQLite adapter later for desktop runtime.
