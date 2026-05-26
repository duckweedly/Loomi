# Tasks: M11 Tool Catalog Visibility

**Input**: Design documents from `specs/013-tool-catalog-visibility/`

**Tests**: Required. This feature changes backend API and Settings UI behavior.

## Phase 1: Setup

- [X] T001 Create M11 Spec Kit artifacts in `specs/013-tool-catalog-visibility/`
- [X] T002 Update `.specify/feature.json` to point to `specs/013-tool-catalog-visibility`

## Phase 2: Backend

- [X] T003 Add backend HTTP tests for `GET /v1/tools/catalog` deterministic entries and no secret-looking fields in `internal/httpapi/tools_test.go`
- [X] T004 Add runtime catalog unit tests for all allowlisted tools and risk metadata in `internal/runtime/tools_test.go`
- [X] T005 Implement runtime tool catalog entries in `internal/runtime/tools.go`
- [X] T006 Add `GET /v1/tools/catalog` route and handler in `internal/httpapi/server.go` and `internal/httpapi/tools.go`

## Phase 3: Frontend

- [X] T007 Add domain/API client types and real API mapping tests for tool catalog in `web/src/realApiClient.test.ts`
- [X] T008 Implement real/mock API client `getToolCatalog` support in `web/src/apiClient.ts`, `web/src/realApiClient.ts`, and `web/src/mockApiClient.ts`
- [X] T009 Add Settings Tools tests for read-only catalog rendering in `web/src/components/SettingsView.runtime.test.tsx`
- [X] T010 Implement Settings Tools panel and state wiring in `web/src/components/SettingsView.tsx`, `web/src/state.ts`, and `web/src/App.tsx`
- [X] T011 Update settings catalog category status/copy for Tools in `web/src/components/settingsCatalog.ts`

## Phase 4: Documentation

- [X] T012 Add architecture doc in `docs-site/src/content/docs/architecture/tool-catalog.md`
- [X] T013 Add API doc in `docs-site/src/content/docs/api/tool-catalog.md`
- [X] T014 Add runbook in `docs-site/src/content/docs/runbooks/local-m11.md`
- [X] T015 Add devlog in `docs-site/src/content/docs/devlog/2026-05-26-m11-tool-catalog.md`
- [X] T016 Update roadmap and Spec Kit workflow docs in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`

## Phase 5: Validation

- [X] T017 Run `go test ./...`
- [X] T018 Run `bun test --cwd web`
- [X] T019 Run `bun run --cwd web build`
- [X] T020 Run `bun run build` from `docs-site/`
- [X] T021 Run `git diff --check`
- [X] T022 Perform browser smoke for Settings Tools catalog and console errors
