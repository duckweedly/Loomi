# Tasks: M29 Multi-agent Runtime Foundation

**Input**: Design documents from `specs/037-multi-agent-runtime-foundation/`

**Tests**: TDD required.

## Phase 1: Setup

- [X] T001 Create M29 Spec Kit artifacts and move active feature pointers.

## Phase 2: Foundation

- [X] T002 [P] Add productdata tests for agent catalog metadata, Work/Chat RunContext filtering, and argument validation.
- [X] T003 Add agent tool constants, validation routing, persona allowlist entries, catalog metadata, and event grouping.
- [X] T004 Add agent task record model and in-memory service methods for spawn/list/complete.
- [X] T004a Add PostgreSQL `agent_tasks` migration, repository methods, and cross-thread no-leak coverage.

## Phase 3: User Story 1 - Spawn Agent Task

- [X] T005 [P] Add agent executor tests for spawn bounds, role validation, redaction, and safe summary.
- [X] T006 Implement `AgentToolExecutor` spawn.
- [X] T007 [P] Add ToolBroker and worker tests for approved agent spawn and provider continuation.
- [X] T008 Route agent tools through ToolBroker, runtime resolution, queued runner, and worker.
- [X] T009 [P] Add HTTP smoke for agent spawn approve -> execute -> final.

## Phase 4: User Story 2 - List and Complete Agent Tasks

- [X] T010 [P] Add agent executor tests for list/complete scope, unknown ids, limits, and bounded result summaries.
- [X] T011 Implement agent list/complete.
- [X] T012 [P] Add bounded loop HTTP smoke for agent.spawn -> agent.list -> agent.complete -> final.

## Phase 5: User Story 3 - Safe and Visible Coordination

- [X] T013 [P] Add safety tests for Chat-mode rejection, oversized fields, denied/stopped/terminal no-exec, duplicate/out-of-scope calls, and unsupported args.
- [X] T014 Implement safe no-exec/no-autonomy metadata and rejection paths.
- [X] T015 [P] Add frontend Settings Tools tests for agent metadata.
- [X] T016 Add frontend mock/client support for agent tool group metadata.
- [X] T017 [P] Add RunRail agent lifecycle tests.
- [X] T018 Update RunRail copy and mock runtime data for agent lifecycle rows.

## Phase 6: Documentation & Validation

- [X] T019 [P] Add multi-agent runtime architecture docs.
- [X] T020 [P] Add multi-agent runtime API/tool contract docs.
- [X] T021 [P] Add M29 runbook and devlog.
- [X] T022 Update roadmap/spec-kit docs and runbook index.
- [X] T023 Run `go test ./...`
- [X] T024 Run `bun test --cwd web`
- [X] T025 Run `bun run --cwd web build`
- [X] T026 Run `bun run --cwd docs-site build`
- [X] T027 Run `git diff --check`
- [X] T028 Perform browser smoke for Settings Tools agent tools, RunRail agent lifecycle visibility, and console errors.
