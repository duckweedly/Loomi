# Tasks: M9 Safe Workspace Write Tools

**Input**: Design documents from `specs/011-safe-workspace-write-tools/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), [contracts/](./contracts/)

**Tests**: Required. This feature introduces local filesystem mutation and must be test-first.

## Phase 1: Setup

- [X] T001 Create M9 Spec Kit artifacts in `specs/011-safe-workspace-write-tools/`
- [X] T002 Update `.specify/feature.json` to point to `specs/011-safe-workspace-write-tools`

## Phase 2: Backend Foundation

- [X] T003 [P] Add runtime tests for `workspace.write_file` validation, safe write, sensitive path denial, missing parent denial, and symlink escape denial in `internal/runtime/tools_test.go`
- [X] T004 [P] Add runtime tests for `workspace.edit` exact single replacement, missing match failure, duplicate match failure, and no-mutation failures in `internal/runtime/tools_test.go`
- [X] T005 [P] Add productdata tests proving write/edit tools require approval and validate argument shape in `internal/productdata/service_test.go`
- [X] T006 Implement `workspace.write_file` and `workspace.edit` tool definitions and schema validation in `internal/runtime/tools.go`
- [X] T007 Harden workspace path resolution for mutation, including symlink escape rejection in `internal/runtime/tools.go`
- [X] T008 Implement bounded UTF-8 write and exact edit execution in `internal/runtime/tools.go`
- [X] T009 Add productdata tool-name constants and argument validation in `internal/productdata/models.go`
- [X] T010 Wire approved write/edit execution into existing worker tool dispatch in `internal/runtime/worker.go`
- [X] T011 Add worker tests for approved write/edit terminal events in `internal/runtime/worker_test.go`

## Phase 3: Frontend UI

- [X] T012 [P] Add frontend runtime adapter test for workspace write/edit result summaries in `web/src/runtime/executionAdapter.test.ts`
- [X] T013 [P] Add ToolCallCard tests for write/edit summary rendering in `web/src/components/ToolCallCard.test.tsx`
- [X] T014 Update ToolCallCard summary formatting for write/edit result payloads in `web/src/components/ToolCallCard.tsx`

## Phase 4: Documentation

- [X] T015 [P] Add architecture doc in `docs-site/src/content/docs/architecture/workspace-write-tools.md`
- [X] T016 [P] Add API/tool contract doc in `docs-site/src/content/docs/api/workspace-write-tools.md`
- [X] T017 [P] Add local runbook in `docs-site/src/content/docs/runbooks/local-m9.md`
- [X] T018 [P] Add devlog in `docs-site/src/content/docs/devlog/2026-05-26-m9-workspace-write-tools.md`
- [X] T019 Update roadmap and Spec Kit workflow docs in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`

## Phase 5: Validation

- [X] T020 Run `go test ./...`
- [X] T021 Run `bun test --cwd web`
- [X] T022 Run `bun run --cwd web build`
- [X] T023 Run `bun run build` from `docs-site/`
- [X] T024 Run `git diff --check`
- [X] T025 Perform browser smoke for workspace write/edit tool states and console errors
