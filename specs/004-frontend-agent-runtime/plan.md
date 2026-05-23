# Implementation Plan: Frontend Agent Runtime Skeleton

**Branch**: `004-frontend-agent-runtime` | **Date**: 2026-05-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/004-frontend-agent-runtime/spec.md`

## Summary

M3.5 turns Loomi's frontend from a mostly static shell into a mock-driven Agent runtime skeleton while the backend M4/M5 run/event/SSE and LLM gateway are still under construction. The plan adds a clear Chat Canvas state model, deterministic success/failure mock run scripts, one shared execution adapter boundary, and visible linkage between Chat Canvas, Run Timeline, and the Agent state motion badge. The implementation remains frontend-only and must not introduce real run persistence, SSE, LLM calls, workers, tools, or production runtime behavior.

## Technical Context

**Language/Version**: TypeScript/React/Vite in `web/`; Bun 1.3+ for tests/build; Go backend remains out of scope for this feature except for real API capability honesty.

**Primary Dependencies**: Existing React/Vite frontend, existing `fetch`-backed API seam, existing mock/real client split, existing Run Rail/Chat Canvas/AgentStateMotion components, existing Bun test style.

**Storage**: In-memory frontend mock state only for runtime scripts; no new database tables, migrations, local storage, or persisted run/event data.

**Testing**: `bun test ./web/src/**/*.test.ts "web/vite.config.test.ts"` for unit/contract-style tests; `bun run --cwd web build`; browser smoke for mock success/failure flow and backend capability unavailable state; docs validation with `bun run --cwd docs-site build` when docs are updated.

**Target Platform**: Local web renderer and Electron development shell on macOS/Darwin; localhost browser-compatible environments.

**Project Type**: Frontend product shell milestone with a future backend adapter boundary.

**Performance Goals**: User-submitted messages appear in the Chat Canvas within 300 ms in mock mode; mock script steps should feel immediate but observable, with deterministic timings suitable for tests; UI remains responsive while scripts run.

**Constraints**: No real run/event/SSE implementation; no LLM gateway; no worker queue; no tool execution; no new frontend framework; no hidden fallback from configured real API mode to mock run behavior; product UI microcopy stays sparse; learning documentation is Chinese and explanatory.

**Scale/Scope**: One selected Chat thread, one active mock run per selected thread for the first slice, deterministic success/failure/stopped scripts, mode-specific recent thread lists preserved. Work-mode runtime behavior remains deferred.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. The runtime skeleton uses Loomi's own state names, Chinese learning documentation, and existing Loomi visual language; it does not copy another product's expression layer.
- **II. Runnable Vertical Slices**: PASS. The feature is demonstrable as a frontend-only vertical slice: submit a mock Chat message, observe Chat Canvas state, Timeline events, Agent badge motion, and success/failure completion.
- **III. Core Flow Before Platform Complexity**: PASS. This is explicitly a frontend M3.5 bridge before M4/M5; it does not pull forward backend run persistence, SSE, LLM gateway, tools, workers, desktop runtime, attachments, RAG, or plugins.
- **IV. Observable Agent Execution**: PASS. The feature deepens observability by making mock run lifecycle events visible and linking them to chat state and agent motion.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Real API mode must honestly show backend capability unavailable when run/event support is absent and must not silently fallback to mock execution.
- **Technical Constraints**: PASS. The plan reuses existing web boundaries and introduces focused frontend state/adapters rather than frameworks or broad abstractions.
- **Development Workflow**: PASS. The feature has a Spec Kit spec and this plan defines research, data model, contracts, quickstart, and docs targets before task generation.
- **Documentation Definition of Done**: PASS. Implementation must update docs-site architecture/devlog/runbook or spec-kit pages for the state model and adapter boundary, then validate docs build.

## Project Structure

### Documentation (this feature)

```text
specs/004-frontend-agent-runtime/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── execution-adapter.md
│   ├── runtime-events.md
│   └── chat-canvas-states.md
└── tasks.md
```

### Source Code (repository root)

```text
web/src/
├── apiClient.ts
├── domain.ts
├── state.ts
├── mockApiClient.ts
├── realApiClient.ts
├── runtime/
│   ├── executionAdapter.ts
│   ├── mockExecutionAdapter.ts
│   ├── realExecutionAdapter.ts
│   ├── chatCanvasState.ts
│   └── runtimeScripts.ts
├── components/
│   ├── ChatCanvas.tsx
│   ├── Composer.tsx
│   ├── RunRail.tsx
│   ├── AgentStateMotion.tsx
│   └── ThreadSidebar.tsx
├── *.test.ts
└── components/*.test.ts

docs-site/src/content/docs/
├── architecture/frontend-agent-runtime.md
├── runbooks/frontend-runtime-smoke.md
├── spec-kit/workflow.md
└── devlog/2026-05-23-m3-5-frontend-agent-runtime.md
```

**Structure Decision**: M3.5 creates a `web/src/runtime/` boundary because runtime execution semantics now have a clearer responsibility than the existing thread/message API clients. `apiClient.ts`, `mockApiClient.ts`, and `realApiClient.ts` remain responsible for M3 durable thread/message data. New runtime adapters own mock scripts and future run/event capability so mock and real can share one UI state machine. Chat Canvas state derivation is kept as a pure module so tests can cover all visible states without needing browser rendering.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Use a frontend-only runtime adapter boundary for M3.5 rather than extending M3 `ApiClient` with future backend semantics.
- Model Chat Canvas states as a pure derived state so every visible state is testable without UI rendering.
- Use deterministic mock scripts for success/failure/stopped flows so screenshots, tests, and learning docs remain reproducible.
- Treat real API missing run/event support as `backend-unavailable`, not as a mock fallback.

## Phase 1: Design Summary

Design artifacts are generated:

- [data-model.md](./data-model.md) defines Chat Canvas State, Runtime Run, Runtime Event, Runtime Script, Assistant Draft, Execution Adapter, Backend Capability State, and Stale Event Guard.
- [contracts/execution-adapter.md](./contracts/execution-adapter.md) defines the shared mock/real frontend execution boundary.
- [contracts/runtime-events.md](./contracts/runtime-events.md) defines user-visible runtime event vocabulary and state transitions.
- [contracts/chat-canvas-states.md](./contracts/chat-canvas-states.md) defines the visible Chat Canvas state contract.
- [quickstart.md](./quickstart.md) defines local validation, browser smoke, and docs validation for the M3.5 slice.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart demonstrates success, failure, stopped, no-thread, empty-thread, loading/error, and backend-unavailable states in mock/real-mode conditions.
- **Core Flow Before Platform Complexity**: PASS. Contracts are frontend-only and explicitly defer persisted run/event/SSE, LLM, tools, workers, desktop runtime, and platform features.
- **Observable Agent Execution**: PASS. Runtime events are first-class user-visible milestones that drive Timeline, Chat Canvas, and AgentStateMotion.
- **Safety/Data Boundaries**: PASS. Real mode does not silently use mock execution when backend capability is absent.
- **Documentation**: PASS. Docs targets and validation commands are identified.

## Complexity Tracking

No constitution violations. No additional runtime framework, state-management library, backend service, event transport, model provider, tool system, worker, or desktop runtime abstraction is justified for M3.5.
