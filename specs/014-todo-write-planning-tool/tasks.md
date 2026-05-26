# Tasks: M12 Todo Write Planning Tool

**Input**: Design documents from `specs/014-todo-write-planning-tool/`

**Tests**: Required. This feature changes tool validation, execution, catalog, and frontend rendering behavior.

## Phase 1: Setup

- [X] T001 Create M12 Spec Kit artifacts in `specs/014-todo-write-planning-tool/`
- [X] T002 Update `.specify/feature.json` to point to `specs/014-todo-write-planning-tool`

## Phase 2: Backend TDD

- [X] T003 Add productdata tests for `runtime.todo_write` request validation in `internal/productdata/service_test.go`
- [X] T004 Add runtime tests for todo argument normalization, result counts, and catalog metadata in `internal/runtime/tools_test.go`
- [X] T005 Add worker test for approved `runtime.todo_write` execution in `internal/runtime/worker_test.go`
- [X] T006 Implement productdata allowlist and bounded argument validation in `internal/productdata/models.go`
- [X] T007 Implement runtime todo tool definition, normalization, execution, and catalog entry in `internal/runtime/tools.go`
- [X] T008 Wire worker tool lookup for `runtime.todo_write` in `internal/runtime/worker.go`

## Phase 3: Frontend

- [X] T009 Add ToolCallCard test for todo_write summaries in `web/src/components/ToolCallCard.test.tsx`
- [X] T010 Add `runtime.todo_write` to mock tool catalog in `web/src/mockApiClient.ts`

## Phase 4: Documentation

- [X] T011 Add architecture doc in `docs-site/src/content/docs/architecture/todo-write-planning-tool.md`
- [X] T012 Add API doc in `docs-site/src/content/docs/api/todo-write-planning-tool.md`
- [X] T013 Add runbook in `docs-site/src/content/docs/runbooks/local-m12.md`
- [X] T014 Add devlog in `docs-site/src/content/docs/devlog/2026-05-26-m12-todo-write-planning-tool.md`
- [X] T015 Update roadmap and Spec Kit workflow docs in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`

## Phase 5: Validation

- [X] T016 Run `go test ./...`
- [X] T017 Run `bun test --cwd web`
- [X] T018 Run `bun run --cwd web build`
- [X] T019 Run `bun run build` from `docs-site/`
- [X] T020 Run `git diff --check`
- [X] T021 Perform browser smoke for Settings Tools catalog and todo_write visibility
