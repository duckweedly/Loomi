# Tasks: M5.5 Settings Placeholder

**Input**: Design documents from `/specs/007-settings-placeholder/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/settings-ui.md, quickstart.md

**Tests**: Included because the spec and quickstart define measurable settings behavior, placeholder safety, state preservation, frontend validation, browser smoke, and documentation validation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing. This task list is only for `007-settings-placeholder`; `006-streaming-chat-runtime` remains separate and must not be mixed in.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare shared settings vocabulary, shell state seams, and documentation targets before story behavior.

- [x] T001 Define settings category and row data structures in `web/src/components/settingsCatalog.ts`
- [x] T002 [P] Add settings catalog tests for required categories and mock/working statuses in `web/src/components/settingsCatalog.test.ts`
- [x] T003 Add settings view state to `web/src/useWorkspaceShellState.ts`
- [x] T004 [P] Add shell state tests for opening/closing Settings and preserving existing workspace state in `web/src/useWorkspaceShellState.test.ts`
- [x] T005 [P] Add M5.5 docs stubs in `docs-site/src/content/docs/architecture/settings-placeholder.md`, `docs-site/src/content/docs/runbooks/settings-placeholder.md`, and `docs-site/src/content/docs/devlog/2026-05-24-m5-5-settings-placeholder.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared component/state boundaries that MUST be complete before user story implementation.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T006 Create Settings view component shell with category navigation and content slots in `web/src/components/SettingsView.tsx`
- [x] T007 [P] Add Settings view visual contract tests for sidebar/content/card landmarks in `web/src/components/SettingsView.layout.test.tsx`
- [x] T008 Wire sidebar Settings menu action to open the Settings view in `web/src/components/ThreadSidebar.tsx` and `web/src/App.tsx`
- [x] T009 Add Settings view styling primitives for desktop-style categories, cards, rows, badges, and controls in `web/src/styles.css`
- [x] T010 Add current-session settings state for default workspace mode and selected mock runtime scenario in `web/src/state.ts` or `web/src/useWorkspaceShellState.ts`
- [x] T011 [P] Add tests for settings state defaults and non-persistence assumptions in `web/src/state.runtime.test.ts` or `web/src/useWorkspaceShellState.test.ts`

**Checkpoint**: Foundation ready - settings can be opened, closed, styled, and reasoned about without story-specific content.

---

## Phase 3: User Story 1 - Open a desktop-style settings surface (Priority: P1) MVP

**Goal**: Users can open a reference-inspired settings surface with left category navigation and right grouped content while preserving workspace context.

**Independent Test**: Select Settings from the existing app shell and verify a two-column settings layout opens with General selected by default, a back affordance, and grouped cards.

### Tests for User Story 1

- [x] T012 [P] [US1] Add App-level test for opening Settings from sidebar and returning to workspace in `web/src/App.settings.test.tsx`
- [x] T013 [P] [US1] Add Settings view test for default General category selection and back affordance in `web/src/components/SettingsView.navigation.test.tsx`
- [x] T014 [P] [US1] Add category navigation test for required primary, Agent Core, and management groups in `web/src/components/SettingsView.navigation.test.tsx`

### Implementation for User Story 1

- [x] T015 [US1] Render Settings view instead of Chat Canvas when settings mode is active in `web/src/App.tsx`
- [x] T016 [US1] Implement Settings sidebar groups, active category state, and category switching in `web/src/components/SettingsView.tsx`
- [x] T017 [US1] Implement General category grouped cards and placeholder sections for non-selected categories in `web/src/components/SettingsView.tsx`
- [x] T018 [US1] Add Settings menu item click behavior to open Settings instead of only showing a submenu arrow in `web/src/components/ThreadSidebar.tsx`
- [x] T019 [US1] Verify Settings opens in mock desktop UI using `specs/007-settings-placeholder/quickstart.md`

**Checkpoint**: User Story 1 is independently functional and demoable as the M5.5 MVP.

---

## Phase 4: User Story 2 - Adjust currently supported local settings (Priority: P2)

**Goal**: Settings exposes only current real local/session controls as working rows and shows runtime/provider state safely as read-only.

**Independent Test**: Change default workspace mode and mock runtime scenario, then verify subsequent local behavior reflects the choices while provider state remains read-only and secret-free.

### Tests for User Story 2

- [x] T020 [P] [US2] Add settings state test for default workspace mode changing future local behavior in `web/src/useWorkspaceShellState.test.ts` or `web/src/state.test.ts`
- [x] T021 [P] [US2] Add settings state test for mock runtime scenario applying only to future mock sends in `web/src/state.runtime.test.ts`
- [x] T022 [P] [US2] Add Settings view test for read-only data source, backend capability, and provider capability rows in `web/src/components/SettingsView.runtime.test.tsx`
- [x] T023 [P] [US2] Add Settings view safety test asserting provider secrets and key fields are absent in `web/src/components/SettingsView.runtime.test.tsx`

### Implementation for User Story 2

- [x] T024 [US2] Add default workspace mode selection control to General settings in `web/src/components/SettingsView.tsx`
- [x] T025 [US2] Connect default workspace mode setting to future local create/open behavior in `web/src/App.tsx` and `web/src/useWorkspaceShellState.ts`
- [x] T026 [US2] Add mock runtime scenario selection control to General settings in `web/src/components/SettingsView.tsx`
- [x] T027 [US2] Connect mock runtime scenario setting to existing `selectRuntimeScript` flow in `web/src/App.tsx` and `web/src/state.ts`
- [x] T028 [US2] Add read-only runtime status rows for data source mode, backend capability, stream state, and selected thread/run status in `web/src/components/SettingsView.tsx`
- [x] T029 [US2] Add read-only provider capability summary loading path without exposing secrets in `web/src/realApiClient.ts`, `web/src/state.ts`, and `web/src/components/SettingsView.tsx`
- [x] T030 [US2] Verify working settings and runtime/provider rows with `specs/007-settings-placeholder/quickstart.md`

**Checkpoint**: User Stories 1 and 2 both work independently, and actual settings are visibly distinct from read-only status rows.

---

## Phase 5: User Story 3 - Preview future settings areas safely (Priority: P3)

**Goal**: Users can navigate future settings categories as mock/placeholder sections without triggering real side effects.

**Independent Test**: Select every placeholder category and verify it shows mock/preview copy, disabled controls, and no external/provider/tool/backend write behavior.

### Tests for User Story 3

- [x] T031 [P] [US3] Add Settings placeholder category test for all future categories in `web/src/components/SettingsView.placeholders.test.tsx`
- [x] T032 [P] [US3] Add placeholder safety test for disabled/mock controls and no secret-entry fields in `web/src/components/SettingsView.placeholders.test.tsx`
- [x] T033 [P] [US3] Add placeholder navigation state preservation test in `web/src/components/SettingsView.placeholders.test.tsx`

### Implementation for User Story 3

- [x] T034 [US3] Implement placeholder panels for Appearance, Providers, Connectors, Plugins, Skill, MCP, Notebook, Memory, Activity Recorder, Context, Safety, Tools, Routes, About, and Advanced in `web/src/components/SettingsView.tsx`
- [x] T035 [US3] Add mock/preview/disabled badges and safe copy for placeholder settings in `web/src/components/SettingsView.tsx`
- [x] T036 [US3] Ensure placeholder controls do not call provider, tool, connector, file, or backend write paths in `web/src/components/SettingsView.tsx`
- [x] T037 [US3] Verify placeholder category navigation with `specs/007-settings-placeholder/quickstart.md`

**Checkpoint**: All M5.5 user stories are independently functional and future settings remain non-executing placeholders.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, browser smoke, and final consistency checks across M5.5.

- [x] T038 [P] Update settings architecture documentation in `docs-site/src/content/docs/architecture/settings-placeholder.md`
- [x] T039 [P] Update settings runbook with mock desktop, settings navigation, placeholder safety, and real API visibility smoke in `docs-site/src/content/docs/runbooks/settings-placeholder.md`
- [x] T040 [P] Add M5.5 development log with completed scope, validation results, known limitations, and next steps in `docs-site/src/content/docs/devlog/2026-05-24-m5-5-settings-placeholder.md`
- [x] T041 Update Spec Kit workflow/status references for `007-settings-placeholder` in `docs-site/src/content/docs/spec-kit/workflow.md` and `docs-site/src/content/docs/roadmap/current-status.md`
- [x] T042 Run frontend tests with `bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts`
- [x] T043 Run frontend build with `bun run --cwd web build`
- [x] T044 Run docs build with `bun run --cwd docs-site build`
- [x] T045 Perform mock desktop Settings browser smoke using `bun run --cwd web desktop:dev` and `specs/007-settings-placeholder/quickstart.md`
- [x] T046 Perform real API visibility smoke when local API is available, or record exact blocker in `docs-site/src/content/docs/devlog/2026-05-24-m5-5-settings-placeholder.md`
- [x] T047 Perform final review against `specs/007-settings-placeholder/contracts/settings-ui.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion - blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion - MVP settings shell.
- **User Story 2 (Phase 4)**: Depends on Foundational completion and can use US1 Settings shell.
- **User Story 3 (Phase 5)**: Depends on Foundational completion and category catalog; can be implemented after or alongside US2 once shell exists.
- **Polish (Phase 6)**: Depends on desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: No dependency on US2 or US3 after Foundation; delivers visible settings surface MVP.
- **US2 (P2)**: Uses US1 shell and adds working local/session settings plus read-only runtime/provider status.
- **US3 (P3)**: Uses US1 shell/category navigation and adds safe placeholder content for future areas.

### Within Each User Story

- Tests first for contract and behavior coverage.
- Catalog/state seams before Settings UI rendering.
- Settings shell before category-specific rows.
- Working settings before placeholder polish when validating actual behavior.
- Story smoke validation at each checkpoint.

### Parallel Opportunities

- Setup catalog tests, shell state tests, and docs stubs T002, T004, T005 can run in parallel.
- US1 tests T012-T014 can run in parallel before implementation.
- US2 tests T020-T023 can run in parallel before implementation.
- US3 tests T031-T033 can run in parallel before implementation.
- Documentation tasks T038-T040 can run in parallel after behavior stabilizes.

---

## Parallel Example: User Story 1

```bash
Task: "Add App-level test for opening Settings from sidebar and returning to workspace in web/src/App.settings.test.tsx"
Task: "Add Settings view test for default General category selection and back affordance in web/src/components/SettingsView.navigation.test.tsx"
Task: "Add category navigation test for required primary, Agent Core, and management groups in web/src/components/SettingsView.navigation.test.tsx"
```

## Parallel Example: User Story 2

```bash
Task: "Add settings state test for default workspace mode changing future local behavior in web/src/useWorkspaceShellState.test.ts"
Task: "Add settings state test for mock runtime scenario applying only to future mock sends in web/src/state.runtime.test.ts"
Task: "Add Settings view tests for read-only runtime/provider rows and secret absence in web/src/components/SettingsView.runtime.test.tsx"
```

## Parallel Example: User Story 3

```bash
Task: "Add Settings placeholder category test for all future categories in web/src/components/SettingsView.placeholders.test.tsx"
Task: "Add placeholder safety test for disabled/mock controls and no secret-entry fields in web/src/components/SettingsView.placeholders.test.tsx"
Task: "Add placeholder navigation state preservation test in web/src/components/SettingsView.placeholders.test.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 setup.
2. Complete Phase 2 foundation.
3. Complete Phase 3 US1.
4. Stop and validate Settings opens, shows the two-column shell, and returns to workspace.
5. Demo the visible settings surface before wiring working rows or placeholders.

### Incremental Delivery

1. Setup + Foundation -> settings catalog, shell state, Settings component, styles ready.
2. US1 -> visible desktop-style Settings shell.
3. US2 -> actual local/session settings and read-only runtime/provider status.
4. US3 -> safe placeholder categories for future settings areas.
5. Polish -> docs, builds, browser smoke, validation log.

### Parallel Team Strategy

With multiple developers:

1. One developer owns settings catalog/state foundation.
2. One developer owns SettingsView layout/navigation and tests.
3. One developer owns working local settings/runtime status rows.
4. One developer owns placeholder category content and docs updates.

## Notes

- This task list intentionally excludes `specs/006-streaming-chat-runtime`.
- Provider draft secrets must not be written into docs examples, tests, persistent settings, backend writes, provider calls, or frontend state as raw strings; only masked key presence may be retained in session-local UI state.
- Placeholder controls must not execute tools, providers, connectors, file operations, backend writes, or external actions.
- [P] tasks must touch different files or be coordinated to avoid same-file conflicts.
- Do not add settings persistence, backend provider secret storage, account/team settings, tool permissions, memory/RAG controls, activity recorder controls, or route management unless the plan is updated first.
