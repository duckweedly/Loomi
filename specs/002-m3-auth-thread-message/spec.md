# Feature Specification: M3 Auth, Thread, and Message

**Feature Branch**: `[002-m3-auth-thread-message]`

**Created**: 2026-05-23

**Status**: Draft

**Input**: User description: "Clarify M3 scope: build the minimal product data layer on top of M2 with local identity, users, threads, messages, real/mock frontend API switching, structured diagnostics, migration/readiness/rollback contracts, and documented future boundaries; do not pull in run/event/SSE, LLM, tools, worker, or desktop runtime."

## Clarifications

### Session 2026-05-23

- Q: Which local identity strategy should M3 use? → A: Fixed local dev user, no session table.
- Q: Should sending a user message automatically create an assistant placeholder in M3? → A: No; persist only the user message.
- Q: Should M3 support message idempotency with client message identifiers? → A: Yes; optional client_message_id returns the existing message on duplicate.
- Q: If a real API base is configured but unavailable, should the web shell fall back to mock data? → A: No; show a recoverable error.
- Q: Should M3 provide seed/demo data? → A: Yes; provide an explicit seed command or script, with no demo data in migrations.
- Q: How should M3 documentation and code comments be written for learning? → A: Documentation should be detailed for personal learning, and key code boundaries should include concise Chinese WHY comments.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Use a local identity to own product data (Priority: P1)

As a local Loomi developer, I need every real thread and message to belong to a fixed local development user, so M3 can introduce durable product data without sessions or anonymous records that would need ownership retrofits later.

**Why this priority**: Thread and message data need an owner before the system can safely support listing, editing, and future authorization boundaries. M3 should establish identity ownership without building a production account system.

**Independent Test**: Can be tested by starting the local API, requesting the current identity, creating a thread, creating a message, and confirming both records are associated with the same local user.

**Acceptance Scenarios**:

1. **Given** the M3 service is running with the local database migrated, **When** a developer asks for the current identity, **Then** the system returns a stable local user with a display name and creation metadata.
2. **Given** a local user exists, **When** the developer creates a thread, **Then** the thread is owned by that user.
3. **Given** a thread owned by the local user exists, **When** the developer adds a message to that thread, **Then** the message is linked to the thread and the same owner context.

---

### User Story 2 - Manage real threads without coupling them to runs (Priority: P1)

As a Loomi user using the local web shell, I need to create, view, list, rename, and archive real threads from persistent storage, so the M1 shell can begin replacing mock conversations with durable conversation containers.

**Why this priority**: Threads are the first real product object users interact with. They must be durable and independently useful before run/event/SSE behavior exists.

**Independent Test**: Can be tested by creating a thread, listing threads, opening the created thread, renaming it, archiving it, and confirming refresh preserves the durable state.

**Acceptance Scenarios**:

1. **Given** no real thread exists for the local user, **When** the user creates a thread with a title and mode, **Then** the thread appears in the thread list after refresh.
2. **Given** multiple real threads exist, **When** the user lists threads, **Then** they appear in most-recently-updated order.
3. **Given** a real thread exists, **When** the user updates its title or mode, **Then** the change is persisted and visible after refresh.
4. **Given** a real thread exists, **When** the user archives it, **Then** it no longer appears in the default active thread list but remains recoverable by durable state.
5. **Given** a thread exists, **When** the user inspects its status, **Then** the status describes thread lifecycle only and does not imply any run state.

---

### User Story 3 - Persist and reload real messages (Priority: P1)

As a Loomi user, I need messages in a thread to be stored and reloaded from durable state, so a local conversation survives refresh and no longer depends only on mock data.

**Why this priority**: Messages are the core content of Loomi's product data layer. M3 should answer who said what in which thread, while leaving agent execution for M4 and later.

**Independent Test**: Can be tested by opening a thread, adding a user message, refreshing the page, and confirming the message still appears in stable order.

**Acceptance Scenarios**:

1. **Given** a real thread exists, **When** the user opens it, **Then** messages for that thread load in stable creation order.
2. **Given** a real thread is open, **When** the user sends a user message, **Then** the message is saved as a complete text message and the thread's updated time changes.
3. **Given** a user message was saved, **When** the user refreshes the web shell, **Then** the message still appears in the same thread.
4. **Given** the system saves a message, **When** the message is inspected, **Then** it contains final text content only and no streaming delta, tool call, run event, or model output semantics.

---

### User Story 4 - Switch the web shell between real API and mock fallback (Priority: P2)

As a Loomi developer, I need the web shell to use real thread/message APIs when configured and keep mock behavior when the backend is not configured, so M3 improves the product path without breaking the M1 demonstration shell.

**Why this priority**: M1 remains the visible shell and must keep working without a backend. M3 should connect only the thread/message path to real data and keep run/timeline behavior explicitly out of scope.

**Independent Test**: Can be tested by running the web shell without an API base configuration and confirming mock data still appears, then running with a real API base configuration and confirming threads/messages come from the backend.

**Acceptance Scenarios**:

1. **Given** no real API base is configured, **When** the developer opens the web shell, **Then** the existing mock thread, message, and run timeline demonstration remains usable.
2. **Given** a real API base is configured and the M3 API is ready, **When** the developer opens the web shell, **Then** thread list and messages are loaded from the real API.
3. **Given** the web shell is using the real API, **When** the user sends a message, **Then** the message is persisted through the backend and appears after refresh.
4. **Given** the web shell is using the real API, **When** the user opens the run timeline or debug rail, **Then** those surfaces remain mock, empty, or explicitly deferred rather than pretending real run/event data exists.

---

### User Story 5 - Carry M2 operational contracts forward (Priority: P2)

As a maintainer, I need M3 to preserve and extend M2's hard contracts for structured diagnostics, migrations, readiness, rollback, and documented future boundaries, so each milestone remains verifiable and does not blur into later platform capabilities.

**Why this priority**: M3 introduces real business tables and API behavior. It must not weaken the operational guarantees established in M2.

**Independent Test**: Can be tested by applying the M3 migration, verifying readiness fails before the M3 schema exists and passes after migration, rolling the migration back and reapplying it, and checking documentation lists deferred run/tool/event/catalog capabilities.

**Acceptance Scenarios**:

1. **Given** the database is only at the M2 schema baseline, **When** readiness is checked for M3, **Then** the system reports not ready because the M3 schema is absent.
2. **Given** the M3 migration is applied, **When** readiness is checked, **Then** the system reports ready when dependencies and schema state are usable.
3. **Given** the M3 migration was applied, **When** it is rolled back and reapplied, **Then** the schema workflow completes without manual cleanup.
4. **Given** an M3 API error occurs, **When** the client receives the error, **Then** the response includes a stable error code, human-readable message, and request identifier without leaking secrets.
5. **Given** the M3 documentation is reviewed, **When** a contributor looks for deferred capabilities, **Then** run/event/SSE, LLM gateway, tools, worker, desktop runtime, attachments, RAG, and catalog-style extension boundaries are explicitly marked for later milestones.

### Edge Cases

- Backend is not configured in the web shell: mock fallback must remain available.
- Backend is configured but not ready: the web shell must show a recoverable error and must not automatically fall back to mock data.
- A user tries to access a thread that does not belong to the current identity: the system must reject access without exposing whether other users' records exist.
- A thread is archived: default thread lists must exclude it while preserving durable state for future recovery behavior.
- A message is sent twice due to double-click or retry: when the same client message identifier is provided, the system must return the existing message instead of creating a duplicate.
- A message is empty or whitespace only: the system must reject it with a clear error.
- A thread title is empty or too long: the system must reject or normalize it according to a documented rule.
- The database is migrated only through M2: M3 readiness must not incorrectly report ready.
- The API returns timestamps: persisted values must be machine-readable timestamps, while display formatting belongs to the web shell.
- The frontend still has mock run data: M3 must not present mock run/event data as real backend execution.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide one fixed local development user for M3 requests and expose that identity to the web shell without requiring a session table or user-selecting request header.
- **FR-002**: The system MUST persist user-owned threads with title, mode, lifecycle status, creation time, update time, and archive state.
- **FR-003**: The system MUST support creating, listing, retrieving, updating, and archiving threads for the current identity.
- **FR-004**: Thread listing MUST default to active threads ordered by most recently updated first.
- **FR-005**: Thread lifecycle status MUST be separate from run status and MUST NOT imply agent execution state.
- **FR-006**: The system MUST persist messages within a thread with role, complete text content, creation time, and optional metadata.
- **FR-007**: The system MUST support listing messages for a thread in stable creation order.
- **FR-008**: The system MUST support adding user messages to an owned thread and updating the parent thread's updated time in the same user-visible operation.
- **FR-009**: Message persistence MUST NOT create assistant placeholder messages, run events, streaming deltas, tool call records, model outputs, worker jobs, or LLM requests.
- **FR-010**: The web shell MUST use real thread/message APIs when configured, show recoverable errors if the configured API is unavailable, and retain mock thread/message/run behavior only when no real API base is configured.
- **FR-011**: The web shell MUST keep run timeline/debug behavior clearly mock, empty, or deferred while M3 only connects threads and messages to real data.
- **FR-012**: The system MUST return machine-readable timestamps for real API data, with human-readable relative or clock formatting handled in the web shell.
- **FR-013**: The system MUST provide a migration workflow that applies, rolls back, and reapplies M3 schema changes without manual cleanup.
- **FR-014**: M3 readiness MUST require the M3 schema version in addition to M2 dependency readiness.
- **FR-015**: API error responses MUST include a stable error code, a human-readable message, and a request identifier, and MUST NOT expose secrets.
- **FR-016**: The system MUST reject empty messages and invalid thread updates with clear client-facing errors.
- **FR-017**: The system MUST accept an optional client message identifier and return the existing message instead of creating a duplicate when the same identifier is reused for the same thread and local identity.
- **FR-018**: The system MUST provide an explicit seed command or script for local demo data, and migrations MUST NOT insert demo threads or messages.
- **FR-019**: Documentation MUST record M3 scope, API contracts, migration/readiness/rollback behavior, seed behavior, frontend real/mock switching, and deferred future boundaries for run/event/SSE, LLM, tools, worker, desktop runtime, attachments, RAG, and catalog-style extension capabilities in enough detail to support personal learning and later review.
- **FR-020**: Key code boundaries and non-obvious invariants MUST include concise Chinese WHY comments, while routine code and self-explanatory operations should remain uncommented.

### Key Entities

- **Local Identity**: A fixed local development user context used to own threads and messages in M3; not a session system, user-switching mechanism, or production account system.
- **User**: The durable owner record for threads and messages; includes display identity and lifecycle metadata.
- **Thread**: A durable conversation container owned by a user; has title, mode, lifecycle status, timestamps, and archive state independent of run execution.
- **Message**: A durable complete-text entry within a thread; M3 persists user-authored messages only, with role, content, timestamps, optional metadata, and no assistant placeholder, streaming, tool, or run semantics.
- **Client Message Identifier**: An optional client-provided idempotency key that prevents duplicate user message persistence.
- **API Error**: A structured error response containing a code, message, and request identifier.
- **Schema Revision**: The M3 migration state required for readiness and rollback/reapply verification.
- **Frontend Data Source Mode**: The web shell's active choice between real API thread/message data and mock fallback.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can complete local M3 setup, apply migrations, start the API, and load real threads/messages in the web shell in under 15 minutes on a prepared machine.
- **SC-002**: After creating a thread and sending a user message, refreshing the web shell shows the same thread and message from durable state.
- **SC-003**: With the database at only the M2 schema baseline, readiness reports not ready; after applying M3 migration, readiness reports ready when dependencies are usable.
- **SC-004**: The M3 migration can be applied, rolled back, and reapplied locally without manual database cleanup.
- **SC-005**: At least 95% of thread/message API error cases return a structured error with code, message, and request identifier.
- **SC-006**: The web shell still opens with mock data when no real API base is configured.
- **SC-007**: M3 documentation allows a new contributor to identify within 10 minutes which capabilities are implemented now and which are deferred to M4 or later.
- **SC-008**: M3 documentation explains the identity, thread/message, readiness, seed, frontend switching, and deferred-boundary decisions clearly enough that the project owner can use it as a learning reference without re-reading implementation code.

## Assumptions

- M2 API, local PostgreSQL, migration, readiness, structured diagnostics, and documentation patterns are available as the foundation for M3.
- M3 remains local-development focused and does not create production authentication, multi-user organization management, hosted operations, or release packaging.
- A single fixed local development user is sufficient for M3 as long as durable records carry user ownership and all access goes through one identity resolution boundary.
- Threads and messages are the only product data made real in M3; run/event/SSE and execution observability start in M4.
- Attachments, RAG, file uploads, context references, catalog/marketplace concepts, tool calls, and model outputs are deferred and may be documented as future boundaries.
- Existing M1 mock UI behavior should remain available for demonstration and development when the real API is not configured.
- Local demo data should be created only through an explicit seed command or script, not through schema migrations.
- Loomi documentation may be more detailed than product UI copy because it is also used as a personal learning record.
- Chinese code comments are appropriate at key boundaries or non-obvious invariants when they explain why the boundary exists; routine what-the-code-does comments should still be avoided.
