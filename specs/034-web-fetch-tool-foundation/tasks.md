# Tasks: M26 Web Fetch Tool Foundation

**Input**: Design documents from `specs/034-web-fetch-tool-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD required. Each story includes test tasks before implementation tasks.

**Organization**: Tasks are grouped by independently testable user story.

## Phase 1: Setup

**Purpose**: Move active Spec Kit context from M25 to M26.

- [X] T001 Confirm M25 tasks are complete and record no remaining incomplete checkboxes in `specs/033-mcp-management-lsp-readonly/tasks.md`
- [X] T002 Update `.specify/feature.json` and `AGENTS.md` Spec Kit plan reference to `specs/034-web-fetch-tool-foundation/plan.md`

---

## Phase 2: Foundational

**Purpose**: Add shared web tool identity, catalog metadata, and mode filtering.

- [X] T003 [P] Add productdata tests for `web.fetch` catalog metadata and Work/Chat RunContext filtering in `internal/productdata/tool_catalog_test.go` and `internal/productdata/service_test.go`
- [X] T004 Add `web.fetch` constants, validation routing, persona allowlist entries, catalog metadata, and safe event argument preview in `internal/productdata/models.go`, `internal/productdata/builtin_personas.go`, and `internal/productdata/tool_catalog.go`
- [X] T005 [P] Add frontend Settings Tools tests for `web.fetch` metadata in `web/src/components/SettingsView.tools.test.tsx`
- [X] T006 Add frontend domain/mock/client support for `web` tool group metadata in `web/src/domain.ts`, `web/src/mockApiClient.ts`, and `web/src/components/SettingsView.tsx`

---

## Phase 3: User Story 1 - Fetch a Public Web Page Safely (Priority: P1)

**Goal**: Approved Work mode `web.fetch` executes one bounded public HTTP(S) read and returns a safe summary.

**Independent Test**: A Work mode run requests `web.fetch`, approval executes exactly one bounded fetch, events show approval/execution/success, and provider continuation completes.

- [X] T007 [P] [US1] Add web executor tests for allowed HTTP(S) fetch, title/excerpt extraction, byte truncation, timeout clamp, unsupported content type, and safe metadata in `internal/runtime/web_tools_test.go`
- [X] T008 [US1] Implement bounded read-only web fetch execution in `internal/runtime/web_tools.go`
- [X] T009 [P] [US1] Add ToolBroker web dispatch tests in `internal/runtime/tool_broker_test.go`
- [X] T010 [US1] Route web tools through ToolBroker and runtime tool resolution in `internal/runtime/tool_broker.go`, `internal/runtime/tools.go`, and `internal/runtime/queued_runner.go`
- [X] T011 [P] [US1] Add worker tests proving approve-before-exec, provider continuation, and one web fetch execution only in `internal/runtime/worker_test.go`
- [X] T012 [P] [US1] Add HTTP smoke for web fetch approve -> execute -> final in `internal/httpapi/web_fetch_tool_smoke_test.go`

---

## Phase 4: User Story 2 - Keep Web Fetch Out of Chat and Private Networks (Priority: P2)

**Goal**: Unsafe mode, URL, redirect, and lifecycle states fail before network execution.

**Independent Test**: Chat mode and unsafe/private targets are rejected before approval/execution or before reading redirected bodies.

- [X] T013 [P] [US2] Add safety tests for Chat-mode rejection, unsafe schemes, credentialed URLs, localhost/private/link-local hosts, blocked redirects, denied/stopped/terminal no-exec, and duplicate/out-of-scope calls in `internal/productdata/web_models_test.go`, `internal/runtime/web_tools_test.go`, `internal/runtime/worker_test.go`, and existing lifecycle guard tests
- [X] T014 [US2] Implement web validation, DNS/private-network checks, redirect validation, and safe no-exec metadata paths in `internal/runtime/web_tools.go`, `internal/runtime/gateway.go`, and `internal/productdata/service.go`

---

## Phase 5: User Story 3 - Make Web Fetch Visible in Settings and RunRail (Priority: P3)

**Goal**: Settings and RunRail show web risk/scope/read-only status and safe lifecycle metadata.

**Independent Test**: Settings Tools and RunRail render `web.fetch` metadata without raw response body, cookie, credential, secret, or local-path leakage.

- [X] T015 [P] [US3] Add RunRail web lifecycle tests in `web/src/components/RunRail.runtime.test.ts`
- [X] T016 [US3] Update RunRail copy and mock runtime data for web lifecycle rows in `web/src/components/RunRail.tsx`, `web/src/runtime/runtimeScripts.ts`, and `web/src/mockData.ts`

---

## Phase 6: Documentation & Validation

**Purpose**: Record M26 behavior and prove the slice.

- [X] T017 [P] Add web fetch architecture docs in `docs-site/src/content/docs/architecture/web-fetch-tool.md`
- [X] T018 [P] Add web fetch API/tool contract docs in `docs-site/src/content/docs/api/web-fetch-tool.md`
- [X] T019 [P] Add M26 runbook and devlog in `docs-site/src/content/docs/runbooks/local-m26-web-fetch-tool.md` and `docs-site/src/content/docs/devlog/2026-05-26-m26-web-fetch-tool.md`
- [X] T020 Update roadmap/spec-kit docs and runbook index in `docs-site/src/content/docs/roadmap/current-status.md`, `docs-site/src/content/docs/spec-kit/workflow.md`, and `docs-site/src/content/docs/runbooks/index.md`
- [X] T021 Run `go test ./...`
- [X] T022 Run `bun test --cwd web`
- [X] T023 Run `bun run --cwd web build`
- [X] T024 Run `bun run --cwd docs-site build`
- [X] T025 Run `git diff --check`
- [X] T026 Perform browser smoke for Settings Tools `web.fetch`, RunRail web lifecycle visibility, and console errors

---

## Dependencies & Execution Order

- Phase 1 -> Phase 2 -> US1 -> US2 -> US3 -> Documentation/Validation.
- US1 is independently useful and can ship before browser/search/artifact expansion.
- US2 depends on web executor validation hooks.
- US3 depends on catalog identity and mock runtime metadata.

## Parallel Opportunities

- T003 and T005 can run together.
- T007, T009, T011, and T012 are separate test files once `web.fetch` constants exist.
- T015 can start after mock metadata shape is defined.
- T017, T018, and T019 can be drafted in parallel with final code review.

## Implementation Strategy

1. Complete Spec Kit context switch and catalog identity first.
2. Use RED -> GREEN for executor URL validation and safe fetch result.
3. Route through ToolBroker/worker only after executor behavior is bounded.
4. Add safety rejection tests before broadening UI claims.
5. Update docs and run full validation before claiming M26 completion.
