# Tasks: Formal Interface Shell Redesign

**Input**: Design documents from `/specs/039-formal-interface-shell-redesign/`

**Prerequisites**: plan.md, spec.md

**Tests**: Browser smoke and existing automated tests are required for this UI shell slice.

**Organization**: Tasks are grouped by user story to keep this first round independently reviewable.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the explicit Spec Kit and documentation scope for UI-01.

- [x] T001 Create UI-01 Spec Kit artifacts in `specs/039-formal-interface-shell-redesign/spec.md`, `specs/039-formal-interface-shell-redesign/plan.md`, and `specs/039-formal-interface-shell-redesign/tasks.md`
- [x] T002 Update current feature pointer in `AGENTS.md` and `.specify/feature.json`
- [x] T003 [P] Add docs-site record for the formal shell redesign under `docs-site/src/content/docs/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define first-round visual tokens and layout mechanics shared by all UI stories.

- [x] T004 Update shell visual tokens in `web/src/styles.css` for light desktop background, rounded window, subtle border, shadow, white canvas, and compact widths
- [x] T005 Update session-local shell defaults in `web/src/useWorkspaceShellState.ts` for the first-open light shell direction

**Checkpoint**: Shared shell tokens and defaults are ready.

---

## Phase 3: User Story 1 - Open a Formal Desktop Shell (Priority: P1) MVP

**Goal**: The app opens into the requested first-round light desktop shell.

**Independent Test**: Open local web and inspect the overall proportions: narrow sidebar, white canvas, centered content, fixed bottom composer.

- [x] T006 [US1] Update outer app/window grid in `web/src/App.tsx` and `web/src/styles.css`
- [x] T007 [US1] Redesign the narrow sidebar layout in `web/src/components/ThreadSidebar.tsx` and `web/src/styles.css`
- [x] T008 [US1] Redesign the main white chat canvas and centered message column in `web/src/components/ChatCanvas.tsx` and `web/src/styles.css`
- [x] T009 [US1] Add narrow viewport protections in `web/src/styles.css`

**Checkpoint**: User Story 1 is visually reviewable in the browser.

---

## Phase 4: User Story 2 - Preserve Existing Conversation Controls (Priority: P2)

**Goal**: Chat and Work still use the existing composer and run controls after the layout change.

**Independent Test**: Type in Chat mode, type and submit in Work mode, and verify active-run Stop remains visible.

- [x] T010 [US2] Rework the fixed bottom composer shell in `web/src/components/Composer.tsx` and `web/src/styles.css` without removing existing submit, Stop, retry, regenerate, provider, persona, or Work in Folder entries
- [x] T011 [US2] Keep provider-unavailable and backend capability notices visible above the fixed composer in `web/src/components/ChatCanvas.tsx` and `web/src/styles.css`
- [x] T012 [US2] Update or add focused web tests only if existing control contracts break in `web/src/components/*.test.tsx` or `web/src/components/*.test.ts`

**Checkpoint**: Chat/Work controls remain functional.

---

## Phase 5: User Story 3 - Keep Secondary Panels Usable (Priority: P3)

**Goal**: Settings, Tools, and RunRail are not redesigned but remain accessible.

**Independent Test**: Open Settings > Tools and RunRail in the browser.

- [x] T013 [US3] Verify RunRail and right-panel containment with the new shell in `web/src/App.tsx` and `web/src/styles.css`
- [x] T014 [US3] Verify Settings > Tools stays readable with the new shell in `web/src/styles.css`

**Checkpoint**: Secondary panels are not broken by shell changes.

---

## Phase 6: Validation & Closeout

**Purpose**: Prove the first-round shell is ready for user review and leave refinements clearly out of scope.

- [x] T015 Run `go test ./...`
- [x] T016 Run `bun test --cwd web`
- [x] T017 Run `bun run --cwd web build`
- [x] T018 Run `bun run --cwd docs-site build`
- [x] T019 Run `git diff --check`
- [x] T020 Browser smoke: open local web, check layout proportions, Chat input, Work input/run, active-run Stop, Settings > Tools, console error count 0, and save screenshot path

## Dependencies & Execution Order

- Phase 1 before Phase 2.
- Phase 2 before User Stories 1-3.
- User Story 1 before User Story 2 because the composer and controls depend on the shell layout.
- User Story 3 after User Stories 1-2 because secondary panels need the final shell containment.
- Validation after all selected stories.

## Implementation Strategy

Complete the first-round shell only. Stop after browser smoke and validation. Do not mark later pixel polish, icon refinement, density tuning, advanced responsive behavior, Settings redesign, Tools redesign, RunRail redesign, backend/runtime/tool/provider/memory/database changes, or M38/activity recorder work as completed.
