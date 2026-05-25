# Feature Specification: Persona Skill Foundation

**Feature Branch**: `[015-persona-skill-foundation]`

**Created**: 2026-05-25

**Status**: Draft

**Input**: User description: "M10 Persona/Skill foundation minimal slice: define persona data model; support built-in persona config sync to DB; persona includes name, description, system prompt, model route, allowed tool names, reasoning mode, budget summary, and version; thread/run can select or inherit persona; RunContext records the persona snapshot/version used by the current run; Timeline/debug can show safe persona summary; frontend only needs a minimal persona selector or read-only display; prioritize a truly verifiable run-to-persona path. Non-goals: full Skill marketplace, plugin install, MCP, Memory/RAG, Sandbox/Desktop Runtime, multi-agent, exposing raw persona prompt in normal Timeline, new large framework/permission system. docs-site must be updated during implementation."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Sync built-in personas into durable state (Priority: P1)

As a Loomi developer validating M10 configuration behavior, I want built-in persona definitions to sync into durable storage with versions, so that runs can refer to stable persona records instead of hardcoded request memory.

**Why this priority**: This is the smallest prerequisite for M10 Persona DB and keeps the slice grounded in durable product data before any frontend or marketplace work.

**Independent Test**: Start or invoke the sync path with at least one built-in persona definition, then verify the DB contains the persona fields, active version, safe summary, and a deterministic update path when the built-in definition changes.

**Acceptance Scenarios**:

1. **Given** a built-in persona config with name, description, system prompt, model route, allowed tool names, reasoning mode, budget summary, and version, **When** Loomi syncs built-ins, **Then** durable persona records and active versions are created or updated idempotently.
2. **Given** the same built-in persona config is synced twice, **When** the second sync runs, **Then** no duplicate active persona or version is created.
3. **Given** a built-in persona changes its version or safe fields, **When** sync runs again, **Then** new runs can resolve the new active version while old run snapshots remain attributable to their original version.

---

### User Story 2 - Run with selected or inherited persona (Priority: P2)

As a user starting a run, I want Loomi to use the persona selected for the run or inherited from the thread/default, so that model route, allowed tools, reasoning mode, budget summary, and system prompt are applied consistently.

**Why this priority**: The feature is not complete until persona selection affects the real worker/RunContext path rather than only existing as configuration data.

**Independent Test**: Select a persona or use the default inherited persona, create a run, and verify the worker RunContext includes the persona snapshot/version before runtime invocation and uses persona-controlled model route and allowed tool names.

**Acceptance Scenarios**:

1. **Given** a thread has a selected persona, **When** a run is created without an explicit run persona override, **Then** the run inherits the thread persona and records the resolved persona snapshot/version.
2. **Given** a run is created with an explicit persona selection, **When** the worker prepares RunContext, **Then** the run-specific persona wins over the thread/default persona.
3. **Given** no thread or run persona is selected, **When** a run starts, **Then** Loomi resolves the configured default built-in persona and records its snapshot/version.
4. **Given** a selected persona is inactive, missing, or has no usable active version, **When** the run starts, **Then** the worker fails safely before provider/runtime invocation with a redacted explanation.

---

### User Story 3 - Observe safe persona summary in Timeline/debug (Priority: P3)

As a user or developer reading a run Timeline/debug panel, I want to see which persona influenced the run without exposing hidden prompt text, so that run behavior is explainable and safe to inspect.

**Why this priority**: Constitution requires observable agent execution, but Persona/Skill introduces prompt and tool policy data that must be summarized carefully.

**Independent Test**: Create a run with a resolved persona, open Timeline/debug from live SSE and history replay, and verify the persona name/version/model route/reasoning/budget/tool-count summary is visible while the raw system prompt is absent.

**Acceptance Scenarios**:

1. **Given** a run uses a persona, **When** Timeline/debug renders live events, **Then** it shows a safe persona summary with name, version, model route label, reasoning mode, budget summary, and allowed tool names/count.
2. **Given** the browser refreshes after run completion, **When** history replay renders the run, **Then** the same safe persona summary is visible.
3. **Given** persona system prompt exists in durable storage and RunContext memory, **When** normal Timeline/debug metadata is inspected, **Then** the raw prompt text is not exposed.

### Edge Cases

- Built-in config contains an unknown allowed tool name.
- Built-in config version is unchanged but safe display fields change.
- Built-in config version changes while older runs still reference the previous snapshot.
- A thread references a persona that was deactivated after selection.
- A run references a persona id that belongs to a different user or scope.
- Persona model route points at an unavailable provider/model.
- Persona allowed tools conflict with the existing runtime MVP tool allowlist.
- Run creation occurs before built-in sync has completed.
- The worker loses job ownership after resolving persona but before runtime invocation.
- Persona prompt contains sensitive or instruction-like content that must not be persisted into normal Timeline metadata.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST define a durable Persona data model with name, description, system prompt, model route, allowed tool names, reasoning mode, budget summary, version, source, active status, and safe summary fields.
- **FR-002**: Loomi MUST support idempotent synchronization of built-in persona configuration into durable storage.
- **FR-003**: Built-in persona synchronization MUST preserve version history so old runs can identify the persona version they used.
- **FR-004**: Persona system prompts MUST be available to runtime context only through the resolved persona snapshot and MUST NOT be exposed in normal Timeline/debug summaries.
- **FR-005**: Thread creation or thread settings MUST support selecting a persona, or inheriting the default persona when no explicit selection exists.
- **FR-006**: Run creation MUST support selecting a persona override, or inheriting the thread/default persona when no run override exists.
- **FR-007**: RunContext MUST include the resolved persona snapshot/version for the current run before provider/runtime invocation.
- **FR-008**: The resolved persona snapshot MUST include safe fields needed by runtime behavior: persona id, version, name, description, model route, allowed tool names, reasoning mode, and budget summary.
- **FR-009**: Runtime tool resolution MUST intersect persona allowed tool names with Loomi's existing runtime tool allowlist; persona configuration MUST NOT introduce new executable tool families.
- **FR-010**: Runtime model route selection MUST use the resolved persona model route when no more specific run route overrides it.
- **FR-011**: Missing, inactive, cross-scope, invalid, or unsynced persona references MUST fail safely before provider/runtime invocation when no default persona can be resolved.
- **FR-012**: Timeline/debug and Background task surfaces MUST show only a safe persona summary and MUST NOT include raw persona system prompt text, provider credentials, raw provider payloads, raw tool result payloads, file contents, shell output, or hidden local state.
- **FR-013**: The frontend MUST provide either a minimal persona selector for run/thread creation or a read-only display of the resolved persona; the selected path MUST be enough to browser-smoke a real run using a persona.
- **FR-014**: The feature MUST reuse existing M6/M9 worker/job, RunContext, pipeline stage, SSE/history replay, provider route, and MVP tool boundaries without adding a new queue, Skill marketplace, plugin install flow, MCP, Memory/RAG, Sandbox/Desktop Runtime, multi-agent orchestration, or broad permission framework.
- **FR-015**: Documentation MUST update docs-site architecture, API/event contract or runbook pages, roadmap/current-status, devlog, and Spec Kit references for the Persona/Skill foundation slice.

### Key Entities *(include if feature involves data)*

- **Persona**: Durable behavior configuration visible to users through safe summary fields and linked to one or more immutable or versioned persona snapshots.
- **Persona Version**: A versioned built-in or future user-defined persona body containing system prompt, model route, allowed tool names, reasoning mode, budget summary, and safe display metadata.
- **Persona Selection**: A thread-level or run-level choice that resolves to a concrete persona version at run start.
- **Persona Snapshot**: The exact persona version copied or referenced by a run so the worker can prepare RunContext and old runs remain attributable.
- **RunContext Persona Summary**: Safe runtime metadata that records the current run's persona id, version, name, model route label, allowed tool names/count, reasoning mode, and budget summary without prompt text.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of built-in persona sync tests create or update durable persona records idempotently without duplicate active versions.
- **SC-002**: 100% of targeted run creation tests resolve run persona in priority order: explicit run selection, thread selection, default built-in persona.
- **SC-003**: 100% of worker RunContext tests record the resolved persona snapshot/version before provider/runtime invocation.
- **SC-004**: 100% of Timeline/debug tests show safe persona summary from both live SSE and history replay while asserting raw system prompt text is absent.
- **SC-005**: A browser smoke can create a run with a selected or default persona and verify Timeline/debug persona summary plus RunContext persona version.
- **SC-006**: The requested validation plan, including backend tests, related web runtime/UI tests, web build, docs-site build, and browser smoke, is documented in quickstart/tasks.

## Assumptions

- M9 RunContext/Pipeline foundation is the baseline and already prepares durable context in the worker path.
- The first M10 slice may store built-in persona config in repository files and sync it locally; editing personas in an admin UI is out of scope.
- The first M10 slice should include at least one default built-in persona so run creation has a deterministic fallback.
- Existing provider/model route and runtime tool allowlist remain authoritative boundaries; persona only narrows or selects within existing capabilities.
- Documentation changes are part of implementation done, but this planning pass stops before implementation until the user confirms.
