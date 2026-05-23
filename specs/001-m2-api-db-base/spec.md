# Feature Specification: M2 API and Database Base

**Feature Branch**: `[001-m2-api-db-base]`

**Created**: 2026-05-23

**Status**: Draft

**Input**: User description: "实现 M2 具体的细节你还是要清晰一下"

## Clarifications

### Session 2026-05-23

- Q: When persistence is unavailable, should the M2 service fail startup or start with readiness reporting not ready? → A: Service starts; readiness is not ready.
- Q: Should the M2 initial schema baseline include future core business tables? → A: Only schema version baseline; no business tables.
- Q: Should M2 persistence support only local development or deployment-ready environments? → A: Local development only.
- Q: What minimum observability level should M2 provide? → A: Structured logs with request id.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Verify the service is alive and ready (Priority: P1)

As a developer preparing the next Loomi milestone, I need a minimal service boundary that can clearly report whether the process is alive and whether its required persistence dependency is usable, so later milestones can build on a trustworthy base instead of hidden mock assumptions.

**Why this priority**: M2 exists to move Loomi from a front-end-only mock shell into a real service boundary. Liveness and readiness checks are the smallest visible proof that the boundary is runnable and observable.

**Independent Test**: Can be fully tested by starting the service in a prepared local environment, checking liveness, checking readiness with persistence available, and checking readiness again when persistence is unavailable.

**Acceptance Scenarios**:

1. **Given** the service process has started, **When** a developer checks whether the process is alive, **Then** the system reports an alive state without requiring the persistence dependency to be available.
2. **Given** the service process has started and required persistence is reachable, **When** a developer checks readiness, **Then** the system reports ready and includes a clear non-secret status summary.
3. **Given** the service process has started but required persistence is unavailable, **When** a developer checks readiness, **Then** the system reports not ready with a non-secret reason while the alive check still confirms the process can respond.

---

### User Story 2 - Configure and start the M2 service predictably (Priority: P1)

As a developer, I need the M2 service to read its required runtime settings, reject invalid startup conditions, and document the expected local values, so every later milestone can run from the same baseline.

**Why this priority**: A service that starts only on one machine or silently falls back to unsafe defaults will make later API, database, event, and worker work unreliable.

**Independent Test**: Can be fully tested by starting the service with valid settings, starting it with a required setting missing or invalid, and confirming the failure is clear and does not expose secrets.

**Acceptance Scenarios**:

1. **Given** all required local settings are present and valid, **When** a developer starts the M2 service, **Then** the service starts and exposes its current environment and status through safe diagnostics.
2. **Given** a required setting is missing or invalid, **When** a developer starts the M2 service, **Then** startup fails fast with a clear, actionable, non-secret error.
3. **Given** settings include sensitive values, **When** startup or readiness fails, **Then** error messages and logs do not reveal the sensitive value.

---

### User Story 3 - Establish reversible persistent schema management (Priority: P2)

As a maintainer, I need the persistent store to have a repeatable initial schema workflow with both apply and rollback verification, so future user, thread, message, run, and event models can be added without ad-hoc database changes.

**Why this priority**: M2 should not create the later data models yet, but it must prove that Loomi can safely evolve persistent state before those models arrive.

**Independent Test**: Can be fully tested by applying the initial schema workflow to an empty local store, confirming the schema state, rolling it back, and applying it again without manual repair.

**Acceptance Scenarios**:

1. **Given** an empty local persistent store, **When** a maintainer applies the M2 schema workflow, **Then** the system records the schema version baseline without creating business tables, and readiness can confirm the store is usable.
2. **Given** the initial schema workflow has been applied, **When** a maintainer rolls it back, **Then** the system returns to a clean pre-M2 schema state.
3. **Given** the schema workflow was applied, rolled back, and applied again, **When** a maintainer runs the documented verification, **Then** the workflow completes without manual cleanup.

---

### User Story 4 - Preserve the existing M1 shell while adding the service base (Priority: P3)

As a Loomi product developer, I need the existing desktop-feeling mock UI shell to keep working while M2 is added, so the project can gain a real service boundary without regressing the visible M1 demonstration.

**Why this priority**: M1 is the current visible proof of the product direction. M2 should prepare replacement boundaries for real data without forcing the UI to become dependent on incomplete backend features.

**Independent Test**: Can be tested by running the existing UI shell after M2 changes and confirming the primary mock conversation, timeline, and right-panel states still work.

**Acceptance Scenarios**:

1. **Given** the M2 service base has been added, **When** a developer opens the existing UI shell without using real backend data, **Then** the mock thread, chat, timeline, and tool/debug panels remain usable.
2. **Given** later milestones need to replace mock data with real service calls, **When** a developer inspects the M2 scope, **Then** it is clear which service checks exist now and which user-facing data features remain future work.

### Edge Cases

- Persistence is unavailable at startup: the service must still start, liveness must remain available, and readiness must report not ready with an explicit non-secret reason.
- Persistence becomes unavailable after startup: readiness must change to not ready without falsely reporting that dependent behavior is safe.
- Schema workflow is run more than once: repeat runs must either be safe or fail with an actionable explanation that does not require manual database guessing.
- Rollback is attempted after the initial schema workflow: the workflow must document what will be removed and verify the post-rollback state.
- Configuration contains malformed values: the system must reject them before reporting ready.
- Errors include connection details or credentials: messages and logs must redact sensitive values.
- Existing UI shell is run without the M2 service: M1 mock behavior must remain available until later milestones intentionally switch specific UI paths to real data.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a liveness check that confirms the service process can respond independently of downstream persistence availability.
- **FR-002**: The system MUST start even when persistence is unavailable and provide a readiness check that reports ready only when all M2-required dependencies are available and usable.
- **FR-003**: Readiness failures MUST include an actionable, non-secret reason that distinguishes unavailable persistence, invalid configuration, and missing schema state.
- **FR-004**: The system MUST load required runtime configuration for local M2 operation and reject missing, malformed, or unsafe values before reporting ready.
- **FR-005**: Startup, liveness, readiness, configuration, and schema workflow failures MUST emit structured diagnostic output with a request or operation identifier, while avoiding secrets or full sensitive connection values.
- **FR-006**: The system MUST provide a repeatable schema apply workflow for the initial persistent store baseline that records schema version state without creating future business tables.
- **FR-007**: The system MUST provide a rollback workflow for the initial persistent store baseline and document the expected post-rollback state.
- **FR-008**: The system MUST include a smoke verification path that demonstrates startup, liveness, readiness, schema apply, schema rollback, and failure visibility.
- **FR-009**: The system MUST keep M2 bounded to service foundation and persistence foundation only; authentication, users, threads, messages, runs, events, workers, model calls, tools, and desktop runtime are out of scope.
- **FR-010**: The existing M1 mock UI shell MUST remain runnable and usable without requiring the M2 service to provide real product data.
- **FR-011**: M2 documentation MUST describe local startup, required settings, health/readiness behavior, schema workflow, smoke verification, known limitations, and the exact features deferred to later milestones.
- **FR-012**: The feature MUST leave a clear replacement boundary for later milestones to connect real user, thread, message, run, and event data without rewriting the existing UI shell concept.

### Key Entities

- **Service Status**: Represents whether the M2 service process is alive and able to answer basic checks.
- **Readiness Status**: Represents whether required M2 dependencies and schema state are usable for dependent behavior.
- **Runtime Configuration**: Represents the environment, network, logging, and persistence settings required to start the service safely.
- **Persistent Store Baseline**: Represents the initial durable storage state that future milestones will extend; in M2 it contains schema version state only, not business records.
- **Schema Revision**: Represents the current applied schema state and supports verification, apply, and rollback workflows.
- **Smoke Verification Result**: Represents the observed outcome of the documented M2 validation steps.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can follow the documented local setup and confirm service liveness and readiness in under 10 minutes on a prepared machine.
- **SC-002**: When persistence is unavailable, readiness reports a not-ready state with a non-secret reason within 10 seconds, while liveness remains separately verifiable.
- **SC-003**: The initial schema workflow can be applied, rolled back, and applied again in a local environment without manual cleanup.
- **SC-004**: 100% of documented M2 smoke checks include expected outcomes and can be repeated by another developer.
- **SC-007**: Startup, liveness, readiness, and schema workflow checks produce structured diagnostic records that include a request or operation identifier.
- **SC-005**: The existing M1 mock UI primary flow remains usable after M2 work, including thread list, chat canvas, composer, timeline, and right-side debug/tool panels.
- **SC-006**: A new contributor can read the M2 documentation and identify at least five explicitly deferred capabilities within 10 minutes.

## Assumptions

- M1 is treated as an existing mock/local UI shell that must remain available while M2 adds the first real service boundary.
- M2 persistence support is limited to local development and validation, not production deployment, multi-user access, hosted operations, or release packaging.
- The exact implementation stack, storage driver, migration tooling, directory layout, and endpoint paths belong in the technical plan, not this stakeholder-facing specification.
- Later milestones will introduce authentication, user/thread/message persistence, run events, streaming, worker execution, model gateway, tools, and desktop runtime after the M2 foundation is verified.
- Documentation updates are part of completion because Loomi requires non-trivial development to update the docs site in the same work session.