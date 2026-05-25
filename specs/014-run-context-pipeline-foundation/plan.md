# Implementation Plan: RunContext Pipeline Foundation

**Branch**: `main` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/014-run-context-pipeline-foundation/spec.md`

## Summary

M9 adds the smallest RunContext + Pipeline foundation on top of the closed M8 worker/job queue. The worker prepares durable run context before runtime invocation, resolves only the enabled MVP tool boundary, invokes the existing runtime/model flow, finalizes through existing run/message contracts, and records safe stage trace events for Timeline/debug replay. The design keeps execution linear and deliberately defers Persona/Skill, MCP, Memory/RAG, Sandbox, Desktop Runtime, shell/filesystem/browser automation, multi-agent, Redis, and new queue/platform layers.

## Technical Context

**Language/Version**: Go backend/runtime/worker; TypeScript/React frontend; Starlight docs-site with Bun validation.

**Primary Dependencies**: Existing `internal/productdata` run/thread/message/job/provider/tool-call data boundary; existing `internal/runtime` worker, queued runner, gateway, tools, and pipeline recorder; existing M6/M8 background jobs and M7 tool continuation; existing `web/src/realApiClient.ts`, `web/src/runtime/*`, `RunTimeline`, and `RunRail`.

**Storage**: Existing PostgreSQL tables for runs, run events, threads, messages, background jobs, provider metadata in job/run inputs, and tool-call projection. No new durable queue table is planned for the MVP. A migration is only allowed if implementation finds persisted event metadata cannot represent the stage trace safely.

**Testing**: `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...`; related web runtime/timeline tests; `bun run --cwd web build`; `bun run --cwd docs-site build`; browser smoke for run Timeline/debug trace showing context prepared, tools resolved, runtime invoked, finalized.

**Target Platform**: Local Loomi development environment with Go API/worker, local PostgreSQL, web renderer, and docs-site.

**Project Type**: Local web application plus Go API/backend runtime and durable product data.

**Performance Goals**: RunContext preparation adds no user-visible extra request wait because it occurs in the worker path; stage trace appears in live SSE and history replay within the normal run-event ordering. Local smoke must show all four MVP stages for a completed run.

**Constraints**: Reuse the M6/M8 worker/job queue and M7 tool continuation. No Redis, external queues, hosted multi-worker platform, Persona/Skill, MCP, long-term Memory/RAG, Sandbox, Desktop Runtime, shell/filesystem/browser automation, multi-agent, or broad DAG/workflow abstraction. Secrets and raw provider/tool payloads must not be written to stage events.

**Scale/Scope**: One local worker path, one linear stage list, one executable MVP tool family (`runtime.get_current_time`), one run at a time per thread, local development validation.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. Uses Loomi's RunContext, Pipeline Stage, Run Event, Timeline, and debug terminology without copying external product expression.
- **II. Runnable Vertical Slices**: PASS. The feature produces an end-to-end run that prepares context, resolves tools, invokes runtime, finalizes, and displays persisted stage trace.
- **III. Core Flow Before Platform Complexity**: PASS. M9 follows worker/job and tool-call closure; all later platform capabilities remain non-goals.
- **IV. Observable Agent Execution**: PASS. Stage started/completed/failed records are persisted run events visible through Timeline/debug replay.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Tool resolution is allowlisted and summaries are redacted; no new external action boundary is introduced.
- **Technical Constraints**: PASS. The plan reuses productdata/runtime/web/doc boundaries and avoids new dependencies.
- **Development Workflow**: PASS. Spec, plan, tasks, analyze, then implementation after user confirmation.
- **Documentation Definition of Done**: PASS. Implementation tasks include docs-site architecture/API/runbook/roadmap/devlog/spec-kit updates and docs build.

## Project Structure

### Documentation (this feature)

```text
specs/014-run-context-pipeline-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── run-context-loader.md
│   ├── pipeline-stage-events.md
│   └── frontend-debug-trace.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── productdata/
│   ├── models.go              # RunContext input/result types if kept at data boundary
│   ├── repository.go          # durable reads for run/thread/messages/job/provider route/tool projection
│   ├── repository_test.go
│   ├── service.go             # identity-scoped RunContext loader and validation
│   └── service_test.go
├── runtime/
│   ├── pipeline.go            # linear stage composition and persisted stage events
│   ├── pipeline_test.go
│   ├── queued_runner.go       # route jobs through RunContext + pipeline stages
│   ├── runner.go              # existing runtime/model invocation called from invoke_runtime stage
│   ├── worker.go              # ownership/cancellation integration remains reused
│   ├── gateway.go
│   └── tools.go               # enabled MVP tool summary source
└── httpapi/
    ├── runtime.go             # existing history/SSE/read endpoints expose stage events
    └── runtime_test.go

web/src/
├── realApiClient.ts                         # map safe stage event metadata
├── runtime/
│   ├── realExecutionAdapter.ts              # replay stage trace into runtime state
│   ├── runtimeEventGroups.ts                # debug/timeline grouping for stage events
│   └── *.test.ts
└── components/
    ├── RunTimeline.tsx                      # visible stage trace
    ├── RunRail.tsx                          # debug detail surface if needed
    └── *.test.tsx

docs-site/src/content/docs/
├── architecture/worker-job-pipeline.md      # extend with RunContext/pipeline foundation
├── api/worker-job-pipeline.md               # stage event contract additions
├── runbooks/local-m9.md                     # local validation and browser smoke
├── roadmap/current-status.md
├── spec-kit/workflow.md
└── devlog/2026-05-25-m9-run-context-pipeline-foundation.md
```

**Structure Decision**: Durable context preparation belongs in `internal/productdata` because it reads run/thread/message/job/tool-call/provider state. Runtime orchestration belongs in `internal/runtime` because worker execution, tool resolution, gateway invocation, and finalization already live there. Frontend changes stay in the existing real API adapter, runtime adapter, Timeline, and RunRail so stage trace extends current run-event replay instead of creating a new dashboard.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Build RunContext from durable product data at worker execution time.
- Keep the MVP RunContext fields limited to run, thread, ordered messages, job metadata, provider/model route, and enabled MVP tools.
- Represent the pipeline as a small ordered stage list rather than a general DAG or plugin engine.
- Persist stage trace through existing run events with safe metadata.
- Treat stage failures as terminal run failures unless the existing worker ownership/cancellation path already produces a terminal stopped state.
- Preserve M7 tool-result continuation by invoking existing queued runner/runtime boundaries from the new `invoke_runtime` stage.

## Phase 1: Design Summary

- [data-model.md](./data-model.md) defines RunContext, ContextSource, Pipeline Stage, Pipeline State, Pipeline Trace Event, Tool Resolution Summary, and Stage Failure.
- [contracts/run-context-loader.md](./contracts/run-context-loader.md) defines the loader input, output, validation, redaction, and failure semantics.
- [contracts/pipeline-stage-events.md](./contracts/pipeline-stage-events.md) defines stage names, event ordering, and metadata safety rules.
- [contracts/frontend-debug-trace.md](./contracts/frontend-debug-trace.md) defines Timeline/debug replay expectations.
- [quickstart.md](./quickstart.md) defines validation commands and the browser smoke path requested by the user.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart validates context preparation through final timeline/debug trace.
- **Core Flow Before Platform Complexity**: PASS. Contracts explicitly defer later M9/M10+ capabilities and broad pipeline engines.
- **Observable Agent Execution**: PASS. Stage trace has persisted event contracts and frontend replay expectations.
- **Safety/Data Boundaries**: PASS. Stage metadata is allowlisted/summarized and raw payloads/secrets are excluded.
- **Documentation**: PASS. docs-site update targets and validation are included in tasks.

## Complexity Tracking

No constitution violations. No new queue, dependency, runtime platform, schema expansion, or workflow engine is justified for this foundation slice.
