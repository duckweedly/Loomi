# Migration CLI Contract: M5 LLM Gateway Schema

M5 adds schema version `000004` for model gateway runs and assistant messages.

## Apply

```bash
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version after apply:

```text
4
```

## Readiness

Before M5 migration, `/readyz` may remain ready for M4 but the model gateway capability check must report unavailable or not configured for model-backed runs.

After M5 migration, readiness for model-backed execution requires:

- Database connection available.
- Schema migration version is clean.
- Schema migration version is `4` or later.
- At least one enabled local provider configuration is valid for the selected provider family.

## Rollback

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version after one-step rollback:

```text
3
```

Rollback removes M5 schema allowances for assistant messages and `model_gateway` runs. It should be used only in local development because completed assistant messages created by M5 depend on the M5 role allowance.

## Data Rules

- Migrations must not insert provider credentials.
- Migrations must not insert demo model runs or events.
- Provider secrets are supplied through local configuration outside migrations.
- Run/event history must preserve M4 ordering and ownership rules.
