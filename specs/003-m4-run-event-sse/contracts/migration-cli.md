# Migration CLI Contract: M4 Run/Event Schema

M4 adds schema version `000003` for durable runs and run events.

## Apply

```bash
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version after apply:

```text
3
```

## Readiness

Before M4 migration, `/readyz` should report not ready for the M4 schema requirement.

After M4 migration, `/readyz` should report ready when database dependency and schema state are usable.

M4 readiness requires:

- Database connection available.
- Schema migration version is clean.
- Schema migration version is `3` or later.

## Rollback

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version after one-step rollback:

```text
2
```

Rollback removes M4 run/event tables and makes M4 readiness fail while preserving M3 thread/message schema expectations.

## Reapply

```bash
migrate -path migrations -database "$DATABASE_URL" up 1
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version after reapply:

```text
3
```

## Data Rules

- Migrations must not insert demo runs or events.
- Runs and events are created by API/simulation behavior only.
- Rollback is destructive for M4 run/event data and should be used only in local development.
