# Migration CLI Contract: M3 Auth, Thread, and Message

M3 extends the explicit local migration workflow from M2. The API service must not apply migrations automatically at startup.

## Migration Files

```text
migrations/
├── 000001_schema_baseline.up.sql
├── 000001_schema_baseline.down.sql
├── 000002_m3_auth_thread_message.up.sql
└── 000002_m3_auth_thread_message.down.sql
```

## Required Environment

```bash
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
```

The URL is a local development example. Logs, readiness reasons, and API errors must not print credentials or full connection strings.

## Apply M3 Schema

From the repository root:

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected final version:

```text
2
```

Expected schema additions:

```text
users
threads
messages
```

Rules:

- The M3 up migration creates product-data tables and supporting indexes/constraints.
- The migration must not insert demo threads or demo messages.
- The migration should not create run, run_event, tool, worker, model output, attachment, RAG, desktop runtime, or catalog tables.

## Validate M2-Only Not Ready State

To verify readiness fails before M3 schema exists, apply or roll back to M2 baseline:

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version after rolling back M3:

```text
1
```

With the API running, `GET /readyz` must return HTTP 503 and a schema check failure because M3 requires version `2`.

## Reapply M3 Schema

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected final version:

```text
2
```

After reapply, `GET /readyz` may return HTTP 200 when PostgreSQL is reachable and migration state is clean.

## Rollback Contract

Rolling back one migration from version 2 must remove the M3 business tables and return the database to the M2 schema baseline:

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
```

Expected outcome:

- `users`, `threads`, and `messages` no longer exist.
- Migration version is `1` and clean.
- No manual cleanup is required before reapplying M3.

## Readiness Interaction

`GET /readyz` must report:

- `not_ready` when PostgreSQL cannot be pinged.
- `not_ready` when schema migration state is missing, dirty, or lower than version `2`.
- `ready` only when PostgreSQL ping succeeds and schema version is clean at version `2` or later.

## Prohibited M3 Migration Behavior

- No automatic migration during API startup.
- No demo thread/message insertion.
- No run/event/SSE, LLM, tool, worker, desktop runtime, attachment, RAG, or catalog tables.
- No credential printing in migration smoke output captured by Loomi code.
