# Implementation Plan: M7 Tool Approval Execution Closure

**Branch**: `codex/011-tool-approval-execution` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

## Summary

Complete the smallest safe approval execution loop on top of the merged M7 foundation. Add scoped idempotent approve/deny endpoints, resume approved blocked runs, execute only `runtime.get_current_time`, persist execution and terminal tool events, wire real ToolCallCard actions, map replayed terminal states, and update docs-site.

## Technical Context

**Language/Version**: Go 1.23 backend; TypeScript/React/Vite frontend; Bun for web/docs.

**Primary Dependencies**: Existing Go HTTP API, `internal/productdata` repository/service, `internal/runtime` worker/tool registry/stream, existing PostgreSQL schema and migrations, existing React runtime adapter, ToolCallCard, RunRail, Timeline, and API client. No new dependency is required.

**Storage**: Reuse existing `tool_calls`, run state, job state, and run_events from M7 foundation. No new migration is expected unless current schema lacks a required state field discovered during implementation.

**Testing**: Targeted red-first tests for productdata approve/deny, HTTP routes, worker execution, stream replay, redaction, frontend API client, ToolCallCard actions, and adapter mapping. Final validation commands are the user-specified backend packages, frontend tests, `bun run --cwd web build`, and `bun run --cwd docs-site build`.

**Target Platform**: Local macOS development with Go API, local worker, React web shell, local database test boundaries as already used by the repo.

**Constraints**: Only `runtime.get_current_time` executes. Timezone is omitted or `UTC`. No shell, filesystem, arbitrary network, MCP, browser automation, multi-tool concurrency, multi-agent loop, secret persistence, or approval bypass.

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. Uses Loomi tool approval language and existing UI.
- **Runnable Vertical Slices**: PASS. The slice runs request -> approval -> execution/denial -> SSE -> UI.
- **Core Flow Before Platform Complexity**: PASS. Completes M7 without pulling in MCP, sandbox, desktop runtime, memory, or multi-agent behavior.
- **Observable Agent Execution**: PASS. All approval/execution transitions are persisted run events and visible in ToolCallCard, RunRail, and Timeline.
- **Safety, Permissions, and Data Boundaries**: PASS. Approval is explicit, execution allowlist is fixed, and result/error summaries are redacted.
- **Documentation Definition of Done**: PASS. Required docs-site pages are updated and docs build is part of validation.

## Project Structure

```text
specs/011-tool-approval-execution/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/requirements.md
├── contracts/http-m7-approval-execution.md
├── contracts/frontend-tool-actions.md
└── tasks.md
```

```text
internal/productdata/
├── repository.go              # Atomic approve/deny/execution transitions and terminal guards
├── repository_test.go
├── service.go                 # Decision/event use cases and redacted terminal event writes
└── service_test.go

internal/httpapi/
├── runtime.go                 # Approve/deny handlers under scoped thread/run/tool-call path
├── runtime_test.go
└── server.go                  # Route registration

internal/runtime/
├── worker.go                  # Resume approved blocked runs and execute current-time tool
├── jobs.go                    # Minimal wake/resume hook if needed
├── tools.go                   # Existing current-time executor/redaction boundary
├── tools_test.go
├── worker_test.go
└── stream_test.go

web/src/
├── realApiClient.ts
├── realApiClient.test.ts
├── runtime/realExecutionAdapter.ts
├── runtime/realExecutionAdapter.test.ts
├── components/ToolCallCard.tsx
├── components/ToolCallCard.test.tsx
├── components/RunRail.tsx
└── components/RunTimeline.tsx

docs-site/src/content/docs/
├── architecture/tool-call-approval.md
├── api/tool-call-approval.md
├── runbooks/local-m7.md
├── devlog/2026-05-25-m7-approval-execution.md
└── roadmap/current-status.md
```

## Phase 0: Research Summary

See [research.md](./research.md). Decisions: reuse current schema/projection, implement decisions in productdata for atomicity, resume via existing worker/job pipeline, execute only current-time, finalize denial as stopped, and rely on run events/SSE as the UI source of truth.

## Phase 1: Design Summary

See [data-model.md](./data-model.md), [contracts/http-m7-approval-execution.md](./contracts/http-m7-approval-execution.md), [contracts/frontend-tool-actions.md](./contracts/frontend-tool-actions.md), and [quickstart.md](./quickstart.md).

## Post-Design Constitution Check

PASS. The design completes a runnable vertical slice, keeps capability scope narrow, records observable lifecycle events, and preserves explicit permission and redaction boundaries.

## Complexity Tracking

No constitution violations. If implementation discovers existing M6 worker resume hooks are insufficient, only a minimal approved-tool wake hook may be added; no separate queue or broad worker rewrite is justified.
