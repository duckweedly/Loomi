# Tasks: Local Codex Execution Bridge

**Input**: Design documents from `/specs/028-local-codex-execution-bridge/`

## Phase 1: Setup

- [x] T001 Update AGENTS.md current Spec Kit feature pointer to specs/028-local-codex-execution-bridge/plan.md
- [x] T002 Create feature artifacts under specs/028-local-codex-execution-bridge/

## Phase 2: Backend Tests

- [x] T003 [P] [US1] Add temp CODEX_HOME detect/enable available-supported capability test in internal/httpapi/runtime_test.go
- [x] T004 [P] [US2] Add fixture Local Codex gateway provider assistant message test in internal/runtime/gateway_test.go
- [x] T005 [P] [US2] Add Chat HTTP smoke for create thread/message/run -> worker -> final assistant message in internal/httpapi/runtime_test.go
- [x] T006 [P] [US3] Add redaction canary tests for API response, run events, assistant metadata, and logs/test output in internal/httpapi/runtime_test.go
- [x] T007 [P] [US3] Add unsupported local provider start-run rejection test in internal/httpapi/runtime_test.go
- [x] T008 [P] [US3] Add concurrent enable/disable/list/save/check test in internal/httpapi/runtime_test.go
- [x] T009 [P] [US3] Add OpenAI-compatible save/check regression assertion in internal/httpapi/runtime_test.go

## Phase 3: Backend Implementation

- [x] T010 [US1] Add Local Codex credential snapshot loader and redaction-safe capability helpers in internal/runtime/local_provider_detection.go
- [x] T011 [US2] Implement Local Codex runtime.Provider bridge in internal/runtime/local_codex_provider.go
- [x] T012 [US1] Register enabled Local Codex with the existing Gateway and provider list in internal/httpapi/runtime.go and internal/runtime/gateway.go
- [x] T013 [US3] Keep provider request/response metadata redacted in internal/runtime/gateway.go
- [x] T014 [US3] Make localProviderEnablements/provider access concurrency-safe across list/save/check/enable/disable in internal/httpapi/server.go and internal/httpapi/runtime.go
- [x] T015 [US3] Keep HTTP start-run guard rejecting unsupported/unavailable local provider states in internal/httpapi/runtime.go

## Phase 4: Web Tests

- [x] T016 [P] [US3] Add Chat available/supported test proving no provider unavailable warning and Composer can send in web/src/components/ChatCanvas.states.test.ts
- [x] T017 [P] [US3] Add unsupported/unavailable Local Codex warning copy tests in web/src/components/ChatCanvas.states.test.ts
- [x] T018 [P] [US3] Add Settings detect/enable/disable/supported state tests in web/src/components/SettingsView.runtime.test.tsx
- [x] T019 [P] [US3] Add API client mapping redaction tests in web/src/realApiClient.test.ts

## Phase 5: Web Implementation

- [x] T020 [US3] Update provider readiness mapping for supported/unavailable/unsupported Local Codex in web/src/runtime/backendCapabilityStatus.ts
- [x] T021 [US3] Update Local Codex Chat warning copy in web/src/i18n.ts and Chat state consumers
- [x] T022 [US3] Update Settings supported-state labels/actions in web/src/components/SettingsView.tsx
- [x] T023 [US3] Ensure API clients map no secret/path fields in web/src/realApiClient.ts and web/src/mockApiClient.ts

## Phase 6: Documentation & Validation

- [x] T024 [P] Add docs-site architecture/api/runbook/devlog pages for M20
- [x] T025 [P] Update docs-site roadmap/current-status and spec-kit/workflow for M20
- [x] T026 Run go test ./...
- [x] T027 Run bun test --cwd web
- [x] T028 Run bun run --cwd web build
- [x] T029 Run bun run --cwd docs-site build
- [x] T030 Run git diff --check

## Dependencies & Execution Order

- Phase 1 must complete before tests and implementation.
- Backend tests should fail before backend implementation.
- Web tests should fail before web implementation.
- Docs can proceed after behavior is settled.
- Validation runs last.

## Implementation Strategy

MVP is US1 + US2: enable Local Codex as available/supported and send one Chat message through worker/gateway. US3 safety tests run alongside and remain required before completion.
