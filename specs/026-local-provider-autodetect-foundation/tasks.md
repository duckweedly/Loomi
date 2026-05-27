# Tasks: Local Provider Autodetect Foundation

**Input**: Design documents from `/specs/026-local-provider-autodetect-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required. Use test-first implementation for detector, HTTP API, and Settings UI behavior.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

**Status**: Complete candidate

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Pin feature context and docs targets.

- [x] T001 Update AGENTS.md current Spec Kit feature pointer to specs/026-local-provider-autodetect-foundation/plan.md
- [x] T002 [P] Add docs placeholders for architecture/API/runbook/devlog in docs-site/src/content/docs/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared local provider capability model.

- [x] T003 [P] Add local provider detection API contract types in internal/runtime/local_provider_detection.go
- [x] T004 [P] Add frontend local provider detection domain/client types in web/src/domain.ts and web/src/realApiClient.ts

**Checkpoint**: Foundation ready for detector/API/UI stories.

---

## Phase 3: User Story 1 - Detect local provider availability safely (Priority: P1) MVP

**Goal**: Detector returns safe Claude Code and Codex capabilities from temp fixtures.

**Independent Test**: `go test ./internal/runtime -run TestLocalProvider`

### Tests for User Story 1

- [x] T005 [P] [US1] Add Claude primaryApiKey fixture test in internal/runtime/local_provider_detection_test.go
- [x] T006 [P] [US1] Add Claude settings env fixture test in internal/runtime/local_provider_detection_test.go
- [x] T007 [P] [US1] Add Claude apiKeyHelper unsupported/no-execution test in internal/runtime/local_provider_detection_test.go
- [x] T008 [P] [US1] Add Codex auth.json API key fixture test in internal/runtime/local_provider_detection_test.go
- [x] T009 [P] [US1] Add Codex OAuth token fixture test in internal/runtime/local_provider_detection_test.go
- [x] T010 [P] [US1] Add CODEX_API_KEY env precedence test in internal/runtime/local_provider_detection_test.go
- [x] T011 [P] [US1] Add missing files and temp HOME isolation tests in internal/runtime/local_provider_detection_test.go

### Implementation for User Story 1

- [x] T012 [US1] Implement Claude Code detector in internal/runtime/local_provider_detection.go
- [x] T013 [US1] Implement Codex detector in internal/runtime/local_provider_detection.go
- [x] T014 [US1] Implement shared redaction and safe model candidate handling in internal/runtime/local_provider_detection.go

**Checkpoint**: User Story 1 detector works independently.

---

## Phase 4: User Story 2 - Expose a safe read-only API surface (Priority: P2)

**Goal**: Backend endpoint returns safe local provider detection without side effects.

**Independent Test**: `go test ./internal/httpapi -run TestLocalProviderDetection`

### Tests for User Story 2

- [x] T015 [P] [US2] Add endpoint safe provider test in internal/httpapi/runtime_test.go
- [x] T016 [P] [US2] Add endpoint secret/path exclusion test in internal/httpapi/runtime_test.go
- [x] T017 [P] [US2] Add disabled/unsupported stable status test in internal/httpapi/runtime_test.go

### Implementation for User Story 2

- [x] T018 [US2] Add GET /v1/local-provider-detections route in internal/httpapi/server.go
- [x] T019 [US2] Add handler and response mapping in internal/httpapi/runtime.go

**Checkpoint**: User Story 2 API works independently.

---

## Phase 5: User Story 3 - Show safe Settings provider status (Priority: P3)

**Goal**: Settings > Providers shows Local Claude Code and Local Codex status without enabling them.

**Independent Test**: `bun test --cwd web src/components/SettingsView.runtime.test.tsx src/realApiClient.test.ts src/state.test.ts`

### Tests for User Story 3

- [x] T020 [P] [US3] Add real API local detection mapping test in web/src/realApiClient.test.ts
- [x] T021 [P] [US3] Add state test proving local detections are separate from configured providers in web/src/state.test.ts
- [x] T022 [P] [US3] Add Settings provider UI source/render tests in web/src/components/SettingsView.runtime.test.tsx

### Implementation for User Story 3

- [x] T023 [US3] Add ApiClient local provider detection method in web/src/apiClient.ts and web/src/mockApiClient.ts
- [x] T024 [US3] Load local provider detections in web/src/state.ts without changing providerCapabilities
- [x] T025 [US3] Render Local Claude Code and Local Codex autodetect cards in web/src/components/SettingsView.tsx
- [x] T026 [US3] Add localized detected/not detected/explicit opt-in/no secrets copy in web/src/i18n.ts

**Checkpoint**: User Story 3 Settings UI works independently.

---

## Phase 6: Documentation & Closeout

**Purpose**: Docs, Spec Kit status, and validation.

- [x] T027 [P] Update architecture doc in docs-site/src/content/docs/architecture/local-provider-autodetect.md
- [x] T028 [P] Update API doc in docs-site/src/content/docs/api/local-provider-autodetect.md
- [x] T029 [P] Update runbook in docs-site/src/content/docs/runbooks/local-m18-5-provider-autodetect.md
- [x] T030 [P] Update devlog in docs-site/src/content/docs/devlog/2026-05-25-m18-5-local-provider-autodetect.md
- [x] T031 Update roadmap status in docs-site/src/content/docs/roadmap/current-status.md
- [x] T032 Update Spec Kit workflow status in docs-site/src/content/docs/spec-kit/workflow.md
- [x] T033 Mark specs/026-local-provider-autodetect-foundation/spec.md and tasks.md as complete candidate after validation
- [x] T034 Run go test ./...
- [x] T035 Run bun test --cwd web
- [x] T036 Run bun run --cwd web build
- [x] T037 Run bun run --cwd docs-site build
- [x] T038 Run git diff --check

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Phase 1
- **US1 (Phase 3)**: Depends on Phase 2
- **US2 (Phase 4)**: Depends on US1 detector
- **US3 (Phase 5)**: Depends on US2 API contract
- **Documentation & Closeout (Phase 6)**: Depends on implemented stories

### User Story Dependencies

- **US1**: No dependency after foundation
- **US2**: Requires US1 detector
- **US3**: Requires US2 endpoint shape

### Parallel Opportunities

- T002-T004 can be prepared in parallel.
- US1 tests T005-T011 target one test file but independent cases.
- Docs T027-T030 can be drafted in parallel after implementation behavior is known.

## Implementation Strategy

1. Establish docs/spec context.
2. Write detector tests and watch them fail.
3. Implement minimal detector.
4. Write endpoint tests and watch them fail.
5. Implement endpoint.
6. Write web tests and watch them fail.
7. Implement Settings read-only display.
8. Update docs and run full validation.
