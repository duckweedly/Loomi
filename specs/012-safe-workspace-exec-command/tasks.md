# Tasks: M10 Safe Workspace Exec Command

**Input**: Design documents from `specs/012-safe-workspace-exec-command/`

**Tests**: Required. This feature introduces local command execution and must be test-first.

## Phase 1: Setup

- [X] T001 Create M10 Spec Kit artifacts in `specs/012-safe-workspace-exec-command/`
- [X] T002 Update `.specify/feature.json` to point to `specs/012-safe-workspace-exec-command`

## Phase 2: Backend

- [X] T003 Add runtime tests for safe argv command execution, cwd containment, dangerous command rejection, timeout, and output truncation in `internal/runtime/tools_test.go`
- [X] T004 Add productdata tests proving exec command requires approval and validates argument shape in `internal/productdata/service_test.go`
- [X] T005 Implement `workspace.exec_command` definition, argument normalization, and validation in `internal/runtime/tools.go`
- [X] T006 Implement argv-only command execution with timeout and bounded stdout/stderr in `internal/runtime/tools.go`
- [X] T007 Add productdata tool-name constant and argument validation in `internal/productdata/models.go`
- [X] T008 Wire approved exec command execution through worker workspace root in `internal/runtime/worker.go`
- [X] T009 Add worker tests for approved exec command terminal events in `internal/runtime/worker_test.go`

## Phase 3: Frontend

- [X] T010 Add frontend runtime adapter test for exec command result summaries in `web/src/runtime/executionAdapter.test.ts`
- [X] T011 Add ToolCallCard test for exec command summary rendering in `web/src/components/ToolCallCard.test.tsx`

## Phase 4: Documentation

- [X] T012 Add architecture doc in `docs-site/src/content/docs/architecture/workspace-exec-command.md`
- [X] T013 Add API/tool contract doc in `docs-site/src/content/docs/api/workspace-exec-command.md`
- [X] T014 Add local runbook in `docs-site/src/content/docs/runbooks/local-m10.md`
- [X] T015 Add devlog in `docs-site/src/content/docs/devlog/2026-05-26-m10-workspace-exec-command.md`
- [X] T016 Update roadmap and Spec Kit workflow docs in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`

## Phase 5: Validation

- [X] T017 Run `go test ./...`
- [X] T018 Run `bun test --cwd web`
- [X] T019 Run `bun run --cwd web build`
- [X] T020 Run `bun run build` from `docs-site/`
- [X] T021 Run `git diff --check`
- [X] T022 Perform browser smoke for exec command tool states and console errors
