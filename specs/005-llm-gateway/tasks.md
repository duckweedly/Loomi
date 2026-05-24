# Tasks: M5 LLM Gateway

**Input**: Design documents from `/specs/005-llm-gateway/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Included because the M5 spec and quickstart define measurable validation for provider capability, run/event/SSE behavior, frontend runtime states, and documentation validation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing. This task list is only for `005-llm-gateway`; `006-streaming-chat-runtime` is intentionally excluded.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare M5 schema, configuration, and docs targets without implementing story behavior yet.

- [x] T001 Create M5 migration files `migrations/000004_m5_llm_gateway.up.sql` and `migrations/000004_m5_llm_gateway.down.sql`
- [x] T002 [P] Add local model provider configuration fields in `internal/config/config.go`
- [x] T003 [P] Add model provider configuration tests in `internal/config/config_test.go`
- [x] T004 [P] Add M5 provider and gateway type skeletons in `internal/runtime/providers.go`
- [x] T005 [P] Add M5 gateway runner skeleton in `internal/runtime/gateway.go`
- [x] T006 [P] Add M5 docs stubs in `docs-site/src/content/docs/architecture/llm-gateway.md`, `docs-site/src/content/docs/api/llm-gateway.md`, `docs-site/src/content/docs/runbooks/local-m5.md`, and `docs-site/src/content/docs/devlog/2026-05-23-m5-llm-gateway.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core persistence, provider abstraction, and HTTP wiring that MUST be complete before any user story can be implemented.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T007 Update M5 migration to allow `messages.role = 'assistant'` and `runs.source = 'model_gateway'` in `migrations/000004_m5_llm_gateway.up.sql`
- [x] T008 Add rollback for assistant role and model gateway source constraints in `migrations/000004_m5_llm_gateway.down.sql`
- [x] T009 Update message and run domain models for assistant messages and `model_gateway` runs in `internal/productdata/models.go`
- [x] T010 Update repository persistence for assistant messages and model gateway runs in `internal/productdata/repository.go`
- [x] T011 [P] Add repository tests for assistant messages, `model_gateway` source, and rollback expectations in `internal/productdata/repository_test.go`
- [x] T012 Update product service rules for appending one assistant message per completed model run in `internal/productdata/service.go`
- [x] T013 [P] Add product service tests for assistant message persistence and duplicate prevention in `internal/productdata/service_test.go`
- [x] T014 Define provider capability, provider request, provider event, and redacted error types in `internal/runtime/providers.go`
- [x] T015 Implement provider selection and local configuration loading for Anthropic, OpenAI, Gemini, and OpenAI-compatible custom providers in `internal/runtime/providers.go`
- [x] T016 [P] Add provider selection and redaction tests in `internal/runtime/providers_test.go`
- [x] T017 Add model gateway runner orchestration for run creation, context loading, stream consumption, stop checks, and terminal state handling in `internal/runtime/gateway.go`
- [x] T018 [P] Add gateway orchestration tests using fake providers in `internal/runtime/gateway_test.go`
- [x] T019 Extend run event helpers with M5 normalized event types in `internal/runtime/runner.go`
- [x] T020 [P] Add normalized event ordering tests in `internal/runtime/runner_test.go`
- [x] T021 Wire provider capability endpoints and model gateway run source routing in `internal/httpapi/runtime.go`
- [x] T022 [P] Add HTTP tests for `/v1/model-providers`, `/v1/model-providers/check`, and `source=model_gateway` validation in `internal/httpapi/runtime_test.go`
- [x] T023 Ensure API server construction injects provider configuration and model gateway runner in `internal/httpapi/server.go`
- [x] T024 Update frontend domain/API types for provider capability, `model_gateway`, and M5 event types in `web/src/domain.ts` and `web/src/apiClient.ts`
- [x] T025 [P] Add frontend API type and parser tests for M5 provider/run/event payloads in `web/src/realApiClient.test.ts`

**Checkpoint**: Foundation ready - M5 stories can now be implemented without mixing in 006 frontend-specific work.

---

## Phase 3: User Story 1 - Receive a model-backed assistant response (Priority: P1) MVP

**Goal**: A user submits a message in an existing thread and receives streaming assistant output and a final assistant message from a real model gateway path.

**Independent Test**: Submit a durable user message, start a `model_gateway` run with a configured provider, stream `model_output_delta` events, and verify the final assistant response appears in thread history exactly once.

### Tests for User Story 1

- [x] T026 [P] [US1] Add provider stream normalization tests for successful text deltas and final completion in `internal/runtime/gateway_test.go`
- [x] T027 [P] [US1] Add HTTP integration test for creating a `model_gateway` run from an existing message in `internal/httpapi/runtime_test.go`
- [x] T028 [P] [US1] Add frontend real adapter test for model-gateway run creation and `model_output_delta` consumption in `web/src/runtime/realExecutionAdapter.test.ts`
- [x] T029 [P] [US1] Add Chat Canvas state test for streaming model output and final assistant message behavior in `web/src/components/ChatCanvas.states.test.ts`

### Implementation for User Story 1

- [x] T030 [US1] Implement current-thread request context loading from existing messages in `internal/runtime/gateway.go`
- [x] T031 [US1] Implement Anthropic streaming request builder and text delta parser in `internal/runtime/providers.go`
- [x] T032 [US1] Implement OpenAI streaming request builder and text delta parser in `internal/runtime/providers.go`
- [x] T033 [US1] Implement Gemini streaming request builder and text chunk parser in `internal/runtime/providers.go`
- [x] T034 [US1] Implement OpenAI-compatible custom provider request builder using configurable base URL, API key, and model in `internal/runtime/providers.go`
- [x] T035 [US1] Convert provider text deltas to `model_output_delta` run events in `internal/runtime/gateway.go`
- [x] T036 [US1] Persist `model_output_completed`, final assistant message, and `run_completed` atomically enough for local validation in `internal/runtime/gateway.go` and `internal/productdata/service.go`
- [x] T037 [US1] Update `POST /v1/threads/{thread_id}/runs` to accept `message_id`, `source=model_gateway`, and `provider_id` in `internal/httpapi/runtime.go`
- [x] T038 [US1] Update real API client to start model-gateway runs from durable user message ids in `web/src/realApiClient.ts`
- [x] T039 [US1] Update real execution adapter to apply `model_output_delta` and `model_output_completed` events in `web/src/runtime/realExecutionAdapter.ts`
- [x] T040 [US1] Update Chat Canvas rendering for model-gateway streaming assistant output in `web/src/components/ChatCanvas.tsx`
- [ ] T041 [US1] Verify US1 API/SSE flow from quickstart sections 5-8 using `specs/005-llm-gateway/quickstart.md`

**Checkpoint**: User Story 1 is independently functional and demoable as the M5 MVP.

---

## Phase 4: User Story 2 - Understand model execution progress and failures (Priority: P2)

**Goal**: Users and developers can see redacted provider errors, timeouts, rate limits, refusals, cancellation, and unavailable states without losing conversation usability.

**Independent Test**: Run successful, stopped, unavailable, misconfigured, timeout, rate-limit, refusal, empty-response, and provider-error scenarios and verify each emits a distinct user-safe execution state.

### Tests for User Story 2

- [x] T042 [P] [US2] Add fake provider tests for provider error, timeout, rate limit, refusal, and empty response mapping in `internal/runtime/gateway_test.go`
- [x] T043 [P] [US2] Add HTTP tests for provider unavailable and misconfigured responses in `internal/httpapi/runtime_test.go`
- [x] T044 [P] [US2] Add frontend tests for provider failure states and no mock fallback in `web/src/runtime/realExecutionAdapter.test.ts`
- [x] T045 [P] [US2] Add timeline rendering tests for M5 provider failures in `web/src/components/RunTimeline.runtime.test.ts`

### Implementation for User Story 2

- [x] T046 [US2] Implement provider capability listing and check behavior with redacted messages in `internal/runtime/providers.go`
- [x] T047 [US2] Implement `/v1/model-providers` and `/v1/model-providers/check` handlers in `internal/httpapi/runtime.go`
- [x] T048 [US2] Map provider errors, timeouts, rate limits, refusals, and empty responses to stable run error codes in `internal/runtime/gateway.go`
- [x] T049 [US2] Ensure provider API keys, Authorization headers, raw request payloads, and raw provider error bodies are excluded from run events in `internal/runtime/gateway.go`
- [x] T050 [US2] Implement cooperative stop handling for active model-gateway runs in `internal/runtime/gateway.go`
- [x] T051 [US2] Update frontend API client with provider capability calls in `web/src/realApiClient.ts`
- [x] T052 [US2] Update runtime adapter to surface provider unavailable, misconfigured, timeout, rate-limit, refusal, and stopped states in `web/src/runtime/realExecutionAdapter.ts`
- [x] T053 [US2] Update Run Timeline provider-progress and redacted-failure rows in `web/src/components/RunTimeline.tsx`
- [x] T054 [US2] Update Chat Canvas failure, stopped, refusal, and provider-unavailable states in `web/src/components/ChatCanvas.tsx`
- [ ] T055 [US2] Verify failure and stop smoke cases from quickstart sections 9-10 using `specs/005-llm-gateway/quickstart.md`

**Checkpoint**: User Stories 1 and 2 both work independently, and failure states are visible without leaking secrets.

---

## Phase 5: User Story 3 - Preserve safety boundaries for future tool use (Priority: P3)

**Goal**: Model output that requests tool/function use is visible as a non-executed boundary event, with no external action performed.

**Independent Test**: Use fake provider output containing Anthropic tool-use content, OpenAI function-call deltas, Gemini function-call style content, and custom OpenAI-compatible tool calls; verify `tool_call_blocked` events and no tool execution.

### Tests for User Story 3

- [x] T056 [P] [US3] Add provider parser tests for Anthropic tool-use content blocks in `internal/runtime/providers_test.go`
- [x] T057 [P] [US3] Add provider parser tests for OpenAI and OpenAI-compatible function-call deltas in `internal/runtime/providers_test.go`
- [x] T058 [P] [US3] Add provider parser tests for Gemini tool/function-call style content in `internal/runtime/providers_test.go`
- [x] T059 [P] [US3] Add gateway tests asserting tool-like output emits `tool_call_blocked` and performs no external action in `internal/runtime/gateway_test.go`
- [x] T060 [P] [US3] Add frontend tests for tool-boundary event display in `web/src/components/RunTimeline.runtime.test.ts`

### Implementation for User Story 3

- [x] T061 [US3] Normalize Anthropic tool-use content into non-executed provider events in `internal/runtime/providers.go`
- [x] T062 [US3] Normalize OpenAI and OpenAI-compatible function-call deltas into non-executed provider events in `internal/runtime/providers.go`
- [x] T063 [US3] Normalize Gemini tool/function-call style content into non-executed provider events in `internal/runtime/providers.go`
- [x] T064 [US3] Convert provider tool-like output to `tool_call_blocked` run events with safe summaries in `internal/runtime/gateway.go`
- [x] T065 [US3] Ensure tool/function-call arguments are not executed and sensitive arguments are not persisted in `internal/runtime/gateway.go`
- [x] T066 [US3] Update frontend runtime mapping for `tool_call_blocked` events in `web/src/runtime/realExecutionAdapter.ts`
- [x] T067 [US3] Update Run Timeline and Chat Canvas to show tool execution is outside M5 scope in `web/src/components/RunTimeline.tsx` and `web/src/components/ChatCanvas.tsx`
- [ ] T068 [US3] Verify tool-boundary smoke case from quickstart section 10 using `specs/005-llm-gateway/quickstart.md`

**Checkpoint**: All M5 user stories are independently functional and tool-like output remains non-executing.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and final consistency checks across M5.

- [x] T069 [P] Update API documentation for provider capability, model-gateway run creation, M5 events, and redaction rules in `docs-site/src/content/docs/api/llm-gateway.md`
- [x] T070 [P] Update architecture documentation for gateway boundaries, request context, provider normalization, and tool-boundary behavior in `docs-site/src/content/docs/architecture/llm-gateway.md`
- [x] T071 [P] Update local runbook with provider configuration, custom provider example, smoke commands, and failure cases in `docs-site/src/content/docs/runbooks/local-m5.md`
- [x] T072 [P] Add M5 development log with completed scope, validation results, and known limitations in `docs-site/src/content/docs/devlog/2026-05-23-m5-llm-gateway.md`
- [x] T073 Update Spec Kit workflow/status references for `005-llm-gateway` in `docs-site/src/content/docs/spec-kit/workflow.md` and `docs-site/src/content/docs/roadmap/current-status.md`
- [x] T074 Run Go validation with `go test ./...`
- [x] T075 Run frontend tests with `bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts`
- [x] T076 Run frontend build with `bun run --cwd web build`
- [x] T077 Run docs build with `bun run --cwd docs-site build`
- [ ] T078 Perform browser smoke for real API model gateway and mock mode in `web/` (mock mode passed; real API model gateway still needs local DB/provider credentials)
- [x] T079 Record final validation notes and any local blockers in `docs-site/src/content/docs/devlog/2026-05-23-m5-llm-gateway.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion - blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion - MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational completion and can use US1 gateway flow, but failure classification can be tested with fake providers independently.
- **User Story 3 (Phase 5)**: Depends on Foundational completion and provider parser seams; can be implemented after or alongside US2.
- **Polish (Phase 6)**: Depends on desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: No dependency on US2 or US3 after Foundation; delivers model-backed assistant response MVP.
- **US2 (P2)**: Uses the same gateway seams as US1; can be validated independently with fake provider failure cases.
- **US3 (P3)**: Uses provider parser seams and gateway event conversion; can be validated independently with fake tool-like provider events.

### Within Each User Story

- Tests first for contract and behavior coverage.
- Provider parsers before gateway conversion.
- Gateway conversion before HTTP/frontend integration.
- Frontend adapter before Chat Canvas/Timeline display.
- Story smoke validation at the checkpoint.

### Parallel Opportunities

- Setup skeleton/config/docs tasks T002-T006 can run in parallel.
- Repository, config, runtime type, and HTTP tests T011, T013, T016, T018, T020, T022, and T025 can run in parallel after their target files exist.
- US1 tests T026-T029 can run in parallel before implementation.
- US2 tests T042-T045 can run in parallel before implementation.
- US3 parser and UI tests T056-T060 can run in parallel before implementation.
- Documentation tasks T069-T072 can run in parallel after behavior is stable.

---

## Parallel Example: User Story 1

```bash
Task: "Add provider stream normalization tests for successful text deltas and final completion in internal/runtime/gateway_test.go"
Task: "Add HTTP integration test for creating a model_gateway run from an existing message in internal/httpapi/runtime_test.go"
Task: "Add frontend real adapter test for model-gateway run creation and model_output_delta consumption in web/src/runtime/realExecutionAdapter.test.ts"
Task: "Add Chat Canvas state test for streaming model output and final assistant message behavior in web/src/components/ChatCanvas.states.test.ts"
```

## Parallel Example: User Story 2

```bash
Task: "Add fake provider tests for provider error, timeout, rate limit, refusal, and empty response mapping in internal/runtime/gateway_test.go"
Task: "Add HTTP tests for provider unavailable and misconfigured responses in internal/httpapi/runtime_test.go"
Task: "Add frontend tests for provider failure states and no mock fallback in web/src/runtime/realExecutionAdapter.test.ts"
Task: "Add timeline rendering tests for M5 provider failures in web/src/components/RunTimeline.runtime.test.ts"
```

## Parallel Example: User Story 3

```bash
Task: "Add provider parser tests for Anthropic tool-use content blocks in internal/runtime/providers_test.go"
Task: "Add provider parser tests for OpenAI and OpenAI-compatible function-call deltas in internal/runtime/providers_test.go"
Task: "Add provider parser tests for Gemini tool/function-call style content in internal/runtime/providers_test.go"
Task: "Add frontend tests for tool-boundary event display in web/src/components/RunTimeline.runtime.test.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 setup.
2. Complete Phase 2 foundation.
3. Complete Phase 3 US1.
4. Stop and validate the model-backed response flow with quickstart sections 5-8.
5. Demo the end-to-end path before adding richer failure/tool-boundary behavior.

### Incremental Delivery

1. Setup + Foundation -> migrations, provider config, gateway seams, and HTTP/frontend types ready.
2. US1 -> successful streaming model-backed assistant response.
3. US2 -> observable redacted failures, stop, provider status, and no mock fallback.
4. US3 -> safe non-executed tool boundary events.
5. Polish -> docs, builds, tests, browser smoke, validation log.

### Parallel Team Strategy

With multiple developers:

1. One developer owns migration/productdata/httpapi foundation.
2. One developer owns runtime provider/gateway fake-provider tests and parsers.
3. One developer owns frontend adapter/Chat Canvas/Timeline mapping.
4. One developer owns docs/runbook/devlog updates after contracts stabilize.

## Notes

- This task list intentionally excludes `specs/006-streaming-chat-runtime`.
- Provider secrets must remain local configuration and must not be written into docs, events, tests, or frontend code.
- [P] tasks must touch different files or be coordinated to avoid same-file conflicts.
- Each user story has an independent smoke checkpoint.
- Do not add provider SDKs or settings UI unless the plan is updated first.
