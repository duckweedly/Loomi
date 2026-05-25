# Research: M14 Memory Management Audit UX

## Decision: Reuse M13 productdata memory/audit state

**Rationale**: M14 needs visibility and control, not a new memory subsystem. Existing memory entries, write proposals, tombstones, run snapshot events, and audit metadata already encode the lifecycle M14 should show.

**Alternatives considered**: A separate UI-only history store was rejected because it would fabricate history and drift from backend truth. A new event bus was rejected as worker/platform scope creep.

## Decision: Add the smallest scoped audit read endpoint if needed

**Rationale**: The UI needs a real safe history. If current endpoints do not expose it, a read-only scoped endpoint can project existing memory events without changing write semantics.

**Alternatives considered**: Packing history into every memory list response was rejected because list and audit have different pagination/filter concerns. Showing only local UI session events was rejected as fake history.

## Decision: Filter only on grounded fields

**Rationale**: The user requested thread/workspace/source_run/source_type from fields that are actually landed. M14 should expose the smallest available combination rather than visual controls with no backend meaning.

**Alternatives considered**: A broad filter drawer with inactive controls was rejected because it creates a misleading management surface.

## Decision: Deletion remains tombstone plus confirmation

**Rationale**: M13 established tombstone delete for safety and auditability. M14 adds UI confirmation so accidental clicks do not immediately delete user memory.

**Alternatives considered**: Hard delete was rejected because it would remove audit evidence. Single-click delete was rejected because it is unsafe for a trust boundary.

## Decision: Safe projections only

**Rationale**: Memory management and audit surfaces are user-readable, but must not leak raw memory, secrets, provider traces, tool output, or local paths.

**Alternatives considered**: Returning raw stored content and relying on UI truncation was rejected because redaction must be enforced at the API boundary.

## Decision: Treat review findings as M14 blockers

**Rationale**: A management UI cannot be trusted if thread-scoped entries cannot be read/deleted, terminal-run memory audit disappears, redaction misses common local-path/provider-output forms, or list/search filters differ by endpoint.

**Alternatives considered**: Deferring these to a later hardening pass was rejected because they directly affect M14's visible contract and browser smoke.
