# Tasks: M13 MCP Call Tool Bridge

**Input**: Design documents from `specs/015-mcp-call-tool-bridge/`

**Tests**: Required. This feature changes tool validation, execution, catalog, and frontend rendering behavior.

## Phase 1: Setup

- [X] T001 Create M13 Spec Kit artifacts in `specs/015-mcp-call-tool-bridge/`
- [X] T002 Update `.specify/feature.json` to point to `specs/015-mcp-call-tool-bridge`

## Phase 2: Backend TDD

- [X] T003 Add productdata tests for `mcp.call_tool` request validation in `internal/productdata/service_test.go`
- [X] T004 Add runtime tests for MCP call normalization, result summary, and catalog metadata in `internal/runtime/tools_test.go`
- [X] T005 Add worker test for approved `mcp.call_tool` execution in `internal/runtime/worker_test.go`
- [X] T006 Update HTTP catalog test expectation in `internal/httpapi/tools_test.go`
- [X] T007 Implement productdata allowlist and bounded MCP argument validation in `internal/productdata/models.go`
- [X] T008 Implement runtime MCP tool definition, normalization, execution, and catalog entry in `internal/runtime/tools.go`
- [X] T009 Wire worker tool lookup for `mcp.call_tool` in `internal/runtime/worker.go`

## Phase 3: Frontend

- [X] T010 Add ToolCallCard test for MCP summaries in `web/src/components/ToolCallCard.test.tsx`
- [X] T011 Add `mcp.call_tool` to mock tool catalog in `web/src/mockApiClient.ts`

## Phase 4: Documentation

- [X] T012 Add architecture doc in `docs-site/src/content/docs/architecture/mcp-call-tool-bridge.md`
- [X] T013 Add API doc in `docs-site/src/content/docs/api/mcp-call-tool-bridge.md`
- [X] T014 Add runbook in `docs-site/src/content/docs/runbooks/local-m13.md`
- [X] T015 Add devlog in `docs-site/src/content/docs/devlog/2026-05-26-m13-mcp-call-tool-bridge.md`
- [X] T016 Update roadmap and Spec Kit workflow docs in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`

## Phase 5: Validation

- [X] T017 Run `go test ./...`
- [X] T018 Run `bun test --cwd web`
- [X] T019 Run `bun run --cwd web build`
- [X] T020 Run `bun run build` from `docs-site/`
- [X] T021 Run `git diff --check`
- [X] T022 Perform browser smoke for Settings Tools catalog and mcp.call_tool visibility
