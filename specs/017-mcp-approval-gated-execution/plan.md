# Implementation Plan: MCP Approval-Gated Execution

**Branch**: `main` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/017-mcp-approval-gated-execution/spec.md`

## Summary

M12 turns M11's discovered local stdio MCP ToolSpec candidates into a minimal approval-gated execution loop. A provider-requested namespaced MCP tool must pass M11 discovery state and M10 persona allowed-tools resolution, enter the existing M7 tool-call projection and approve/deny flow, execute exactly once through the M6 worker under ownership/lease/cancel guards, persist only redacted arguments/result/error/audit run events, and perform one M7-style provider continuation. The slice explicitly defers remote MCP, OAuth, marketplace/plugin install, DB-managed MCP server admin, shell/filesystem/browser automation, automatic execution, complex sandboxing, admin UI, and multi-tool loops.

## Technical Context

**Language/Version**: Go backend/runtime/worker; TypeScript/React frontend for replay labels; Starlight docs-site with Bun validation.

**Primary Dependencies**: Existing `internal/productdata` run/thread/tool-call/persona boundaries, `internal/runtime` ToolSpec/ToolRegistry and provider continuation model, M7 approval APIs and tool-call projection, M6 worker/job ownership and cancellation, M9 RunContext/pipeline events, M10 persona allowed-tools snapshots, M11 MCP discovery/candidate mapping, SSE/history replay, and existing Timeline/debug grouping.

**Storage**: Reuse or extend the existing tool-call projection and run-event metadata with safe MCP fields. Persist only safe identity, namespaced tool name, discovery candidate version or schema hash, redacted argument summary/hash, approval/execution status, redacted result/error summaries, timestamps, and ownership-safe audit metadata. Do not persist raw env, args, command paths, stdout/stderr, tokens, credentials, secret-looking paths, raw result payloads, file contents, shell output, browser state, or desktop captured data.

**Testing**: Go backend tests for resolution, projection, approve/deny, redaction, continuation context, and run events. Go worker tests for ownership, lease, stop/cancel, retry/recovery, stdio lifecycle, and no duplicate execution. Frontend replay tests for live-style and history events. Docs validation with `bun run --cwd docs-site build`.

**Target Platform**: Local Loomi development environment with Go API/worker, local PostgreSQL where current tool-call/run-event tests require it, web renderer, and docs-site.

**Project Type**: Local web application plus Go API/backend runtime and durable product data.

**Performance Goals**: Approval projection remains idempotent for repeated provider requests. MCP stdio execution is bounded by per-call timeout. Worker recovery avoids duplicate process startup. Continuation remains one extra provider call.

**Constraints**: Only already-discovered local stdio MCP candidates. No remote MCP, HTTP/SSE/OAuth MCP, marketplace/plugin install, admin UI, DB-managed server config, automatic execution, shell/filesystem/browser automation, complex sandbox, or multi-step tool loop. MCP server output is untrusted data.

**Scale/Scope**: One namespaced MCP tool call per run, one approval decision, one worker-owned stdio invocation, one redacted result/error, one provider continuation, and replayable events.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. The plan uses Loomi's ToolSpec, RunContext, worker, approval, and Timeline terminology and does not copy external product expression.
- **II. Runnable Vertical Slices**: PASS. The MVP is a demonstrable approval-blocked MCP call that executes once after approval and returns one continuation.
- **III. Core Flow Before Platform Complexity**: PASS. M12 follows M7/M9/M10/M11 foundations and explicitly defers remote MCP, marketplace, sandbox, admin UI, automation tools, and multi-tool loops.
- **IV. Observable Agent Execution**: PASS. Approval, denial, execution start/success/failure, cancellation, and continuation are persisted as replayable redacted events.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Execution is approval-gated, scoped, audited, redacted, and treats MCP output as untrusted data.
- **Technical Constraints**: PASS. The plan reuses existing productdata/runtime/worker/httpapi/web/docs boundaries and avoids new platform layers.
- **Development Workflow**: PASS. The design artifacts have moved through implementation; follow-up changes must keep spec, tasks, docs, and validation evidence in sync.
- **Documentation Definition of Done**: PASS. Tasks include docs-site architecture/API/runbook/roadmap/devlog/spec-kit updates and docs build.

## Project Structure

### Documentation (this feature)

```text
specs/017-mcp-approval-gated-execution/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── mcp-approval-gate.md
│   ├── mcp-worker-execution.md
│   ├── mcp-continuation.md
│   └── mcp-redaction-events.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── productdata/
│   ├── models.go                 # tool-call projection fields/statuses if extension is needed
│   ├── repository.go             # scoped projection/event persistence and idempotent reads/writes
│   ├── service.go                # approve/deny and persona/candidate policy checks if owned here
│   └── *_test.go
├── runtime/
│   ├── tools.go                  # MCP candidate executable resolution through ToolRegistry
│   ├── queued_runner.go          # approval-gated execution orchestration facade
│   ├── mcp_stdio.go              # bounded local stdio process invocation for approved calls
│   ├── mcp_redaction.go          # argument/result/error redaction helpers
│   ├── gateway.go                # redacted MCP result continuation context
│   ├── queued_runner.go          # ownership/lease/cancel guards around approved MCP execution
│   ├── worker.go                 # queued worker processing entrypoint
│   └── *_test.go
└── httpapi/
    ├── runtime.go                # existing tool-call read/approve/deny surfaces if metadata expands
    └── runtime_test.go

web/src/
├── realApiClient.ts
├── runtime/
│   ├── realExecutionAdapter.ts
│   └── runtimeEventGroups.ts
└── components/
    ├── RunTimeline.tsx
    └── RunRail.tsx

docs-site/src/content/docs/
├── architecture/mcp-approval-gated-execution.md
├── api/mcp-approval-gated-execution.md
├── api/tool-call-approval.md
├── runbooks/local-m12-mcp-approval-execution.md
├── roadmap/current-status.md
├── spec-kit/workflow.md
└── devlog/2026-05-25-m12-mcp-approval-gated-execution.md
```

**Structure Decision**: Approval and persistence reuse the M7 tool-call projection and productdata/run-event boundary. MCP candidate resolution stays near runtime ToolRegistry and M10/M11 context preparation. Stdio process execution is worker-owned inside the existing `internal/runtime` runner/worker boundary and bounded behind a runtime facade. Frontend changes are limited to existing event replay and Timeline/debug group mapping.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Reuse M7 approval projection and approve/deny APIs as the only entry to MCP execution.
- Gate execution by both M11 discovery state and M10 persona allowed-tools snapshot.
- Treat MCP execution as at-most-once within M12 worker recovery constraints.
- Bound stdio process lifecycle and redact all process details before persistence or continuation.
- Continue the provider exactly once with redacted MCP result data.
- Defer remote MCP, OAuth, marketplace, admin UI, DB-managed servers, sandbox, automation, and multi-tool loops.

## Phase 1: Design Summary

- [data-model.md](./data-model.md) defines the projection, attempt, stdio invocation, redacted result, continuation, and audit/event entities.
- [contracts/mcp-approval-gate.md](./contracts/mcp-approval-gate.md) defines the approval entry contract.
- [contracts/mcp-worker-execution.md](./contracts/mcp-worker-execution.md) defines worker ownership, lease, cancellation, retry/recovery, and process lifecycle behavior.
- [contracts/mcp-continuation.md](./contracts/mcp-continuation.md) defines single continuation input/output and tool-loop rejection.
- [contracts/mcp-redaction-events.md](./contracts/mcp-redaction-events.md) defines event metadata and forbidden fields.
- [quickstart.md](./quickstart.md) defines validation commands and smoke expectations for the implementation session.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. US1 alone blocks MCP execution behind approval; US1+US2 executes once safely; US1+US2+US3 completes the one-continuation loop.
- **Core Flow Before Platform Complexity**: PASS. The design does not add remote MCP, OAuth, sandbox, marketplace, admin UI, DB-managed server config, or multi-tool loops.
- **Observable Agent Execution**: PASS. Events cover approval, execution, worker, cancellation, failure, continuation, and replay states.
- **Safety/Data Boundaries**: PASS. Raw MCP process data never crosses persistence/UI/continuation boundaries.
- **Documentation**: PASS. docs-site update targets and docs build are included in tasks.

## Complexity Tracking

No constitution violations. The small MCP executor facade is justified because M12 needs one approved local stdio invocation while keeping M7 approval, M6 worker ownership, M9 RunContext, M10 persona, and M11 discovery boundaries intact.
