# Data Model: M2 API and Database Base

M2 contains service foundation entities rather than user-facing business records. It must not create users, threads, messages, runs, run events, workers, tools, or desktop runtime records.

## Runtime Configuration

Represents the settings required to start and validate the local M2 service.

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| `APP_ENV` | string | yes | `local`, `test`, or `development`; M2 docs use `local` |
| `HTTP_ADDR` | string | yes | Host and port bind address; local default `127.0.0.1:8080` |
| `DATABASE_URL` | secret string | yes | Must parse as a PostgreSQL connection URL; never logged in full |
| `LOG_LEVEL` | string | yes | `debug`, `info`, `warn`, or `error`; local default `info` |
| `READINESS_TIMEOUT_SECONDS` | integer | yes | Positive integer; local default `5`; must keep readiness failure under 10 seconds |

Validation:

- Missing required values fail startup with a redacted diagnostic message.
- Malformed values fail startup with a redacted diagnostic message.
- Secret fields expose only a redacted marker such as `[redacted]` in logs and errors.

## Service Status

Represents whether the process can respond to liveness checks.

| Field | Type | Rules |
|-------|------|-------|
| `status` | enum | `alive` |
| `service` | string | `loomi-api` |
| `environment` | string | Safe runtime environment name |
| `request_id` | string | Per-request identifier returned in diagnostics |

State:

```text
process started -> alive
```

The liveness status does not depend on PostgreSQL availability.

## Readiness Status

Represents whether M2 dependencies are usable.

| Field | Type | Rules |
|-------|------|-------|
| `status` | enum | `ready` or `not_ready` |
| `service` | string | `loomi-api` |
| `environment` | string | Safe runtime environment name |
| `request_id` | string | Per-request identifier returned in diagnostics |
| `checks` | array | One entry per dependency check |

State transitions:

```text
not_ready -> ready       when config, database ping, and schema version checks pass
ready -> not_ready       when database ping or schema version check fails later
```

## Dependency Check

Represents an individual readiness check.

| Field | Type | Rules |
|-------|------|-------|
| `name` | enum | `config`, `database`, `schema` |
| `status` | enum | `ok` or `failed` |
| `reason` | string | Non-secret human-readable reason; omitted or empty when `ok` |

Validation:

- `database` fails when PostgreSQL ping fails within the readiness timeout.
- `schema` fails when the schema version baseline is missing or dirty.
- Reasons must not include full credentials or sensitive connection strings.

## Persistent Store Baseline

Represents the local PostgreSQL schema state that future milestones can extend.

| Field | Type | Rules |
|-------|------|-------|
| `version` | integer | `1` after M2 baseline apply |
| `dirty` | boolean | Must be `false` for readiness to pass |
| `business_tables_created` | boolean | Must be `false` in M2 |

Lifecycle:

```text
empty store -> baseline applied (version 1, dirty false)
baseline applied -> baseline rolled back (no M2 schema version applied)
baseline rolled back -> baseline applied again
```

## Schema Revision

Represents a migration file pair.

| Field | Type | Rules |
|-------|------|-------|
| `version` | integer | Sequential, starting at `000001` |
| `name` | string | `schema_baseline` for M2 |
| `direction` | enum | `up` or `down` |
| `creates_business_tables` | boolean | `false` for M2 |

## Diagnostic Record

Represents structured output for startup, health, readiness, config, and schema workflow checks.

| Field | Type | Rules |
|-------|------|-------|
| `level` | enum | `debug`, `info`, `warn`, `error` |
| `message` | string | Human-readable summary |
| `request_id` | string | Required for HTTP checks |
| `operation_id` | string | Required for startup and migration smoke commands |
| `component` | string | Example: `config`, `http`, `database`, `migration` |
| `error` | string | Redacted when present |

Validation:

- At least one of `request_id` or `operation_id` must be present.
- Sensitive values must be redacted.

## Smoke Verification Result

Represents the documented outcome of the M2 validation flow.

| Field | Type | Rules |
|-------|------|-------|
| `step` | string | Startup, liveness, readiness, migration up, migration down, web build, docs build |
| `expected` | string | Expected observable result |
| `actual` | string | Recorded by the developer during verification |
| `passed` | boolean | True only when actual matches expected |
