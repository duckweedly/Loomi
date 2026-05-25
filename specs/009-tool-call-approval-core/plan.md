# Implementation Plan: M7 Tool Call Approval Core

**Branch**: `009-tool-call-approval-core` | **Date**: 2026-05-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/009-tool-call-approval-core/spec.md`

## Summary

M7 upgrades Loomi from M5's non-executed tool boundary to a minimal, auditable, approval-gated internal tool invocation slice. It adds a first-class tool-call lifecycle, run_event mapping, a minimal current-state projection for idempotent approval and worker resume, schema validation for untrusted model arguments, approve/deny APIs, one safest MVP tool (`runtime.get_current_time`), worker block/resume semantics on top of M6, ToolCallCard/RunRail/Timeline display states, and documentation of safety boundaries. The slice does not add shell, filesystem, arbitrary network, MCP, browser automation, multi-agent behavior, long-term memory/RAG, or full multi-step agent loops.

## Technical Context

**Language/Version**: Go 1.23 for API/runtime/worker execution; TypeScript/React/Vite in `web/`; Bun for frontend/docs commands.

**Primary Dependencies**: Existing Go standard library HTTP/runtime primitives; existing `pgx/v5` PostgreSQL boundary; existing M3 thread/message service; existing M4 run/event/history-first SSE stream; existing M5 provider-normalized model gateway and non-executed tool boundary; existing M6 durable background jobs, local worker, leases, recovery, cancel, diagnostics, and pipeline events; existing React runtime adapter, ToolCallCard, RunRail, Timeline, Settings, and Background tasks entry points. No provider SDK, external queue, sandbox, MCP, browser automation, or plugin dependency is required.

**Storage**: PostgreSQL through existing migrations. M7 should add migration `000006_m7_tool_call_approval` for a minimal `tool_calls` projection and any run/job status extension needed for `blocked_on_tool_approval` or equivalent worker state. Run events remain the audit/replay contract.

**Testing**: `go test ./...`; targeted backend tests for tool schema validation, redaction, lifecycle event ordering, approve/deny idempotency, worker block/resume, cancellation, and duplicate execution prevention; frontend tests for real execution adapter, ToolCallCard, RunRail, and Timeline grouping; `bun test` for relevant web tests; `bun run --cwd web build`; local API/SSE smoke for fake/model tool request, approval wait, approve, deny, execution, cancellation, reconnect replay, and unsupported tool failure; browser smoke for ToolCallCard approval/result UI; `bun run --cwd docs-site build` when docs are updated.

**Target Platform**: Local macOS/Darwin development, local Go API at `127.0.0.1:8080`, web renderer and Electron-compatible frontend shell, local PostgreSQL instance.

**Project Type**: Local web application with Go API/backend runtime, durable product data, local worker execution, and React frontend.

**Performance Goals**: 95% of tool-request events become visible through history-first SSE within 2 seconds after provider normalization in local validation; repeated same approve or deny decisions remain idempotent across 10 repeated requests; approved MVP tool calls reach visible succeeded/failed/cancelled state within 10 seconds after user decision in local smoke; two-worker approval resume tests produce no duplicate execution attempts or terminal tool events across 50 consecutive attempts.

**Constraints**: One active run per thread remains enforced. Run events remain the user-visible execution contract. Tool arguments are untrusted and must be schema-validated. Event/result/UI payloads must be redacted. Tools requiring approval must not auto-execute. M7 permits only no-side-effect internal tools and implements only `runtime.get_current_time` as the MVP executable tool. Shell, file read/write, arbitrary network, MCP, browser automation, multi-agent orchestration, long-term memory/RAG, secret exposure, raw provider payload persistence, and approval bypass are out of scope. M6 worker pipeline should be extended only through minimal block/resume interfaces.

**Scale/Scope**: Local-development M7 vertical slice; one local API process may host local workers; one local PostgreSQL instance; fixed local development user; one executable allowlisted tool; exactly one executable tool call per run for MVP, with duplicate ids or multiple simultaneous tool calls failing safely; full multi-step tool-result model continuation is designed as a boundary but may be deferred.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. M7 uses Loomi's own Tool Call, Tool Definition, Approval Decision, Run Event, and Worker Block/Resume language and does not copy external product expression, branding, private interfaces, or non-public structures.
- **II. Runnable Vertical Slices**: PASS. The plan produces an end-to-end slice: model requests an allowlisted tool, Loomi records and displays approval-required state, user approves or denies, worker resumes approved execution, a redacted result/error/cancel state is persisted, and UI/history replay shows the lifecycle.
- **III. Core Flow Before Platform Complexity**: PASS. Tool calling follows the staged roadmap after M5/M6 foundations. Desktop runtime, shell/file/network tools, MCP, browser automation, multi-agent execution, memory/RAG, settings provider management, sandbox, and full tool loops remain deferred.
- **IV. Observable Agent Execution**: PASS. Every tool lifecycle transition is persisted as a run event, replayed by history-first SSE, and surfaced in ToolCallCard, RunRail, and Timeline.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Approval is explicit, arguments are untrusted and schema-validated, only redacted payloads are persisted, and MVP execution is limited to a no-side-effect internal tool.
- **Technical Constraints**: PASS. The plan reuses the existing Go API, PostgreSQL migration workflow, productdata persistence, runtime gateway/worker boundaries, run/event/SSE stream, and frontend runtime adapter instead of adding a broad tool platform.
- **Development Workflow**: PASS. Spec, plan, research, data model, contracts, quickstart, checklist, and tasks are generated under `specs/009-tool-call-approval-core` before implementation.
- **Documentation Definition of Done**: PASS. Implementation must update docs-site architecture, API, runbook, devlog, roadmap current status, and validate docs build.

## Project Structure

### Documentation (this feature)

```text
specs/009-tool-call-approval-core/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── http-m7.openapi.yaml
│   ├── tool-lifecycle-events.md
│   ├── worker-approval-resume.md
│   ├── frontend-tool-ui.md
│   └── docs-update-plan.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
└── loomi-api/
    ├── main.go                         # Reuse M6 worker startup; no new binary required
    └── main_test.go

internal/
├── config/
│   ├── config.go                       # Optional M7 tool approval config defaults if needed
│   └── config_test.go
├── db/
│   ├── readiness.go                    # Require schema version including M7 migration
│   └── readiness_test.go
├── httpapi/
│   ├── runtime.go                      # Approve/deny/read tool-call handlers layered on thread/run scope
│   ├── runtime_test.go
│   └── server.go                       # Route registration and dependency injection
├── productdata/
│   ├── models.go                       # Tool Call, approval/execution status, event constants
│   ├── repository.go                   # tool_calls persistence, idempotent transitions, scoped reads
│   ├── repository_test.go
│   ├── service.go                      # schema validation, redaction, approve/deny, event recording use cases
│   └── service_test.go
└── runtime/
    ├── gateway.go                      # Convert provider tool requests into M7 tool-call records/events
    ├── providers.go                    # Preserve provider normalization; no raw payload persistence
    ├── runner.go                       # Run state and terminal guard integration
    ├── worker.go                       # Observe blocked/resumable tool approval states through M6 loop
    ├── jobs.go                         # Minimal resume/wake behavior for approved tool calls
    ├── pipeline.go                     # Tool lifecycle pipeline event recording if split from service
    ├── tools.go                        # Allowlisted internal tool definitions and executor boundary
    └── *_test.go

migrations/
├── 000006_m7_tool_call_approval.up.sql
└── 000006_m7_tool_call_approval.down.sql

web/src/
├── domain.ts                           # M7 tool lifecycle types and statuses
├── realApiClient.ts                    # Read/approve/deny tool-call API methods
├── runtime/
│   ├── executionAdapter.ts             # Tool event mapping for history/live replay
│   ├── realExecutionAdapter.ts
│   └── *.test.ts
├── components/
│   ├── ToolCallCard.tsx                # Approval controls and lifecycle/result/error rendering
│   ├── RunTimeline.tsx                 # Tool event grouping distinct from model stream
│   ├── RunRail.tsx                     # Tool wait/execution/result summary states
│   └── *.test.tsx
└── state.ts                            # Tool-call view model state and terminal guards

docs-site/src/content/docs/
├── architecture/tool-call-approval.md
├── api/tool-call-approval.md
├── runbooks/local-m7.md
├── devlog/2026-05-24-m7-tool-call-approval.md
└── roadmap/current-status.md
```

**Structure Decision**: M7 keeps durable current-state and idempotency logic in `internal/productdata` because tool calls are owned product execution records. Runtime execution and allowlisted tool definitions live in `internal/runtime` because they are execution behavior behind the gateway/worker boundary. HTTP handlers remain scoped under existing thread/run APIs. Frontend changes extend the existing runtime adapter, ToolCallCard, RunRail, and Timeline instead of creating a separate tool dashboard. Documentation updates are planned under docs-site but are implementation tasks, not part of this planning-only request.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Use a minimal allowlisted internal tool registry.
- Choose `runtime.get_current_time` as the only MVP executable tool.
- Persist a minimal `tool_calls` projection in addition to run events.
- Model approval wait as a worker-blocked run/job state on top of M6, not a new queue.
- Require schema validation before approval and execution.
- Keep full multi-step tool-result model continuation out of MVP while defining the result boundary.
- Use distinct tool event types and UI grouping.
- Redact all persisted arguments, results, and errors.

## Phase 1: Design Summary

Design artifacts are generated:

- [data-model.md](./data-model.md) defines Tool Call, Tool Definition, Approval Decision, Tool Result, Run Event extensions, Worker Block/Resume State, and Tool Result Context Boundary.
- [contracts/http-m7.openapi.yaml](./contracts/http-m7.openapi.yaml) defines draft run-event, tool-call read, approve, and deny API expectations.
- [contracts/tool-lifecycle-events.md](./contracts/tool-lifecycle-events.md) defines event ordering, metadata, terminal states, and frontend grouping contract.
- [contracts/worker-approval-resume.md](./contracts/worker-approval-resume.md) defines M6 worker wait/block/resume behavior and cancellation/recovery semantics.
- [contracts/frontend-tool-ui.md](./contracts/frontend-tool-ui.md) defines ToolCallCard, RunRail, Timeline, and adapter mapping expectations.
- [contracts/docs-update-plan.md](./contracts/docs-update-plan.md) defines required docs-site pages and documentation validation.
- [quickstart.md](./quickstart.md) defines local migration, fake/model tool request, approve/deny, execution, cancellation, replay, browser smoke, validation commands, and documentation validation.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart validates request -> approval_required -> approve/deny -> execute/no-execute -> result/error/cancel -> history replay -> UI grouping.
- **Core Flow Before Platform Complexity**: PASS. Only `runtime.get_current_time` is executable; shell/files/network/MCP/browser/multi-agent/RAG/full loops remain out of scope.
- **Observable Agent Execution**: PASS. All lifecycle states map to run events and UI grouping contracts.
- **Safety/Data Boundaries**: PASS. Schema validation, redaction, idempotent approval, cancellation precedence, and non-goals are explicit in data model and contracts.
- **Documentation**: PASS. docs-site targets and validation commands are specified.

## Complexity Tracking

No constitution violations. A minimal `tool_calls` projection is justified because event-only state would make concurrent approve/deny idempotency and worker resume fragile. No external queue, sandbox, plugin platform, MCP, shell/filesystem/network tool, standalone worker binary, or full multi-step loop is justified for M7.
