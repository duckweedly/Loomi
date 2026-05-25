# Tasks: M14 Memory Management Audit UX

**Input**: Design documents from `specs/021-memory-management-audit-ux/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required by the user. This task set records completed prep/blocker foundation, US1 management UX, US2 audit history UX, and seeded browser smoke.

**Organization**: Tasks are grouped by independently testable user stories.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish M14 Spec Kit artifacts and context pointer.

- [x] T001 Create `specs/021-memory-management-audit-ux/spec.md`
- [x] T002 Create `specs/021-memory-management-audit-ux/plan.md`
- [x] T003 [P] Create `specs/021-memory-management-audit-ux/research.md`
- [x] T004 [P] Create `specs/021-memory-management-audit-ux/data-model.md`
- [x] T005 [P] Create `specs/021-memory-management-audit-ux/contracts/memory-management-api.md`
- [x] T006 [P] Create `specs/021-memory-management-audit-ux/quickstart.md`
- [x] T007 [P] Create `specs/021-memory-management-audit-ux/checklists/requirements.md`
- [x] T008 Update `.specify/feature.json` and `AGENTS.md` to point at `specs/021-memory-management-audit-ux/plan.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Inspect current memory API/UI types so implementation extends existing boundaries.

- [x] T009 Review existing memory API/service contracts in `internal/httpapi/memory.go` and `internal/productdata/`
- [x] T010 Review current Settings > Memory UI/client paths in `web/src/realApiClient.ts` and `web/src/components/SettingsView.tsx`
- [x] T011 Review existing docs-site memory pages before editing `docs-site/src/content/docs/`

---

## Phase 3: Prep Blockers - Safe API Contract Foundation

**Goal**: Fix review findings that would make M14 management/audit UX unsafe or impossible to trust.

**Independent Test**: Backend/client tests prove thread-scoped detail/delete, terminal-run audit retention, expanded redaction, and unified list/search filters.

- [x] T012 [P] Add backend blocker tests in `internal/httpapi/memory_test.go`
- [x] T013 [P] Add real API client contract mapping tests in `web/src/realApiClient.test.ts`
- [x] T014 [P] Add MemoryPanel error state test in `web/src/components/MemoryPanel.test.tsx`
- [x] T015 Fix thread-scoped memory detail/delete authorization in `internal/productdata/`
- [x] T016 Preserve memory audit for terminal source runs in `internal/productdata/`
- [x] T017 Expand redaction coverage for `/home`, Windows paths, stdout/stderr, and provider traces in `internal/productdata/models.go`
- [x] T018 Unify list/search/detail/audit API shape in `internal/httpapi/memory.go`
- [x] T019 Update `web/src/realApiClient.ts` contract mapping for safe management/audit projections
- [x] T020 Add low-risk MemoryPanel load/search/delete error display in `web/src/components/MemoryPanel.tsx`

---

## Phase 4: User Story 1 - Manage approved memories (Priority: P1, Complete)

**Goal**: Settings > Memory supports real list/search/filter/detail/delete-confirm flows.

**Independent Test**: Browser seeded entry smoke plus web tests can list, search/filter, inspect detail, and confirm delete for memory entries.

- [x] T021 [US1] Add Settings > Memory search/filter/detail/delete-confirm tests in `web/src/components/SettingsView.layout.test.tsx`
- [x] T022 [US1] Update `web/src/components/SettingsView.tsx` for grounded search/filter controls
- [x] T023 [US1] Add memory detail drawer/modal in `web/src/components/SettingsView.tsx`
- [x] T024 [US1] Add delete confirmation flow in `web/src/components/SettingsView.tsx`
- [x] T025 [US1] Cover loading/null/empty/error/tombstoned states in Settings > Memory tests

---

## Phase 5: User Story 2 - Review safe memory history (Priority: P2, Complete)

**Goal**: Users can see real safe memory history for proposal/approval/denial/delete/snapshot events.

**Independent Test**: Backend/client/UI tests show real audit history and prove unsafe metadata is redacted.

- [x] T026 [P] [US2] Add Settings > Memory history tests in `web/src/components/SettingsView.layout.test.tsx`
- [x] T027 [US2] Render memory audit/history in `web/src/components/SettingsView.tsx` using `GET /v1/memory/audit`
- [x] T028 [US2] Verify UI never fabricates audit items when backend endpoint is unavailable

---

## Phase 6: Documentation

**Purpose**: Keep docs-site aligned with M14 behavior.

- [x] T029 Update management/audit flow in `docs-site/src/content/docs/architecture/memory-foundation.md`
- [x] T030 Update endpoint/payload docs in `docs-site/src/content/docs/api/memory-foundation.md`
- [x] T031 Update local validation steps in `docs-site/src/content/docs/runbooks/local-m13-memory.md`
- [x] T032 [P] Add M14 devlog in `docs-site/src/content/docs/devlog/2026-05-25-m14-memory-management-audit-ux.md`
- [x] T033 Update status/workflow docs in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`

---

## Phase 7: Validation & Browser Smoke

**Purpose**: Prove implementation and docs.

- [x] T034 Run manual cross-artifact consistency review or `/speckit-analyze`
- [x] T035 Run `go test ./...`
- [x] T036 Run `bun test --cwd web`
- [x] T037 Run `bun run --cwd web build`
- [x] T038 Run `bun run --cwd docs-site build`
- [x] T039 Run `git diff --check`
- [x] T040 Browser seeded entry smoke for full M14 list, search/filter, detail, delete confirmation, and audit history flows

## Dependencies & Execution Order

- Setup and foundational review precede code changes.
- Prep blockers precede full US1/US2 UI work.
- US1 can complete independently before US2 once prep blockers are green.
- US2 UI depends on real productdata audit/event availability and safe projection decisions.
- Documentation follows implemented behavior.
- Validation and browser smoke depend on all implementation tasks.

## Implementation Strategy

1. Point Spec Kit at M14.
2. Fix prep blockers and API/client contract shape with tests.
3. Document the complete M14 UX/API contract.
4. Implement full Settings > Memory management/history UI.
5. Run required validation plus seeded browser smoke before claiming full UX complete candidate.
