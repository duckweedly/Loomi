# Implementation Plan: M2 API and Database Base

**Branch**: `main` | **Date**: 2026-05-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/001-m2-api-db-base/spec.md`

## Summary

M2 establishes Loomi's first real backend boundary: a local-development API service with liveness/readiness checks, safe runtime configuration, structured diagnostics, PostgreSQL connectivity, and a reversible schema-version baseline. The existing M1 mock UI remains runnable while M2 prepares the service and persistence seams that M3/M4 will extend with real product data and run events.

## Technical Context

**Language/Version**: Go 1.23.12 for the M2 service; TypeScript/React remains in `web/` for M1 validation; Bun 1.3.13 for the docs site

**Primary Dependencies**: Go standard library `net/http`, `log/slog`, `context`, `database/sql`-free service code; `github.com/jackc/pgx/v5/pgxpool` for PostgreSQL pooling; `golang-migrate/migrate` CLI or Docker image for SQL migration execution

**Storage**: Local PostgreSQL for development only; M2 schema contains migration/version state only and creates no future business tables

**Testing**: `go test ./...` for service/config/db behavior; migration up/down smoke commands; `bun run build` in `web/` to confirm M1 mock UI still builds; `bun run build` in `docs-site/` when docs are updated

**Target Platform**: Local development on macOS/Darwin arm64 first, with generic localhost-compatible commands documented for future contributors

**Project Type**: Web-service foundation plus existing web UI shell

**Performance Goals**: Liveness responds in under 1 second locally; readiness reports dependency failure in under 10 seconds; local setup verifies liveness/readiness in under 10 minutes on a prepared machine

**Constraints**: Service must start even when PostgreSQL is unavailable; readiness must report not ready until config, database connectivity, and schema version state are usable; diagnostics must be structured and redact secrets; no auth, thread/message, run/event, worker, LLM, tool, desktop runtime, or production deployment work in M2

**Scale/Scope**: Single local developer environment; one API process; one local PostgreSQL instance; one schema baseline migration pair; no multi-user or hosted operations

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. M2 implements Loomi's own service foundation and does not copy external branding, UI expression, prompts, private names, or non-public interfaces.
- **II. Runnable Vertical Slices**: PASS. The slice is runnable through local service startup, liveness/readiness checks, migration up/down, and documented smoke verification.
- **III. Core Flow Before Platform Complexity**: PASS. M2 is the prescribed API/database base and explicitly defers auth, thread/message, run/event/SSE, LLM, tools, workers, and desktop runtime.
- **IV. Observable Agent Execution**: PASS. M2 does not execute agents yet, but it establishes structured diagnostics and request/operation identifiers required by future observable runs.
- **V. Safety, Permissions, and Data Boundaries**: PASS. M2 redacts secrets, keeps write behavior limited to local schema migration, and documents failure visibility.
- **Technical Constraints**: PASS. The plan keeps API boundaries clear for the existing `web/` mock UI and adds Go service foundations under reserved backend directories.
- **Development Workflow**: PASS. Spec and clarify are complete; this plan generates research, data model, contracts, and quickstart before tasks/implementation.
- **Documentation Definition of Done**: PASS. Implementation tasks must update `docs-site/src/content/docs/` and validate docs with `bun run build`.

## Project Structure

### Documentation (this feature)

```text
specs/001-m2-api-db-base/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── http-health.openapi.yaml
│   └── migration-cli.md
└── tasks.md             # Created later by /speckit-tasks
```

### Source Code (repository root)

```text
cmd/
└── loomi-api/
    └── main.go                 # API process entry point

internal/
├── config/
│   ├── config.go               # Runtime config loading, validation, redaction
│   └── config_test.go
├── db/
│   ├── pool.go                 # PostgreSQL pool creation and ping readiness checks
│   ├── readiness.go            # Database/schema readiness helpers
│   └── readiness_test.go
├── diagnostics/
│   ├── logger.go               # slog JSON logger, request/operation id helpers
│   └── logger_test.go
└── httpapi/
    ├── server.go               # HTTP server wiring
    ├── health.go               # /healthz and /readyz handlers
    └── health_test.go

services/
└── api/
    └── README.md               # Service boundary notes and local commands

migrations/
├── 000001_schema_baseline.up.sql
└── 000001_schema_baseline.down.sql

compose.yaml                    # Local PostgreSQL service for development
.env.example                    # Safe local configuration example
```

### Documentation Site Updates During Implementation

```text
docs-site/src/content/docs/api/index.md
docs-site/src/content/docs/architecture/api-db-base.md
docs-site/src/content/docs/runbooks/index.md
docs-site/src/content/docs/spec-kit/workflow.md
docs-site/src/content/docs/devlog/2026-05-23-m2-api-db-base.md
```

**Structure Decision**: M2 uses the existing reserved backend directories. `cmd/loomi-api` owns process startup, `internal/` owns private service packages, `services/api` documents the service boundary, and `migrations/` holds SQL schema revision files. The existing `web/` project is not restructured; it is only validated to ensure M1 mock behavior still builds.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). All technical unknowns from the plan context are resolved:

- Go standard library HTTP server is sufficient for M2 and avoids premature routing abstractions.
- `pgxpool` is the PostgreSQL pool because it is the current Go-native pgx pool API and supports direct ping checks.
- `golang-migrate` SQL files provide reversible up/down schema workflows without embedding migration execution into the API process.
- Structured JSON diagnostics use `log/slog` with request and operation identifiers.
- Local PostgreSQL via `compose.yaml` is enough for M2 because production and hosted deployment are out of scope.

## Phase 1: Design Summary

Design artifacts are generated:

- [data-model.md](./data-model.md) defines runtime configuration, liveness/readiness status, dependency checks, schema revision, and diagnostics records.
- [contracts/http-health.openapi.yaml](./contracts/http-health.openapi.yaml) defines the `/healthz` and `/readyz` HTTP contracts.
- [contracts/migration-cli.md](./contracts/migration-cli.md) defines the migration command contract and expected version behavior.
- [quickstart.md](./quickstart.md) defines local setup, smoke checks, migration up/down, and M1/docs validation commands.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart demonstrates service startup, health checks, database readiness failure/success, migration up/down, and UI/docs validation.
- **Core Flow Before Complexity**: PASS. No future business tables or platform capabilities are introduced.
- **Observable Execution Foundation**: PASS. Logs and readiness payloads include request/operation identifiers for future tracing.
- **Safety/Data Boundaries**: PASS. Secrets are redacted and local-only persistence scope is explicit.
- **Documentation**: PASS. Required docs-site pages and validation commands are identified for implementation.

## Complexity Tracking

No constitution violations. No additional runtime layers, routers, ORMs, auth systems, workers, or deployment abstractions are justified for M2.
