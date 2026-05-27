# Tasks: M28 Artifact Runtime Foundation

**Input**: Design documents from `specs/036-artifact-runtime-foundation/`

**Tests**: TDD required.

## Phase 1: Setup

- [X] T001 Create M28 Spec Kit artifacts and move active feature pointers.

## Phase 2: Foundation

- [X] T002 [P] Add productdata tests for artifact catalog metadata, Work/Chat RunContext filtering, and argument validation.
- [X] T003 Add artifact tool constants, validation routing, persona allowlist entries, catalog metadata, and event grouping.
- [X] T004 Add artifact record model and in-memory service methods for create/read/list.
- [X] T004a Add PostgreSQL `artifacts` migration, repository methods, and cross-thread no-leak coverage.

## Phase 3: User Story 1 - Create Text Artifact

- [X] T005 [P] Add artifact executor tests for create_text bounds, UTF-8 validation, redaction, and safe summary.
- [X] T006 Implement `ArtifactToolExecutor` create_text.
- [X] T007 [P] Add ToolBroker and worker tests for approved artifact create and provider continuation.
- [X] T008 Route artifact tools through ToolBroker, runtime resolution, queued runner, and worker.
- [X] T009 [P] Add HTTP smoke for artifact create approve -> execute -> final.

## Phase 4: User Story 2 - Read and List Artifacts

- [X] T010 [P] Add artifact executor tests for read/list scope, unknown ids, limits, and bounded excerpts.
- [X] T011 Implement artifact read/list.
- [X] T012 [P] Add bounded loop HTTP smoke for create_text -> list -> read -> final.

## Phase 5: User Story 3 - Safe and Visible Non-executable Runtime

- [X] T013 [P] Add safety tests for Chat-mode rejection, oversized content, invalid UTF-8, denied/stopped/terminal no-exec, duplicate/out-of-scope calls, and unsupported args.
- [X] T014 Implement safe no-exec metadata and rejection paths.
- [X] T015 [P] Add frontend Settings Tools tests for artifact metadata.
- [X] T016 Add frontend mock/client support for artifact tool group metadata.
- [X] T017 [P] Add RunRail artifact lifecycle tests.
- [X] T018 Update RunRail copy and mock runtime data for artifact lifecycle rows.

## Phase 6: Documentation & Validation

- [X] T019 [P] Add artifact runtime architecture docs.
- [X] T020 [P] Add artifact runtime API/tool contract docs.
- [X] T021 [P] Add M28 runbook and devlog.
- [X] T022 Update roadmap/spec-kit docs and runbook index.
- [X] T023 Run `go test ./...`
- [X] T024 Run `bun test --cwd web`
- [X] T025 Run `bun run --cwd web build`
- [X] T026 Run `bun run --cwd docs-site build`
- [X] T027 Run `git diff --check`
- [X] T028 Perform browser smoke for Settings Tools artifact tools, RunRail artifact lifecycle visibility, and console errors.
