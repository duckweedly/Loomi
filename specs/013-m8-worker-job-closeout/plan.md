# Implementation Plan: M8 Worker Job Closeout

**Branch**: `main` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/013-m8-worker-job-closeout/spec.md`

## Summary

M8 closeout audits the original Worker + Job Queue roadmap against the already implemented M6 Worker Job Pipeline. The audit finds that jobs table, transactional enqueue, worker claim, lease renewal, failed terminal state, crash recovery, immediate API acknowledgement, and lost-lock ownership guards are already covered. The smallest remaining gap is that recovery emitted retry scheduling but did not move the next claim time forward, so this plan adds local retry backoff only and records the closeout in docs.

## Technical Context

**Language/Version**: Go 1.23 for backend/runtime; TypeScript/React remains unchanged; Bun for docs-site validation.

**Primary Dependencies**: Existing PostgreSQL-backed productdata repository, in-memory service test double, runtime worker/job coordinator, and docs-site Starlight project. No new dependency.

**Storage**: Existing `background_jobs.scheduled_at`, `attempt_count`, `max_attempts`, lease, and ownership fields from migration `000005_m6_worker_job_pipeline`.

**Testing**: Targeted `go test ./internal/productdata ./internal/runtime`, requested `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...`, and `bun run --cwd docs-site build`.

**Target Platform**: Local Loomi development environment.

**Project Type**: Go API/runtime with docs-site documentation.

**Performance Goals**: Recovered work remains visible and claimable shortly after lease expiry; local retry backoff stays within the existing M6 recovery smoke expectations.

**Constraints**: No Redis, external queue, new worker platform, M9 RunContext/Pipeline, MCP, Memory, Desktop Runtime, or frontend behavior changes.

**Scale/Scope**: Closeout/audit feature plus one minimal retry/backoff patch.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. The feature uses Loomi roadmap terminology and audits internal behavior.
- **II. Runnable Vertical Slices**: PASS. The patch is verified by targeted retry/recovery tests and package validation.
- **III. Core Flow Before Platform Complexity**: PASS. The plan closes worker/job queue before M9 and explicitly avoids later platform capabilities.
- **IV. Observable Agent Execution**: PASS. Retry scheduling remains visible through existing job retry events and diagnostics.
- **V. Safety, Permissions, and Data Boundaries**: PASS. No new external writes or secret-bearing surfaces are introduced.
- **Technical Constraints**: PASS. Existing Go/PostgreSQL/service boundaries are reused.
- **Development Workflow**: PASS. Spec, plan, tasks, optional analysis, implementation, validation, and docs are generated in one closeout session.
- **Documentation Definition of Done**: PASS. Roadmap/current-status and devlog are updated, then docs-site build is run.

## Project Structure

### Documentation (this feature)

```text
specs/013-m8-worker-job-closeout/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   └── m8-audit.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/productdata/
├── repository.go        # PostgreSQL stale-lease recovery schedules next retry with backoff
├── service.go           # In-memory service mirrors retry backoff behavior
├── repository_test.go   # Contract coverage for backed-off retry scheduling
└── service_test.go      # In-memory behavior coverage for backed-off retry scheduling

docs-site/src/content/docs/
├── roadmap/current-status.md
└── devlog/2026-05-25-m8-worker-job-closeout.md
```

**Structure Decision**: The closeout belongs in `internal/productdata` because retry scheduling is durable job state. Runtime worker code already consumes `scheduled_at` through claim behavior, so no runtime or frontend changes are needed.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Treat `specs/008-worker-job-pipeline` and its code/tests as the existing M8-equivalent implementation.
- Count "retry/backoff" as behavior only when `scheduled_at` delays the next claim.
- Patch only recovery scheduling; do not change worker loops, queue topology, or frontend states.
- Document M8 as closed by M6 plus this closeout patch, without claiming M9.

## Phase 1: Design Summary

- [data-model.md](./data-model.md) defines audit items, background jobs, worker leases, retry schedules, and closeout records.
- [contracts/m8-audit.md](./contracts/m8-audit.md) records the audit matrix and coverage evidence.
- [quickstart.md](./quickstart.md) lists validation commands and expected outcomes.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. The closeout produces testable retry/backoff behavior and docs evidence.
- **Core Flow Before Platform Complexity**: PASS. No M9 or platform expansion is included.
- **Observable Agent Execution**: PASS. Retry events now point at the real next scheduled claim time.
- **Safety/Data Boundaries**: PASS. Redaction and ownership guard behavior are preserved.
- **Documentation**: PASS. Roadmap and devlog updates are planned and validated.

## Complexity Tracking

No constitution violations. No new dependencies or runtime layers are justified.
