# Tasks: M21 Workspace Read Tools

**Input**: Design documents from `/specs/029-workspace-read-tools/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required by user request and specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Phase 1: Setup

- [X] T001 Update AGENTS.md current Spec Kit feature pointer to `specs/029-workspace-read-tools/plan.md`
- [X] T002 [P] Review Arkloop read-only mechanism references without copying code or expression
- [X] T003 [P] Add Speckit docs links/status for 029 workspace read tools

---

## Phase 2: Foundational

- [X] T004 Add workspace tool names/catalog metadata and persona allowlist support in `internal/productdata/models.go`, `internal/productdata/tool_catalog.go`, and `internal/productdata/builtin_personas.go`
- [X] T005 Add workspace tool argument validation and event metadata group/source mapping in `internal/productdata/models.go` and `internal/productdata/service.go`
- [X] T006 Add workspace root resolution, path boundary validation, sensitive denylist, bounded read, glob, and grep implementation in `internal/runtime/workspace_tools.go`
- [X] T007 Wire workspace tools through `internal/runtime/tools.go`, `internal/runtime/tool_broker.go`, `internal/runtime/gateway.go`, and `internal/runtime/queued_runner.go`

---

## Phase 3: User Story 1 - Approved Workspace Reading (Priority: P1)

**Goal**: Work mode can request/approve/execute glob/read/grep and continue through provider result flow.

**Independent Test**: Backend smoke proves approval-required before execution and success after approval for all three tools.

- [X] T008 [P] [US1] Add backend tests for workspace tool definitions and argument normalization in `internal/runtime/tools_test.go`
- [X] T009 [P] [US1] Add ToolBroker tests proving workspace tools use one broker entrypoint in `internal/runtime/tool_broker_test.go`
- [X] T010 [US1] Add backend smoke for approved glob/read/grep continuation in `internal/httpapi/workspace_read_tools_smoke_test.go`
- [X] T011 [US1] Implement approved workspace tool execution until T008-T010 pass

---

## Phase 4: User Story 2 - Workspace Boundary Protection (Priority: P2)

**Goal**: Workspace tools deny traversal, outside absolute paths, symlink escape, and sensitive files without leaking content.

**Independent Test**: Backend smoke/unit tests cover allowed fixture content plus denied sensitive/outside/symlink cases.

- [X] T012 [P] [US2] Add workspace root/path/sensitive/symlink tests in `internal/runtime/workspace_tools_test.go`
- [X] T013 [US2] Implement path boundary and sensitive deny behavior until T012 passes
- [X] T014 [US2] Extend backend smoke to assert denial after approval and no sensitive content leakage

---

## Phase 5: User Story 3 - Operator Visibility (Priority: P3)

**Goal**: Settings and timelines display workspace tool catalog and event states safely.

**Independent Test**: Web tests verify workspace catalog rows, read-only risk metadata, no absolute path exposure, and timeline state rendering.

- [X] T015 [P] [US3] Add web Settings > Tools catalog tests for workspace group/risk/no absolute path in `web/src/components/SettingsView.tools.test.tsx`
- [X] T016 [P] [US3] Add timeline rendering tests for workspace tool events in `web/src/components/RunTimeline.runtime.test.tsx`
- [X] T017 [US3] Update Settings/Timeline/domain rendering to show workspace tools safely

---

## Phase 6: Documentation & Validation

- [X] T018 [P] Update docs-site architecture/api/runbook/devlog/current-status/spec-kit workflow pages for M21
- [X] T019 Run `go test ./...`
- [X] T020 Run `bun test --cwd web`
- [X] T021 Run `bun run --cwd web build`
- [X] T022 Run `bun run --cwd docs-site build`
- [X] T023 Run `git diff --check`

---

## Dependencies & Execution Order

- Setup and Foundational tasks block user stories.
- US1 is MVP and must pass before UI is called complete.
- US2 can proceed after foundational implementation and shares files with US1, so run sequentially in this session.
- US3 depends on backend catalog/event metadata shape.
- Documentation and validation run after code behavior is complete.

## Parallel Opportunities

- T002 and T003 can run in parallel.
- T008, T009, and T012 target separate test files.
- T015 and T016 target separate frontend test coverage.
- T018 docs pages can be updated after backend metadata names settle.

## Implementation Strategy

1. Complete foundational backend metadata and executor path.
2. Prove approved execution and boundary denials with backend tests/smoke.
3. Update web catalog/timeline rendering.
4. Update docs-site and run required validation.
