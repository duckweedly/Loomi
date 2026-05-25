# Research: Real Testing Console & Background UX

## Decision 1: Reuse existing provider readiness rather than introduce new provider storage

**Decision**: Provider Test Console displays backend configured providers from existing readiness/configuration surfaces. Local draft remains browser-only.

**Rationale**:
- M6.5 is productization, not settings secret storage.
- Avoids persisting API keys.
- Keeps backend env configuration as source of truth for real calls.

**Alternatives considered**:
- Add final provider secret storage.
- Let local draft override real calls.
- Store API keys in browser session.

## Decision 2: Treat Test connection as per-provider UI action

**Decision**: The UI exposes Test connection per configured provider row. If existing backend readiness is aggregate-only, map aggregate readiness back to provider rows without adding broad storage or secret-management APIs.

**Rationale**:
- Users think about testing a visible provider row.
- M6.5 should improve user clarity without pulling new backend provider storage forward.
- This preserves existing M5.5/M6 surfaces.

**Alternatives considered**:
- Add a new provider-specific backend API.
- Only expose global refresh readiness.
- Hide test action until a per-provider API exists.

## Decision 3: Gate only real provider-dependent modes

**Decision**: Provider readiness blocks `real_api` and `model_gateway`, but not mock mode.

**Rationale**:
- Mock mode is useful for development even without providers.
- Real modes would otherwise imply generation when no provider can run.

**Alternatives considered**:
- Global provider readiness blocking all Chat.
- Showing provider warning only in Settings.

## Decision 4: Background tasks panel is read-only

**Decision**: The right panel observes worker job state and events only.

**Rationale**:
- M6.5 should improve visibility.
- Mutating controls would drift toward M7/tool execution/approval semantics.
- Read-only panel is safer and easier to test.

**Alternatives considered**:
- Manual retry/recover/cancel controls in Background tasks.
- Worker management UI.

## Decision 5: Background task snapshot priority

**Decision**: Select snapshot by selected Chat run job, then empty state. Cross-run active job discovery is deferred.

**Rationale**:
- Chat context should drive the visible state when available.
- Empty state avoids implying hidden jobs.

**Alternatives considered**:
- Always show cross-run latest job.
- Only show selected run job.
- Show all jobs as a queue management interface.

## Decision 6: Recovering as observation state

**Decision**: Display backend `recovering` status if it exists; otherwise derive read-only recovering state from `job_recovering` events.

**Rationale**:
- User should see recovery behavior.
- Avoids rewriting M6 job data model.
- Keeps recovering visible in both panel and Composer state polish.

**Alternatives considered**:
- Add recovering to backend job model unconditionally.
- Hide recovery unless backend status exists.

## Decision 7: Timeline labels preserve raw event identity and stream order

**Decision**: UI uses readable localized labels while preserving raw event type and incoming order.

**Rationale**:
- Developers need to map UI state to worker logs/events.
- Users need readable grouping.
- M6.5 should not rewrite event history.

**Alternatives considered**:
- Raw event-only timeline.
- Fully hiding technical event names.
- Reordering events by UI severity or category.

## Decision 8: No large i18n dependency

**Decision**: Extend existing localization mechanism.

**Rationale**:
- Scope is copy completion, not localization architecture.
- Avoids dependency churn.

**Alternatives considered**:
- Adding a new i18n framework.
- Hardcoding English-only text.

## Decision 9: Secret display must be sanitized at display boundary

**Decision**: Provider error messages rendered in UI must be sanitized and tested.

**Rationale**:
- External provider errors may include request metadata.
- The UI must never leak API keys.

**Alternatives considered**:
- Trusting provider errors as display-safe.
- Showing raw backend exceptions in the console.
