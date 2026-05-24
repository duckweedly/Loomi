# Implementation Plan: M6 Worker Job Pipeline

**Branch**: `007-settings-placeholder` | **Date**: 2026-05-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/008-worker-job-pipeline/spec.md`

## Summary

M6 moves Loomi run execution from request-scoped work to durable background execution. The slice adds a database-backed job queue, local worker loop, lease-based ownership, retry/recovery, persisted cancellation, minimal linear pipeline stages, worker/queue diagnostics, and timeline events while preserving existing M4/M5 run/event/SSE and message contracts. The feature artifacts intentionally live under `specs/008-worker-job-pipeline` even though the current local git branch name is still `007-settings-placeholder`.

## Technical Context

**Language/Version**: Go 1.23 for API/runtime/worker execution; TypeScript/React/Vite in `web/`; Bun for frontend/docs commands.

**Primary Dependencies**: Existing Go standard library runtime primitives and HTTP stack; existing `pgx/v5` PostgreSQL boundary; existing M3 thread/message service; existing M4 run/event/SSE stream; existing M5 model gateway and provider-normalized runtime events; existing React runtime adapter and shell components. No external queue service is required for the first M6 slice.

**Storage**: PostgreSQL through existing migrations. M6 needs migration `000005_m6_worker_job_pipeline` for durable background jobs, lease/ownership fields, retry state, and any run/status extensions needed for queued/recovering execution.

**Testing**: `go test ./...`; targeted backend tests for job claim, lease renewal, recovery, retry exhaustion, stop behavior, diagnostics, and idempotent terminal writes; `bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts`; `bun run --cwd web build`; local API/SSE smoke for queued acknowledgement, reconnect, recovery, cancellation, and diagnostics; browser smoke for real API queued/running/recovering/stopped/failed/completed states; `bun run --cwd docs-site build` when docs are updated.

**Target Platform**: Local macOS/Darwin development, local Go API at `127.0.0.1:8080`, web renderer and Electron-compatible frontend shell, local PostgreSQL instance.

**Project Type**: Local web application with Go API/backend runtime, durable product data, local worker execution, and React frontend.

**Performance Goals**: 95% of run-start actions acknowledge queued or running status within 2 seconds; recoverable stale work resumes or fails visibly within 30 seconds after lease expiry in local validation; queued cancellation prevents normal execution in 100% of smoke attempts; two-worker claim validation produces no duplicate terminal events or final assistant messages across 50 consecutive attempts.

**Constraints**: One active run per thread remains enforced. Run/event history remains the user-visible execution contract. Worker execution is local-development scoped and may run in the API process for the first slice. Provider credentials and raw provider failures must never appear in job, event, diagnostic, or frontend output. Tool execution, desktop activity capture, sandboxing, channels, plugins, multi-agent orchestration, RAG/memory, and full workflow DAGs remain out of scope.

**Scale/Scope**: Local-development M6 vertical slice; one local API process may host one or more local workers for validation; one local PostgreSQL instance; fixed local development user; durable jobs for run execution only; minimal linear pipeline stages; no hosted multi-tenant worker fleet.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. M6 uses Loomi's own Run, Background Job, Worker Lease, Pipeline Step, and Run Event terminology and does not copy external product expression, branding, private interfaces, or non-public structures.
- **II. Runnable Vertical Slices**: PASS. The plan produces an end-to-end slice: start a run, receive queued/running acknowledgement, process work through a worker, recover or cancel as needed, persist events, and show terminal outcome in the existing timeline.
- **III. Core Flow Before Platform Complexity**: PASS. Worker/job queue/pipeline follows M5 in the staged roadmap and explicitly keeps desktop runtime, real tool execution, activity capture, sandbox, plugins, multi-agent orchestration, RAG/memory, and full DAG orchestration deferred.
- **IV. Observable Agent Execution**: PASS. Queueing, worker claim, lease renewal, pipeline stages, retries, recovery, cancellation, failures, and final outcomes are persisted as run events and visible through existing timeline/debug paths.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Credentials remain backend-local, diagnostic/event metadata is redacted, user/provider content remains data rather than instructions, and no new external write/tool/desktop permission boundary is introduced.
- **Technical Constraints**: PASS. The plan reuses existing Go API, pgx/PostgreSQL, migration workflow, runtime, run/event/SSE, provider gateway, and frontend adapter seams instead of adding an external queue framework or broad platform runtime.
- **Development Workflow**: PASS. The spec is complete, planning artifacts are generated under the independent `specs/008-worker-job-pipeline` directory, and Phase 2 tasks can be generated after this plan.
- **Documentation Definition of Done**: PASS. Implementation must update docs-site architecture, API, runbook, Spec Kit status, and devlog pages, then validate the docs site build.

## Project Structure

### Documentation (this feature)

```text
specs/008-worker-job-pipeline/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── http-m6.openapi.yaml
│   ├── worker-queue.md
│   ├── pipeline-events.md
│   └── migration-cli.md
└── tasks.md             # Created by /speckit-tasks
```

### Source Code (repository root)

```text
cmd/
├── loomi-api/
│   ├── main.go                         # Wire M6 queue/worker runtime into local API startup
│   └── main_test.go
└── loomi-seed/
    ├── main.go
    └── main_test.go

internal/
├── config/
│   ├── config.go                       # Local worker enable/pause/lease/retry settings
│   └── config_test.go
├── db/
│   ├── readiness.go                    # Require clean schema version >= 5 for M6 worker readiness
│   └── readiness_test.go
├── httpapi/
│   ├── runtime.go                      # Background-capable run start/stop/history/stream behavior
│   ├── diagnostics.go                  # Redacted worker queue diagnostics endpoint if split from health
│   ├── runtime_test.go
│   ├── health_test.go
│   └── server.go
├── productdata/
│   ├── models.go                       # Background job, lease, run status, diagnostic value types
│   ├── repository.go                   # Durable job claim/lease/retry/idempotency persistence
│   ├── repository_test.go
│   ├── service.go                      # Identity-scoped run/job use cases
│   └── service_test.go
└── runtime/
    ├── runner.go                       # Route run execution through background worker boundary
    ├── worker.go                       # Local worker loop and processing lifecycle
    ├── jobs.go                         # Job claim/lease/recovery coordinator
    ├── pipeline.go                     # Minimal linear pipeline stage recorder
    ├── gateway.go                      # Existing M5 provider gateway invoked by worker when configured
    ├── stream.go                       # Existing history-then-live event broadcaster
    └── *_test.go

migrations/
├── 000005_m6_worker_job_pipeline.up.sql
└── 000005_m6_worker_job_pipeline.down.sql

web/src/
├── domain.ts                           # Queued/recovering worker event domain additions
├── realApiClient.ts                    # Worker queue diagnostics and background run states if exposed to UI
├── mockApiClient.ts                    # Explicit mock behavior remains separate
├── runtime/
│   ├── executionAdapter.ts             # Runtime state mapping for queued/recovering/stopping
│   ├── realExecutionAdapter.ts
│   └── *.test.ts
├── components/
│   ├── ChatCanvas.tsx                  # Real background execution states in existing canvas
│   ├── RunTimeline.tsx                 # Queue/worker/recovery/cancel events
│   ├── RunRail.tsx                     # Queued/running/recovering/stopped/failed/completed motion states
│   └── *.test.tsx
└── state.ts                            # Guard terminal states and stale stream updates
```

### Documentation Site Updates During Implementation

```text
docs-site/src/content/docs/architecture/worker-job-pipeline.md
docs-site/src/content/docs/api/worker-job-pipeline.md
docs-site/src/content/docs/runbooks/local-m6.md
docs-site/src/content/docs/spec-kit/workflow.md
docs-site/src/content/docs/roadmap/current-status.md
docs-site/src/content/docs/devlog/2026-05-24-m6-worker-job-pipeline.md
```

**Structure Decision**: M6 keeps product execution data in `internal/productdata` because jobs, leases, run states, and events are durable user-visible execution records. Worker orchestration lives under `internal/runtime` so HTTP handlers remain request boundaries and future standalone workers can reuse the same job/runner logic. Frontend changes stay inside existing runtime adapter, Chat Canvas, Run Rail, and Run Timeline paths so queued/recovering/stopped/failed states extend the current run/event UI instead of creating a parallel job dashboard.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Use a durable database-backed job queue for the M6 vertical slice.
- Keep the initial worker model inside the local API process, with clear worker boundaries.
- Claim jobs through time-bounded worker leases.
- Enforce idempotency at job, run, event, and final-message boundaries.
- Model cancellation as a persisted stop request observed at safe boundaries.
- Add minimal pipeline steps as observable execution stages, not a full orchestration engine.
- Extend existing run/event APIs and frontend runtime adapter instead of adding a parallel job UI.
- Add local diagnostics for queue and worker readiness.
- Use migration `000005_m6_worker_job_pipeline` for background job persistence.

## Phase 1: Design Summary

Design artifacts are generated:

- [data-model.md](./data-model.md) defines Run, Background Job, Worker, Worker Lease, Run Event, Pipeline Step, and Queue Diagnostics.
- [contracts/http-m6.openapi.yaml](./contracts/http-m6.openapi.yaml) defines background-capable run start/read/stop/history/stream and worker queue diagnostics expectations.
- [contracts/worker-queue.md](./contracts/worker-queue.md) defines job creation, claim, lease renewal, recovery, cancellation, completion, and diagnostics semantics.
- [contracts/pipeline-events.md](./contracts/pipeline-events.md) defines M6 timeline events, pipeline step names, frontend runtime mapping, and failure visibility.
- [contracts/migration-cli.md](./contracts/migration-cli.md) defines M6 migration, readiness, local validation, and rollback expectations.
- [quickstart.md](./quickstart.md) defines local migration, API/worker startup, queued run, history replay, multi-worker claim safety, recovery, cancellation, diagnostics, browser smoke, validation commands, and rollback/reapply checks.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart validates migration, worker-enabled startup, queued acknowledgement, history-first replay, multi-worker claim safety, recovery, cancellation, diagnostics, browser states, and rollback/reapply.
- **Core Flow Before Platform Complexity**: PASS. Contracts keep the first pipeline linear and defer real tool execution, desktop runtime, activity capture, sandboxing, channels, plugins, multi-agent orchestration, RAG/memory, and full workflow DAGs.
- **Observable Agent Execution**: PASS. Queue, worker, lease, recovery, retry, cancellation, pipeline, failure, and terminal states are durable run events and mapped into the existing frontend runtime model.
- **Safety/Data Boundaries**: PASS. Job/event/diagnostic metadata is explicitly redacted; provider credentials remain backend-local; user/provider payloads are treated as data; no new external action boundary is introduced.
- **Documentation**: PASS. Documentation targets and validation commands are identified.

## Complexity Tracking

No constitution violations. No external queue service, standalone worker binary, admin console, full pipeline DAG, plugin runtime, desktop runtime, tool execution layer, sandbox, production auth layer, or hosted multi-worker fleet is justified for the first M6 slice.
