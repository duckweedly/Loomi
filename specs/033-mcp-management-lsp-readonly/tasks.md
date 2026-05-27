# Tasks: M25 MCP Management + LSP Read-only Foundation

**Input**: Design documents from `specs/033-mcp-management-lsp-readonly/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD required. Each story includes test tasks before implementation tasks.

**Organization**: Tasks are grouped by independently testable user story.

## Phase 1: Setup

**Purpose**: Move active Spec Kit context from M24 to M25.

- [X] T001 Confirm M24 tasks are complete and record no remaining incomplete checkboxes in `specs/032-sandbox-exec-command-tools/tasks.md`
- [X] T002 Update `.specify/feature.json` and `AGENTS.md` Spec Kit plan reference to `specs/033-mcp-management-lsp-readonly/plan.md`

---

## Phase 2: Foundational

**Purpose**: Add shared MCP status and LSP identity/catalog metadata.

- [X] T003 [P] Add productdata tests for LSP catalog entries and Work/Chat RunContext filtering in `internal/productdata/tool_catalog_test.go` and `internal/productdata/service_test.go`
- [X] T004 Add LSP tool constants, validation routing, persona allowlist entries, catalog metadata, and safe event argument preview in `internal/productdata/models.go`, `internal/productdata/builtin_personas.go`, `internal/productdata/tool_catalog.go`, and `internal/productdata/service.go`
- [X] T005 [P] Add frontend Settings MCP and LSP catalog tests in `web/src/components/SettingsView.mcp.test.tsx` and `web/src/components/SettingsView.tools.test.tsx`
- [X] T006 Add MCP status frontend domain/client/mock state and LSP catalog rendering metadata in `web/src/domain.ts`, `web/src/realApiClient.ts`, `web/src/mockApiClient.ts`, and `web/src/components/SettingsView.tsx`

---

## Phase 3: User Story 1 - Inspect Local MCP Servers (Priority: P1)

**Goal**: Settings > MCP renders safe read-only status for local stdio MCP servers.

**Independent Test**: API and UI expose empty/configured/succeeded/failed MCP status without secrets or raw config.

- [X] T007 [P] [US1] Add runtime tests for safe MCP server status summaries from local config plus discovery metadata in `internal/runtime/mcp_management_test.go`
- [X] T008 [US1] Implement safe MCP status summary builder in `internal/runtime/mcp_management.go`
- [X] T009 [P] [US1] Add HTTP API tests for `GET /v1/mcp/servers` empty/configured/redacted states in `internal/httpapi/mcp_management_test.go`
- [X] T010 [US1] Add read-only MCP status API route and handler in `internal/httpapi/server.go` and related API files
- [X] T011 [US1] Implement Settings > MCP real read-only panel states in `web/src/components/SettingsView.tsx`

---

## Phase 4: User Story 2 - Expose Read-only LSP Tools (Priority: P2)

**Goal**: Approve and execute bounded read-only LSP diagnostics/symbols/references through the tool runtime.

**Independent Test**: A Work mode run requests an LSP tool, approval executes exactly one bounded read-only query, events show approval/execution/success, and provider continuation completes.

- [X] T012 [P] [US2] Add LSP executor tests for symbols, references, diagnostics, bounds, UTF-8 safety, and workspace-relative results in `internal/runtime/lsp_tools_test.go`
- [X] T013 [US2] Implement bounded read-only LSP tool execution in `internal/runtime/lsp_tools.go`
- [X] T014 [P] [US2] Add ToolBroker LSP dispatch tests in `internal/runtime/tool_broker_test.go`
- [X] T015 [US2] Route LSP tools through ToolBroker and runtime tool resolution in `internal/runtime/tool_broker.go` and `internal/runtime/tools.go`
- [X] T016 [P] [US2] Add worker tests proving approve-before-exec, provider continuation, and one LSP execution only in `internal/runtime/worker_test.go`
- [X] T017 [P] [US2] Add HTTP smoke for LSP approve -> execute -> final in `internal/httpapi/lsp_readonly_smoke_test.go`

---

## Phase 5: User Story 3 - Keep LSP and MCP Safe and Visible (Priority: P3)

**Goal**: Show LSP/MCP risk, scope, read-only status, and failure states without leaks.

**Independent Test**: Unsafe LSP paths/modes do not execute, and Settings/RunRail render safe metadata.

- [X] T018 [P] [US3] Add LSP safety tests for Chat-mode rejection, invalid arguments, traversal, absolute path, symlink escape, sensitive path, denied/stopped no-exec, and terminal no-op behavior in `internal/productdata/models_test.go`, `internal/runtime/lsp_tools_test.go`, `internal/runtime/gateway_test.go`, and `internal/runtime/worker_test.go`
- [X] T019 [US3] Implement LSP validation and safe no-exec metadata paths in `internal/runtime/lsp_tools.go`, `internal/runtime/gateway.go`, and `internal/productdata/service.go`
- [X] T020 [P] [US3] Add RunRail LSP lifecycle tests in `web/src/components/RunRail.runtime.test.ts`
- [X] T021 [US3] Update RunRail copy and mock runtime data for LSP lifecycle rows in `web/src/components/RunRail.tsx` and `web/src/mockData.ts`

---

## Phase 6: Documentation & Validation

**Purpose**: Record M25 behavior and prove the slice.

- [X] T022 [P] Add MCP/LSP architecture docs in `docs-site/src/content/docs/architecture/mcp-management-lsp-readonly.md`
- [X] T023 [P] Add MCP/LSP API docs in `docs-site/src/content/docs/api/mcp-management-lsp-readonly.md`
- [X] T024 [P] Add M25 runbook and devlog in `docs-site/src/content/docs/runbooks/local-m25-mcp-lsp-readonly.md` and `docs-site/src/content/docs/devlog/2026-05-26-m25-mcp-lsp-readonly.md`
- [X] T025 Update roadmap/spec-kit docs and runbook index in `docs-site/src/content/docs/roadmap/current-status.md`, `docs-site/src/content/docs/spec-kit/workflow.md`, and `docs-site/src/content/docs/runbooks/index.md`
- [X] T026 Run `go test ./...`
- [X] T027 Run `bun test --cwd web`
- [X] T028 Run `bun run --cwd web build`
- [X] T029 Run `bun run --cwd docs-site build`
- [X] T030 Run `git diff --check`
- [X] T031 Perform browser smoke for Settings MCP, Settings Tools LSP, RunRail LSP visibility, and console errors

---

## Dependencies & Execution Order

- Phase 1 -> Phase 2 -> US1 -> US2 -> US3 -> Documentation/Validation.
- US1 is independently useful and can ship before executable LSP tools.
- US2 depends on catalog identity and workspace scope reuse.
- US3 hardens mode/path safety and visible audit metadata before M25 closeout.

## Parallel Opportunities

- T003 and T005 can run together.
- T007 and T009 are separate backend test files once the status contract is defined.
- T012, T014, T016, and T017 are separate test files once LSP constants exist.
- T020 can start after mock metadata is defined.
- T022, T023, and T024 can be drafted in parallel with final code review.

## Implementation Strategy

1. Complete Spec Kit context switch and catalog identity first.
2. Use RED -> GREEN for MCP read-only status API.
3. Use RED -> GREEN for one LSP executor behavior at a time.
4. Add safety rejection tests before broadening UI claims.
5. Update docs and run full validation before claiming M25 completion.
