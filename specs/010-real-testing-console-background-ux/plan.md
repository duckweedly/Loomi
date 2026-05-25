# Implementation Plan: Real Testing Console & Background UX

**Feature**: `010-real-testing-console-background-ux`  
**Date**: 2026-05-24  
**Phase**: M6.5 productization  
**Spec**: `specs/010-real-testing-console-background-ux/spec.md`  
**Current roadmap context**: M6 worker job pipeline and provider readiness are already merged; M6.5 must reuse M5.5/M6 provider readiness, runtime mode, worker job, and timeline surfaces.

## Summary

Productize the existing real testing path by turning provider readiness and worker job internals into clear read-only user surfaces:

1. Provider Test Console in Settings.
2. Provider readiness gating in Chat/Composer for real modes.
3. Read-only Background tasks panel.
4. Runtime/worker/job event groups in RunRail/Timeline.
5. Composer/ChatCanvas status polish.
6. English/Chinese copy completion.
7. Local real testing documentation and validation coverage.

## Constitution Alignment

This feature aligns with `.specify/memory/constitution.md`:

- **Runnable Vertical Slices**: each user story produces visible UI state, event display, or docs/runbook output.
- **Core Flow Before Platform Complexity**: M6.5 productizes provider/job observability and does not pull M7 tool call, approval, or tool execution forward.
- **Observable Agent Execution**: worker/job state, diagnostics, events, errors, cancellation, retry, and recovery become visible.
- **Safety, Permissions, and Data Boundaries**: API keys and provider credentials must not be displayed or persisted.
- **Documentation as Done**: docs-site architecture, runbook, devlog, and roadmap updates are included.

## Constraints

- Do not build M7 tool call.
- Do not build approval.
- Do not build tool execution protocol.
- Do not persist API keys.
- Do not add final settings secret storage.
- Do not rewrite the M6 job data model unless a bug is confirmed.
- Do not rewrite frontend runtime architecture.
- Do not add a large i18n dependency.
- Documentation updates are part of done.

## Technical Strategy

Use a thin adapter/productization layer over existing surfaces:

- Existing provider readiness source becomes Provider Test Console data.
- Existing backend configured provider data remains authoritative.
- Existing local draft state remains browser-session-only.
- Existing runtime capability/mode controls Chat/Composer gating.
- Existing run/job/SSE/timeline events feed Background tasks and RunRail.
- Existing i18n mechanism receives new keys; no new large dependency.
- Existing tests are expanded around states and copy.

### Provider Test Action

Provider Test Console treats Test connection as a per-configured-provider UI action. The implementation should first reuse the existing backend provider readiness/check surface. If the backend readiness surface is aggregate rather than per-provider, the UI should map the aggregate result back onto configured provider rows without adding broad provider storage, secret persistence, or final settings secret-management APIs.

### Background Task Snapshot Selection

The Background tasks panel chooses its read-only snapshot in this order:

1. Job associated with the currently selected Chat run.
2. Empty state if no selected run job is available.

Cross-run active worker job discovery is deferred until a richer right-panel data contract exists.

This panel observes state only; it does not claim, retry, recover, cancel, or otherwise mutate worker jobs.

### Recovering State

`recovering` is supported as a UI observation state. If the backend exposes it as a job status, the UI displays that status. If not, the UI derives recovering from `job_recovering` events. This must not trigger a M6 job data model rewrite unless implementation discovers a real bug.

### Timeline Event Ordering

RunRail/Timeline and Background tasks preserve incoming event stream order for M6.5. Unknown/future worker or job events receive a readable fallback label while retaining raw event type for debugging. M6.5 does not reorder event history.

## Milestones

### Milestone 1: Provider Test Console

- Surface configured backend providers.
- Add provider test action and status rendering.
- Add empty provider env guidance.
- Add local draft vs configured provider explanation.
- Add secret sanitization expectations to UI/tests.

### Milestone 2: Chat/Composer Readiness Gating

- Derive provider availability from existing readiness state.
- In real modes, show provider unavailable CTA instead of generation.
- Ensure mock mode remains unaffected.
- Clarify button availability and status labels.

### Milestone 3: Background Tasks Panel

- Replace placeholder with read-only observation panel.
- Show selected run job or empty state.
- Show worker diagnostics summary.
- Show latest worker/job events.
- Add empty state.

### Milestone 4: RunRail / Timeline Event Productization

- Group runtime, worker, and job events.
- Add readable labels for `job_claimed`, `lease_renewed`, `job_recovering`, `job_retry_scheduled`, `job_attempt_failed`, `job_retry_exhausted`, cancellation, worker diagnostics, and unknown events.

### Milestone 5: Copy and i18n

- Complete English/Chinese strings for Provider Test Console, Background tasks, runtime/worker/job event groups, provider unavailable, backend capability, Composer states, and disabled reasons.

### Milestone 6: Documentation and Tests

- Add architecture docs.
- Add local real testing runbook.
- Update devlog and roadmap.
- Add provider, background task, i18n, mock/real mode, and Composer state tests.
- Run code validation and docs build.

## Risk Register

| Risk | Impact | Mitigation |
|------|--------|------------|
| Existing provider API lacks all display fields | Console may need incomplete states | Use available fields first; mark missing optional fields as unavailable rather than inventing config |
| Provider error may include secret-like text | Secret leakage | Sanitize before display and add tests |
| Existing provider check is aggregate-only | UI action may appear per-provider but data is aggregate | Map aggregate result back to provider rows without adding storage APIs |
| Job event naming differs from UI expectations | Missing timeline labels | Map known M6 event types explicitly; include fallback unknown-event label |
| Background panel implies control | User may expect mutations | Read-only copy and no mutating controls |
| Provider gating blocks mock mode | Regression | Add explicit mock-mode tests |
| Composer state matrix grows confusing | UI inconsistency | Define state table and test each state |

## Test Strategy

### Unit / component tests

- Provider configured list rendering.
- Provider missing empty state.
- Provider check success/failure states.
- Secret sanitization in provider error display.
- Local draft copy and no backend write expectation.
- Composer state labels and buttons.
- Timeline event label mapping and fallback.
- Background tasks empty/current/event states.
- i18n English/Chinese key coverage.

### Deterministic Worker Event Fixtures

Tests should use deterministic fixtures or mocked event streams for retry, recovery, cancellation, failed, and dead states. Manual real-provider failure testing is useful, but automated coverage must not depend on causing real provider outages or waiting for nondeterministic worker timing.

### Integration / route tests

- Settings > Providers opens console.
- Open Settings > Providers CTA from Chat/Composer.
- Mock mode remains usable with no provider.
- Real mode shows provider unavailable guidance.
- Background tasks displays run/job event state from test fixture.

### Manual validation

- Start API with provider env.
- Start web.
- Open Settings > Providers.
- Test provider connection.
- Send real model message.
- Observe queued run, worker job, SSE timeline.
- Validate cancel/recovery/failure state paths.

## Documentation Plan

Update:

- `docs-site/src/content/docs/architecture/provider-test-console.md`
- `docs-site/src/content/docs/architecture/background-task-ux.md`
- `docs-site/src/content/docs/runbooks/provider-testing.md`
- `docs-site/src/content/docs/runbooks/local-m6.md`
- `docs-site/src/content/docs/devlog/2026-05-24-m6-5-real-testing-console.md`
- `docs-site/src/content/docs/roadmap/current-status.md`

## Validation Commands

Exact commands should be confirmed from current package scripts during implementation. Expected validation categories:

- frontend typecheck/lint/test command
- API/backend test command if provider readiness or worker state code is touched
- web/UI test command for Settings/Chat/Background tasks
- docs build from `docs-site/`: `bun run build`

If a command cannot run, record the exact reason in devlog.
