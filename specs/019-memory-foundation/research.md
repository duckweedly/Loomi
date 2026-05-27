# Research: M13 Memory Foundation

## Decision: Use PG-first memory storage and search

**Rationale**: The requested first version explicitly names `memory_entries`, `memory_search`, and `memory_write`, while also excluding vector DB, embeddings, and a RAG system. PostgreSQL fits Loomi's current staged architecture, migration/testing practices, and local development loop.

**Alternatives considered**:

- Vector DB or embedding search: rejected because it is an explicit non-goal for this slice.
- External memory service: rejected because it adds provider/security complexity before the PG slice proves value.
- File-backed memory: rejected because deletion, audit, and scoped query behavior are harder to validate consistently.

## Decision: Attach memory through RunContext as a safe snapshot

**Rationale**: M9 established RunContext as the worker/runtime handoff. Memory should be a bounded, redacted input to that context rather than an ad hoc prompt mutation. This preserves observability and lets Timeline/debug show memory availability without raw sensitive content.

**Alternatives considered**:

- Inject memory directly into provider prompts from the gateway: rejected because it bypasses RunContext/Pipeline visibility.
- Let frontend send memory with run creation: rejected because memory must be durable, scoped, server-authorized, and safe against client tampering.

## Decision: Approval-gate agent-proposed memory writes

**Rationale**: Agent writes can accidentally store secrets, incorrect facts, or sensitive context. Pending proposals with redacted previews give users explicit control and support audit/idempotency.

**Alternatives considered**:

- Auto-save every agent proposal: rejected because it violates user control and privacy boundaries.
- Disable agent writes entirely: rejected because the Memory Foundation acceptance requires the agent to be able to write memory eventually.
- Persona-level blanket permission: rejected because persona scope is not per-memory user consent.

## Decision: Tombstone user deletions

**Rationale**: Deletion must immediately exclude memories from search and RunContext while preserving safe audit evidence that a deletion occurred. Tombstones also make repeated delete requests idempotent.

**Alternatives considered**:

- Hard delete only: rejected because audit and idempotency evidence would be lost.
- Soft delete with raw content retained indefinitely: rejected because deleted content should not remain available to normal product surfaces.

## Decision: Redact before API/UI/RunContext/event surfaces

**Rationale**: Memory content may contain secrets, raw tool output, provider traces, local paths, or browser/activity data. Redaction must happen before content crosses into snapshots, API responses, event replay, documentation examples, or provider input.

**Alternatives considered**:

- Store raw content and redact only in UI: rejected because RunContext and audit replay would still have unsafe source data.
- Rely on the model not to reveal sensitive memory: rejected because memory is untrusted data and must not become an instruction or leakage path.

## Decision: Minimal MemoryProvider boundary, PG implementation only

**Rationale**: A small provider interface prevents RunContext/search/write callers from depending on PG details, but implementing multiple providers now would add complexity not needed for the first slice.

**Alternatives considered**:

- No provider boundary: rejected because future provider work would spread storage assumptions across runtime and API layers.
- Implement OpenViking now: rejected because the user explicitly excluded it from this round.

## Decision: Defer memory distillation

**Rationale**: Automated summarization needs retention rules, source selection, deletion handling, conflict handling, and model-cost safety. First slice should store/retrieve explicit approved memories only.

**Alternatives considered**:

- Background scheduled distillation: rejected because worker/job queue work is out of scope and existing queues should not be redesigned.
- On-run automatic summary generation: rejected because it would blur approval-gated write boundaries.

## Decision: Treat memory content and queries as untrusted data

**Rationale**: A saved memory can contain prompt injection, stale claims, malicious tool text, or user-provided instructions. Search results must be context, not commands.

**Alternatives considered**:

- Trust approved memories as instructions: rejected because approval means reusable data, not policy authority.
- Strip all user-authored detail: rejected because memory would lose value; instead use redacted safe summaries and scope controls.
