# Tasks: Memory Provider Foundation

**Input**: Design documents from `/specs/042-memory-provider-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Phase 1: Plan And Test Baseline

- [X] T001 Confirm existing memory API and Settings tests still run or document current unrelated failures. Baseline: `bun test web/src/components/SettingsView.runtime.test.tsx web/src/memory.test.ts` passes; `go test ./internal/productdata ./internal/httpapi ./internal/runtime` has unrelated existing failures in work-mode tool scoping and web.fetch smoke.
- [X] T002 Add failing productdata tests for default local provider config, unknown-provider degraded fallback, and diagnostic redaction.
- [X] T003 Add failing HTTP API tests for `GET /v1/memory/provider` and `PUT /v1/memory/provider`.
- [X] T004 Add failing runtime test for safe memory readiness metadata during run preparation.
- [X] T005 Add failing web tests for Settings > Memory backend-derived provider status.

## Phase 2: Backend Provider Foundation

- [X] T006 Add memory provider config/status/readiness models in `internal/productdata/models.go`.
- [X] T007 Implement provider config persistence and normalization in `internal/productdata/service.go` and repository boundary.
- [X] T008 Implement provider status resolution and redaction in `internal/productdata`.
- [X] T009 Add memory provider HTTP handlers in `internal/httpapi/memory.go`.
- [X] T010 Wire provider routes in `internal/httpapi/server.go` without changing existing memory endpoints.
- [X] T011 Add safe memory readiness projection in `internal/runtime` run preparation path.

## Phase 3: Settings UI

- [X] T012 Add provider status/update types and API client methods in `web/src`.
- [X] T013 Update `web/src/components/SettingsView.tsx` to show backend memory provider state and refresh/update actions.
- [X] T014 Keep existing memory management list/search/detail/audit/delete behavior unchanged.

## Phase 4: Documentation And Validation

- [X] T015 Update docs-site architecture/API/runbook/devlog pages for memory provider foundation and non-goals.
- [X] T016 Run focused Go validation.
- [X] T017 Run focused web validation.
- [X] T018 Run docs-site build.
- [X] T019 Run browser smoke for Settings > Memory.
