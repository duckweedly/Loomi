# Tasks: M23 Workspace Mutation Tools

**Input**: Design documents from `specs/031-workspace-mutation-tools/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD required. Each story includes test tasks before implementation tasks.

**Organization**: Tasks are grouped by independently testable user story.

## Phase 1: Setup

**Purpose**: Confirm the feature boundary and current workspace read-tool baseline.

- [X] T001 Confirm M22 tasks are complete and record no remaining incomplete checkboxes in `specs/030-bounded-agent-loop-todo-foundation/tasks.md`
- [X] T002 Update `AGENTS.md` Spec Kit plan reference to `specs/031-workspace-mutation-tools/plan.md`

---

## Phase 2: Foundational

**Purpose**: Add shared mutation tool identity and safe catalog metadata.

- [X] T003 [P] Add productdata tests for `workspace.write_file` and `workspace.edit` catalog entries in `internal/productdata/tool_catalog_test.go`
- [X] T004 Add mutation tool constants, safe argument keys, persona allowlist entries, and catalog metadata in `internal/productdata/models.go`, `internal/productdata/builtin_personas.go`, and `internal/productdata/tool_catalog.go`
- [X] T005 [P] Add frontend catalog tests for write-capable workspace mutation rows in `web/src/components/SettingsView.tools.test.tsx`
- [X] T006 Update frontend tool metadata types and mock catalog rows in `web/src/domain.ts`, `web/src/mockApiClient.ts`, and `web/src/mockData.ts`

---

## Phase 3: User Story 1 - Write New Workspace Files (Priority: P1)

**Goal**: Approve and execute one bounded new text file write under the workspace root.

**Independent Test**: A Work mode run requests `workspace.write_file`, approval creates one file, events show approval/execution/success, and pre-approval/denied/stopped paths do not write.

- [X] T007 [P] [US1] Add workspace write executor tests for new file, existing target rejection, traversal, symlink escape, sensitive path, invalid UTF-8, and size limit in `internal/runtime/workspace_tools_test.go`
- [X] T008 [US1] Implement `workspace.write_file` execution in `internal/runtime/workspace_tools.go`
- [X] T009 [P] [US1] Add ToolBroker mutation dispatch tests in `internal/runtime/tool_broker_test.go`
- [X] T010 [US1] Route mutation tools through ToolBroker and runtime tool resolution in `internal/runtime/tool_broker.go` and `internal/runtime/tools.go`
- [X] T011 [P] [US1] Add gateway/worker tests proving mutation tools are approval-gated, Work-mode enabled, Chat-mode rejected, and stopped/denied paths do not write in `internal/runtime/gateway_test.go` and `internal/runtime/worker_test.go`
- [X] T012 [US1] Preserve approval-gated execution and bounded continuation behavior for `workspace.write_file` in `internal/runtime/gateway.go` and `internal/runtime/queued_runner.go`
- [X] T013 [P] [US1] Add HTTP smoke for `workspace.write_file` approve -> execute -> final in `internal/httpapi/workspace_mutation_tools_smoke_test.go`

---

## Phase 4: User Story 2 - Edit Existing Workspace Files (Priority: P2)

**Goal**: Approve and execute one bounded exact replacement in an existing text file.

**Independent Test**: A Work mode run requests `workspace.edit`, approval changes the target exactly once, and ambiguous/missing/too-large edits fail without mutation.

- [X] T014 [P] [US2] Add workspace edit executor tests for exact replacement, missing old text, duplicate old text, sensitive path, binary/invalid UTF-8, and size limit in `internal/runtime/workspace_tools_test.go`
- [X] T015 [US2] Implement `workspace.edit` execution in `internal/runtime/workspace_tools.go`
- [X] T016 [P] [US2] Add gateway/worker tests for approval-gated edit, retry idempotency, and terminal no-op behavior in `internal/runtime/gateway_test.go` and `internal/runtime/worker_test.go`
- [X] T017 [US2] Preserve safe result metadata and terminal retry guards for `workspace.edit` in `internal/runtime/queued_runner.go`
- [X] T018 [P] [US2] Add HTTP smoke for `workspace.edit` approve -> execute -> final in `internal/httpapi/workspace_mutation_tools_smoke_test.go`

---

## Phase 5: User Story 3 - Show Mutation Risk and Audit Trail (Priority: P3)

**Goal**: Make mutation tools visibly distinct from read tools in Settings and runtime timeline.

**Independent Test**: Web tests render mutation catalog rows and RunRail lifecycle rows with risk/approval/write-capable metadata and no secret/root leakage.

- [X] T019 [P] [US3] Add RunRail mutation lifecycle tests in `web/src/components/RunRail.runtime.test.ts`
- [X] T020 [US3] Update RunRail and API event mapping copy for workspace mutation rows in `web/src/components/RunRail.tsx` and `web/src/realApiClient.ts`
- [X] T021 [P] [US3] Extend mock runtime data with mutation catalog/timeline evidence in `web/src/mockData.ts`

---

## Phase 6: Documentation & Validation

**Purpose**: Record the M23 behavior and prove the slice.

- [X] T022 [P] Add workspace mutation architecture docs in `docs-site/src/content/docs/architecture/workspace-mutation-tools.md`
- [X] T023 [P] Add workspace mutation API docs in `docs-site/src/content/docs/api/workspace-mutation-tools.md`
- [X] T024 [P] Add M23 runbook and devlog in `docs-site/src/content/docs/runbooks/local-m23-workspace-mutation-tools.md` and `docs-site/src/content/docs/devlog/2026-05-26-m23-workspace-mutation-tools.md`
- [X] T025 Update roadmap/spec-kit docs and runbook index in `docs-site/src/content/docs/roadmap/current-status.md`, `docs-site/src/content/docs/spec-kit/workflow.md`, and `docs-site/src/content/docs/runbooks/index.md`
- [X] T026 Run `go test ./...`
- [X] T027 Run `bun test --cwd web`
- [X] T028 Run `bun run --cwd web build`
- [X] T029 Run `bun run --cwd docs-site build`
- [X] T030 Run `git diff --check`
- [X] T031 Perform browser smoke for Settings/RunRail mutation visibility and console errors

---

## Dependencies & Execution Order

- Phase 1 -> Phase 2 -> US1 -> US2 -> US3 -> Documentation/Validation.
- US1 is the MVP and can be validated independently before US2.
- US2 depends on shared executor policy from US1 but must remain independently testable.
- US3 can start after mutation metadata exists.

## Parallel Opportunities

- T003 and T005 can run together.
- T007, T009, T011, and T013 are separate test files once constants exist.
- T014, T016, and T018 are separate test files after US1.
- T022, T023, and T024 can be drafted in parallel with final code review.

## Implementation Strategy

1. Complete catalog identity first.
2. Use RED -> GREEN for `workspace.write_file`.
3. Use RED -> GREEN for `workspace.edit`.
4. Add frontend visibility after backend metadata stabilizes.
5. Update docs and run full validation before claiming completion.
