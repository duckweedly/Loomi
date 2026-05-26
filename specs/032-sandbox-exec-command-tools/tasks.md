# Tasks: M24 Sandbox Exec Command Tools

**Input**: Design documents from `specs/032-sandbox-exec-command-tools/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD required. Each story includes test tasks before implementation tasks.

**Organization**: Tasks are grouped by independently testable user story.

## Phase 1: Setup

**Purpose**: Move active Spec Kit context from M23 to M24.

- [X] T001 Confirm M23 tasks are complete and record no remaining incomplete checkboxes in `specs/031-workspace-mutation-tools/tasks.md`
- [X] T002 Update `.specify/feature.json` and `AGENTS.md` Spec Kit plan reference to `specs/032-sandbox-exec-command-tools/plan.md`

---

## Phase 2: Foundational

**Purpose**: Add shared sandbox exec identity and catalog metadata.

- [X] T003 [P] Add productdata tests for `sandbox.exec_command` catalog entry and Work/Chat RunContext filtering in `internal/productdata/tool_catalog_test.go` and `internal/productdata/service_test.go`
- [X] T004 Add sandbox tool constants, safe argument keys, persona allowlist entries, catalog metadata, and safe event argument preview in `internal/productdata/models.go`, `internal/productdata/builtin_personas.go`, `internal/productdata/tool_catalog.go`, and `internal/productdata/service.go`
- [X] T005 [P] Add frontend catalog tests for exec-capable sandbox rows in `web/src/components/SettingsView.tools.test.tsx`
- [X] T006 Update frontend tool metadata rendering and mock catalog row in `web/src/components/SettingsView.tsx` and `web/src/mockApiClient.ts`

---

## Phase 3: User Story 1 - Run One Approved Command (Priority: P1)

**Goal**: Approve and execute one safe bounded argv command under the workspace root.

**Independent Test**: A Work mode run requests `sandbox.exec_command`, approval runs exactly one command, events show approval/execution/success, and provider continuation completes.

- [X] T007 [P] [US1] Add sandbox executor tests for successful argv execution, non-zero exit, output truncation, and bounded command validation in `internal/runtime/sandbox_tools_test.go`
- [X] T008 [US1] Implement bounded `sandbox.exec_command` execution in `internal/runtime/sandbox_tools.go`
- [X] T009 [P] [US1] Add ToolBroker sandbox dispatch tests in `internal/runtime/tool_broker_test.go`
- [X] T010 [US1] Route sandbox tools through ToolBroker and runtime tool resolution in `internal/runtime/tool_broker.go` and `internal/runtime/tools.go`
- [X] T011 [P] [US1] Add worker tests proving approve-before-exec, provider continuation, and one execution only in `internal/runtime/worker_test.go`
- [X] T012 [US1] Preserve approval-gated execution and terminal retry guards for `sandbox.exec_command` in `internal/runtime/queued_runner.go`
- [X] T013 [P] [US1] Add HTTP smoke for `sandbox.exec_command` approve -> execute -> final in `internal/httpapi/sandbox_exec_command_smoke_test.go`

---

## Phase 4: User Story 2 - Enforce Command Safety Boundaries (Priority: P2)

**Goal**: Reject unsafe command requests before spawn.

**Independent Test**: Unsafe requests fail without command side effects and without leaking host absolute roots or secrets.

- [X] T014 [P] [US2] Add executor tests for empty argv, shell-form, file-reading command, destructive command, cwd traversal, absolute cwd, denied sensitive cwd, path-bearing `ls`, unsafe git, and model-supplied env rejection in `internal/runtime/sandbox_tools_test.go`
- [X] T015 [US2] Implement tiny command allowlist and reject unsafe commands before spawn in `internal/runtime/sandbox_tools.go`
- [X] T016 [P] [US2] Add gateway/worker tests proving Chat-mode rejection, stopped/denied paths do not execute, and terminal no-op behavior in `internal/runtime/gateway_test.go` and `internal/runtime/worker_test.go`
- [X] T017 [US2] Preserve safe request/result metadata and no-exec terminal paths in `internal/runtime/gateway.go`, `internal/runtime/queued_runner.go`, and `internal/productdata/service.go`

---

## Phase 5: User Story 3 - Show Exec Risk and Audit Trail (Priority: P3)

**Goal**: Make sandbox exec visibly distinct from workspace read/mutation tools.

**Independent Test**: Web tests render sandbox exec catalog rows and RunRail lifecycle rows with risk/approval/exec-capable metadata and no secret/root leakage.

- [X] T018 [P] [US3] Add RunRail sandbox exec lifecycle tests in `web/src/components/RunRail.runtime.test.ts`
- [X] T019 [US3] Update RunRail copy for sandbox exec rows in `web/src/components/RunRail.tsx`
- [X] T020 [P] [US3] Extend mock runtime data with sandbox exec catalog/timeline evidence in `web/src/mockData.ts`

---

## Phase 6: Documentation & Validation

**Purpose**: Record M24 behavior and prove the slice.

- [X] T021 [P] Add sandbox exec architecture docs in `docs-site/src/content/docs/architecture/sandbox-exec-command.md`
- [X] T022 [P] Add sandbox exec API docs in `docs-site/src/content/docs/api/sandbox-exec-command.md`
- [X] T023 [P] Add M24 runbook and devlog in `docs-site/src/content/docs/runbooks/local-m24-sandbox-exec-command.md` and `docs-site/src/content/docs/devlog/2026-05-26-m24-sandbox-exec-command.md`
- [X] T024 Update roadmap/spec-kit docs and runbook index in `docs-site/src/content/docs/roadmap/current-status.md`, `docs-site/src/content/docs/spec-kit/workflow.md`, and `docs-site/src/content/docs/runbooks/index.md`
- [X] T025 Run `go test ./...`
- [X] T026 Run `bun test --cwd web`
- [X] T027 Run `bun run --cwd web build`
- [X] T028 Run `bun run --cwd docs-site build`
- [X] T029 Run `git diff --check`
- [X] T030 Perform browser smoke for Settings/RunRail sandbox exec visibility and console errors

---

## Dependencies & Execution Order

- Phase 1 -> Phase 2 -> US1 -> US2 -> US3 -> Documentation/Validation.
- US1 is the MVP and can be validated independently before US2.
- US2 hardens the same executor before the feature can be considered complete.
- US3 can start after catalog metadata exists.

## Parallel Opportunities

- T003 and T005 can run together.
- T007, T009, T011, and T013 are separate test files once constants exist.
- T014 and T016 are separate test files after US1.
- T021, T022, and T023 can be drafted in parallel with final code review.

## Implementation Strategy

1. Complete catalog identity first.
2. Use RED -> GREEN for safe `sandbox.exec_command` execution.
3. Add safety rejection tests before broadening behavior.
4. Add frontend visibility after backend metadata stabilizes.
5. Update docs and run full validation before claiming M24 completion.
