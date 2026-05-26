# Implementation Plan: M22 Bounded Agent Loop + Todo Foundation

**Branch**: `017-mcp-approval-gated-execution` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/030-bounded-agent-loop-todo-foundation/spec.md`

## Summary

M22 removes the current one-tool-call ceiling by adding a bounded, sequential continuation loop for Work mode runs. The implementation reuses the existing Gateway, worker, RunContext, ToolBroker, approval, and run-event model: each provider-requested tool call still pauses independently for approval, only one tool may be pending/executing at a time, and a small per-run loop limit prevents unbounded autonomy. M22 also adds safe todo state as replayable Work mode metadata so operators can see the agent's current plan before mutation and shell tools arrive.

## Technical Context

**Language/Version**: Go backend, TypeScript/React frontend, Starlight docs site.

**Primary Dependencies**: Existing `internal/productdata` run/event/tool-call/job context, `internal/runtime` Gateway/Worker/QueuedRunRouter/ToolBroker, current Bun/Vite web stack. No new third-party runtime dependency.

**Storage**: Existing PostgreSQL and in-memory service models for runs, events, tool calls, and background jobs. No new table unless implementation proves existing event metadata cannot replay todo state.

**Testing**: `go test ./...`, backend smoke/unit tests, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, `git diff --check`, plus browser smoke for Work Plan/timeline replay when UI behavior changes.

**Target Platform**: Local Loomi backend and web shell.

**Project Type**: Web application with Go API/worker backend and React frontend.

**Performance Goals**: Bounded loop smoke completes within normal backend test time; event replay remains linear in run event count for the small configured loop limit.

**Constraints**: Sequential only; one pending/executing tool call at a time; per-tool approval required; small loop limit; safe metadata only; no workspace write/edit, shell/code execution, browser automation, web search/fetch, artifact creation, multi-agent, marketplace, or unbounded autonomous loop.

**Scale/Scope**: One Work mode run executing a bounded number of sequential tool calls against currently enabled tools. M21 workspace read tools are the primary fixture tools for this slice.

## Constitution Check

- **I. Mechanism Parity, Original Expression**: PASS. Arkloop is a mechanism reference only; Loomi keeps its own Gateway, ToolBroker, RunContext, event, todo, and UI language.
- **II. Runnable Vertical Slices**: PASS. The MVP is a smokeable `workspace.glob -> approve -> result -> workspace.read -> approve -> result -> final` run plus UI replay.
- **III. Core Flow Before Platform Complexity**: PASS. This deepens the existing run/event/tool path before write, shell, browser, web, desktop, or multi-agent work.
- **IV. Observable Agent Execution**: PASS. The feature is centered on persisted loop/todo events and timeline/debug replay.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Every tool call remains approval-gated; loop limit and redaction are required.

Post-design check: PASS. No constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/030-bounded-agent-loop-todo-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── bounded-agent-loop.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/productdata/
├── models.go
├── service.go
├── repository.go
└── *_test.go

internal/runtime/
├── gateway.go
├── queued_runner.go
├── tool_broker.go
└── *_test.go

internal/httpapi/
└── *_smoke_test.go

web/src/
├── workModeProjection.ts
├── components/WorkPlanView.tsx
├── components/RunRail.tsx
├── domain.ts
└── *_test.tsx

docs-site/src/content/docs/
├── architecture/
├── api/
├── runbooks/
├── devlog/
├── roadmap/
└── spec-kit/
```

**Structure Decision**: Keep loop state and todo metadata on the existing run/event/tool-call boundary. Backend loop control belongs in `internal/runtime` because Gateway/QueuedRunRouter currently own provider continuation and approved tool execution. Durable validation and redaction belong in `internal/productdata` because both memory and Postgres services must replay the same safe metadata. Frontend projection stays in Work mode/timeline components rather than adding a separate task system.

## Phase 0: Research

See [research.md](./research.md).

## Phase 1: Design & Contracts

See [data-model.md](./data-model.md), [contracts/bounded-agent-loop.md](./contracts/bounded-agent-loop.md), and [quickstart.md](./quickstart.md).

## Complexity Tracking

No constitution violations.
