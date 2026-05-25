# Tasks: Real Testing Console & Background UX

**Input**: `specs/010-real-testing-console-background-ux/spec.md`  
**Plan**: `specs/010-real-testing-console-background-ux/plan.md`  
**Feature**: M6.5 real testing console and background UX  
**Date**: 2026-05-24

## User Stories

- **US1 Provider Test Console**: Local tester can open Settings > Providers and clearly see configured backend providers, test readiness, see sanitized failures, and understand local draft does not affect real calls.
- **US2 Chat Provider Readiness**: User in Chat/Composer sees provider unavailable guidance for real modes, while mock mode remains usable.
- **US3 Background Tasks Observer**: User can inspect current run/job state, worker diagnostics, latest worker/job events, and empty state in a read-only right panel.
- **US4 Timeline Event Productization**: User can understand M6 runtime/worker/job events in RunRail/Timeline with readable grouped labels.
- **US5 Composer State Polish**: User sees correct Composer/ChatCanvas text and button availability for queued/running/retrying/recovering/stopped/failed/closed/cancelled states.
- **US6 Local Real Testing Documentation**: Developer can follow docs to configure env, start API/web, test providers, send real messages, and observe jobs/timeline/failure states.

## Phase 1: Setup

- [X] T001 Read `.specify/memory/constitution.md` and record implementation constraints before editing any code
- [X] T002 Read `specs/008-worker-job-pipeline/plan.md` to verify M6 worker/job event names and state semantics
- [X] T003 Locate current Settings > Providers files and tests in the web app
- [X] T004 Locate current provider readiness API/client hook and provider check behavior
- [X] T005 Locate current local provider draft state and confirm whether it is session-only
- [X] T006 Locate current Chat, Composer, ChatCanvas, runtime capability, and provider readiness integration files
- [X] T007 Locate current Background tasks placeholder/right panel files
- [X] T008 Locate current RunRail/Timeline event rendering files
- [X] T009 Locate current English and Chinese i18n message files
- [X] T010 Locate current frontend test commands, backend/API test commands, UI test commands, and docs build command from package scripts
- [X] T011 Confirm docs-site content structure and whether Starlight nav/sidebar requires explicit registration

## Phase 2: Foundational

- [X] T012 Confirm configured provider row fields in the existing provider settings module: id, family, model, base URL, status, message, and last checked timestamp if available
- [X] T013 Confirm provider-dependent runtime capabilities in the existing Composer/runtime integration module: `real_api` and `model_gateway` require provider readiness; `mock` does not
- [X] T014 Confirm Background tasks snapshot source priority in the existing panel module: selected Chat run job, then empty state; cross-run active job discovery is deferred
- [X] T015 Confirm Timeline event label inputs in the existing RunRail/Timeline module: raw event type, group, message/detail, timestamp, and severity if available
- [X] T016 Confirm no new final provider secret storage, API key persistence, M7 tool call, approval, or tool execution protocol is introduced in the implementation plan
- [X] T017 Confirm all new user-visible copy will use the existing i18n mechanism, not a new i18n dependency

## Phase 3: US1 Provider Test Console

**Goal**: Settings > Providers becomes a Provider Test Console that distinguishes backend configured providers from local draft and supports connection testing.

**Independent Test Criteria**:
- With backend providers, console displays id/family/model/base URL/status/message.
- With no backend providers, console displays env setup empty state.
- Test connection shows checking/success/failed.
- Failure messages are sanitized.
- Local draft copy states browser-session-only and does not affect real calls.
- Local draft edits do not write backend config.

### Tests

- [X] T018 [P] [US1] Add configured provider render test in existing Settings Providers test file
- [X] T019 [P] [US1] Add no-provider empty state test in existing Settings Providers test file
- [X] T020 [P] [US1] Add local draft explanatory copy test in existing Settings Providers test file
- [X] T021 [P] [US1] Add provider Test connection checking-state test in existing Settings Providers test file
- [X] T022 [P] [US1] Add provider Test connection success-state test in existing Settings Providers test file
- [X] T023 [P] [US1] Add provider Test connection failure-state sanitization test in existing Settings Providers test file
- [X] T024 [US1] Add local draft does-not-write-backend-config test in existing Settings Providers test file

### Implementation

- [X] T025 [US1] Rename Settings > Providers surface to Provider Test Console in existing Settings Providers component
- [X] T026 [US1] Render configured backend providers with id, family, model, base URL, status, and message in existing Settings Providers component
- [X] T027 [US1] Render no-provider env configuration empty state in existing Settings Providers component
- [X] T028 [US1] Render Local draft section with browser-session-only, unsaved, and no-real-call-impact copy in existing Settings Providers component
- [X] T029 [US1] Wire Test connection as a per-provider UI action using the existing provider readiness/check behavior; if backend readiness is aggregate-only, map the aggregate result back to provider rows without adding provider secret storage APIs
- [X] T030 [US1] Render provider check `checking`, `success`, and `failed` states in existing Settings Providers component
- [X] T031 [US1] Sanitize provider failure messages at the provider console display boundary so bearer tokens, API key query params, and `sk-`-style token values are never rendered
- [X] T032 [US1] Prevent local draft changes from calling backend provider config write behavior in existing provider draft handlers
- [X] T033 [US1] Run focused Settings Providers tests and fix only US1 regressions

## Phase 4: US2 Chat Provider Readiness

**Goal**: Chat/Composer makes provider unavailable obvious in real modes and does not block mock mode.

**Independent Test Criteria**:
- `real_api` unavailable shows provider unavailable message and Settings CTA.
- `model_gateway` unavailable shows provider unavailable message and Settings CTA.
- Provider unavailable state does not show generating.
- mock mode remains usable without provider readiness.

### Tests

- [X] T034 [P] [US2] Add `real_api` provider unavailable Composer/Chat test in existing Chat or Composer test file
- [X] T035 [P] [US2] Add `model_gateway` provider unavailable Composer/Chat test in existing Chat or Composer test file
- [X] T036 [P] [US2] Add mock mode not blocked by missing provider test in existing Chat or Composer test file
- [X] T037 [P] [US2] Add Open Settings > Providers CTA test in existing Chat or Composer integration test file

### Implementation

- [X] T038 [US2] Apply provider readiness gate only to `real_api` and `model_gateway` in existing Composer/runtime integration file
- [X] T039 [US2] Render provider unavailable message in existing Composer or ChatCanvas component
- [X] T040 [US2] Add Open Settings > Providers CTA using existing Settings navigation mechanism
- [X] T041 [US2] Disable Send while provider is unavailable in real provider-dependent modes
- [X] T042 [US2] Ensure provider unavailable state does not render generating/spinner state in Composer or ChatCanvas
- [X] T043 [US2] Ensure mock mode bypasses provider readiness gate in existing Composer/runtime integration file
- [X] T044 [US2] Run focused Chat/Composer provider readiness tests and fix only US2 regressions

## Phase 5: US3 Background Tasks Observer

**Goal**: Right-side Background tasks panel becomes a read-only observer for current run/job, diagnostics, latest events, and empty state.

**Independent Test Criteria**:
- Empty state appears when no selected Chat run job evidence exists.
- Selected Chat run/job state appears when job evidence is available.
- queued/leased/retrying/completed/failed/cancelled/dead/recovering labels render.
- Worker diagnostics summary appears when available.
- Latest worker/job events appear when available.
- No mutating controls are introduced.

### Tests

- [X] T045 [P] [US3] Add Background tasks empty-state test in existing Background tasks panel test file
- [X] T046 [P] [US3] Add current run/job state test in existing Background tasks panel test file
- [X] T047 [P] [US3] Add worker diagnostics summary test in existing Background tasks panel test file
- [X] T048 [P] [US3] Add latest worker/job events test in existing Background tasks panel test file
- [X] T049 [P] [US3] Add read-only no mutating controls test in existing Background tasks panel test file
- [X] T050 [P] [US3] Add Background tasks snapshot priority test for selected Chat run job and empty state in existing Background tasks panel test file
- [X] T051 [P] [US3] Add recovering observation state test for backend `recovering` status or derived `job_recovering` event in existing Background tasks panel test file

### Implementation

- [X] T052 [US3] Replace Background tasks placeholder with read-only panel structure in existing right-side panel component
- [X] T053 [US3] Render no-background-task empty state in existing Background tasks panel component
- [X] T054 [US3] Render Background tasks snapshot using priority order: selected Chat run job, then empty state
- [X] T055 [US3] Add job status labels for queued, leased, retrying, completed, failed, cancelled, and dead in existing Background tasks panel component or shared status label mapper
- [X] T056 [US3] Add recovering observation label from backend `recovering` status or derived `job_recovering` event without changing the M6 job data model
- [X] T057 [US3] Render worker diagnostics summary fields when available in existing Background tasks panel component
- [X] T058 [US3] Render latest worker/job events using existing event data in existing Background tasks panel component
- [X] T059 [US3] Remove or avoid any retry/recover/cancel worker job controls in Background tasks panel
- [X] T060 [US3] Run focused Background tasks tests and fix only US3 regressions

## Phase 6: US4 Timeline Event Productization

**Goal**: RunRail/Timeline shows M6 worker/job/runtime events with readable grouped labels.

**Independent Test Criteria**:
- Event groups runtime/worker/job/diagnostics render.
- All requested M6 events render with readable labels.
- Unknown worker/job events have a safe fallback.
- Raw event identity remains visible or accessible for debugging.
- Incoming event stream order is preserved.

### Tests

- [X] T061 [P] [US4] Add timeline event group label tests in existing RunRail/Timeline test file
- [X] T062 [P] [US4] Add timeline label tests for `job_claimed` and `lease_renewed` in existing RunRail/Timeline test file
- [X] T063 [P] [US4] Add timeline label tests for `job_recovering` and `job_retry_scheduled` in existing RunRail/Timeline test file
- [X] T064 [P] [US4] Add timeline label tests for `job_attempt_failed` and `job_retry_exhausted` in existing RunRail/Timeline test file
- [X] T065 [P] [US4] Add timeline label tests for cancellation and worker diagnostics events in existing RunRail/Timeline test file
- [X] T066 [P] [US4] Add unknown worker/job event fallback test in existing RunRail/Timeline test file
- [X] T067 [P] [US4] Add out-of-order worker/job event display test that preserves stream order and readable fallback labels in existing RunRail/Timeline test file

### Implementation

- [X] T068 [US4] Add runtime, worker, job, and diagnostics event group labels in existing RunRail/Timeline event label mapper
- [X] T069 [US4] Add readable labels for `job_claimed` and `lease_renewed` in existing RunRail/Timeline event label mapper
- [X] T070 [US4] Add readable labels for `job_recovering` and `job_retry_scheduled` in existing RunRail/Timeline event label mapper
- [X] T071 [US4] Add readable labels for `job_attempt_failed` and `job_retry_exhausted` in existing RunRail/Timeline event label mapper
- [X] T072 [US4] Add readable labels for cancellation and worker diagnostics events in existing RunRail/Timeline event label mapper
- [X] T073 [US4] Preserve raw event type in existing RunRail/Timeline event details or accessible metadata
- [X] T074 [US4] Add unknown worker/job event fallback label in existing RunRail/Timeline event label mapper
- [X] T075 [US4] Preserve incoming worker/job event order in RunRail/Timeline display and avoid rewriting event history during M6.5 productization
- [X] T076 [US4] Run focused RunRail/Timeline tests and fix only US4 regressions

## Phase 7: US5 Composer State Polish

**Goal**: Composer/ChatCanvas button state and text are correct for queued/running/retrying/recovering/stopped/failed/closed/cancelled.

**Independent Test Criteria**:
- queued/running/retrying/recovering states disable or enable buttons correctly.
- stopped/failed/closed/cancelled text is clear.
- retry/regenerate availability matches current state.
- provider unavailable never looks like generating.

### Tests

- [X] T077 [P] [US5] Add Composer queued and running state tests in existing Composer test file
- [X] T078 [P] [US5] Add Composer retrying and recovering state tests in existing Composer test file
- [X] T079 [P] [US5] Add Composer stopped, failed, closed, and cancelled text tests in existing Composer test file
- [X] T080 [P] [US5] Add retry/regenerate availability tests from the Composer State Matrix in existing Composer test file
- [X] T081 [P] [US5] Add provider unavailable not-generating regression test in existing Composer or ChatCanvas test file

### Implementation

- [X] T082 [US5] Add explicit Composer state label mapping in existing Composer state module or component
- [X] T083 [US5] Apply queued and running button availability rules in existing Composer component
- [X] T084 [US5] Apply retrying and recovering button availability rules in existing Composer component
- [X] T085 [US5] Apply stopped, failed, closed, and cancelled display text in existing Composer or ChatCanvas component
- [X] T086 [US5] Apply retry/regenerate availability exactly from the Composer State Matrix in existing Composer component
- [X] T087 [US5] Add disabled reason action logic for provider unavailable, generation in progress, retry already scheduled, recovery in progress, and no valid prompt; do not implement non-retryable failure display until backend retryability exists
- [X] T088 [US5] Reconcile Composer provider unavailable display with US2 provider gate behavior
- [X] T089 [US5] Run focused Composer/ChatCanvas tests and fix only US5 regressions

## Phase 8: Cross-cutting i18n

**Goal**: English and Chinese strings cover all new product surfaces and states.

**Dependencies**: US1-US5 copy keys should be known before this phase.

- [X] T090 [P] Add English Provider Test Console strings in existing English i18n message file
- [X] T091 [P] Add Chinese Provider Test Console strings in existing Chinese i18n message file
- [X] T092 [P] Add English Background tasks strings in existing English i18n message file
- [X] T093 [P] Add Chinese Background tasks strings in existing Chinese i18n message file
- [X] T094 [P] Add English runtime/worker/job/diagnostics event group strings in existing English i18n message file
- [X] T095 [P] Add Chinese runtime/worker/job/diagnostics event group strings in existing Chinese i18n message file
- [X] T096 [P] Add English provider unavailable, backend capability, and Composer state strings in existing English i18n message file
- [X] T097 [P] Add Chinese provider unavailable, backend capability, and Composer state strings in existing Chinese i18n message file
- [X] T098 Run i18n key coverage, typecheck, or nearest existing localization validation command
- [X] T099 Check new Provider Test Console, Background tasks, Timeline, and Composer user-visible strings use existing i18n message keys instead of hardcoded English-only text

## Phase 9: US6 Local Real Testing Documentation

**Goal**: Documentation explains how to run and verify the real testing path.

**Independent Test Criteria**:
- Docs explain provider env startup, Settings provider test, real message send, Background tasks, SSE timeline, cancel/recovery/failure.
- Docs clarify configured providers vs local draft.
- Docs clarify M6.5 boundaries and M7 exclusions.
- Docs build passes.

### Documentation Tasks

- [X] T100 [P] [US6] Create `docs-site/src/content/docs/architecture/provider-test-console.md`
- [X] T101 [P] [US6] Create `docs-site/src/content/docs/architecture/background-task-ux.md`
- [X] T102 [P] [US6] Create `docs-site/src/content/docs/runbooks/provider-testing.md`
- [X] T103 [US6] Update `docs-site/src/content/docs/runbooks/local-m6.md`
- [X] T104 [P] [US6] Create `docs-site/src/content/docs/devlog/2026-05-24-m6-5-real-testing-console.md`
- [X] T105 [US6] Update `docs-site/src/content/docs/roadmap/current-status.md`
- [X] T106 [US6] Update docs-site sidebar/nav file if current docs structure requires explicit registration
- [X] T107 [US6] Run `bun run build` from `docs-site/` and record result in devlog

## Final Phase: Validation & Scope Audit

- [X] T108 Run frontend lint/typecheck validation command and record result in devlog
- [X] T109 Run frontend unit/component tests for Settings, Chat/Composer, Background tasks, RunRail/Timeline, and i18n
- [X] T110 Run backend/API tests if provider check, readiness, worker event source, or API code changed
- [X] T111 Run UI/e2e tests if existing harness covers Settings/Chat flows
- [X] T112 Manually validate Settings > Providers provider check success/failure without exposing API keys
- [X] T113 Manually validate real_api/model_gateway provider unavailable guidance and Settings CTA
- [X] T114 Manually validate mock mode remains usable without provider readiness
- [X] T115 Manually validate Background tasks empty/current-job/latest-events states
- [X] T116 Manually validate RunRail/Timeline labels for all requested M6 events
- [X] T117 Manually validate Composer queued/running/retrying/recovering/stopped/failed/closed/cancelled states
- [X] T118 Verify automated tests cover retry, recovery, cancellation, failed, and dead states using deterministic fixtures or mocked event streams, not only manual real-provider failure paths
- [X] T119 Verify no M7 tool call, approval flow, tool execution protocol, API key persistence, final settings secret storage, broad M6 job data model rewrite, frontend runtime rewrite, or large i18n dependency was introduced
- [X] T120 Update `docs-site/src/content/docs/devlog/2026-05-24-m6-5-real-testing-console.md` with exact validation commands, pass/fail results, and known limitations

## Dependencies

```text
Phase 1 Setup
  -> Phase 2 Foundational
    -> US1 Provider Test Console
    -> US2 Chat Provider Readiness
    -> US3 Background Tasks Observer
    -> US4 Timeline Event Productization
    -> US5 Composer State Polish
      -> Cross-cutting i18n
      -> US6 Documentation
        -> Final Validation
```

### Story dependencies

- **US1** can start after Phase 2.
- **US2** can start after Phase 2, but benefits from US1 CTA target existing.
- **US3** can start after Phase 2.
- **US4** can start after Phase 2 and may share label mapping with US3.
- **US5** can start after US2 provider gate shape is clear.
- **US6** can start after US1-US5 behavior stabilizes, but architecture docs can draft in parallel.
- **Final validation** depends on all implementation/docs phases.

## Parallel Opportunities

### After Phase 2

These can run in parallel if separate agents avoid overlapping files:

```text
US1 Provider Test Console
US3 Background Tasks Observer
US4 Timeline Event Productization
```

US2 and US5 can also run in parallel only if ownership of Composer/ChatCanvas files is coordinated.

### US1 parallel test work

```text
T018 configured provider render test
T019 no-provider empty state test
T020 local draft explanatory copy test
T021 checking-state test
T022 success-state test
T023 failure sanitization test
```

### US3 parallel test work

```text
T045 empty-state test
T046 current run/job state test
T047 worker diagnostics summary test
T048 latest worker/job events test
T049 read-only controls test
T050 snapshot priority test
T051 recovering observation test
```

### US4 parallel test work

```text
T061 event group tests
T062 job_claimed/lease_renewed tests
T063 recovering/retry scheduled tests
T064 attempt failed/retry exhausted tests
T065 cancellation/diagnostics tests
T066 unknown fallback test
T067 out-of-order stream-order test
```

### i18n parallel work

English and Chinese message additions can be parallelized only if they edit separate locale files:

```text
T090, T092, T094, T096
T091, T093, T095, T097
```

### docs parallel work

```text
T100 provider-test-console architecture doc
T101 background-task-ux architecture doc
T102 provider-testing runbook
T104 devlog draft
```

## MVP Scope

Recommended MVP is **US1 + US2**:

1. Provider Test Console makes backend provider readiness visible.
2. Chat/Composer stops misleading users in real modes when provider is unavailable.
3. Mock mode remains unaffected.

Then add US3/US4 for worker observability, US5 for polish, US6/docs before completion.
