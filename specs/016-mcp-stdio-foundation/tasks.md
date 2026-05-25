# Tasks: MCP Stdio Foundation

**Input**: Design documents from `specs/016-mcp-stdio-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Required by quickstart: Go tests for config validation, discovery parser, redaction, tool mapping, persona references, and RunContext MCP availability; web tests only if Timeline/debug UI mapping is touched.

**Organization**: Tasks are grouped by user story so discovery, mapping, and observability can be implemented and validated independently.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the narrow MCP foundation without entering execution.

- [x] T001 Review existing runtime tool registry and ToolSpec names in `internal/runtime/tools.go`
- [x] T002 Review existing RunContext/pipeline safe summary code in `internal/runtime/pipeline.go`
- [x] T003 Review existing persona allowed-tools validation in `internal/productdata/service.go`
- [x] T004 Review existing Timeline/debug event mapping in `web/src/runtime/runtimeEventGroups.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define shared MCP types, validation, redaction, and non-execution boundary used by every story.

- [x] T005 Create MCP config and discovery types in `internal/runtime/mcp_config.go`
- [x] T006 [P] Add MCP safety redaction helpers in `internal/runtime/mcp_config.go`
- [x] T007 [P] Add tests for local stdio config acceptance and HTTP/SSE/OAuth/remote rejection in `internal/runtime/mcp_config_test.go`
- [x] T008 [P] Add tests proving env, args, stderr, tokens, credentials, and secret-looking paths are redacted in `internal/runtime/mcp_config_test.go`
- [x] T009 Define read-only MCP ToolSpec candidate fields in `internal/runtime/mcp_tools.go`
- [x] T010 Document the non-executable MCP boundary in `internal/runtime/mcp_tools.go`

**Checkpoint**: Foundation ready - user story implementation can start.

---

## Phase 3: User Story 1 - Configure and discover a local MCP server (Priority: P1) MVP

**Goal**: Validate explicit local stdio config and run bounded discovery/list-tools without executing any MCP tool.

**Independent Test**: A local stdio config fixture produces discovered tool candidates or a redacted failure; no tool execution path is called.

### Tests for User Story 1

- [x] T011 [P] [US1] Add discovery parser tests for successful list-tools output in `internal/runtime/mcp_discovery_test.go`
- [x] T012 [P] [US1] Add discovery failure tests for timeout, missing command, invalid JSON, oversized schema, and invalid schema in `internal/runtime/mcp_discovery_test.go`
- [x] T013 [P] [US1] Add process cleanup/session lifecycle tests for discovery-only behavior in `internal/runtime/mcp_discovery_test.go`

### Implementation for User Story 1

- [x] T014 [US1] Implement bounded stdio discovery/list-tools session lifecycle in `internal/runtime/mcp_discovery.go`
- [x] T015 [US1] Implement discovery result safe summaries and redacted failure codes in `internal/runtime/mcp_discovery.go`
- [x] T016 [US1] Ensure discovery code has no MCP tool invocation path in `internal/runtime/mcp_discovery.go`
- [x] T017 [US1] Add a local discovery fixture or test double in `internal/runtime/testdata/mcp_stdio_list_tools.json`

**Checkpoint**: US1 can be validated independently through config + discovery tests.

---

## Phase 4: User Story 2 - Expose discovered MCP tools as read-only ToolRegistry candidates (Priority: P2)

**Goal**: Map discovered MCP schemas into namespaced non-executable ToolSpec candidates and allow persona references without enabling execution.

**Independent Test**: Valid and conflicting MCP schemas map to stable namespaced read-only candidates; persona allowed-tools references remain disabled by default.

### Tests for User Story 2

- [x] T018 [P] [US2] Add ToolSpec mapping tests for `mcp.<server_slug>.<tool_name>` naming in `internal/runtime/mcp_tools_test.go`
- [x] T019 [P] [US2] Add conflict tests for internal tool name collisions and duplicate MCP tool names in `internal/runtime/mcp_tools_test.go`
- [x] T020 [P] [US2] Add persona allowed-tools reference tests for discovered non-executable MCP candidates in `internal/productdata/service_test.go`

### Implementation for User Story 2

- [x] T021 [US2] Implement MCP schema to read-only ToolSpec candidate mapping in `internal/runtime/mcp_tools.go`
- [x] T022 [US2] Integrate read-only MCP candidates into ToolRegistry summaries without adding executors in `internal/runtime/tools.go`
- [x] T023 [US2] Extend persona allowed-tools validation to recognize discovered MCP candidate names as non-executable references in `internal/productdata/service.go`
- [x] T024 [US2] Ensure MCP candidates cannot override or replace internal tools in `internal/runtime/mcp_tools.go`

**Checkpoint**: US2 can be validated independently through mapping and persona reference tests.

---

## Phase 5: User Story 3 - Observe MCP availability and safety errors in RunContext and Timeline (Priority: P3)

**Goal**: Surface safe MCP discovery status, candidate availability, disabled execution state, and redacted safety errors in RunContext and Timeline/debug.

**Independent Test**: RunContext contains safe MCP availability metadata and Timeline/debug shows discovery success/failure labels from live or replayed metadata when UI mapping is touched.

### Tests for User Story 3

- [x] T025 [P] [US3] Add RunContext MCP availability summary tests in `internal/runtime/pipeline_test.go`
- [x] T026 [P] [US3] Add redacted safety error event tests in `internal/runtime/pipeline_test.go`
- [x] T027 [P] [US3] Add frontend runtime event grouping tests for MCP labels in `web/src/runtime/runtimeEventGroups.test.ts`
- [x] T028 [P] [US3] Add Timeline/debug rendering tests for MCP labels in `web/src/components/RunTimeline.runtime.test.ts`

### Implementation for User Story 3

- [x] T029 [US3] Add MCP availability safe summary to RunContext preparation in `internal/runtime/pipeline.go`
- [x] T030 [US3] Emit or map MCP discovery success/failure/safety metadata through existing run-event summary flow in `internal/runtime/pipeline.go`
- [x] T031 [US3] Map MCP discovery labels in `web/src/runtime/runtimeEventGroups.ts`
- [x] T032 [US3] Render MCP discovery and non-executable state in `web/src/components/RunTimeline.tsx` or `web/src/components/RunRail.tsx`

**Checkpoint**: US3 can be validated independently through RunContext tests and optional frontend tests.

---

## Phase 6: Documentation & Validation

**Purpose**: Keep docs-site and Spec Kit status aligned with the planned implementation.

- [x] T033 [P] Create architecture documentation in `docs-site/src/content/docs/architecture/mcp-stdio-foundation.md`
- [x] T034 [P] Create API/event documentation in `docs-site/src/content/docs/api/mcp-stdio-foundation.md` or extend `docs-site/src/content/docs/api/tool-call-approval.md`
- [x] T035 [P] Create local runbook in `docs-site/src/content/docs/runbooks/local-m11-mcp.md`
- [x] T036 Update roadmap status in `docs-site/src/content/docs/roadmap/current-status.md`
- [x] T037 Update Spec Kit workflow references in `docs-site/src/content/docs/spec-kit/workflow.md`
- [x] T038 Add devlog entry in `docs-site/src/content/docs/devlog/2026-05-25-m11-mcp-stdio-foundation.md`
- [x] T039 Run backend validation from `specs/016-mcp-stdio-foundation/quickstart.md`
- [x] T040 Run web validation from `specs/016-mcp-stdio-foundation/quickstart.md` if UI files changed
- [x] T041 Run `bun run --cwd docs-site build`
- [x] T042 Record validation results and known limitations in `specs/016-mcp-stdio-foundation/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational.
- **User Story 2 (Phase 4)**: Depends on Foundational and benefits from US1 discovery fixtures.
- **User Story 3 (Phase 5)**: Depends on Foundational and can use mocked discovery summaries; full smoke benefits from US1 and US2.
- **Documentation & Validation (Phase 6)**: Runs after desired stories are complete.

### User Story Dependencies

- **US1 (P1)**: Required MVP for real discovery.
- **US2 (P2)**: Can start after Foundational using parser fixtures; integrates best after US1.
- **US3 (P3)**: Can start after Foundational with mocked summaries; full RunContext smoke should follow US1/US2.

### Parallel Opportunities

- T006-T008 can run in parallel after T005.
- T011-T013 can run in parallel before T014-T017.
- T018-T020 can run in parallel before T021-T024.
- T025-T028 can run in parallel before T029-T032.
- T033-T035 can run in parallel with late implementation once contracts are stable.

## Parallel Example: User Story 1

```bash
Task: "Add discovery parser tests for successful list-tools output in internal/runtime/mcp_discovery_test.go"
Task: "Add discovery failure tests for timeout, missing command, invalid JSON, oversized schema, and invalid schema in internal/runtime/mcp_discovery_test.go"
Task: "Add process cleanup/session lifecycle tests for discovery-only behavior in internal/runtime/mcp_discovery_test.go"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Setup and Foundational tasks.
2. Implement config validation, redaction, and discovery-only session lifecycle.
3. Validate with Go tests proving no tool execution.
4. Stop and review before mapping or UI expansion if risk appears.

### Incremental Delivery

1. US1: local stdio discovery/list-tools.
2. US2: read-only ToolSpec candidates and persona references.
3. US3: RunContext availability plus Timeline/debug labels.
4. Docs and validation evidence.

### Boundary Reminder

No task may implement real MCP tool execution, automatic execution, shell/filesystem/browser automation, remote MCP, marketplace/plugin install, or approval bypass. Future execution remains design-only until a later approved spec.
