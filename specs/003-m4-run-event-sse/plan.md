# Implementation Plan: M4 Run, Event, and SSE

**Branch**: `main` | **Date**: 2026-05-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/003-m4-run-event-sse/spec.md`

## Summary

M4 adds the first real execution-observability layer on top of M3's local identity, threads, and messages. It introduces durable runs, ordered run events, history-then-live event streaming, best-effort cooperative stop, and a deterministic local simulated run so the web shell can observe real execution state without pulling in LLM Gateway, tool calling, worker queues, desktop runtime, or production auth.

## Technical Context

**Language/Version**: Go 1.23.0 for the API and deterministic local run simulator; TypeScript/React/Vite in `web/`; Bun 1.3+ for web/docs validation

**Primary Dependencies**: Existing Go standard library HTTP stack (`net/http`, `context`, `encoding/json`), existing pgx/PostgreSQL boundary, existing migration workflow, browser `EventSource`-compatible SSE for live run events, existing React/Vite frontend API seam

**Storage**: Local PostgreSQL; migration version `000003` adds M4 `runs` and `run_events` tables plus indexes/constraints; migrations remain schema-only and do not insert demo runs/events

**Testing**: `go test ./...` for run/event service, repository, readiness, SSE handlers, and deterministic simulator; `bun test ./web/src/*.test.ts ./web/src/components/*.test.ts` for frontend state/event mapping; `bun run --cwd web build`; local API smoke for start/history/stream/stop; `bun run --cwd docs-site build` when docs are changed

**Target Platform**: Local development on macOS/Darwin and localhost-compatible environments

**Project Type**: Local web-service API plus existing web/desktop-feeling shell

**Performance Goals**: First visible run status/event within 2 seconds; cooperative stop terminal state within 3 seconds in the local smoke path; stream interruption shown within 2 seconds; event ordering stable across refresh/reconnect

**Constraints**: Fixed local identity only; one active run per thread while allowing active runs across different threads; deterministic local simulated execution only; no LLM requests, tool execution, worker queue, desktop runtime, production auth, attachments, RAG, or plugin runtime; event stream must deliver persisted history before live events; user-controlled event payload text is data, not instructions; structured errors must not leak secrets

**Scale/Scope**: One local API process, one local PostgreSQL instance, one fixed local development user, local runs/events only, deterministic local simulator with a small event sequence per run, and no hosted multi-user operations

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. M4 uses Loomi's own Run, Run Event, Event Stream, Stop Request, and Execution Boundary terms; it does not copy external product branding, UI expression, private interfaces, or non-public structures.
- **II. Runnable Vertical Slices**: PASS. M4 is demonstrable by applying migration 000003, starting a run, observing persisted history and live stream updates, refreshing to recover the timeline, and issuing cooperative stop.
- **III. Core Flow Before Platform Complexity**: PASS. Run/event/SSE is the constitution's next stage after M3; LLM gateway, tools, workers, desktop runtime, attachments, RAG, plugins, and production auth remain deferred.
- **IV. Observable Agent Execution**: PASS. M4 directly implements observable run lifecycle, persisted events, SSE updates, errors, stop outcomes, and timeline/debug visibility.
- **V. Safety, Permissions, and Data Boundaries**: PASS. M4 keeps fixed local ownership, treats event payload content as data, redacts secrets from user-facing errors, and avoids tool/file/desktop permissions until later milestones.
- **Technical Constraints**: PASS. The plan reuses existing Go API, pgx/PostgreSQL, migration, diagnostics, and frontend API seam boundaries without adding a router, ORM, queue, model provider, or desktop runtime.
- **Development Workflow**: PASS. The feature has a spec and clarifications; this plan produces research, data model, contracts, and quickstart before tasks/implementation.
- **Documentation Definition of Done**: PASS. Implementation must update docs-site architecture, API, runbook, Spec Kit status, and devlog pages, then validate the docs site.

## Project Structure

### Documentation (this feature)

```text
specs/003-m4-run-event-sse/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── http-m4.openapi.yaml
│   ├── sse-run-events.md
│   ├── migration-cli.md
│   └── frontend-run-data-source.md
└── tasks.md             # Created by /speckit-tasks
```

### Source Code (repository root)

```text
cmd/
└── loomi-api/
    └── main.go                         # Wire M4 product runtime service into the local API

internal/
├── db/
│   └── readiness.go                    # Require clean schema version >= 3 for M4 readiness
├── httpapi/
│   ├── runtime.go                      # `/v1/threads/{thread_id}/runs`, run read/stop, and event stream handlers
│   ├── runtime_test.go
│   └── server.go                       # Register M4 run/event routes
├── productdata/
│   ├── models.go                       # Shared thread/message plus new run/event value types if kept in one domain package
│   ├── repository.go                   # pgx-backed run/event persistence and active-run constraints
│   ├── repository_test.go
│   ├── service.go                      # Identity-scoped run/event use cases
│   └── service_test.go
└── runtime/
    ├── simulator.go                    # Deterministic local simulated run event sequence
    ├── simulator_test.go
    ├── stream.go                       # In-process event broadcaster for history-then-live SSE delivery
    └── stream_test.go

migrations/
├── 000003_m4_run_event_sse.up.sql
└── 000003_m4_run_event_sse.down.sql

web/src/
├── realApiClient.ts                    # Fetch/SSE-backed run/event client methods
├── mockApiClient.ts                    # Preserve mock behavior when no real API base is configured
├── state.ts                            # Select current run, consume run events, guard stale stream updates
├── domain.ts                           # Run/event domain types aligned with M4 categories
└── components/
    ├── ChatCanvas.tsx                  # Show real run progress/final/error states for selected thread
    ├── RunRail.tsx                     # Bind agent state motion to real run/event state
    └── RunTimeline.tsx                 # Render real persisted/live M4 events
```

### Documentation Site Updates During Implementation

```text
docs-site/src/content/docs/architecture/run-event-sse.md
docs-site/src/content/docs/api/run-event-sse.md
docs-site/src/content/docs/runbooks/local-m4.md
docs-site/src/content/docs/spec-kit/workflow.md
docs-site/src/content/docs/devlog/2026-05-23-m4-run-event-sse.md
```

**Structure Decision**: M4 reuses M3's local API, identity, product data, and PostgreSQL boundaries. Run/event persistence stays near `internal/productdata` because runs are product execution data owned by local identity and thread. Deterministic execution and stream fan-out live under `internal/runtime` so later worker/job execution can replace the simulator without rewriting HTTP handlers or frontend contracts. Frontend code consumes the same run/event domain shape as the future real adapter and keeps mock mode available only when no real API base is configured.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). All plan unknowns are resolved:

- M4 uses deterministic local simulated runs as the first runnable execution source.
- Active run concurrency is one active run per thread, with independent active runs allowed across different threads.
- Persisted run events are the source of truth; SSE is a live projection.
- SSE clients receive persisted history before live events.
- Initial event categories are `lifecycle`, `progress`, `message`, `error`, and `final`.
- Stop is best-effort cooperative, not hard worker interruption.
- M4 stays local-development scoped and keeps production auth/LLM/tools/workers deferred.
- M4 readiness requires schema version `000003` or later.
- Event payload text is data and must not be treated as instructions.

## Phase 1: Design Summary

Design artifacts are generated:

- [data-model.md](./data-model.md) defines Run, Run Event, Event Stream Cursor, Stop Request, Deterministic Local Simulation, Stream State, and M4 Schema Revision.
- [contracts/http-m4.openapi.yaml](./contracts/http-m4.openapi.yaml) defines run start/read/stop/history endpoints and structured error responses.
- [contracts/sse-run-events.md](./contracts/sse-run-events.md) defines history-then-live SSE event delivery and reconnect behavior.
- [contracts/migration-cli.md](./contracts/migration-cli.md) defines M4 migration apply/version/rollback/reapply expectations.
- [contracts/frontend-run-data-source.md](./contracts/frontend-run-data-source.md) defines frontend real/mock run data source behavior and stale stream guards.
- [quickstart.md](./quickstart.md) defines local setup, schema readiness checks, run/event API smoke, SSE smoke, stop smoke, frontend smoke, rollback/reapply, and docs validation.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart demonstrates schema readiness failure before M4, readiness success after M4 migration, run start, event history, SSE history-then-live behavior, cooperative stop, frontend real run/event rendering, and docs validation.
- **Core Flow Before Platform Complexity**: PASS. Contracts explicitly defer LLM, tools, workers, desktop runtime, attachments, RAG, plugins, production auth, and hosted multi-user operations.
- **Observable Execution Boundary**: PASS. Runs and events are durable and streamed, with lifecycle/progress/message/error/final categories and terminal outcomes.
- **Safety/Data Boundaries**: PASS. Ownership remains local-identity scoped, event payload text is treated as data, errors are structured and redacted, and no tool/file/desktop permissions are introduced.
- **Documentation**: PASS. Documentation targets and validation commands are identified.

## Complexity Tracking

No constitution violations. No additional router, ORM, queue, LLM provider, worker runtime, desktop runtime, production auth layer, or plugin platform is justified for M4.
