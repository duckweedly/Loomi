# Implementation Plan: Real Usage Readiness

**Branch**: `040-real-usage-readiness` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/040-real-usage-readiness/spec.md`

## Summary

Close UI-02 by making Loomi's existing UI states honest and usable for real work: clear Mock/Real/provider status, honest Work folder limitation, Work task panel that only uses real plan metadata, human-first tool event labels, approval-blocked visibility, simplified sidebar actions, and mode-specific composer copy. The feature is frontend presentation and documentation only; it does not change backend/runtime/provider/tool execution/database behavior or add capabilities.

## Technical Context

**Language/Version**: TypeScript, React, Vite, CSS; Go backend validation unchanged.

**Primary Dependencies**: Existing React app under `web/`, `@lobehub/ui`, `lucide-react`, Bun test/build scripts, Starlight docs site.

**Storage**: No persisted storage changes. Shell behavior remains session-local UI state.

**Testing**: `bun test --cwd web`, focused component tests, `bun run --cwd web build`, `go test ./...`, `bun run --cwd docs-site build`, `git diff --check`, browser smoke.

**Target Platform**: Web-first desktop shell with existing Electron preload boundary unchanged.

**Project Type**: Web/desktop-feeling application shell.

**Performance Goals**: Use simple derived state and render-time formatting only; no new polling, subscriptions, provider calls, or runtime work.

**Constraints**: No backend/runtime/provider/tool/database/M38 changes. No new tool capability. No fake enabled controls for missing Work folder picker. No pixel-level redesign.

**Scale/Scope**: Existing UI components and tests: `App.tsx`, `state.ts`, `ChatCanvas.tsx`, `Composer.tsx`, `WorkPlanView.tsx`, `RunRail.tsx`, `ThreadSidebar.tsx`, `ToolCallCard.tsx`, `workModeProjection.ts`, `styles.css`, and focused tests.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Mechanism Parity, Original Expression**: Pass. This feature uses Loomi-specific copy and mechanics, not copied brand or private expression.
- **Runnable Vertical Slices**: Pass. The slice is visible in automated component tests and browser smoke.
- **Core Flow Before Platform Complexity**: Pass. It improves existing UI readiness and explicitly avoids M38/activity recorder or new tool capability.
- **Observable Agent Execution**: Pass. It makes run/tool/approval state easier to understand.
- **Safety, Permissions, and Data Boundaries**: Pass. Missing folder picker is shown honestly; no new file access, provider, or tool execution path is added.

## Project Structure

### Documentation (this feature)

```text
specs/040-real-usage-readiness/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code

```text
web/src/
├── App.tsx
├── state.ts
├── workModeProjection.ts
├── styles.css
└── components/
    ├── ChatCanvas.tsx
    ├── Composer.tsx
    ├── ThreadSidebar.tsx
    ├── ToolCallCard.tsx
    ├── WorkPlanView.tsx
    └── RunRail.tsx
```

**Structure Decision**: Keep all changes inside existing UI state and component boundaries. Use small pure helper functions for labels and projections rather than adding a new state manager or runtime layer.

## Phase 0: Research

- **Decision**: Treat Mock/Real/provider state as UI copy derived from existing `dataSourceMode`, provider capability, and backend capability state.
  **Rationale**: These signals already exist and changing runtime selection would be out of scope.
  **Alternatives considered**: Add a new readiness API; rejected because the feature must not change backend/API behavior.

- **Decision**: Show Work folder as an honest limitation state instead of a disabled button.
  **Rationale**: No safe directory picker contract exists in this slice.
  **Alternatives considered**: Wire Electron or browser file picker; rejected as new capability/safety scope.

- **Decision**: WorkPlanView only treats explicit work metadata as plan/todos/artifacts.
  **Rationale**: User messages and generic tool events are not plan metadata and should not be presented as a plan.
  **Alternatives considered**: Continue deriving steps from messages; rejected because it misrepresents real task state.

- **Decision**: RunRail and ToolCallCard use human-first labels with technical detail retained in secondary text.
  **Rationale**: Real users need readable progress while developers still need raw tool details.
  **Alternatives considered**: Hide raw tool names entirely; rejected because debugging and auditability need them.

## Phase 1: Design

### Data Model

No persisted data model changes. Existing `Thread`, `Run`, `RunEvent`, `ToolCall`, provider capability, and Work projection types are reused.

### Contracts

No backend/API contract changes.

Frontend UI contract:

- `ChatCanvas` displays runtime reality, provider guidance, approval blocked notice, WorkPlanView, and mode-specific Composer configuration.
- `Composer` accepts mode and data source labels and renders a limitation state for missing Work folder selection.
- `workModeProjection` only uses explicit work metadata for steps/todos/artifacts.
- `RunRail` formats tool-call events with human-first labels and details.
- `ToolCallCard` formats tool cards with human-first tool labels.
- `ThreadSidebar` owns a session-local search query over currently supplied mode-filtered threads.
- `state.createThread` accepts the selected mode for future thread creation without changing backend semantics.

### Quickstart / Smoke

1. Start the local web app.
2. Open the app in a browser.
3. Verify Mock/demo state is visible.
4. Type in Chat and Work composers.
5. Verify Composer does not show fake Work folder, attachment, persona/provider selector, or voice controls.
6. Confirm the sidebar does not show the removed search field or bottom action cluster.
7. Inspect RunRail tool events and confirm human-first labels.
8. Confirm approval-blocked state shows waiting copy and Approve/Deny/Stop.
9. Open Settings > Providers and Settings > Tools.
10. Verify browser console error count is 0 and save a screenshot.

## Constitution Check - Post Design

Pass. The design remains frontend-only, improves observable execution, does not add permissions or tools, and includes automated plus browser validation.

## Complexity Tracking

No constitution violations.
