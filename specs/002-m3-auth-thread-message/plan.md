# Implementation Plan: M3 Auth, Thread, and Message

**Branch**: `main` | **Date**: 2026-05-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/002-m3-auth-thread-message/spec.md`

## Summary

M3 builds the minimal durable product data layer on top of the M2 API/database base: a fixed local development identity, durable users, threads, user-authored messages, idempotent message creation, real/mock frontend API switching, M3 schema readiness, explicit seed data, and structured API errors. The implementation reuses the M2 Go HTTP service, pgx/PostgreSQL boundary, migration workflow, diagnostics style, and docs-site discipline while explicitly deferring run/event/SSE, LLM, tools, workers, desktop runtime, attachments, RAG, and catalog-style extension capabilities.

## Technical Context

**Language/Version**: Go 1.23.0 for the API and seed command; TypeScript/React/Vite in `web/`; Bun 1.3+ for web/docs validation

**Primary Dependencies**: Existing Go standard library HTTP stack (`net/http`, `context`, `log/slog`, `encoding/json`), `github.com/jackc/pgx/v5/pgxpool` for PostgreSQL access, explicit `golang-migrate` SQL files, existing React/Vite frontend with browser `fetch`

**Storage**: Local PostgreSQL; migration version `000002` adds M3 `users`, `threads`, and `messages` tables plus indexes/constraints; migrations remain schema-only and do not insert demo threads or messages

**Testing**: `go test ./...` for identity, repository/service validation, readiness, diagnostics, and HTTP handlers; migration up/down/reapply smoke commands; `bun run build` in `web/`; browser smoke for mock mode and real API mode; `bun run build` in `docs-site/` when docs are changed

**Target Platform**: Local development on macOS/Darwin and localhost-compatible environments

**Project Type**: Local web-service API plus existing web/desktop-feeling shell

**Performance Goals**: Local `/readyz` dependency failures stay within the existing 1-10 second readiness timeout; local thread/message CRUD responses should be visibly immediate for a single-developer dataset; full local M3 setup and smoke verification should complete in under 15 minutes on a prepared machine

**Constraints**: Fixed local identity only; no session table, user picker, or user-selecting request header; no automatic migration on API startup; no automatic mock fallback when a real API base is configured; messages are complete final user text only; M3 must not create assistant placeholders, run events, streaming deltas, tool calls, worker jobs, model outputs, LLM requests, desktop runtime behavior, attachments, RAG, or catalog extension surfaces; structured errors must include stable code, human message, and request id without secrets; key non-obvious boundaries get concise Chinese WHY comments

**Scale/Scope**: One local API process, one local PostgreSQL instance, one fixed local development user, local thread/message data only, active-thread default list, direct retrieval for archived durable state, and no hosted multi-user operations

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. M3 defines Loomi-specific identity/thread/message language and does not copy external branding, UI expression, private interfaces, or non-public product structure.
- **II. Runnable Vertical Slices**: PASS. The slice is demonstrable by applying M3 migration, starting the API, reading local identity, creating/listing/updating/archiving a thread, creating/reloading an idempotent user message, switching the web shell between mock and real API modes, and validating docs.
- **III. Core Flow Before Platform Complexity**: PASS. Auth/thread/message is the constitution's next staged milestone after M2; run/event/SSE, LLM, tools, workers, desktop runtime, attachments, RAG, and catalog concepts remain deferred.
- **IV. Observable Agent Execution**: PASS. M3 does not execute agents; it preserves the future observability boundary by keeping thread lifecycle separate from run status and by excluding run/event/model/tool semantics from message persistence.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Ownership is explicit through a fixed local user boundary, cross-owner access returns non-enumerating errors, diagnostics redact secrets, and configured-real-API failures stay visible rather than silently falling back to mock data.
- **Technical Constraints**: PASS. The plan reuses the existing Go API layout, PostgreSQL pool, migration workflow, frontend API seam, and docs-site structure without introducing frameworks, ORMs, sessions, queues, or production deployment layers.
- **Development Workflow**: PASS. The feature has a complete spec and clarification checklist; this plan produces research, data model, contracts, and quickstart before tasks/implementation.
- **Documentation Definition of Done**: PASS. Implementation must update `docs-site/src/content/docs/` for architecture, API, runbook, Spec Kit status, and devlog, then validate the docs site.

## Project Structure

### Documentation (this feature)

```text
specs/002-m3-auth-thread-message/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── http-m3.openapi.yaml
│   ├── migration-cli.md
│   ├── seed-cli.md
│   └── frontend-data-source.md
└── tasks.md             # Created by /speckit-tasks
```

### Source Code (repository root)

```text
cmd/
├── loomi-api/
│   └── main.go                         # Wire M2 health/readiness plus M3 product routes
└── loomi-seed/
    └── main.go                         # Explicit local demo seed command; not run by migrations

internal/
├── config/
│   ├── config.go                       # Reuse M2 config; keep DATABASE_URL redaction
│   └── config_test.go
├── db/
│   ├── pool.go                         # Reuse pgxpool creation
│   ├── readiness.go                    # Require clean schema version >= 2 for M3 readiness
│   └── readiness_test.go
├── diagnostics/
│   ├── logger.go                       # Reuse JSON diagnostics and request/operation ids
│   └── logger_test.go
├── identity/
│   ├── local.go                        # Fixed local user id/display name and EnsureLocalUser boundary
│   └── local_test.go
├── productdata/
│   ├── models.go                       # User, Thread, Message, validation constants, states
│   ├── repository.go                   # pgx-backed SQL persistence and idempotency queries
│   ├── repository_test.go
│   ├── service.go                      # Identity-scoped thread/message use cases and transactions
│   └── service_test.go
└── httpapi/
    ├── errors.go                       # Structured API error envelope with request id
    ├── health.go                       # Existing M2 health/readiness response shape
    ├── product.go                      # `/v1/me`, `/v1/threads`, and message handlers
    ├── product_test.go
    └── server.go                       # Route registration

migrations/
├── 000001_schema_baseline.up.sql
├── 000001_schema_baseline.down.sql
├── 000002_m3_auth_thread_message.up.sql
└── 000002_m3_auth_thread_message.down.sql

web/src/
├── apiClient.ts                        # Data-source selector and shared client interface
├── mockApiClient.ts                    # Existing mock thread/message/run behavior
├── realApiClient.ts                    # Fetch-backed M3 thread/message client
├── domain.ts                           # Separate thread lifecycle from run status
├── App.tsx                             # Load real/mock data, show recoverable real-API errors
├── components/
│   ├── ChatCanvas.tsx                  # Render durable messages without assuming run events
│   ├── Composer.tsx                    # Send user messages with client_message_id
│   └── ThreadSidebar.tsx               # Create/list/rename/archive real threads
└── useWorkspaceShellState.ts           # Existing shell panel state remains UI-only
```

### Documentation Site Updates During Implementation

```text
docs-site/src/content/docs/architecture/auth-thread-message.md
docs-site/src/content/docs/api/thread-message.md
docs-site/src/content/docs/runbooks/local-m3.md
docs-site/src/content/docs/spec-kit/workflow.md
docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md
```

**Structure Decision**: M3 keeps M2's `cmd/`, `internal/`, `migrations/`, and `web/` boundaries. Identity is isolated under `internal/identity` so the fixed local user can later be replaced by real auth without changing every handler. Durable thread/message use cases live in `internal/productdata` because they are the first real product-data layer and must stay separate from future run/event execution. The frontend splits mock and real clients behind one API seam so the M1 demo shell remains available only when no real API base is configured.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). All plan unknowns are resolved:

- M3 continues with Go standard library HTTP and pgxpool; no router, ORM, code generator, or session framework is introduced.
- The fixed local identity is a durable `users` row ensured by the API/seed boundary, not demo data inserted by migrations and not a session table.
- M3 schema version `000002` owns `users`, `threads`, and `messages`; text IDs avoid adding a UUID extension or dependency for local-only scope.
- Message idempotency uses an optional `client_message_id` plus a partial unique index scoped by `thread_id`, `user_id`, and identifier.
- HTTP APIs use `/v1` JSON contracts and a structured error envelope containing `code`, `message`, and `request_id`.
- Frontend real/mock switching uses `VITE_LOOMI_API_BASE_URL`; if configured, fetch errors remain visible and do not fall back to mock data.
- Seed behavior is an explicit command (`go run ./cmd/loomi-seed`) that creates deterministic local demo data without migration side effects.

## Phase 1: Design Summary

Design artifacts are generated:

- [data-model.md](./data-model.md) defines Local Identity, User, Thread, Message, Client Message Identifier, API Error, Schema Revision, Seed Data Set, and Frontend Data Source Mode.
- [contracts/http-m3.openapi.yaml](./contracts/http-m3.openapi.yaml) defines the `/v1/me`, `/v1/threads`, and `/v1/threads/{thread_id}/messages` HTTP contracts.
- [contracts/migration-cli.md](./contracts/migration-cli.md) defines the M3 migration apply/version/rollback/reapply contract.
- [contracts/seed-cli.md](./contracts/seed-cli.md) defines the explicit local seed command and idempotent demo data rules.
- [contracts/frontend-data-source.md](./contracts/frontend-data-source.md) defines frontend environment switching and no-fallback real API behavior.
- [quickstart.md](./quickstart.md) defines local setup, schema readiness checks, API smoke tests, idempotency checks, frontend mock/real smoke, rollback/reapply, and docs validation.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart demonstrates schema readiness failure at M2-only baseline, readiness success after M3 migration, product API CRUD, idempotent message creation, seed command, frontend mock mode, frontend real API mode, and docs validation.
- **Core Flow Before Platform Complexity**: PASS. Contracts explicitly omit run/event/SSE, LLM, tools, worker, desktop runtime, attachments, RAG, and catalog surfaces.
- **Observable Execution Boundary**: PASS. Thread lifecycle is modeled as `active`/`archived`; run status stays a separate future/mocked concern.
- **Safety/Data Boundaries**: PASS. API access is identity-scoped, cross-owner access uses non-enumerating `not_found`, and errors carry request ids without secrets.
- **Documentation**: PASS. Implementation docs targets and validation commands are identified.

## Complexity Tracking

No constitution violations. No additional runtime framework, ORM, auth/session system, background worker, model provider, desktop runtime, or deployment abstraction is justified for M3.
