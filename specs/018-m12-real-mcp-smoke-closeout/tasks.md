# Tasks: M12 Real MCP Smoke Closeout

**Input**: Design documents from `/specs/018-m12-real-mcp-smoke-closeout/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required because this feature is a closeout evidence slice for a safety boundary.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Phase 1: Setup

**Purpose**: Pin the closeout feature and confirm existing boundaries.

- [X] T001 Create Spec Kit feature artifacts in `specs/018-m12-real-mcp-smoke-closeout/`
- [X] T002 Update active Spec Kit feature pointer in `AGENTS.md`

---

## Phase 2: Foundational

**Purpose**: Establish reusable smoke fixture coverage without adding platform capability.

- [X] T003 [P] Add real local stdio MCP discovery-and-call fixture in `internal/httpapi/mcp_real_smoke_test.go`
- [X] T004 [P] Add provider continuation assertions in `internal/httpapi/mcp_real_smoke_test.go`

**Checkpoint**: Fixture can prove discovery/list-tools and approved tools/call with real Content-Length framing.

---

## Phase 3: User Story 1 - Prove Local MCP Approval Execution (Priority: P1) MVP

**Goal**: One local smoke proves the M12 chain from discovery through approval, execution, redaction, continuation, and final.

**Independent Test**: `go test ./internal/httpapi -run TestM12RealLocalMCPApprovalSmoke`

### Tests for User Story 1

- [X] T005 [US1] Add M12.5 real local MCP approval smoke in `internal/httpapi/mcp_real_smoke_test.go`

### Implementation for User Story 1

- [X] T006 [US1] Configure smoke worker from `LOOMI_MCP_SERVERS_JSON` in `internal/httpapi/mcp_real_smoke_test.go`
- [X] T007 [US1] Assert event ordering, one tools/call, redaction, and final assistant message in `internal/httpapi/mcp_real_smoke_test.go`

**Checkpoint**: User Story 1 is complete when the targeted Go smoke passes.

---

## Phase 4: User Story 2 - Document Closeout Evidence (Priority: P2)

**Goal**: Docs-site records the closeout evidence, validation commands, browser limitation, and non-goals.

**Independent Test**: `bun run --cwd docs-site build`

### Implementation for User Story 2

- [X] T008 [P] [US2] Update M12 runbook evidence in `docs-site/src/content/docs/runbooks/local-m12-mcp-approval-execution.md`
- [X] T009 [P] [US2] Update M12 devlog closeout evidence in `docs-site/src/content/docs/devlog/2026-05-25-m12-mcp-approval-gated-execution.md`
- [X] T010 [P] [US2] Update roadmap status in `docs-site/src/content/docs/roadmap/current-status.md`
- [X] T011 [US2] Update Spec Kit workflow status in `docs-site/src/content/docs/spec-kit/workflow.md`

**Checkpoint**: User Story 2 is complete when docs build and clearly preserve M12.5 scope.

---

## Phase 5: Validation

**Purpose**: Run required closeout validation and mark evidence.

- [X] T012 Run `go test ./...`
- [X] T013 Run `bun test --cwd web`
- [X] T014 Run `bun run --cwd web build`
- [X] T015 Run `bun run --cwd docs-site build`
- [X] T016 Run `git diff --check`
- [X] T017 Record browser smoke status or exact skipped reason in `docs-site/src/content/docs/devlog/2026-05-25-m12-mcp-approval-gated-execution.md`

---

## Dependencies & Execution Order

- Phase 1 precedes all work.
- Phase 2 precedes User Story 1.
- User Story 1 can complete before docs updates.
- User Story 2 depends on actual smoke evidence from User Story 1.
- Validation runs after both stories.

## Parallel Opportunities

- T003 and T004 can be developed together inside the same smoke file but must be reconciled sequentially before T005.
- T008, T009, and T010 can be drafted in parallel after smoke results are known.

## Implementation Strategy

1. Complete the targeted smoke first.
2. Update docs with exact evidence and limitations.
3. Run full required validation.
4. Keep all broader MCP platform work out of scope.
