# Tasks: M2 API and Database Base

**Input**: Design documents from `specs/001-m2-api-db-base/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Included because the specification defines independent tests, smoke verification, readiness failure behavior, migration rollback/reapply, and M1 UI preservation as acceptance criteria.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel after dependencies are met because it touches different files
- **[Story]**: User story label for story phases only
- Every task includes exact file paths

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the Go backend foundation and local development scaffolding.

- [X] T001 Create root Go module in `go.mod`
- [X] T002 Create local environment example in `.env.example`
- [X] T003 Create local PostgreSQL compose service in `compose.yaml`
- [X] T004 [P] Create API service directory note in `services/api/README.md`
- [X] T005 [P] Create backend package directories under `internal/config/`, `internal/diagnostics/`, `internal/db/`, and `internal/httpapi/`
- [X] T006 [P] Create API command directory under `cmd/loomi-api/`
- [X] T007 [P] Create migration directory under `migrations/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add shared configuration, diagnostics, database abstractions, and server wiring required by all user stories.

**CRITICAL**: No user story work should begin until this phase is complete.

- [X] T008 Write runtime configuration tests in `internal/config/config_test.go`
- [X] T009 Implement runtime configuration loader, validation, and database URL redaction in `internal/config/config.go`
- [X] T010 Run configuration tests with `go test ./internal/config`
- [X] T011 [P] Write diagnostics helper tests in `internal/diagnostics/logger_test.go`
- [X] T012 [P] Implement structured logger, request ID, operation ID, and redaction helpers in `internal/diagnostics/logger.go`
- [X] T013 Run diagnostics tests with `go test ./internal/diagnostics`
- [X] T014 Add pgx dependency to `go.mod` and `go.sum`
- [X] T015 [P] Implement PostgreSQL pool creation in `internal/db/pool.go`
- [X] T016 [P] Define database readiness checker interface and check result types in `internal/db/readiness.go`
- [X] T017 Create HTTP server wiring in `internal/httpapi/server.go`
- [X] T018 Run package compilation check with `go test ./...`

**Checkpoint**: Shared backend foundation compiles and config/diagnostics tests pass.

---

## Phase 3: User Story 1 - Verify the service is alive and ready (Priority: P1) MVP

**Goal**: Developers can start the service, confirm process liveness, and see readiness accurately reflect PostgreSQL/schema availability.

**Independent Test**: Start the service, call `/healthz`, call `/readyz` with persistence available, then call `/readyz` with persistence unavailable and confirm `503 not_ready` while `/healthz` stays alive.

### Tests for User Story 1

- [X] T019 [P] [US1] Write HTTP health/readiness handler tests in `internal/httpapi/health_test.go`
- [X] T020 [P] [US1] Write database readiness unit tests with fake checker in `internal/db/readiness_test.go`

### Implementation for User Story 1

- [X] T021 [US1] Implement `/healthz` and `/readyz` handlers in `internal/httpapi/health.go`
- [X] T022 [US1] Implement database ping and schema readiness behavior in `internal/db/readiness.go`
- [X] T023 [US1] Implement API process startup without requiring PostgreSQL availability in `cmd/loomi-api/main.go`
- [X] T024 [US1] Run health/readiness tests with `go test ./internal/httpapi ./internal/db`
- [X] T025 [US1] Manually verify `/healthz` returns HTTP 200 and `/readyz` returns HTTP 503 when PostgreSQL is stopped using `cmd/loomi-api/main.go` and `compose.yaml`

**Checkpoint**: User Story 1 is independently functional and demonstrates alive/ready separation.

---

## Phase 4: User Story 2 - Configure and start the M2 service predictably (Priority: P1)

**Goal**: The service loads required local settings, rejects invalid startup configuration, and emits safe structured diagnostics.

**Independent Test**: Start with valid settings, then start with missing/malformed settings and confirm fast, redacted, actionable failure output.

### Tests for User Story 2

- [X] T026 [P] [US2] Extend configuration validation tests for invalid `APP_ENV`, `HTTP_ADDR`, `LOG_LEVEL`, and `READINESS_TIMEOUT_SECONDS` in `internal/config/config_test.go`
- [X] T027 [P] [US2] Add startup diagnostic behavior test coverage in `internal/diagnostics/logger_test.go`

### Implementation for User Story 2

- [X] T028 [US2] Harden runtime configuration validation for all M2 fields in `internal/config/config.go`
- [X] T029 [US2] Wire startup diagnostics with operation IDs and redacted database URL output in `cmd/loomi-api/main.go`
- [X] T030 [US2] Run config and diagnostics tests with `go test ./internal/config ./internal/diagnostics`
- [X] T031 [US2] Manually verify missing or malformed configuration fails startup without printing full `DATABASE_URL` using `.env.example` and `cmd/loomi-api/main.go`

**Checkpoint**: User Story 2 is independently functional and startup behavior is predictable.

---

## Phase 5: User Story 3 - Establish reversible persistent schema management (Priority: P2)

**Goal**: The local persistent store has a repeatable schema baseline workflow that can apply, roll back, and re-apply without creating business tables.

**Independent Test**: Apply the migration to an empty local PostgreSQL store, verify version `1`, roll back one migration, re-apply it, and verify readiness passes only when schema version is present and clean.

### Tests for User Story 3

- [X] T032 [P] [US3] Add schema readiness test cases for missing, dirty, and valid schema state in `internal/db/readiness_test.go`

### Implementation for User Story 3

- [X] T033 [US3] Create schema baseline up migration in `migrations/000001_schema_baseline.up.sql`
- [X] T034 [US3] Create schema baseline down migration in `migrations/000001_schema_baseline.down.sql`
- [X] T035 [US3] Ensure schema readiness checks version and dirty state without requiring business tables in `internal/db/readiness.go`
- [X] T036 [US3] Run DB tests with `go test ./internal/db`
- [X] T037 [US3] Manually verify migration apply, version, rollback, and re-apply using `migrations/000001_schema_baseline.up.sql`, `migrations/000001_schema_baseline.down.sql`, and `compose.yaml`
- [X] T038 [US3] Manually verify `/readyz` returns ready only after baseline schema is applied using `cmd/loomi-api/main.go`

**Checkpoint**: User Story 3 is independently functional and schema workflow is reversible.

---

## Phase 6: User Story 4 - Preserve the existing M1 shell while adding the service base (Priority: P3)

**Goal**: The existing M1 mock UI shell remains runnable and no UI path is forced onto incomplete M2 backend data.

**Independent Test**: Build the existing `web/` project and confirm M1 mock thread, chat, composer, timeline, and right-side panels remain independent of M2 service availability.

### Implementation for User Story 4

- [X] T039 [US4] Confirm no M2 backend dependency is introduced into `web/src/apiClient.ts`
- [X] T040 [US4] Confirm mock data remains available in `web/src/mockData.ts`
- [X] T041 [US4] Run existing web tests or build validation from `web/package.json`
- [X] T042 [US4] Build M1 shell with `cd web && bun run build`

**Checkpoint**: User Story 4 is independently validated and the M1 shell remains usable.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Update documentation, validate contracts, and run full smoke checks.

- [X] T043 [P] Update API documentation for `/healthz` and `/readyz` in `docs-site/src/content/docs/api/index.md`
- [X] T044 [P] Update M2 architecture documentation in `docs-site/src/content/docs/architecture/api-db-base.md`
- [X] T045 [P] Update local validation runbook in `docs-site/src/content/docs/runbooks/index.md`
- [X] T046 [P] Update Spec Kit artifact references in `docs-site/src/content/docs/spec-kit/workflow.md`
- [X] T047 Create M2 development log in `docs-site/src/content/docs/devlog/2026-05-23-m2-api-db-base.md`
- [X] T048 Validate HTTP contract against implementation manually using `specs/001-m2-api-db-base/contracts/http-health.openapi.yaml`
- [X] T049 Validate migration command contract manually using `specs/001-m2-api-db-base/contracts/migration-cli.md`
- [X] T050 Run full Go validation with `go test ./...`
- [X] T051 Run docs validation with `cd docs-site && bun run build`
- [X] T052 Run web validation with `cd web && bun run build`
- [X] T053 Run quickstart smoke sequence from `specs/001-m2-api-db-base/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user stories.
- **US1 (Phase 3)**: Depends on Foundational; MVP scope.
- **US2 (Phase 4)**: Depends on Foundational; can run after or alongside US1 after shared server/config packages exist.
- **US3 (Phase 5)**: Depends on Foundational and benefits from US1 readiness handler behavior.
- **US4 (Phase 6)**: Depends on M2 code changes being present; can validate independently before final polish.
- **Polish (Phase 7)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other user stories after Foundational.
- **User Story 2 (P1)**: No dependency on other user stories after Foundational, but shares `cmd/loomi-api/main.go` with US1.
- **User Story 3 (P2)**: Depends on DB readiness abstractions from Foundational and readiness behavior from US1.
- **User Story 4 (P3)**: Depends on implementation choices not modifying `web/` runtime behavior.

### Parallel Opportunities

- Setup directory/document tasks T004-T007 can run in parallel.
- Diagnostics package tasks T011-T013 can run in parallel with DB package setup T015-T016 after `go.mod` exists.
- US1 tests T019-T020 can run in parallel.
- US2 tests T026-T027 can run in parallel.
- Documentation updates T043-T046 can run in parallel after implementation stabilizes.

---

## Parallel Example: User Story 1

```bash
Task: "T019 [US1] Write HTTP health/readiness handler tests in internal/httpapi/health_test.go"
Task: "T020 [US1] Write database readiness unit tests with fake checker in internal/db/readiness_test.go"
```

## Parallel Example: User Story 2

```bash
Task: "T026 [US2] Extend configuration validation tests in internal/config/config_test.go"
Task: "T027 [US2] Add startup diagnostic behavior tests in internal/diagnostics/logger_test.go"
```

## Parallel Example: Documentation Polish

```bash
Task: "T043 Update docs-site/src/content/docs/api/index.md"
Task: "T044 Update docs-site/src/content/docs/architecture/api-db-base.md"
Task: "T045 Update docs-site/src/content/docs/runbooks/index.md"
Task: "T046 Update docs-site/src/content/docs/spec-kit/workflow.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 Setup.
2. Complete Phase 2 Foundational.
3. Complete Phase 3 US1.
4. Stop and verify `/healthz` and `/readyz` behavior independently.

### Incremental Delivery

1. Add service foundation and compile/test it.
2. Add US1 health/readiness behavior for MVP.
3. Add US2 configuration and startup diagnostics.
4. Add US3 migration workflow and schema readiness.
5. Validate US4 web shell preservation.
6. Finish docs and full smoke validation.

### Validation Gates

- After Phase 2: `go test ./...`
- After US1: `go test ./internal/httpapi ./internal/db` plus manual health/readiness checks
- After US2: `go test ./internal/config ./internal/diagnostics` plus invalid config startup check
- After US3: migration up/down/reapply and `/readyz` ready check
- After US4: `cd web && bun run build`
- Final: `go test ./...`, `cd docs-site && bun run build`, `cd web && bun run build`, and quickstart smoke sequence

## Notes

- M2 must not create users, threads, messages, runs, run events, workers, tools, auth, LLM gateway, or desktop runtime.
- The API process must not auto-apply migrations.
- The service must start even when PostgreSQL is unavailable; `/readyz` carries dependency failure.
- Diagnostics must include request or operation IDs and must not print full secrets.
