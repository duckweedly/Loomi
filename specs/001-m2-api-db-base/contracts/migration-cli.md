# Migration CLI Contract: M2 Schema Baseline

M2 uses an explicit local migration workflow. The API service must not apply migrations automatically at startup.

## Migration Files

```text
migrations/
├── 000001_schema_baseline.up.sql
└── 000001_schema_baseline.down.sql
```

Rules:

- Version numbering starts at `000001`.
- The M2 baseline migration records schema version state only.
- The M2 baseline migration must not create users, threads, messages, runs, run_events, workers, tools, or other business tables.
- The down migration returns the local store to the pre-M2 baseline state.

## Required Environment

```bash
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
```

The URL above is a local development example only. Logs and diagnostics must not print it in full.

## Commands

Apply all migrations:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Expected outcome:

```text
000001/u schema_baseline
```

Check current version:

```bash
migrate -path migrations -database "$DATABASE_URL" version
```

Expected outcome after apply:

```text
1
```

Rollback one migration:

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
```

Expected outcome:

```text
000001/d schema_baseline
```

Re-apply after rollback:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Expected outcome:

```text
000001/u schema_baseline
```

## Readiness Interaction

`GET /readyz` must report:

- `not_ready` when PostgreSQL cannot be pinged.
- `not_ready` when the schema baseline version is absent or dirty.
- `ready` only when PostgreSQL ping succeeds and the baseline schema version is present and clean.

## Prohibited M2 Behavior

- No automatic migration during API startup.
- No production deployment contract.
- No future business tables.
- No credential printing in logs or readiness reasons.
