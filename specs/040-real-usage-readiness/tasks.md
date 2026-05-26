# Tasks: Real Usage Readiness

**Input**: Design documents from `/specs/040-real-usage-readiness/`

**Prerequisites**: plan.md, spec.md

**Tests**: Required. User explicitly requested coverage for removing confusing Chat controls, WorkPlanView empty/plan/approval states, RunRail human labels, simplified Sidebar actions, Composer Chat/Work placeholders, and basic Markdown.

**Organization**: Tasks are grouped by user story so each readiness surface is independently testable.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish UI-02 Spec Kit and documentation scope.

- [x] T001 Create UI-02 Spec Kit artifacts in `specs/040-real-usage-readiness/spec.md`, `specs/040-real-usage-readiness/plan.md`, and `specs/040-real-usage-readiness/tasks.md`
- [x] T002 Update current feature pointer in `.specify/feature.json` and `AGENTS.md`
- [x] T003 [P] Update docs-site workflow/current-status/architecture/devlog records under `docs-site/src/content/docs/`

---

## Phase 2: Foundational Tests (Blocking Prerequisites)

**Purpose**: Lock the requested user-visible behavior before implementation.

- [x] T004 [P] Add runtime header removal and provider warning tests in `web/src/components/ChatCanvas.states.test.ts`
- [x] T005 [P] Add fake Composer control removal and mode placeholder tests in `web/src/components/Composer.test.ts`
- [x] T006 [P] Add WorkPlanView empty/metadata/approval blocked tests in `web/src/components/WorkPlanView.test.tsx`
- [x] T007 [P] Add RunRail human tool label tests in `web/src/components/RunRail.polish.test.ts`
- [x] T008 [P] Add Sidebar removed-entry/status label tests in `web/src/components/ThreadSidebar.actions.test.ts`
- [x] T009 [P] Add ToolCallCard human label tests in `web/src/components/ToolCallCard.test.tsx`

**Checkpoint**: Focused tests fail for missing UI-02 readiness behavior.

---

## Phase 3: User Story 1 - Understand Current Runtime Reality (Priority: P1)

**Goal**: Provider blockers are obvious without the old runtime/debug header.

**Independent Test**: ChatCanvas tests show the old runtime header is absent and provider unavailable still links to Provider Settings.

- [x] T010 [US1] Remove old source-mode runtime header from `web/src/components/ChatCanvas.tsx`
- [x] T011 [US1] Update provider unavailable action copy in `web/src/i18n.ts`
- [x] T012 [US1] Add supporting styles in `web/src/styles.css`

---

## Phase 4: User Story 2 - Use Work Mode Without Fake Controls (Priority: P1)

**Goal**: Work mode shows honest plan states without fake folder controls.

**Independent Test**: Work mode does not render fake folder controls and does not invent steps without plan metadata.

- [x] T013 [US2] Stop deriving Work plan steps from generic user messages in `web/src/workModeProjection.ts`
- [x] T014 [US2] Render empty Work plan metadata state in `web/src/components/WorkPlanView.tsx`
- [x] T015 [US2] Remove disabled Work folder fake button from `web/src/components/Composer.tsx`
- [x] T016 [US2] Remove obsolete Work folder limitation styles in `web/src/styles.css`

---

## Phase 5: User Story 3 - Read Runs Like User Progress (Priority: P2)

**Goal**: Tool and approval states are readable by a normal user.

**Independent Test**: RunRail and ToolCallCard show human-first labels; approval blocked shows waiting state and actions.

- [x] T017 [US3] Add approval waiting notice in `web/src/components/ChatCanvas.tsx`
- [x] T018 [US3] Add human-first tool event formatting in `web/src/components/RunRail.tsx`
- [x] T019 [US3] Add human-first tool card formatting in `web/src/components/ToolCallCard.tsx`
- [x] T020 [US3] Add approval and tool label styles in `web/src/styles.css`

---

## Phase 6: User Story 4 - Basic Sidebar and Composer Are Usable (Priority: P3)

**Goal**: Sidebar, Composer, and message rendering basics feel usable, not static.

**Independent Test**: Sidebar actions, Composer mode placeholders, fake control removal, and basic Markdown rendering are covered by component tests.

- [x] T021 [US4] Remove duplicate sidebar search field in `web/src/components/ThreadSidebar.tsx`
- [x] T021a [US4] Move current-mode new thread action into the titlebar compose button in `web/src/App.tsx`
- [x] T022 [US4] Add mode-specific New Chat/New Work labels in `web/src/components/ThreadSidebar.tsx` and `web/src/i18n.ts`
- [x] T023 [US4] Add readable run status labels to sidebar run dots in `web/src/components/ThreadSidebar.tsx`
- [x] T024 [US4] Pass selected mode into thread creation in `web/src/App.tsx` and `web/src/state.ts`
- [x] T025 [US4] Remove duplicate sidebar search/footer styles in `web/src/styles.css`
- [x] T032 [US4] Add thread rename/delete row menu in `web/src/components/ThreadSidebar.tsx`
- [x] T033 [US4] Remove unimplemented Composer attachment, persona/provider selector, Work folder, and voice controls in `web/src/components/Composer.tsx`
- [x] T034 [US4] Add basic Markdown rendering in `web/src/components/ChatCanvas.tsx`

---

## Phase 7: Validation & Closeout

**Purpose**: Prove the UI-02 readiness slice is complete.

- [x] T026 Run `go test ./...`
- [x] T027 Run `bun test --cwd web`
- [x] T028 Run `bun run --cwd web build`
- [x] T029 Run `bun run --cwd docs-site build`
- [x] T030 Run `git diff --check`
- [x] T031 Browser smoke: verify no chat runtime header, Chat/Work input, no fake Composer controls, simplified Sidebar actions, Markdown rendering, human tool labels, approval blocked controls, Settings Providers/Tools, console error count 0, and screenshot path

## Dependencies & Execution Order

- Phase 1 before Phase 2.
- Phase 2 before implementation phases.
- US1 and US2 can run in parallel after tests.
- US3 depends on existing tool/run metadata and can run independently of sidebar work.
- US4 can run independently after foundational tests.
- Validation after all user stories.

## Implementation Strategy

Finish UI-02 readiness only. Do not advance M38, add tools, change provider/runtime behavior, alter database schema, or perform pixel-level redesign.
