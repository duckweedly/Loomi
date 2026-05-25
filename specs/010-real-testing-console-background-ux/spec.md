# Feature Specification: Real Testing Console & Background UX

**Feature Directory**: `specs/010-real-testing-console-background-ux`  
**Created**: 2026-05-24  
**Status**: Ready for planning/implementation  
**Scope**: M6.5 productization  
**Input**: Productize the real local testing path for provider readiness, Chat/Composer gating, worker jobs, background tasks, Timeline events, i18n, docs, and validation.

## Overview

M5.5/M6 made Loomi's real provider path, provider readiness, Composer/Chat surfaces, and worker job pipeline usable from environment-backed configuration. M6.5 turns those capabilities into a stable real testing workbench: users should know which providers are configured, whether real model calls are available, what background worker/job state is active, why a run is blocked/failing/retrying/recovering, and how to reproduce the local real testing path.

## Clarifications

### Session 2026-05-24

- Q: Provider Test Console 的 Test connection 是逐 provider 测试还是整体 readiness 刷新？ → A: UI 以单个 configured provider 为操作单位；如果现有后端只支持整体 readiness 刷新，前端可将整体结果映射回 provider 行展示，不新增大范围 provider API。
- Q: Background tasks 中 current run/job 如何选择？ → A: M6.5 实现当前选中 Chat run 的 job；如果没有选中 run，则显示 empty state。跨 run 主动发现 active job 需要后续 richer right-panel data contract，不在本切片伪造。
- Q: recovering 是真实 job status 还是 UI 派生状态？ → A: 如果后端暴露 recovering job status 就直接展示；否则从 `job_recovering` event 派生为只读 UI observation state，不重写 M6 job 模型。
- Q: retry/regenerate 可用条件如何收口？ → A: 以 Composer state matrix 为准；provider unavailable、queued、running、retrying、recovering 禁用 retry/regenerate。当前 domain 未暴露 retryability 来源，因此非 retryable failure 的 UI 区分保留到后续 backend retryability contract。
- Q: out-of-order worker/job events 如何处理？ → A: M6.5 保持 SSE/event stream 原始顺序，不做事件历史重排；unknown/future event 使用 fallback label 并保留 raw event type。

## User Story

As a local Loomi tester running the real model path from environment variables, I need a clear workspace that shows whether providers are configured, whether the selected backend capability can call a model, what background run/job state is active, and why failures/retries/recoveries happen, so I can confidently test the real API path without mistaking unavailable providers or background jobs for active generation.

## Goals

1. Upgrade Settings > Providers into a Provider Test Console.
2. Clearly separate backend configured providers from browser-local draft provider fields.
3. Connect provider readiness to Chat/Composer behavior.
4. Upgrade Background tasks from placeholder to read-only observation panel.
5. Productize RunRail/Timeline display for M6 worker/job events.
6. Polish Composer/ChatCanvas state text and button availability.
7. Finish English/Chinese terminology for provider, background task, worker/job/runtime states.
8. Add local real testing runbook.
9. Add test coverage for success, failure, empty, unavailable, mock, i18n, and worker state scenarios.

## Non-Goals

- Do not implement M7 tool calls.
- Do not implement approval flows.
- Do not implement tool execution protocol.
- Do not persist API keys in browser or final settings secret storage.
- Do not redesign the worker job data model unless a real bug is found.
- Do not rewrite frontend runtime architecture.
- Do not add a large i18n dependency.
- Do not make local draft provider settings affect backend model calls.

## Actors

- **Local tester**: starts API/web locally and validates real model behavior.
- **Developer**: diagnoses provider, model gateway, worker, job, retry, and recovery behavior.
- **Mock-mode user**: continues using mock mode without being blocked by provider readiness.

## Assumptions

- Backend provider configuration already comes from environment variables.
- Existing M5.5/M6 provider readiness and worker job APIs/events are reusable.
- Browser-local provider draft state already exists or can remain session-local.
- Mock mode should remain available even when no real provider is configured.
- `real_api` and `model_gateway` modes require provider readiness before implying generation is active.
- Documentation updates are part of done.

## User Scenarios & Acceptance Tests

### Scenario 1: View configured backend providers

Given the API is started with one or more provider environment variables  
When the user opens Settings > Providers  
Then the user sees configured providers with provider id, family, model, base URL, status, and message  
And the UI makes clear these configured providers affect real model calls.

### Scenario 2: No backend provider configured

Given the API is started without provider environment variables  
When the user opens Settings > Providers  
Then the user sees an empty configured-provider state  
And the UI shows readable environment configuration guidance  
And no API key value is displayed.

### Scenario 3: Test provider connection success

Given a configured provider can be reached  
When the user clicks Test connection for that provider row  
Then the provider enters checking state  
And then shows success with a readable message.

### Scenario 4: Test provider connection failure

Given a configured provider fails readiness validation  
When the user clicks Test connection for that provider row  
Then the provider enters checking state  
And then shows failed with a readable sanitized error  
And no API key or secret material is displayed.

### Scenario 5: Local draft does not affect backend calls

Given the user edits provider fields in a local draft area  
When the user sends a real model message  
Then backend configured providers are still the source of truth  
And the UI states that local draft is browser-session-only and not saved.

### Scenario 6: Chat blocks real modes when provider unavailable

Given runtime mode is `real_api` or `model_gateway`  
And provider readiness is unavailable  
When the user opens Chat or Composer  
Then the user sees provider unavailable messaging  
And a CTA opens Settings > Providers  
And the UI does not show a misleading generating state.

### Scenario 7: Mock mode is not blocked

Given runtime mode is mock  
And no real provider is configured  
When the user sends a mock-mode message  
Then Composer/Chat can proceed normally  
And provider unavailable messaging does not block mock mode.

### Scenario 8: Observe current background task state

Given a run has an associated worker job  
When the user opens Background tasks  
Then the panel shows the selected Chat run job when available, or empty state when no selected run job exists  
And the panel shows worker diagnostics summary and latest worker/job events when available.

### Scenario 9: Background tasks empty state

Given there is no selected run job  
When the user opens Background tasks  
Then the panel shows a clear empty state explaining where run/job state will appear.

### Scenario 10: Timeline shows worker/job events

Given worker/job events occur  
When the user opens RunRail/Timeline  
Then job claim, lease renewal, recovery, retry, attempt failure, retry exhausted, cancellation, diagnostics, and unknown future events appear in readable grouped form while preserving raw event identity.

### Scenario 11: Composer state polish

Given a run is queued, running, retrying, recovering, stopped, failed, closed, or cancelled  
When the user views Composer/ChatCanvas  
Then buttons, labels, and retry/regenerate availability match the state matrix.

## Functional Requirements

### Provider Test Console

- **FR-001**: Settings > Providers MUST show backend configured providers separately from local draft provider fields.
- **FR-002**: Each configured provider MUST display provider id, family, model, base URL, status, and message.
- **FR-003**: The console MUST expose Test connection as a per-configured-provider UI action. If the existing backend only supports refreshing aggregate provider readiness, the UI MAY map the aggregate readiness result back to provider rows, but M6.5 MUST NOT introduce broad new provider storage or secret-management APIs.
- **FR-004**: Test connection MUST show `checking`, `success`, and `failed` states for the affected configured provider row.
- **FR-005**: Provider errors displayed in the UI MUST be sanitized and readable. The UI MUST NOT render API keys, bearer tokens, or secret-like provider credentials, even if the backend returns raw provider error text.
- **FR-006**: Empty provider state MUST explain required environment configuration without showing secret values.
- **FR-007**: Local draft provider UI MUST state that it is browser-session-only, unsaved, and does not affect backend calls.
- **FR-008**: Local draft changes MUST NOT write backend provider config or change real model invocation.

### Chat / Composer Provider Readiness

- **FR-009**: Chat/Composer MUST distinguish mock mode from real provider-dependent modes.
- **FR-010**: In `real_api` or `model_gateway`, unavailable provider readiness MUST show clear provider unavailable messaging.
- **FR-011**: Provider unavailable state MUST include an Open Settings > Providers CTA.
- **FR-012**: Provider unavailable state MUST NOT show the model as generating.
- **FR-013**: Mock mode MUST NOT be blocked by missing or failed real provider readiness.

### Background Tasks Panel

- **FR-014**: Background tasks panel MUST be read-only.
- **FR-015**: The panel MUST show the current observable run/job snapshot for the selected Chat run when available, and otherwise show an empty state. Cross-run job discovery is deferred until a richer right-panel data contract exists.
- **FR-016**: The panel MUST show job states: queued, leased, retrying, completed, failed, cancelled, and dead. It MUST also show recovering when the backend exposes it as a job status, or derive a read-only recovering observation state from `job_recovering` events when the backend does not.
- **FR-017**: The panel MUST include worker diagnostics summary when available.
- **FR-018**: The panel MUST show latest worker/job events when available.
- **FR-019**: The panel MUST show a clear empty state when no background task exists.
- **FR-020**: The panel MUST avoid controls that imply mutating worker/job state.

### RunRail / Timeline

- **FR-021**: RunRail/Timeline MUST group runtime, worker, and job events using user-readable labels.
- **FR-022**: Timeline MUST display `job_claimed`.
- **FR-023**: Timeline MUST display `lease_renewed`.
- **FR-024**: Timeline MUST display `job_recovering`.
- **FR-025**: Timeline MUST display `job_retry_scheduled`.
- **FR-026**: Timeline MUST display `job_attempt_failed`.
- **FR-027**: Timeline MUST display `job_retry_exhausted`.
- **FR-028**: Timeline MUST display cancellation events.
- **FR-029**: Timeline MUST display worker diagnostics events.
- **FR-030**: Timeline MUST preserve incoming worker/job event order and avoid rewriting event history during M6.5 productization.
- **FR-031**: Timeline MUST provide a readable fallback label for unknown worker/job events while keeping raw event type visible or accessible.

### Composer / ChatCanvas Polish

- **FR-032**: Composer buttons MUST reflect queued/running/retrying/recovering states.
- **FR-033**: Provider unavailable MUST not appear as active generation.
- **FR-034**: Stopped, failed, closed, and cancelled states MUST use clear user-facing text.
- **FR-035**: Regenerate/retry controls MUST follow the Composer state matrix: regenerate is enabled only when an existing valid prompt/turn can be submitted again; retry is enabled for failed states until backend retryability is exposed; both are disabled during provider unavailable, queued, running, retrying, and recovering states.
- **FR-036**: Disabled retry/regenerate states MUST expose or encode a concise reason for internal action logic, such as provider unavailable, generation in progress, retry already scheduled, recovery in progress, or no valid prompt. M6.5 MUST NOT implement non-retryable failure display until backend retryability is exposed.

### i18n

- **FR-037**: English and Chinese text MUST cover Provider Test Console.
- **FR-038**: English and Chinese text MUST cover Background tasks.
- **FR-039**: English and Chinese text MUST cover runtime, worker, and job event groups.
- **FR-040**: English and Chinese text MUST cover provider unavailable, backend capability, and Composer states.
- **FR-041**: New user-visible strings for Provider Test Console, Background tasks, Timeline, and Composer MUST use the existing i18n mechanism rather than new English-only hardcoded text.

### Documentation

- **FR-042**: Documentation MUST explain Provider Test Console behavior.
- **FR-043**: Documentation MUST explain Background tasks UX.
- **FR-044**: Documentation MUST include local provider testing runbook.
- **FR-045**: Documentation MUST update current roadmap/status.
- **FR-046**: Documentation MUST record validation results and known limitations.

## Key Entities

- **Configured Provider**: Backend-discovered provider configuration that affects real model calls.
- **Local Draft Provider**: Browser-session-only draft fields that do not save or affect backend calls.
- **Provider Check**: A connection/readiness test with checking/success/failed states. It is initiated per provider row in UI and may use aggregate backend readiness if that is the existing backend surface.
- **Backend Capability**: Runtime capability such as mock, real_api, or model_gateway.
- **Run**: A user-visible chat generation attempt or session action.
- **Worker Job**: Background unit associated with a run.
- **Background Task Snapshot**: Read-only UI snapshot selected by priority: selected Chat run job, then empty state. It may include job status, derived observation state, diagnostics, and latest events.
- **Recovering Observation State**: A UI-visible state that is either a backend job status or a derived read-only state from `job_recovering`; it MUST NOT require rewriting the M6 worker job data model.
- **Worker Diagnostics**: Summary of worker health, lease, recovery, retry, or failure information.
- **Worker/Job Event**: Timeline event emitted by the worker job pipeline.
- **Composer State**: User-facing state controlling input, stop, retry, regenerate, and status text.

## Edge Cases

- No configured providers.
- Provider configured but connection fails.
- Provider check returns an error containing secret-like text.
- Local draft exists while backend has no configured provider.
- Local draft has a model/base URL that differs from backend.
- Runtime mode is mock with provider unavailable.
- Runtime mode is real_api/model_gateway with provider unavailable.
- Job exists without current run context.
- Run exists without job event history.
- Worker diagnostics unavailable.
- Job transitions from retrying to dead.
- User cancels while job is queued or leased.
- Worker/job events arrive out of expected order; Timeline and Background tasks preserve raw event order from the stream and use readable fallback labels without rewriting event history.
- Locale changes while status is active.

## Success Criteria

- **SC-001**: A local tester can determine whether real providers are configured within 10 seconds of opening Settings > Providers.
- **SC-002**: A local tester can run a provider connection test and understand success/failure without reading logs.
- **SC-003**: 100% of provider failure displays avoid exposing API keys or secret values.
- **SC-004**: In real provider-dependent modes, users are never shown a generating state when provider readiness is unavailable.
- **SC-005**: Mock-mode chat remains usable when no real provider is configured.
- **SC-006**: A local tester can identify current run/job state from the Background tasks panel without opening developer tools.
- **SC-007**: Timeline displays all M6 worker job event categories in readable grouped form.
- **SC-008**: Composer controls match run/job state across queued, running, retrying, recovering, stopped, failed, closed, cancelled.
- **SC-009**: English and Chinese strings cover all new user-visible states.
- **SC-010**: Runbook enables a developer to complete a real local provider test path from environment setup to Chat validation.
