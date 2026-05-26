# Implementation Plan: Formal Interface Shell Redesign

**Branch**: `039-formal-interface-shell-redesign` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/039-formal-interface-shell-redesign/spec.md`

## Summary

Redesign Loomi's first-round interface shell into a light desktop-feeling app: soft outer background, rounded main window, narrow native-like sidebar, white chat canvas, centered content column, and fixed bottom composer. Preserve existing Chat/Work state, provider warnings, Stop visibility, Settings, Tools, and RunRail. Avoid backend/runtime/provider/tool/memory/database changes.

## Technical Context

**Language/Version**: TypeScript, React, Vite, CSS; Go backend validation unchanged.

**Primary Dependencies**: Existing React app under `web/`, `@lobehub/ui`, `lucide-react`, Bun test/build scripts, Starlight docs site.

**Storage**: N/A for this UI-only slice.

**Testing**: `bun test --cwd web`, `bun run --cwd web build`, `go test ./...`, `bun run --cwd docs-site build`, `git diff --check`, browser smoke.

**Target Platform**: Web-first desktop shell with responsive narrow viewport compatibility.

**Project Type**: Web/desktop-feeling application shell.

**Performance Goals**: No new runtime dependencies or expensive rendering loops; layout should remain CSS-driven.

**Constraints**: Preserve existing runtime flows and controls; do not copy reference product branding or private expression; do not advance M38/activity recorder; no backend or database changes.

**Scale/Scope**: First-round shell redesign across existing App, Sidebar, ChatCanvas, Composer, RunRail compatibility, docs, and tests.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Mechanism Parity, Original Expression**: Pass. The reference is used only for layout proportions and desktop-shell mechanics; Loomi wording, identity, and controls remain its own.
- **Runnable Vertical Slices**: Pass. The slice is validated through browser-visible Chat/Work send paths, Stop visibility, Settings > Tools, console check, and automated commands.
- **Core Flow Before Platform Complexity**: Pass. This is desktop-feeling shell work and explicitly avoids backend/runtime/tool/provider/memory expansion.
- **Observable Agent Execution**: Pass. Existing provider warnings, run context, Stop, and RunRail remain visible.
- **Safety, Permissions, and Data Boundaries**: Pass. No new permissions, file access, database changes, or external write operations.

## Project Structure

### Documentation (this feature)

```text
specs/039-formal-interface-shell-redesign/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```text
web/src/
├── App.tsx
├── styles.css
├── useWorkspaceShellState.ts
└── components/
    ├── ThreadSidebar.tsx
    ├── ChatCanvas.tsx
    ├── Composer.tsx
    └── RunRail.tsx

docs-site/src/content/docs/
├── architecture/
├── devlog/
├── runbooks/
└── spec-kit/
```

**Structure Decision**: Keep the existing React/CSS component structure. Use CSS layout tokens and small prop additions for shell/sidebar controls instead of introducing new state management or dependencies.

## Phase 0: Research

- **Decision**: Use existing `web/src/styles.css` CSS variables and component classes for the shell redesign.
  **Rationale**: The requested work is visual/layout-heavy and the project already centralizes shell styling in one CSS file.
  **Alternatives considered**: Add a design-token package or component library theme override; rejected as too broad for a first-round UI shell.

- **Decision**: Keep Settings, Tools, and RunRail structurally unchanged, adjusting only shell-level containment if needed.
  **Rationale**: User explicitly excluded redesigning these surfaces.
  **Alternatives considered**: Redesign every panel for visual consistency; rejected as too much scope.

- **Decision**: Default the session shell to light theme for this UI-01 slice.
  **Rationale**: The target direction is a light desktop app and existing shell state is session-local.
  **Alternatives considered**: Keep dark default and rely on manual theme toggle; rejected because the first open would not match the requested test direction.

## Phase 1: Design

### Data Model

No persisted data model changes. Existing thread, run, provider, persona, and shell state are reused. Shell state changes remain session-local.

### Contracts

No backend/API contract changes. UI contract for this feature:

- `ThreadSidebar` receives mode selection affordances and keeps thread selection/create/archive/rename/settings callbacks.
- `ChatCanvas` keeps existing runtime props and remains responsible for provider warning, message history, Work projection, active tool calls, and composer.
- `Composer` keeps submit, Stop, retry, regenerate, persona, provider label, and Work in Folder affordances.

### Quickstart / Smoke

1. Start the local web app.
2. Open the app in a browser.
3. Verify overall shell proportions against the reference direction without copying brand/expression.
4. In Chat mode, type into the composer.
5. In Work mode, type and submit to start a run.
6. Verify Stop is visible while active.
7. Open Settings > Tools.
8. Verify console error count is 0 and save a screenshot.

## Constitution Check - Post Design

Pass. The plan remains UI-only, uses existing frontend boundaries, preserves observable run controls, includes docs-site updates, and defines validation before completion.

## Complexity Tracking

No constitution violations.
