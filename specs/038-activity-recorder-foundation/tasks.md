# Tasks: M30 Activity Recorder Foundation

**Input**: Design documents from `specs/038-activity-recorder-foundation/`

**Tests**: TDD required.

## Phase 1: Setup

- [X] T001 Create M30 Spec Kit artifacts and move active feature pointers.

## Phase 2: Foundation

- [ ] T002 [P] Add productdata tests for disabled-by-default status, enable/disable, append rejection while disabled, list bounds, clear idempotency, and redaction in `internal/productdata/activity_recorder_test.go`.
- [ ] T003 Add activity recorder models, ids, service interface, and memory service state in `internal/productdata/models.go` and `internal/productdata/service.go`.
- [ ] T004 Add redaction/bounds helpers for activity summaries and metadata in `internal/productdata/service.go`.

## Phase 3: User Story 1 - Explicit Opt-in

- [ ] T005 [P] Add HTTP tests for status, enable, disable, and append rejection while disabled in `internal/httpapi/activity_recorder_test.go`.
- [ ] T006 Implement `/v1/activity-recorder/status`, `/enable`, and `/disable` handlers in `internal/httpapi/activity_recorder.go` and route them from `internal/httpapi/server.go`.
- [ ] T007 Implement append rejection while disabled and safe error responses in `internal/httpapi/activity_recorder.go`.

## Phase 4: User Story 2 - View Bounded Activity Summaries

- [ ] T008 [P] Add productdata and HTTP tests for append/list ordering, limit cap, supported kinds, redaction, and metadata safety.
- [ ] T009 Implement append/list service methods and HTTP `POST/GET /v1/activity-recorder/events`.
- [ ] T010 [P] Add frontend domain/client/mock tests for Activity Recorder status/events mapping in `web/src/realApiClient.test.ts` and `web/src/mockApiClient.test.ts`.
- [ ] T011 Add frontend domain types and real/mock API client methods in `web/src/domain.ts`, `web/src/apiClient.ts`, `web/src/realApiClient.ts`, and `web/src/mockApiClient.ts`.

## Phase 5: User Story 3 - Settings Control and Cleanup

- [ ] T012 [P] Add Activity Recorder panel rendering tests for disabled/enabled/empty/error/redaction/clear confirmation states in `web/src/components/ActivityRecorderPanel.test.tsx`.
- [ ] T013 Add `ActivityRecorderPanel` and wire Settings > Activity Recorder as a real panel in `web/src/components/ActivityRecorderPanel.tsx`, `web/src/components/SettingsView.tsx`, and `web/src/components/settingsCatalog.ts`.
- [ ] T014 [P] Add state tests for loading status/events, latest-request guards, enable/disable, append refresh, and clear refresh in `web/src/state.test.ts`.
- [ ] T015 Wire Activity Recorder state and handlers through `web/src/state.ts` and `web/src/App.tsx`.
- [ ] T016 Add backend and frontend clear tests, then implement idempotent `DELETE /v1/activity-recorder/events`.

## Phase 6: Documentation & Validation

- [ ] T017 [P] Add Activity Recorder architecture docs in `docs-site/src/content/docs/architecture/activity-recorder-foundation.md`.
- [ ] T018 [P] Add Activity Recorder API contract docs in `docs-site/src/content/docs/api/activity-recorder-foundation.md`.
- [ ] T019 [P] Add M30 runbook and devlog in `docs-site/src/content/docs/runbooks/local-m30-activity-recorder.md` and `docs-site/src/content/docs/devlog/2026-05-26-m30-activity-recorder.md`.
- [ ] T020 Update roadmap/spec-kit docs and runbook index for M30.
- [ ] T021 Run `go test ./...`.
- [ ] T022 Run `bun test --cwd web`.
- [ ] T023 Run `bun run --cwd web build`.
- [ ] T024 Run `bun run --cwd docs-site build`.
- [ ] T025 Run `git diff --check`.
- [ ] T026 Perform browser smoke for Settings > Activity Recorder status, event rows, redaction marker, cleanup affordance visibility, and console errors.

## Dependencies

- Phase 2 blocks backend API work.
- User Story 1 blocks event append/list because append requires opt-in state.
- User Story 2 blocks Settings event list rendering.
- User Story 3 blocks final browser smoke.

## MVP

Complete T002-T011 for API-level opt-in and safe event list before polishing Settings cleanup.
