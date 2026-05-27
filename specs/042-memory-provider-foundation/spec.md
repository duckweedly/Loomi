# Feature Specification: Memory Provider Foundation

**Feature Branch**: `[042-memory-provider-foundation]`

**Created**: 2026-05-26

**Status**: Draft

**Input**: User description: "Bring Loomi memory toward Arkloop-class mechanism parity by adding the first memory provider foundation slice: selectable provider configuration, health/status diagnostics, safe runtime readiness, and Settings visibility, while preserving Loomi's own expression and not yet adding memory tools or distillation."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Choose a memory provider foundation (Priority: P1)

As a Loomi user, I want Memory settings to show whether memory is enabled and which provider is selected, so I can understand the system's memory mode before trusting it in a run.

**Why this priority**: Provider selection is the base contract for every later memory tool, recall, and distillation slice.

**Independent Test**: Start Loomi with the local memory provider available, open Settings > Memory, and verify the selected provider, enablement state, and provider readiness are rendered from backend data.

**Acceptance Scenarios**:

1. **Given** memory is enabled with the local provider, **When** the user opens Settings > Memory, **Then** the page shows the local provider as selected and available.
2. **Given** memory is disabled, **When** the user opens Settings > Memory, **Then** the page shows memory disabled and does not claim semantic recall is available.
3. **Given** an unknown provider value is stored, **When** the backend resolves memory config, **Then** it falls back to the local provider and reports a safe degraded diagnostic.

---

### User Story 2 - Diagnose semantic provider readiness (Priority: P1)

As a Loomi operator, I want provider status to distinguish unconfigured, configured, healthy, unhealthy, and degraded states, so I can fix memory setup without guessing.

**Why this priority**: The current UI can show status-like labels without a real provider contract. The new foundation must make status grounded and actionable.

**Independent Test**: Configure a semantic provider endpoint as missing, unreachable, and healthy in deterministic tests, then verify API and UI status projections match the state without leaking secrets.

**Acceptance Scenarios**:

1. **Given** a semantic provider is selected but required configuration is missing, **When** status is requested, **Then** Loomi returns `unconfigured` with a safe reason.
2. **Given** a semantic provider endpoint is unreachable, **When** status is requested, **Then** Loomi returns `unhealthy` with a redacted error code and no raw request headers.
3. **Given** a semantic provider health check succeeds, **When** status is requested, **Then** Loomi returns `healthy` and records the check time.

---

### User Story 3 - Expose safe runtime readiness to runs (Priority: P2)

As a maintainer, I want each run to know the resolved memory provider state, so future memory tools and recall can reuse a durable, observable readiness boundary.

**Why this priority**: Later memory tools and post-run persistence need the run to know which memory provider was active without re-resolving unsafe config in multiple places.

**Independent Test**: Start a run with memory enabled and disabled, then verify the run context safe summary includes memory provider readiness metadata and no secrets.

**Acceptance Scenarios**:

1. **Given** memory is enabled and healthy, **When** a run context is prepared, **Then** it includes safe provider readiness metadata.
2. **Given** memory is disabled, **When** a run context is prepared, **Then** it marks memory readiness as disabled and does not attempt semantic provider access.
3. **Given** provider resolution fails, **When** a run context is prepared, **Then** the run can continue with memory unavailable and a safe event explains the condition.

---

### User Story 4 - Preserve Loomi's existing memory management behavior (Priority: P2)

As an existing Loomi user, I want current approved memory list, search, detail, audit, and delete behavior to continue working while provider configuration is added.

**Why this priority**: Provider foundation must not regress M13/M14 memory management or user deletion control.

**Independent Test**: Run existing memory API and Settings tests after provider state is introduced and verify approved-entry management still passes.

**Acceptance Scenarios**:

1. **Given** approved local memories exist, **When** provider foundation is enabled, **Then** list/search/detail/delete still use safe summaries and existing scope checks.
2. **Given** audit history exists, **When** the user opens Settings > Memory, **Then** history remains grounded in real audit data.

### Edge Cases

- Provider config is partially filled and must not be treated as healthy.
- Provider health returns a non-success response with a body containing secret-like text.
- Memory is disabled while a run is being prepared.
- Existing local memory entries exist while the selected semantic provider is unavailable.
- A stale frontend status response arrives after a newer refresh request.
- Browser smoke runs without any configured semantic provider.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST expose a memory provider configuration state with at least enabled/disabled, selected provider, commit-after-run preference, configured status, health status, checked timestamp, and safe diagnostic fields.
- **FR-002**: Loomi MUST include the current local memory store as the default provider so existing approved-entry behavior remains available.
- **FR-003**: Loomi MUST represent semantic providers as selectable future-capable provider records without requiring memory tools or distillation in this slice.
- **FR-004**: Loomi MUST normalize unknown or missing provider selections to a safe local-provider default and report a degraded diagnostic.
- **FR-005**: Loomi MUST provide a backend status API that Settings can use without fabricated UI-only state.
- **FR-006**: Loomi MUST ensure provider status responses never include API keys, authorization headers, raw endpoint credentials, raw provider traces, or secret-like values.
- **FR-007**: Loomi MUST add safe memory readiness metadata to run context or run events when a run is prepared.
- **FR-008**: Loomi MUST continue to support existing memory list, search, detail, audit, proposal, approval, denial, and delete behavior.
- **FR-009**: Loomi MUST update Settings > Memory to render provider enablement, selected provider, status, diagnostics, and refresh behavior from backend data.
- **FR-010**: Loomi MUST document provider foundation behavior, status states, validation commands, and known non-goals in the docs site.

### Key Entities *(include if feature involves data)*

- **Memory Provider Config**: User-scoped memory enablement and provider selection state, including safe non-secret metadata and commit-after-run preference.
- **Memory Provider Status**: A safe runtime projection describing whether the selected provider is disabled, local, unconfigured, healthy, unhealthy, or degraded.
- **Memory Readiness Snapshot**: Run-scoped safe metadata recording the resolved memory state at run preparation time.
- **Provider Diagnostic**: Redacted status detail intended to help users fix setup without exposing credentials or raw provider payloads.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Settings > Memory shows backend-derived provider state for enabled, disabled, local, unconfigured, healthy, and unhealthy scenarios.
- **SC-002**: Existing memory API and Settings management tests continue to pass without changing their user-visible behavior.
- **SC-003**: Provider diagnostics expose no secret-like values in API responses, run events, logs used by tests, or frontend state.
- **SC-004**: A run can be prepared successfully when memory is disabled or provider health is unavailable.
- **SC-005**: Documentation clearly marks memory tools, automatic distillation, external semantic storage, and full provider configuration flows as later slices.

## Assumptions

- Loomi keeps the current local approved-entry memory store as the default provider.
- This slice does not implement agent-facing memory tools.
- This slice does not implement automatic conversation distillation or semantic recall.
- This slice does not copy Arkloop UI wording, private names, or visual styling.
- Semantic provider adapters may be represented as configured/unconfigured status records before full read/write integration exists.
