# Tasks: M27 Browser Automation Foundation

**Input**: Design documents from `specs/035-browser-automation-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD required. Each story includes test tasks before implementation tasks.

**Organization**: Tasks are grouped by independently testable user story.

## Phase 1: Setup

**Purpose**: Move active Spec Kit context from M26 to M27.

- [X] T001 Confirm M26 tasks are complete and record no remaining incomplete checkboxes in `specs/034-web-fetch-tool-foundation/tasks.md`
- [X] T002 Update `.specify/feature.json` and `AGENTS.md` Spec Kit plan reference to `specs/035-browser-automation-foundation/plan.md`

---

## Phase 2: Foundational

**Purpose**: Add browser tool identity, catalog metadata, and mode filtering.

- [X] T003 [P] Add productdata tests for browser catalog metadata and Work/Chat RunContext filtering in `internal/productdata/tool_catalog_test.go` and `internal/productdata/service_test.go`
- [X] T004 Add browser tool constants, validation routing, persona allowlist entries, catalog metadata, and safe event metadata grouping in `internal/productdata/models.go`, `internal/productdata/builtin_personas.go`, `internal/productdata/tool_catalog.go`, and `internal/productdata/service.go`
- [X] T005 [P] Add frontend Settings Tools tests for browser metadata in `web/src/components/SettingsView.tools.test.tsx`
- [X] T006 Add frontend mock/client support for `browser` tool group metadata in `web/src/mockApiClient.ts` and `web/src/components/SettingsView.tsx`

---

## Phase 3: User Story 1 - Open and Snapshot a Public Page (Priority: P1)

**Goal**: Approved `browser.open` creates one run-scoped page session and `browser.snapshot` returns safe current page state.

**Independent Test**: A Work mode run requests browser open, approval executes exactly one bounded navigation, events show approval/execution/success, and provider continuation completes with a safe session snapshot.

- [X] T007 [P] [US1] Add browser executor tests for open, snapshot, title/text/link extraction, truncation, unsupported content, and safe metadata in `internal/runtime/browser_tools_test.go`
- [X] T008 [US1] Implement bounded browser session open/snapshot execution in `internal/runtime/browser_tools.go`
- [X] T009 [P] [US1] Add ToolBroker browser dispatch tests in `internal/runtime/tool_broker_test.go`
- [X] T010 [US1] Route browser tools through ToolBroker and runtime tool resolution in `internal/runtime/tool_broker.go`, `internal/runtime/tools.go`, and `internal/runtime/queued_runner.go`
- [X] T011 [P] [US1] Add worker tests proving approve-before-exec, provider continuation, and one browser open execution only in `internal/runtime/worker_test.go`
- [X] T012 [P] [US1] Add HTTP smoke for browser open approve -> execute -> final in `internal/httpapi/browser_automation_smoke_test.go`

---

## Phase 4: User Story 2 - Navigate by Approved Link Click (Priority: P2)

**Goal**: Approved `browser.click_link` navigates exactly one safe link in an existing run-scoped session.

**Independent Test**: Open a page, approve click_link for one link index, and verify session URL/snapshot changes with safe previous/final URL metadata.

- [X] T013 [P] [US2] Add browser click_link tests for safe link navigation, blocked link targets, unknown sessions, out-of-range indexes, and no raw HTML leakage in `internal/runtime/browser_tools_test.go`
- [X] T014 [US2] Implement browser click_link target resolution, session ownership checks, redirect validation, and safe result metadata in `internal/runtime/browser_tools.go`
- [X] T015 [P] [US2] Add bounded loop HTTP smoke for browser.open -> browser.click_link -> browser.snapshot -> final in `internal/httpapi/browser_automation_smoke_test.go`

---

## Phase 5: User Story 3 - Keep Browser Automation Safe and Visible (Priority: P3)

**Goal**: Unsafe modes/targets/lifecycle states do not execute, and Settings/RunRail visibly separate browser tools.

**Independent Test**: Settings Tools and RunRail render browser metadata without raw HTML, cookie, credential, secret, or local-path leakage.

- [X] T016 [P] [US3] Add safety tests for Chat-mode rejection, unsafe schemes, credentialed URLs, localhost/private/link-local hosts, blocked redirects, denied/stopped/terminal no-exec, duplicate/out-of-scope calls, and unsupported browser args in `internal/productdata/browser_models_test.go`, `internal/runtime/browser_tools_test.go`, `internal/runtime/gateway_test.go`, and `internal/runtime/worker_test.go`
- [X] T017 [US3] Implement browser validation and safe no-exec metadata paths in `internal/runtime/browser_tools.go`, `internal/runtime/gateway.go`, and `internal/productdata/service.go`
- [X] T018 [P] [US3] Add RunRail browser lifecycle tests in `web/src/components/RunRail.runtime.test.ts`
- [X] T019 [US3] Update RunRail copy and mock runtime data for browser lifecycle rows in `web/src/components/RunRail.tsx`, `web/src/runtime/runtimeScripts.ts`, and `web/src/mockData.ts`

---

## Phase 6: Documentation & Validation

**Purpose**: Record M27 behavior and prove the slice.

- [X] T020 [P] Add browser automation architecture docs in `docs-site/src/content/docs/architecture/browser-automation-foundation.md`
- [X] T021 [P] Add browser automation API/tool contract docs in `docs-site/src/content/docs/api/browser-automation-foundation.md`
- [X] T022 [P] Add M27 runbook and devlog in `docs-site/src/content/docs/runbooks/local-m27-browser-automation.md` and `docs-site/src/content/docs/devlog/2026-05-26-m27-browser-automation.md`
- [X] T023 Update roadmap/spec-kit docs and runbook index in `docs-site/src/content/docs/roadmap/current-status.md`, `docs-site/src/content/docs/spec-kit/workflow.md`, and `docs-site/src/content/docs/runbooks/index.md`
- [X] T024 Run `go test ./...`
- [X] T025 Run `bun test --cwd web`
- [X] T026 Run `bun run --cwd web build`
- [X] T027 Run `bun run --cwd docs-site build`
- [X] T028 Run `git diff --check`
- [X] T029 Perform browser smoke for Settings Tools browser tools, RunRail browser lifecycle visibility, and console errors

---

## Dependencies & Execution Order

- Phase 1 -> Phase 2 -> US1 -> US2 -> US3 -> Documentation/Validation.
- US1 creates browser session state and is the MVP.
- US2 depends on browser sessions and link summaries.
- US3 depends on catalog/runtime identity and mock lifecycle metadata.

## Parallel Opportunities

- T003 and T005 can run together.
- T007, T009, T011, and T012 are separate test files once browser constants exist.
- T013 and T015 can run after open/snapshot primitives exist.
- T018 can start after mock metadata shape is defined.
- T020, T021, and T022 can be drafted in parallel with final code review.

## Implementation Strategy

1. Complete Spec Kit context switch and catalog identity first.
2. Use RED -> GREEN for browser.open and browser.snapshot before click_link.
3. Add click_link only after session state and safe link extraction are bounded.
4. Add safety rejection tests before broadening UI claims.
5. Update docs and run full validation before claiming M27 completion.
