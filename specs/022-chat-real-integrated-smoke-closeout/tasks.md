# Tasks: M15 Chat Real Integrated Smoke Closeout

**Input**: Design documents from `specs/022-chat-real-integrated-smoke-closeout/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required. M15 is a closeout/evidence slice whose value is the repeatable smoke.

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 Read M15 spec, plan, research, data-model, quickstart, and contract docs before editing code
- [x] T002 [P] Locate current HTTP chat/run, approval, worker, provider fixture, memory snapshot, MCP execution, and event replay seams in `internal/`
- [x] T003 [P] Locate current real API timeline/replay mapping and docs pages that need closeout updates in `web/` and `docs-site/src/content/docs/`

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T004 Add or reuse deterministic provider fixture phases for tool request and final continuation in backend test code under `internal/`
- [x] T005 Add or reuse deterministic local stdio MCP fixture that returns secret-looking data but exposes only redacted summaries in backend test code under `internal/`
- [x] T006 Add M15 sensitive canary assertions covering API payloads, RunContext summaries, run events, tool result summary, final assistant message, and docs examples

---

## Phase 3: User Story 1 - Prove the real chat path reaches approval (Priority: P1) MVP

**Goal**: Real API/service/worker path loads approved memory and blocks on one discovered persona-allowed MCP approval.

**Independent Test**: Gated smoke reaches `blocked_on_tool_approval` with memory snapshot and MCP candidate evidence.

- [x] T007 [US1] Add gated `TestM15ChatRealIntegratedSmoke` setup in `internal/httpapi/chat_real_integrated_smoke_test.go`
- [x] T008 [US1] Seed approved memory and assert `RunContext.MemorySnapshot` enters the run in `internal/httpapi/chat_real_integrated_smoke_test.go`
- [x] T009 [US1] Seed discovered persona-allowed MCP candidate and assert approval-required projection/event evidence in `internal/httpapi/chat_real_integrated_smoke_test.go`

---

## Phase 4: User Story 2 - Approve, execute, continue, and complete (Priority: P2)

**Goal**: HTTP approval resumes worker execution, records redacted result, continues provider, and completes the run.

**Independent Test**: Same gated smoke verifies exactly one approval, MCP execution, continuation, final assistant message, and completed run.

- [x] T010 [US2] Drive approval through the real HTTP approve handler in `internal/httpapi/chat_real_integrated_smoke_test.go`
- [x] T011 [US2] Assert worker executes exactly one MCP `tools/call` and records redacted success in `internal/httpapi/chat_real_integrated_smoke_test.go`
- [x] T012 [US2] Assert provider continuation writes one final assistant message and completes the run in `internal/httpapi/chat_real_integrated_smoke_test.go`

---

## Phase 5: User Story 3 - Replay evidence without leaking secrets (Priority: P3)

**Goal**: Replay contains required milestones and all shareable surfaces stay redacted.

**Independent Test**: Gated smoke fetches replay/history evidence and redaction assertions pass.

- [x] T013 [US3] Assert persisted history contains queued/worker/pipeline, memory, MCP discovery/candidate hash, approval, execution, continuation, and completion events in `internal/httpapi/chat_real_integrated_smoke_test.go`
- [x] T014 [US3] Add or adjust web replay mapping test only if current UI cannot show M15 timeline states in `web/src/realApiClient.test.ts` or runtime tests
- [x] T015 [US3] Document browser smoke status or blocker in M15 docs

---

## Phase 6: Documentation & Validation

- [x] T016 [P] Add M15 devlog in `docs-site/src/content/docs/devlog/`
- [x] T017 [P] Add local M15 chat smoke runbook in `docs-site/src/content/docs/runbooks/`
- [x] T018 Update `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`
- [x] T019 Update related API/architecture docs if implementation reveals behavior differences
- [x] T020 Run `LOOMI_M15_REAL_CHAT_SMOKE=1 go test ./internal/httpapi -run TestM15ChatRealIntegratedSmoke -count=1 -v`
- [x] T021 Run `go test ./...`
- [x] T022 Run `bun test --cwd web`
- [x] T023 Run `bun run --cwd web build`
- [x] T024 Run `bun run --cwd docs-site build`
- [x] T025 Run `git diff --check`

## Dependencies & Execution Order

- Phase 1 before all code/docs edits.
- Phase 2 before user-story assertions.
- US1 before US2; US2 before US3 because the smoke is one integrated chain.
- Documentation can run in parallel with final smoke assertions after behavior is stable.
- Validation runs last.

## Implementation Strategy

Implement the smallest backend smoke that proves the complete chain. Add production code only if existing boundaries cannot expose required safe evidence. Keep docs examples synthetic and redacted.
