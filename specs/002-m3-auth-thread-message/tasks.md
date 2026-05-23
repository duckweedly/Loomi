# Tasks: M3 Auth, Thread, and Message

**Input**: Design documents from `specs/002-m3-auth-thread-message/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Included because the feature spec defines independent tests for each user story and the implementation plan requires Go tests, migration smoke checks, frontend smoke checks, and docs validation.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested as an independent increment after shared setup/foundation work.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel with other `[P]` tasks in the same phase because it touches different files and has no dependency on incomplete tasks.
- **[Story]**: Maps the task to a specific user story from `spec.md`.
- Every task includes exact file paths.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the M3 implementation file seams defined by `plan.md` without adding behavior yet.

- [X] T001 Create M3 backend package files in `internal/identity/local.go`, `internal/productdata/models.go`, `internal/productdata/repository.go`, `internal/productdata/service.go`, `internal/httpapi/errors.go`, and `internal/httpapi/product.go`
- [X] T002 [P] Create M3 backend test files in `internal/identity/local_test.go`, `internal/productdata/repository_test.go`, `internal/productdata/service_test.go`, and `internal/httpapi/product_test.go`
- [X] T003 [P] Create explicit seed command files in `cmd/loomi-seed/main.go` and `cmd/loomi-seed/main_test.go`
- [X] T004 [P] Create frontend data-source files in `web/src/mockApiClient.ts` and `web/src/realApiClient.ts`
- [X] T005 [P] Create M3 documentation page files in `docs-site/src/content/docs/architecture/auth-thread-message.md`, `docs-site/src/content/docs/api/thread-message.md`, `docs-site/src/content/docs/runbooks/local-m3.md`, and `docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add the shared database/schema, model, and HTTP plumbing prerequisites that all user stories depend on.

**CRITICAL**: No user story work should begin until this phase is complete.

- [X] T006 [P] Add M3 up migration for `users`, `threads`, `messages`, constraints, and indexes in `migrations/000002_m3_auth_thread_message.up.sql`
- [X] T007 [P] Add M3 down migration that removes `messages`, `threads`, and `users` in `migrations/000002_m3_auth_thread_message.down.sql`
- [X] T008 [P] Define shared M3 model types, lifecycle constants, validation limits, and ID prefix helpers in `internal/productdata/models.go`
- [X] T009 [P] Implement structured API error envelope helpers with `code`, `message`, and `request_id` in `internal/httpapi/errors.go`
- [X] T010 Add product repository and service interfaces used by HTTP handlers in `internal/productdata/service.go`
- [X] T011 Add product route registration scaffolding without story-specific behavior in `internal/httpapi/server.go`

**Checkpoint**: M3 files, migration pair, shared model constants, structured errors, and route seams exist; user story implementation can start.

---

## Phase 3: User Story 1 - Use a local identity to own product data (Priority: P1) MVP

**Goal**: Expose one fixed local development user and ensure all product data operations can resolve durable ownership without sessions or user-selected headers.

**Independent Test**: Start the API after M3 migration, request `GET /v1/me`, create a thread and message through later stories, and confirm records are associated with `user_local_dev`.

### Tests for User Story 1

- [X] T012 [P] [US1] Add fixed local identity tests for id, display name, and no request-selected identity in `internal/identity/local_test.go`
- [X] T013 [P] [US1] Add user upsert repository tests for durable fixed local user creation in `internal/productdata/repository_test.go`
- [X] T014 [P] [US1] Add current identity service tests for ensuring the local user before returning it in `internal/productdata/service_test.go`
- [X] T015 [P] [US1] Add `GET /v1/me` handler tests for response shape and `request_id` in `internal/httpapi/product_test.go`

### Implementation for User Story 1

- [X] T016 [US1] Implement fixed local identity constants and resolver in `internal/identity/local.go`
- [X] T017 [US1] Implement user upsert and fetch repository methods in `internal/productdata/repository.go`
- [X] T018 [US1] Implement current identity service method in `internal/productdata/service.go`
- [X] T019 [US1] Implement `GET /v1/me` handler in `internal/httpapi/product.go`
- [X] T020 [US1] Wire the product service and `/v1/me` route into `internal/httpapi/server.go` and `cmd/loomi-api/main.go`

**Checkpoint**: `GET /v1/me` returns the stable local development user with creation metadata and a request id.

---

## Phase 4: User Story 2 - Manage real threads without coupling them to runs (Priority: P1)

**Goal**: Support creating, listing, retrieving, renaming, and archiving durable threads owned by the fixed local user, with thread lifecycle separate from run status.

**Independent Test**: Create a thread, list active threads in most-recently-updated order, retrieve it, update title/mode, archive it, confirm active list excludes it, and confirm direct retrieval still works.

### Tests for User Story 2

- [X] T021 [P] [US2] Add thread repository tests for create, active list order, get by owner, update, and archive in `internal/productdata/repository_test.go`
- [X] T022 [P] [US2] Add thread service tests for title/mode validation and owner-scoped thread lifecycle behavior in `internal/productdata/service_test.go`
- [X] T023 [P] [US2] Add thread HTTP handler tests for `POST /v1/threads`, `GET /v1/threads`, `GET /v1/threads/{thread_id}`, `PATCH /v1/threads/{thread_id}`, and `POST /v1/threads/{thread_id}/archive` in `internal/httpapi/product_test.go`

### Implementation for User Story 2

- [X] T024 [US2] Implement thread title, mode, and lifecycle validation helpers in `internal/productdata/models.go`
- [X] T025 [US2] Implement owner-scoped thread repository methods in `internal/productdata/repository.go`
- [X] T026 [US2] Implement create, list, get, update, and archive thread service methods in `internal/productdata/service.go`
- [X] T027 [US2] Implement thread HTTP handlers and request/response mapping in `internal/httpapi/product.go`
- [X] T028 [US2] Register thread routes in `internal/httpapi/server.go`

**Checkpoint**: Thread API is independently usable through curl without creating or implying any run state.

---

## Phase 5: User Story 3 - Persist and reload real messages (Priority: P1)

**Goal**: Persist complete user-authored messages inside owned threads, list them in stable order, and return an existing message on duplicate `client_message_id`.

**Independent Test**: Open a real thread, add a user message, repeat the same request with the same `client_message_id`, refresh/list messages, and confirm only one user message exists with no assistant placeholder or run/event semantics.

### Tests for User Story 3

- [X] T029 [P] [US3] Add message repository tests for stable creation order and partial unique `client_message_id` idempotency in `internal/productdata/repository_test.go`
- [X] T030 [P] [US3] Add message service tests for empty content rejection, duplicate message return, and unchanged thread timestamp on duplicates in `internal/productdata/service_test.go`
- [X] T031 [P] [US3] Add message HTTP handler tests for `GET /v1/threads/{thread_id}/messages`, `POST /v1/threads/{thread_id}/messages`, duplicate idempotency, empty content errors, and unknown thread errors in `internal/httpapi/product_test.go`

### Implementation for User Story 3

- [X] T032 [US3] Implement message content and optional `client_message_id` validation helpers in `internal/productdata/models.go`
- [X] T033 [US3] Implement owner-scoped message repository methods and idempotent insert-or-return behavior in `internal/productdata/repository.go`
- [X] T034 [US3] Implement transactional create-message service behavior that updates parent thread `updated_at` only for new messages in `internal/productdata/service.go`
- [X] T035 [US3] Implement message HTTP handlers and response status selection for created versus duplicate messages in `internal/httpapi/product.go`
- [X] T036 [US3] Register message routes in `internal/httpapi/server.go`

**Checkpoint**: Message API persists only final user text, reloads durable state, and never creates assistant placeholders, runs, events, tool calls, or LLM requests.

---

## Phase 6: User Story 4 - Switch the web shell between real API and mock fallback (Priority: P2)

**Goal**: Keep the existing mock shell available when no API base is configured, and use real M3 thread/message APIs when `VITE_LOOMI_API_BASE_URL` is set without silently falling back on real API failures.

**Independent Test**: Run the web shell without `VITE_LOOMI_API_BASE_URL` and confirm mock data remains usable; run it with the M3 API base and confirm real threads/messages load and persist; stop the API and confirm a recoverable error appears instead of mock fallback.

### Tests for User Story 4

- [X] T037 [P] [US4] Add TypeScript compile coverage for separated thread lifecycle and run status types in `web/src/domain.ts`
- [X] T038 [P] [US4] Add frontend data-source smoke checklist entries for mock mode, real API mode, and configured-real-API error mode in `docs-site/src/content/docs/runbooks/local-m3.md`

### Implementation for User Story 4

- [X] T039 [US4] Separate `Thread` lifecycle status from `RunStatus` and use machine-readable API timestamps in `web/src/domain.ts`
- [X] T040 [US4] Move existing in-memory mock behavior from `web/src/apiClient.ts` into `web/src/mockApiClient.ts`
- [X] T041 [US4] Implement fetch-backed M3 client methods for identity, threads, and messages in `web/src/realApiClient.ts`
- [X] T042 [US4] Implement the `VITE_LOOMI_API_BASE_URL` data-source selector and no-fallback real API error behavior in `web/src/apiClient.ts`
- [X] T043 [US4] Update API loading, selected thread state, and stale-response guards in `web/src/App.tsx`
- [X] T044 [US4] Update thread create, rename, select, and archive interactions in `web/src/components/ThreadSidebar.tsx`
- [X] T045 [US4] Update message send behavior to include a generated `client_message_id` in `web/src/components/Composer.tsx`
- [X] T046 [US4] Update real message rendering and recoverable API error display in `web/src/components/ChatCanvas.tsx`
- [X] T047 [US4] Keep run timeline and debug surfaces mock, empty, or explicitly deferred in `web/src/components/RunRail.tsx` and `web/src/components/RunTimeline.tsx`

**Checkpoint**: The web shell works in mock mode without a backend and uses durable M3 thread/message data when a real API base is configured.

---

## Phase 7: User Story 5 - Carry M2 operational contracts forward (Priority: P2)

**Goal**: Preserve structured diagnostics, M3 readiness, migration rollback/reapply, explicit seed data, structured API errors, and documented deferred boundaries.

**Independent Test**: Apply M2-only schema and see M3 readiness fail; apply M3 schema and see readiness pass; roll back and reapply without manual cleanup; run seed twice without duplicates; verify docs explain implemented and deferred capabilities.

### Tests for User Story 5

- [X] T048 [P] [US5] Add M3 readiness tests for M2-only failure, version 2 success, dirty failure, and redacted schema reasons in `internal/db/readiness_test.go`
- [X] T049 [P] [US5] Add seed command tests for idempotent local user, demo thread, demo message, structured diagnostics, and redacted failure output in `cmd/loomi-seed/main_test.go`
- [X] T050 [P] [US5] Add structured API error tests for stable codes, human messages, request ids, and no secret leakage in `internal/httpapi/product_test.go`

### Implementation for User Story 5

- [X] T051 [US5] Update schema readiness to require clean migration version `2` or later in `internal/db/readiness.go`
- [X] T052 [US5] Implement explicit idempotent local demo seed command in `cmd/loomi-seed/main.go`
- [X] T053 [US5] Ensure all product API errors use the stable envelope from `internal/httpapi/errors.go` in `internal/httpapi/product.go`
- [X] T054 [US5] Document M3 architecture boundaries, ownership, and deferred run/event/LLM/tool/worker/runtime concepts in `docs-site/src/content/docs/architecture/auth-thread-message.md`
- [X] T055 [P] [US5] Document M3 endpoint contracts and structured errors in `docs-site/src/content/docs/api/thread-message.md`
- [X] T056 [P] [US5] Document local M3 setup, migration, seed, frontend switching, rollback, and troubleshooting commands in `docs-site/src/content/docs/runbooks/local-m3.md`
- [X] T057 [P] [US5] Document completed work, validation results, limitations, and next steps in `docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md`
- [X] T058 [US5] Update the current Spec Kit feature status and tasks reference in `docs-site/src/content/docs/spec-kit/workflow.md`

**Checkpoint**: M3 operational behavior is verifiable through readiness, migration smoke, seed smoke, structured errors, and documentation.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and consistency checks across all stories.

- [X] T059 Run `go test ./...` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md`
- [X] T060 Run migration apply, rollback, and reapply smoke commands from `specs/002-m3-auth-thread-message/quickstart.md` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md`
- [X] T061 Run `bun run build` from `web/` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md`
- [X] T062 Perform browser smoke for web mock mode and real API mode and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md`
- [X] T063 Run `bun run build` from `docs-site/` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md`
- [X] T064 Verify `specs/002-m3-auth-thread-message/quickstart.md`, `docs-site/src/content/docs/runbooks/local-m3.md`, and `docs-site/src/content/docs/api/thread-message.md` agree on endpoint names, environment variables, commands, and deferred boundaries

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; can start immediately.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user stories.
- **US1 (Phase 3)**: Depends on Foundational; MVP slice for fixed local identity.
- **US2 (Phase 4)**: Depends on US1 because threads require the fixed durable owner.
- **US3 (Phase 5)**: Depends on US1 and US2 because messages require an owned thread.
- **US4 (Phase 6)**: Depends on US2 and US3 for real thread/message API behavior; mock-mode preservation can be prepared after Foundational.
- **US5 (Phase 7)**: Depends on Foundational for migrations and can progress alongside US1-US4 where tasks touch independent files; final readiness/seed/docs verification depends on backend story completion.
- **Polish (Phase 8)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: MVP; no dependencies on other stories after Foundational.
- **User Story 2 (P1)**: Requires US1 identity ownership.
- **User Story 3 (P1)**: Requires US1 identity and US2 thread container.
- **User Story 4 (P2)**: Requires backend thread/message APIs from US2 and US3 for real API mode.
- **User Story 5 (P2)**: Operationally cross-cutting; readiness and seed can be implemented after Foundational, docs/validation should finish after stories land.

### Within Each User Story

- Tests are listed before implementation tasks.
- Models and validation helpers precede repository/service implementation.
- Repository/service implementation precedes HTTP handlers.
- HTTP handlers precede frontend real API integration.
- Story checkpoint should pass before moving to the next priority story.

---

## Parallel Execution Examples

### User Story 1

```text
Task: T012 Add fixed local identity tests in internal/identity/local_test.go
Task: T013 Add user upsert repository tests in internal/productdata/repository_test.go
Task: T014 Add current identity service tests in internal/productdata/service_test.go
Task: T015 Add GET /v1/me handler tests in internal/httpapi/product_test.go
```

### User Story 2

```text
Task: T021 Add thread repository tests in internal/productdata/repository_test.go
Task: T022 Add thread service tests in internal/productdata/service_test.go
Task: T023 Add thread HTTP handler tests in internal/httpapi/product_test.go
```

### User Story 3

```text
Task: T029 Add message repository tests in internal/productdata/repository_test.go
Task: T030 Add message service tests in internal/productdata/service_test.go
Task: T031 Add message HTTP handler tests in internal/httpapi/product_test.go
```

### User Story 4

```text
Task: T039 Update domain types in web/src/domain.ts
Task: T040 Move mock behavior into web/src/mockApiClient.ts
Task: T041 Implement fetch-backed client in web/src/realApiClient.ts
```

### User Story 5

```text
Task: T048 Add M3 readiness tests in internal/db/readiness_test.go
Task: T049 Add seed command tests in cmd/loomi-seed/main_test.go
Task: T055 Document M3 endpoint contracts in docs-site/src/content/docs/api/thread-message.md
Task: T056 Document local M3 runbook in docs-site/src/content/docs/runbooks/local-m3.md
Task: T057 Document M3 devlog in docs-site/src/content/docs/devlog/2026-05-23-m3-auth-thread-message.md
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1 Setup.
2. Complete Phase 2 Foundational prerequisites.
3. Complete Phase 3 User Story 1.
4. Validate `GET /v1/me` independently.
5. Continue to User Story 2 and User Story 3 for the minimal durable product-data slice.

### Incremental Delivery

1. Setup + Foundational: schema, shared model constants, error envelope, route seams.
2. US1: fixed local identity and durable user ownership.
3. US2: durable thread CRUD/lifecycle without run coupling.
4. US3: durable user messages with idempotency.
5. US4: frontend real/mock switching for thread/message path.
6. US5: operational contracts, seed, readiness, rollback/reapply, and docs.
7. Polish: full quickstart validation and docs build.

### Parallel Team Strategy

1. Complete Setup and Foundational together.
2. Run US1 first to establish ownership.
3. After US1, split backend work by story boundaries where practical:
   - Developer A: US2 thread repository/service/HTTP.
   - Developer B: US3 message repository/service/HTTP after thread interfaces stabilize.
   - Developer C: US5 readiness/seed/docs and US4 frontend mock-client preparation.
4. Integrate US4 after US2/US3 APIs match `contracts/http-m3.openapi.yaml`.

---

## Notes

- Keep M3 comments sparse: only key non-obvious boundaries get concise Chinese WHY comments.
- Do not introduce routers, ORMs, sessions, queues, LLM calls, worker jobs, desktop runtime behavior, attachments, RAG, or catalog concepts.
- Do not make the frontend silently fall back to mock data when `VITE_LOOMI_API_BASE_URL` is configured.
- Migrations must remain schema-only; local demo data belongs only in `cmd/loomi-seed/main.go`.
- Before claiming implementation completion, run `go test ./...`, `bun run build` in `web/`, browser smoke for UI changes, and `bun run build` in `docs-site/`.
