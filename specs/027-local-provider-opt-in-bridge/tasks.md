# Tasks: Local Provider Opt-in Bridge

**Input**: Design documents from `/specs/027-local-provider-opt-in-bridge/`

## Phase 1: Setup

- [x] T001 Update AGENTS.md current Spec Kit feature pointer to specs/027-local-provider-opt-in-bridge/plan.md
- [x] T002 Create feature artifacts under specs/027-local-provider-opt-in-bridge/

## Phase 2: Backend Tests

- [x] T003 [P] Add local provider enable/disable API tests in internal/httpapi/runtime_test.go
- [x] T004 [P] Add model provider list tests for detected-but-not-enabled and safe enabled capability in internal/httpapi/runtime_test.go
- [x] T005 [P] Add unsupported/unavailable enable rejection tests in internal/httpapi/runtime_test.go

## Phase 3: Backend Implementation

- [x] T006 Extend safe ProviderCapability metadata in internal/runtime/providers.go
- [x] T007 Add local provider enablement-to-capability helpers in internal/runtime/local_provider_detection.go
- [x] T008 Add session-local enablement state and enable/disable routes in internal/httpapi/server.go and internal/httpapi/runtime.go
- [x] T009 Ensure model provider listing includes only explicitly enabled local providers in internal/httpapi/runtime.go

## Phase 4: Web Tests

- [x] T010 [P] Add realApiClient mapping/API tests for enable/disable and safe metadata in web/src/realApiClient.test.ts
- [x] T011 [P] Add state tests proving local enable/disable actions update configured providers in web/src/state.test.ts
- [x] T012 [P] Add Settings UI contract tests for enable/disable/unsupported copy in web/src/components/SettingsView.runtime.test.tsx
- [x] T013 [P] Add Chat warning test for enabled-but-unsupported local provider in web/src/components/ChatCanvas.states.test.ts

## Phase 5: Web Implementation

- [x] T014 Extend provider domain/API types in web/src/domain.ts, web/src/apiClient.ts, and web/src/realApiClient.ts
- [x] T015 Add mock client enable/disable behavior in web/src/mockApiClient.ts
- [x] T016 Add state actions for enabling/disabling local providers in web/src/state.ts
- [x] T017 Wire Settings local provider action buttons and copy in web/src/components/SettingsView.tsx and web/src/i18n.ts
- [x] T018 Keep Chat provider warning blocking unsupported local providers in web/src/runtime/backendCapabilityStatus.ts

## Phase 6: Documentation & Validation

- [x] T019 Update docs-site architecture/api/runbook/devlog/current-status/workflow pages
- [x] T020 Run full validation commands and record results
