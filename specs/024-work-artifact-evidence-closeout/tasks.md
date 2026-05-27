# Tasks: M17 Work Artifact Evidence Closeout

**Input**: Design documents from `specs/024-work-artifact-evidence-closeout/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required by feature request.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Phase 1: Setup

- [x] T001 Create `specs/024-work-artifact-evidence-closeout/` Spec Kit artifacts
- [x] T002 Update `.specify/feature.json` and AGENTS current feature pointer to 024

---

## Phase 2: Foundational

- [x] T003 [P] Add M17 local seed evidence test coverage in `cmd/loomi-seed/main_test.go`
- [x] T004 Add explicit M17 local-dev/test seed scenario in `cmd/loomi-seed/main.go`

---

## Phase 3: User Story 1 - Replay evidence into Work Plan View (Priority: P1)

**Goal**: Seeded Work thread renders goal, steps, artifact metadata, and recent progress from current run events.

**Independent Test**: Run the seed path and projection tests; browser smoke opens the seeded Work thread in real API mode.

- [x] T005 [P] [US1] Add projection evidence coverage in `web/src/workModeProjection.test.ts`
- [x] T006 [US1] Ensure Work projection preserves event replay metadata in `web/src/workModeProjection.ts`

---

## Phase 4: User Story 2 - Show safe artifact evidence only (Priority: P2)

**Goal**: Artifact evidence displays safe metadata plus redaction marker and no executable controls.

**Independent Test**: Render artifact cards with unsafe metadata and verify no action controls or unsafe values appear.

- [x] T007 [P] [US2] Add no-executable artifact controls render test in `web/src/components/WorkPlanView.test.tsx`
- [x] T008 [US2] Add redaction marker support in `web/src/domain.ts`, `web/src/workModeProjection.ts`, and `web/src/components/WorkPlanView.tsx`

---

## Phase 5: User Story 3 - Preserve mode isolation during real smoke (Priority: P3)

**Goal**: Chat mode never renders Work Plan View.

**Independent Test**: Existing and added render tests prove Chat mode isolation.

- [x] T009 [P] [US3] Keep Chat mode isolation coverage in `web/src/components/WorkPlanView.test.tsx`

---

## Phase 6: Documentation & Validation

- [x] T010 Update architecture/API docs in `docs-site/src/content/docs/architecture/work-mode-foundation.md` and `docs-site/src/content/docs/api/work-mode-foundation.md`
- [x] T011 [P] Add runbook `docs-site/src/content/docs/runbooks/local-m17-work-artifact-smoke.md`
- [x] T012 [P] Add devlog `docs-site/src/content/docs/devlog/2026-05-25-m17-work-artifact-evidence-closeout.md`
- [x] T013 Update `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`
- [x] T014 Run `go test ./...`
- [x] T015 Run `bun test --cwd web`
- [x] T016 Run `bun run --cwd web build`
- [x] T017 Run `bun run --cwd docs-site build`
- [x] T018 Run `git diff --check`
- [x] T019 Run browser smoke and record evidence

## Dependencies & Execution Order

- Setup before seed/code/doc changes.
- T003 before T004.
- T005/T007 before T006/T008.
- Documentation after implementation.
- Validation after all changes.

## Implementation Strategy

Deliver M17 through the minimal local seed evidence path, frontend safe projection/card rendering, docs, and real browser smoke. Do not add a production event-write API or artifact runtime.
