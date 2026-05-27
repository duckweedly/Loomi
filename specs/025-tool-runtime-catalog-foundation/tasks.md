# Tasks: M18 Tool Runtime + Tool Catalog Foundation

**Input**: `spec.md`, `plan.md`, `research.md`, `data-model.md`, `contracts/`
**Status**: Complete candidate

## Phase 1 - Spec Kit Setup

- [X] T001 Create `specs/025-tool-runtime-catalog-foundation/` with spec, plan, research, data model, contracts, quickstart, tasks, and checklist.
- [X] T002 Point `.specify/feature.json` at `specs/025-tool-runtime-catalog-foundation`.
- [X] T003 Run Spec Kit analyze and remove scope/status conflicts.

## Phase 2 - Productdata and Runtime Tests

- [X] T004 [P] Add unit tests for builtin current-time catalog entry.
- [X] T005 [P] Add unit tests for discovered MCP catalog entry.
- [X] T006 [P] Add unit tests for persona allowlist filtering.
- [X] T007 [P] Add unit tests for not-allowlisted, disabled, and non-executable broker rejection.
- [X] T008 [P] Add unit tests proving builtin and MCP use the same broker entrypoint.
- [X] T009 [P] Add redaction tests for result/event metadata.

## Phase 3 - Catalog and Broker Implementation

- [X] T010 Add catalog enums/models and safe metadata projection.
- [X] T011 Add builtin current-time catalog definition and executor adapter.
- [X] T012 Add MCP candidate catalog mapping from discovery metadata.
- [X] T013 Add `ToolExecutor`, `ToolInvocation`, `ToolResult`, and broker implementation.
- [X] T014 Update RunContext tool resolution to use catalog + persona allowlist + MCP discovery metadata.
- [X] T015 Route approved builtin and MCP worker resume through broker.

## Phase 4 - API and Smoke Tests

- [X] T016 Add read-only `GET /v1/tools/catalog`.
- [X] T017 Add HTTP tests for catalog response and redaction.
- [X] T018 Add deterministic provider builtin approval-to-broker smoke.
- [X] T019 Add deterministic provider MCP discovery/approval-to-broker smoke.
- [X] T020 Verify replay API has unified tool event chain and sensitive canary does not leak.

## Phase 5 - Web UI

- [X] T021 Add Tool catalog domain/API client mapping.
- [X] T022 Add Settings > Tools read-only panel.
- [X] T023 Add web tests for safe catalog rendering and absence of raw args/results/secrets/write controls.

## Phase 6 - Documentation and Validation

- [X] T024 Update architecture docs for tool runtime/catalog/broker boundaries.
- [X] T025 Update API docs for tools catalog API and event payloads.
- [X] T026 Update runbook for local M18 validation.
- [X] T027 Update devlog, roadmap current status, and Spec Kit workflow.
- [X] T028 Run `go test ./...`.
- [X] T029 Run `bun test --cwd web`.
- [X] T030 Run `bun run --cwd web build`.
- [X] T031 Run `bun run --cwd docs-site build`.
- [X] T032 Run `git diff --check`.
