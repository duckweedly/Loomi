# Feature Specification: M13 Memory Foundation

**Feature Branch**: `[019-memory-foundation]`

**Created**: 2026-05-25

**Status**: Implemented

**Input**: User description: "M13 Memory Foundation / 019-memory-foundation. First version PG memory with memory_entries, memory_search, memory_write. RunContext can include a safe memory snapshot. Agent can retrieve historical memories, and writes must be approval-gated or stay inside explicit safety boundaries. User can view/delete memories through minimal API/UI boundaries. Memory distill is design-only for later, not part of the first implementation slice. MemoryProvider abstraction may be planned, but v1 prioritizes PG provider and does not implement OpenViking. Strictly define privacy, safety, deletion, audit, redaction, and user control. Non-goals: vector DB/embedding/RAG system, OpenViking provider, marketplace/plugin, sandbox/browser/activity recorder, multi-agent long-term memory automation, worker/job queue or MCP rewrite."

## Clarifications

### Session 2026-05-25

- Q: Should first-slice writes require explicit approval? -> A: Yes, agent-proposed memory writes are approval-gated unless they are user-initiated deletion/control operations.
- Q: Should v1 use embeddings/vector search? -> A: No, use PostgreSQL structured/text search only; embeddings are out of scope.
- Q: Should deletion be hard delete or soft delete? -> A: User delete creates a tombstone/audit trail and excludes the memory from search/snapshot immediately.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Retrieve safe historical memory in RunContext (Priority: P1)

As a Loomi user, I want the agent to receive a small safe snapshot of relevant approved memory before a run, so that it can use prior context without exposing deleted, unsafe, or unrelated data.

**Why this priority**: Memory is useful only if the current agent run can consume it through the existing RunContext/Pipeline boundary. This is the smallest runnable Memory Foundation slice.

**Independent Test**: Seed approved memories for a user/thread/workspace scope, start a run, and verify the RunContext contains only safe, redacted, non-deleted memory summaries with source metadata and bounded count/size.

**Acceptance Scenarios**:

1. **Given** approved memory entries exist for the current user/thread/workspace scope, **When** RunContext is prepared, **Then** Loomi includes a bounded safe memory snapshot sorted by relevance and recency.
2. **Given** memory entries are deleted, disabled, unsafe, outside scope, or pending approval, **When** RunContext is prepared, **Then** those entries are excluded from the snapshot.
3. **Given** memory content contains secret-looking text, local paths, raw tool output, or provider traces, **When** a snapshot is built, **Then** only redacted safe summaries and metadata are visible in RunContext/debug surfaces.
4. **Given** memory search returns no safe entries, **When** RunContext is prepared, **Then** the run continues with an explicit empty snapshot event rather than failing.

---

### User Story 2 - Approval-gate agent memory writes (Priority: P2)

As a Loomi user supervising an agent, I want proposed memory writes to require approval before becoming reusable context, so that the agent cannot silently store sensitive or low-quality facts.

**Why this priority**: Write safety is the trust boundary. A memory system that can only read static entries is incomplete, but writes must not bypass user control.

**Independent Test**: Run a fixture where the agent proposes a memory write, verify it records a pending approval item and safe preview, then approve or deny it and confirm only approved entries become searchable.

**Acceptance Scenarios**:

1. **Given** an agent proposes a memory write, **When** the proposal is recorded, **Then** Loomi stores a pending `memory_write` request with redacted preview, provenance, scope, and audit metadata, and does not expose it to search or RunContext yet.
2. **Given** the user approves a pending memory write, **When** approval succeeds, **Then** Loomi creates an approved `memory_entry`, records an audit event, and makes it eligible for search/snapshot.
3. **Given** the user denies a pending memory write, **When** denial succeeds, **Then** Loomi records the decision and never exposes that proposal as memory context.
4. **Given** the same proposal or approval request is retried, **When** idempotency keys match, **Then** Loomi does not duplicate memory entries or audit events.

---

### User Story 3 - User can view, search, and delete memory (Priority: P3)

As a Loomi user, I want a minimal memory management surface, so that I can inspect what Loomi remembers, remove entries, and understand why a memory exists.

**Why this priority**: User control is required for privacy and trust, but the first implementation can be a thin API/UI surface after retrieval and write safety are defined.

**Independent Test**: Create several approved memories, list/search them through the minimal API/UI boundary, delete one, and verify it disappears from list/search/RunContext while leaving safe audit evidence.

**Acceptance Scenarios**:

1. **Given** approved memory entries exist, **When** the user opens the memory management surface or calls the list API, **Then** Loomi shows safe title/summary, scope, source run/thread, created time, updated time, and delete action.
2. **Given** the user searches memory, **When** a query is submitted, **Then** Loomi returns only entries visible to that user and scope using PG-backed text/metadata search.
3. **Given** the user deletes a memory entry, **When** deletion succeeds, **Then** Loomi tombstones the entry, records a safe audit event, and excludes it from future search/snapshots immediately.
4. **Given** the user tries to view/delete memory outside their scope, **When** the request is authorized, **Then** Loomi denies it without leaking whether the entry exists.

---

### User Story 4 - Plan future distillation/provider boundaries without implementing them (Priority: P4)

As a Loomi developer, I want the first design to reserve clear seams for MemoryProvider and future distillation, so that PG v1 does not block later providers while still avoiding premature platform complexity.

**Why this priority**: The v1 implementation should not build OpenViking or automated distillation, but the contracts should make their future boundary explicit.

**Independent Test**: Review the plan/data model/contracts and verify PG provider is the only implementation task, while MemoryProvider and distill are documented as design-only future boundaries.

**Acceptance Scenarios**:

1. **Given** the Memory Foundation plan is reviewed, **When** provider scope is inspected, **Then** it identifies PG as the only v1 provider and labels OpenViking as deferred.
2. **Given** distillation is reviewed, **When** implementation tasks are inspected, **Then** there are no tasks for automated summarization, scheduled distill jobs, or multi-agent long-term memory automation.

### Edge Cases

- Memory entry content includes secrets, tokens, Authorization headers, private paths, file contents, tool output, browser/activity data, or provider traces.
- A memory is deleted while a run is being prepared or while a search request is in flight.
- A user requests deletion twice, deletes an already tombstoned entry, or retries after timeout.
- Agent proposes a duplicate, contradictory, oversized, unsafe, or out-of-scope memory write.
- Memory search query is empty, too long, contains prompt-injection text, or matches only tombstoned/private entries.
- RunContext preparation fails to load memory due to PG timeout or query error.
- Older runs/events exist before memory metadata is available.
- Approval and deletion race for the same proposed or approved memory.
- Audit events themselves contain unsafe raw memory content.
- Future distillation attempts to summarize deleted or denied memory.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST add a PG-backed first version of memory with `memory_entries`, `memory_search`, and `memory_write` boundaries.
- **FR-002**: `memory_entries` MUST represent approved reusable memory with scope, provenance, safe summary/content, status, timestamps, and redaction/audit metadata.
- **FR-003**: `memory_search` MUST return only approved, non-deleted entries visible to the requesting user/workspace/thread scope.
- **FR-004**: RunContext MUST include a bounded safe memory snapshot with redacted content, provenance, and exclusion reasons for empty/error states.
- **FR-005**: Memory snapshot construction MUST exclude deleted, denied, pending, unsafe, out-of-scope, or disabled entries.
- **FR-006**: Agent-proposed `memory_write` MUST be approval-gated before becoming an approved memory entry.
- **FR-007**: User-initiated memory delete/control actions MUST NOT require agent approval, but MUST be scoped, authorized, audited, and immediately effective.
- **FR-008**: Memory write approval/denial MUST be idempotent and MUST NOT create duplicate entries or duplicate audit events on retry.
- **FR-009**: Memory entries, proposals, search responses, RunContext snapshots, UI state, events, and docs examples MUST NOT expose secrets, raw provider traces, raw tool output, local activity/browser data, private paths, credentials, or unsafe long raw text.
- **FR-010**: Deleting a memory MUST tombstone it, record a safe audit event, preserve deletion evidence, and exclude it from search and RunContext immediately.
- **FR-011**: Users MUST be able to list, search, inspect safe metadata for, and delete their own approved memory entries through minimal API/UI boundaries.
- **FR-012**: Unauthorized list/search/read/delete requests MUST deny access without leaking whether an out-of-scope memory exists.
- **FR-013**: Memory-related run events and audit events MUST be replayable with safe metadata for search, snapshot load, write proposed, write approved, write denied, and entry deleted states.
- **FR-014**: PG v1 MUST avoid vector DB, embeddings, RAG orchestration, OpenViking implementation, marketplace/plugin, sandbox/browser/activity recorder, multi-agent long-term memory automation, worker/job queue rewrite, and MCP rewrite.
- **FR-015**: The plan MAY define a MemoryProvider abstraction, but first implementation tasks MUST target PG provider only.
- **FR-016**: Memory distillation MUST be documented as future design-only scope and MUST NOT appear as an implementation task in the first slice.
- **FR-017**: Memory search and snapshot loading MUST treat memory content, search queries, provider output, tool output, and MCP/server text as untrusted data, never instructions.
- **FR-018**: Documentation and tasks MUST include data model, API/event contracts, privacy/safety/deletion/audit/redaction boundaries, docs-site planned status, and validation commands.

### Key Entities *(include if feature involves data)*

- **Memory Entry**: Approved reusable memory visible to a user/workspace/thread scope, with safe content/summary, provenance, status, timestamps, and deletion/audit metadata.
- **Memory Search Request**: Scoped query over approved memory using PG-backed text/metadata search and safe filters.
- **Memory Search Result**: Redacted entry preview safe for RunContext, API, and UI surfaces.
- **Memory Write Proposal**: Agent-proposed memory candidate awaiting approval or denial, with redacted preview, provenance, idempotency key, and scope.
- **Memory Approval Decision**: User decision that approves or denies a proposal and records audit metadata.
- **Memory Snapshot**: Bounded set of safe memory results attached to RunContext for one run.
- **Memory Tombstone**: Deleted entry state that preserves audit evidence and prevents future search/snapshot inclusion.
- **Memory Audit Event**: Safe event describing search, snapshot, proposal, approval, denial, and deletion without raw sensitive content.
- **Memory Provider**: Future abstraction boundary for provider-specific storage/search behavior; v1 implementation is PG only.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of RunContext memory tests confirm snapshots contain only approved, non-deleted, scoped, redacted memory entries.
- **SC-002**: 100% of search authorization tests confirm out-of-scope, pending, denied, deleted, and unsafe entries are excluded without existence leaks.
- **SC-003**: 100% of memory write tests confirm agent proposals remain pending until approved and denied proposals never become searchable.
- **SC-004**: 100% of idempotency tests confirm repeated proposal, approval, denial, and delete requests do not duplicate entries or audit events.
- **SC-005**: 100% of redaction tests confirm secrets, credentials, private paths, raw tool/provider output, browser/activity data, and unsafe raw text are absent from snapshots, API responses, UI replay, and docs examples.
- **SC-006**: Deletion tests confirm tombstoned memories disappear from list/search/RunContext immediately while safe audit evidence remains.
- **SC-007**: Minimal UI/API smoke confirms a user can list, search, inspect safe memory metadata, and delete an entry.
- **SC-008**: Spec Kit analysis finds no implementation tasks for vector DB/embedding/RAG, OpenViking, marketplace/plugin, sandbox/browser/activity recorder, multi-agent long-term memory automation, worker/job queue rewrite, MCP rewrite, or automated distillation.

## Assumptions

- Loomi continues to use existing local identity/thread/run boundaries and current authorization patterns.
- PG is available wherever backend memory tests run; no new vector database or external memory service is introduced.
- Memory write proposals originate from agent/runtime events, but approval uses Loomi's explicit user-control pattern rather than silent automatic persistence.
- First-slice memory content is short, explicit, and user-reviewable; complex distillation and ranking are deferred.
- Existing RunContext/Pipeline, run events, Timeline/debug, and docs-site patterns remain the integration points.
- Memory content is always untrusted data, even if previously approved.
