# Tasks: M16 Work Mode Foundation

**Input**: Design documents from `specs/023-work-mode-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required by user request for projection, rendering, mode isolation, event replay, and redaction.

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 Update active Spec Kit reference in `AGENTS.md` and `.specify/feature.json`
- [x] T002 [P] Add M16 spec/design docs in `specs/023-work-mode-foundation/`

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T003 [P] Add Work mode projection types in `web/src/domain.ts`
- [x] T004 Add safe Work projection helper in `web/src/workModeProjection.ts`
- [x] T005 [P] Add Work projection tests in `web/src/workModeProjection.test.ts`
- [x] T006 Seed Work thread plan/artifact metadata in `web/src/mockData.ts`

---

## Phase 3: User Story 1 - See a work plan on a work thread (Priority: P1) MVP

**Goal**: Work mode threads render a Work Plan View from existing messages/run events.

**Independent Test**: Render Work mode with seeded run data and verify goal, steps, status, artifacts, and recent progress are visible.

- [x] T007 [P] [US1] Add Work Plan View component tests in `web/src/components/WorkPlanView.test.tsx`
- [x] T008 [US1] Implement Work Plan View component in `web/src/components/WorkPlanView.tsx`
- [x] T009 [US1] Mount Work Plan View for Work mode only in `web/src/components/ChatCanvas.tsx`
- [x] T010 [US1] Add Work Plan View styles in `web/src/styles.css`

---

## Phase 4: User Story 2 - Keep artifact references safe (Priority: P2)

**Goal**: Artifact references show safe metadata only and redact unsafe values.

**Independent Test**: Projection test with secret-looking artifact metadata does not render raw secret values or executable fields.

- [x] T011 [US2] Add redaction coverage for artifact metadata in `web/src/workModeProjection.test.ts`
- [x] T012 [US2] Render metadata-only artifact cards in `web/src/components/WorkPlanView.tsx`

---

## Phase 5: User Story 3 - Preserve Chat and Work isolation (Priority: P3)

**Goal**: Chat mode remains unchanged while Work mode adds the new surface.

**Independent Test**: Render Chat and Work threads and verify only Work mode contains the Work Plan View.

- [x] T013 [US3] Add Chat/Work isolation render coverage in `web/src/components/WorkPlanView.test.tsx`
- [x] T014 [US3] Verify event replay changes projection status in `web/src/workModeProjection.test.ts`

---

## Phase 6: Documentation & Validation

- [x] T015 [P] Add architecture docs in `docs-site/src/content/docs/architecture/work-mode-foundation.md`
- [x] T016 [P] Add API/payload notes in `docs-site/src/content/docs/api/work-mode-foundation.md`
- [x] T017 [P] Add local runbook in `docs-site/src/content/docs/runbooks/local-m16-work-mode.md`
- [x] T018 Add devlog in `docs-site/src/content/docs/devlog/2026-05-25-m16-work-mode-foundation.md`
- [x] T019 Update roadmap and Spec Kit workflow docs in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`
- [x] T020 Run `bun test --cwd web`
- [x] T021 Run `bun run --cwd web build`
- [x] T022 Run `bun run --cwd docs-site build`
- [x] T023 Run `git diff --check`
- [x] T024 Run browser smoke for Work mode and Chat mode
- [x] T025 Mark completed tasks in `specs/023-work-mode-foundation/tasks.md`

## Dependencies & Execution Order

- Phase 1 before implementation.
- Phase 2 before UI rendering.
- US1 before US2/US3 final verification.
- Documentation can proceed after implementation shape is stable.
- Validation runs last.

## Implementation Strategy

Deliver the MVP through the frontend projection first, then tighten artifact redaction, mode isolation, docs, and browser evidence. Do not add backend APIs unless the projection cannot meet the minimum view from existing data.
