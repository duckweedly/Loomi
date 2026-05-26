# Tasks: M8 Safe Workspace Read Tools

**Input**: Design documents from `specs/010-safe-workspace-read-tools/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), [contracts/](./contracts/)

**Tests**: Required. This feature introduces local filesystem read access and must be test-first.

## Phase 1: Setup

- [X] T001 Create M8 Spec Kit artifacts in `specs/010-safe-workspace-read-tools/`
- [X] T002 Update `.specify/feature.json` to point to `specs/010-safe-workspace-read-tools`

## Phase 2: Backend Foundation

- [X] T003 [P] Add runtime tool tests for `workspace.glob`, `workspace.grep`, and `workspace.read_file` validation in `internal/runtime/tools_test.go`
- [X] T004 [P] Add service tests proving workspace tools require approval and keep M7 idempotency in `internal/productdata/service_test.go`
- [X] T005 Implement workspace read tool definitions and schema validation in `internal/runtime/tools.go`
- [X] T006 Implement workspace root containment and sensitive path denial in `internal/runtime/tools.go`
- [X] T007 Implement bounded glob, grep, and read execution in `internal/runtime/tools.go`
- [X] T008 Wire approved workspace read execution into the existing worker path in `internal/runtime/worker.go`
- [X] T009 Add backend worker/replay tests for workspace read tool terminal events in `internal/runtime/worker_test.go`

## Phase 3: Frontend UI

- [X] T010 [P] Add frontend runtime adapter tests for workspace read tool result summaries in `web/src/runtime/executionAdapter.test.ts`
- [X] T011 [P] Add ToolCallCard tests for workspace read tool labels and bounded result rendering in `web/src/components/ToolCallCard.test.tsx`
- [X] T012 Extend frontend domain/result mapping for workspace read tool summaries in `web/src/domain.ts` and `web/src/runtime/executionAdapter.ts`
- [X] T013 Update ToolCallCard, RunRail, and Timeline copy/states for workspace read tools in `web/src/components/`

## Phase 4: Documentation

- [X] T014 [P] Add architecture doc in `docs-site/src/content/docs/architecture/workspace-read-tools.md`
- [X] T015 [P] Add API/tool contract doc in `docs-site/src/content/docs/api/workspace-read-tools.md`
- [X] T016 [P] Add local runbook in `docs-site/src/content/docs/runbooks/local-m8.md`
- [X] T017 [P] Add devlog in `docs-site/src/content/docs/devlog/2026-05-26-m8-workspace-read-tools.md`
- [X] T018 Update roadmap and Spec Kit workflow docs in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`

## Phase 5: Validation

- [X] T019 Run `go test ./...`
- [X] T020 Run `bun test --cwd web`
- [X] T021 Run `bun run --cwd web build`
- [X] T022 Run `bun run build` from `docs-site/`
- [X] T023 Run `git diff --check`
- [X] T024 Perform browser smoke for page load and console errors
