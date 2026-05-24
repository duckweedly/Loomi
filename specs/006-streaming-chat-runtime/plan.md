# Implementation Plan: Streaming Chat Runtime

**Branch**: `006-streaming-chat-runtime` | **Date**: 2026-05-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/006-streaming-chat-runtime/spec.md`

## Summary

This feature turns the existing frontend runtime skeleton into a more realistic LLM streaming chat experience. The implementation should reuse the current runtime adapter, selected-thread state, run/event/SSE boundary, assistant draft model, Chat Canvas, Composer, Run Timeline/Run Rail, and Thread Sidebar surfaces while adding visible streaming assistant bubbles, grouped timeline/debug events, clearer backend capability status, complete composer recovery actions, and smoother thread/message loading/error states.

## Technical Context

**Language/Version**: TypeScript with React and Vite under `web/`; Bun 1.3+ for frontend tests/build; Go backend changes are out of scope for this frontend planning slice.

**Primary Dependencies**: Existing React/Vite frontend, existing `apiClient.ts` mode selection, existing `realApiClient.ts` SSE/run-event integration, existing `web/src/runtime/` execution adapter boundary, existing Chat Canvas, Composer, Run Timeline/Run Rail, Thread Sidebar, and existing Bun test style.

**Storage**: Existing frontend thread/message/run/event state only. No new database tables, migrations, local storage, or provider credential storage are required by this feature. If implementation discovers the current state shape cannot preserve regenerated assistant attempts, the task should add frontend state fields only after updating this plan/spec.

**Testing**: Unit and component tests with `bun test ./web/src/**/*.test.ts "web/vite.config.test.ts"`; production build with `bun run --cwd web build`; browser smoke through `bun run --cwd web dev` or desktop shell; docs validation with `bun run --cwd docs-site build` when docs-site pages are updated.

**Target Platform**: Local web renderer and Electron development shell on macOS/Darwin; localhost browser-compatible development environment.

**Project Type**: Frontend runtime UX slice for Loomi's web/desktop-feeling shell.

**Performance Goals**: Pending assistant bubble appears within 500 ms of the first assistant-output signal in local validation; message layout remains stable during draft growth; execution mode or capability problem can be identified from the UI within 5 seconds; timeline grouping handles lifecycle/model/worker/error event mixes without hiding the current run outcome.

**Constraints**: No new model provider, worker queue, tool execution, desktop plugin, authentication flow, or broad state-management framework. Configured real API mode must not silently fall back to mock execution. Mock/local/deferred/unavailable modes must be labeled honestly. Regenerate must preserve the previous assistant response and add a new assistant attempt. One selected thread may have at most one pending or active assistant run blocking composer send.

**Scale/Scope**: One selected thread at a time in Chat mode, one active/pending run per selected thread, future-proof grouping for richer M5/M6 events, and status display for mock/local simulated/real/unavailable/recovering modes. Multi-device synchronization, offline composition, real provider setup, worker queue ownership, and Work-mode-specific runtime semantics remain out of scope.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. The feature deepens Loomi's own run/message/timeline language and does not copy another product's brand, icons, copy, or private interaction patterns.
- **II. Runnable Vertical Slices**: PASS. Each user story can be demonstrated independently through frontend-visible states: streaming draft, grouped timeline, capability status, composer recovery, and thread/message state handling.
- **III. Core Flow Before Platform Complexity**: PASS. The feature stays on the web chat timeline and run/event layer. It explicitly does not pull forward LLM provider implementation, tool calling, worker queues, desktop plugins, activity recording, or multi-agent behavior.
- **IV. Observable Agent Execution**: PASS. The plan increases observability by making model deltas, lifecycle, worker/job, error, retry, cancellation, stream disconnect, and recovery states user-visible through chat and timeline surfaces.
- **V. Safety, Permissions, and Data Boundaries**: PASS. The feature exposes unavailable/setup/provider/stream states honestly and does not add new external write operations, secret handling, tool permissions, or hidden backend fallbacks.
- **Technical Constraints**: PASS. The plan reuses existing frontend boundaries and Bun validation. No new dependencies or broad abstractions are required.
- **Development Workflow**: PASS. The feature has a spec, clarification, this plan, research, data model, contracts, quickstart, and clear docs validation expectations.
- **Documentation Definition of Done**: PASS. Implementation must update relevant docs-site pages for changed UI flows, event model, runtime status, and smoke validation.

## Project Structure

### Documentation (this feature)

```text
specs/006-streaming-chat-runtime/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── backend-capability-status.md
│   ├── chat-canvas-streaming.md
│   ├── composer-actions.md
│   ├── runtime-event-groups.md
│   └── thread-message-states.md
└── tasks.md              # Created by /speckit-tasks, not by /speckit-plan
```

### Source Code (repository root)

```text
web/src/
├── apiClient.ts                         # Existing mock/real mode selection boundary
├── domain.ts                            # Existing Thread, Message, Run, RunEvent, AssistantDraft domain types
├── realApiClient.ts                     # Existing real API run/event/SSE mapping seam
├── state.ts                             # Existing selected-thread, message, run, stream, and backend capability state
├── runtime/
│   ├── chatCanvasState.ts               # Visible chat canvas state derivation
│   ├── executionAdapter.ts              # Shared runtime adapter contract
│   ├── mockExecutionAdapter.ts          # Deterministic local/mock runtime scripts
│   ├── realExecutionAdapter.ts          # Real API runtime adapter wrapper
│   └── runtimeScripts.ts                # Mock event/draft scenarios
├── components/
│   ├── ChatCanvas.tsx                   # Message list, assistant draft bubble, runtime states, status chips
│   ├── Composer.tsx                     # Send/stop/retry/regenerate/continue interactions
│   ├── RunRail.tsx                      # Timeline/debug event grouping and run controls
│   ├── RunTimeline.tsx                  # Timeline composition around selected run
│   ├── ThreadSidebar.tsx                # Thread list, selection, loading/error affordances
│   └── AgentStateMotion.tsx             # Agent status linkage to selected run/event state
├── runtime/*.test.ts                    # Runtime state, adapter, script, and event mapping tests
├── components/*.test.ts                 # Chat Canvas, Composer, Run Rail, Timeline, Thread Sidebar tests
└── *.test.ts                            # App/state/api integration-style unit tests

docs-site/src/content/docs/
├── architecture/frontend-agent-runtime.md
├── architecture/run-event-sse.md
├── api/run-event-sse.md
├── runbooks/frontend-runtime-smoke.md
├── spec-kit/workflow.md
└── devlog/2026-05-23-streaming-chat-runtime.md
```

**Structure Decision**: Keep the feature inside the existing web frontend and runtime adapter boundary. The implementation should extend focused state derivation, event grouping, and UI contracts instead of creating a new runtime layer or replacing the adapter/API split. Documentation lives in the existing Starlight docs-site sections because this feature changes runtime behavior, event interpretation, UI flows, and validation runbooks.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Render assistant draft as a first-class transient chat bubble derived from the selected run's assistant draft until finalization.
- Group timeline/debug events by semantic purpose in the frontend so richer backend events can arrive without cluttering the UI.
- Derive backend capability display from mode, adapter capability, stream state, provider/setup errors, and run recovery state without hiding unavailable modes.
- Treat Composer actions as selected-thread run transitions with explicit guards for active runs, failed runs, retry, stop, regenerate, and continuation.
- Preserve previous assistant responses on regenerate and add a new assistant attempt for observability and auditability.
- Reuse existing thread/message/run/event concepts and avoid data model or backend changes unless implementation proves the current frontend state cannot represent required UI states.

## Phase 1: Design Summary

Design artifacts are generated:

- [data-model.md](./data-model.md) defines Thread, Message, Assistant Draft Bubble, Assistant Attempt, Run, Run Event, Timeline Event Group, Backend Capability Status, Composer Input/Action State, and Chat Canvas Presentation State.
- [contracts/chat-canvas-streaming.md](./contracts/chat-canvas-streaming.md) defines visible assistant draft and finalization behavior.
- [contracts/runtime-event-groups.md](./contracts/runtime-event-groups.md) defines event grouping, severity, and mapping rules.
- [contracts/backend-capability-status.md](./contracts/backend-capability-status.md) defines mode/capability status values and precedence.
- [contracts/composer-actions.md](./contracts/composer-actions.md) defines send, stop, retry, regenerate, and continue interaction contracts.
- [contracts/thread-message-states.md](./contracts/thread-message-states.md) defines selected-thread, loading, error, retry, history, and recovery presentation contracts.
- [quickstart.md](./quickstart.md) defines automated validation, browser smoke, and docs validation for the slice.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart covers streaming success, failure, stop, reconnect/recovery, event grouping, capability statuses, composer actions, and thread/message states.
- **Core Flow Before Platform Complexity**: PASS. Contracts stay frontend-first and explicitly defer providers, workers, tools, desktop plugins, and cross-device/offline behavior.
- **Observable Agent Execution**: PASS. Design treats run events, model deltas, retry/cancellation, provider errors, and regenerated attempts as explainable timeline/chat state rather than hidden control flow.
- **Safety/Data Boundaries**: PASS. Status contracts separate backend unavailable, stream disconnected, provider unavailable, and model failure; no secrets or external writes are introduced.
- **Documentation**: PASS. Docs targets and validation commands are identified.

## Complexity Tracking

No constitution violations. No new dependency, backend service, database migration, provider integration, worker queue, desktop runtime layer, or global state-management framework is justified for this feature.
