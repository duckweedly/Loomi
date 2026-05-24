# Implementation Plan: M5.5 Settings Placeholder

**Branch**: `[main]` | **Date**: 2026-05-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/007-settings-placeholder/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build a temporary M5.5 Settings surface in the existing Loomi desktop-feeling web shell. The slice delivers a reference-inspired two-column settings view with Loomi-owned categories and grouped setting cards. Currently supported behavior is limited to session-local settings and read-only runtime/provider status; future platform areas appear as clearly labeled mock placeholders.

## Technical Context

**Language/Version**: TypeScript with React in the existing `web/` app; documentation in Starlight markdown.

**Primary Dependencies**: Existing React/Vite/Bun frontend stack, existing lucide icon usage, existing Loomi state/hooks/components. No new dependency planned.

**Storage**: Session-local React state only for M5.5. No database, backend settings persistence, or browser secret storage.

**Testing**: Bun tests for frontend state/components; Vite/TypeScript build; docs-site build.

**Target Platform**: Loomi desktop-feeling web shell and Electron dev shell.

**Project Type**: Frontend web/desktop UI slice with documentation updates.

**Performance Goals**: Settings opens and becomes identifiable in under 5 seconds during manual smoke; category switching should feel immediate for local UI state.

**Constraints**: Do not expose provider secrets. Do not execute tools, provider calls, connectors, file writes, or backend write operations from placeholder controls. Preserve current workspace context when entering/leaving Settings. Keep real settings scoped to current local/session behavior.

**Scale/Scope**: One Settings view with required categories, General working controls, About/status support, and placeholder/mock panels for deferred areas.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Mechanism Parity, Original Expression**: PASS. The screenshot is layout direction only; plan requires Loomi-owned labels and copy.
- **Runnable Vertical Slices**: PASS. M5.5 produces a visible settings surface and smokeable controls.
- **Core Flow Before Platform Complexity**: PASS. Real provider management, persistence, tools, worker routes, memory/RAG, and activity recorder are explicitly deferred.
- **Observable Agent Execution**: PASS. Runtime/model gateway state is displayed read-only; no hidden execution behavior is introduced.
- **Safety, Permissions, and Data Boundaries**: PASS. Provider secrets stay out of browser settings; placeholders do not execute external actions.
- **Documentation Workflow**: PASS. docs-site updates and docs build are planned as part of done.

## Project Structure

### Documentation (this feature)

```text
specs/007-settings-placeholder/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── settings-ui.md
└── tasks.md
```

### Source Code (repository root)

```text
web/src/
├── App.tsx
├── useWorkspaceShellState.ts
├── state.ts
├── components/
│   ├── ThreadSidebar.tsx
│   ├── SettingsView.tsx
│   ├── SettingsView.test.tsx or SettingsView.*.test.ts
│   ├── settingsMenuItems.ts
│   └── settingsMenuItems.test.ts
└── styles.css

docs-site/src/content/docs/
├── architecture/
├── runbooks/
├── devlog/
└── spec-kit/
```

**Structure Decision**: Implement in the existing `web/` app as a frontend UI slice. Add a Settings-specific component and small state additions rather than introducing a new settings subsystem or persistence layer.

## Complexity Tracking

No constitution violations or justified complexity exceptions.

## Phase 0: Research

Completed in [research.md](./research.md).

Decisions:

- Settings is an in-app desktop-style surface, not a separate modal/window.
- Actual settings are session-local controls only.
- Provider/model gateway state is read-only and redacted.
- Future settings areas are explicit placeholders.
- Documentation and smoke validation are part of done.

## Phase 1: Design & Contracts

Completed artifacts:

- [data-model.md](./data-model.md): Settings Category, Setting Section, Setting Row, Local Settings State, Runtime Capability Summary, Placeholder Setting.
- [contracts/settings-ui.md](./contracts/settings-ui.md): entry, layout, category, working-row, placeholder-safety, and visual contracts.
- [quickstart.md](./quickstart.md): mock desktop UI, settings navigation, working controls, placeholders, real API visibility smoke, validation commands.

## Constitution Check (Post-Design)

- **Mechanism Parity, Original Expression**: PASS. Contract uses reference as layout direction only and requires Loomi-owned copy.
- **Runnable Vertical Slices**: PASS. Quickstart defines visible UI smoke and working local controls.
- **Core Flow Before Platform Complexity**: PASS. All platform-heavy areas remain placeholder/mock.
- **Observable Agent Execution**: PASS. Runtime/provider capability remains visible without new hidden execution.
- **Safety, Permissions, and Data Boundaries**: PASS. Placeholder safety contract prevents external actions and secret exposure.
- **Documentation Workflow**: PASS. Implementation must update docs-site and run docs build.

## Phase 2: Task Planning Approach

The next `/speckit-tasks` step should prioritize:

1. Settings shell/navigation and route/view state.
2. General category working controls.
3. Runtime/provider read-only status rows.
4. Placeholder category panels with safe disabled/mock behavior.
5. Tests for state preservation, working controls, placeholder safety, and visual contract copy.
6. Docs-site updates and browser smoke.
