# Data Model: M13 Memory Foundation

## Memory Entry

Approved reusable memory eligible for search and RunContext snapshots.

**Fields**:

- `id`: stable memory identifier.
- `scope_type`: user, workspace, thread, or future narrower scope.
- `scope_id`: owner scope identifier.
- `created_by_user_id`: user who approved or created the entry.
- `source_thread_id`: optional thread provenance.
- `source_run_id`: optional run provenance.
- `source_event_id`: optional event/write proposal provenance.
- `title`: short safe label.
- `summary`: redacted safe summary used by API/UI/RunContext.
- `content`: approved memory content after redaction/safety checks; not raw tool/provider output.
- `tags`: optional simple labels for filtering.
- `status`: `approved`, `tombstoned`, `disabled`.
- `safety_state`: `safe`, `redacted`, `blocked`.
- `content_hash`: duplicate/idempotency helper.
- `created_at`, `updated_at`, `deleted_at`: lifecycle timestamps.
- `deleted_by_user_id`: set for tombstoned entries.
- `delete_reason`: optional safe reason code.

**Rules**:

- Only `approved` entries may appear in search or RunContext.
- `tombstoned` entries are excluded immediately.
- Unsafe raw data is never returned through normal API/UI/RunContext surfaces.
- Duplicate detection uses scope plus normalized content hash, not raw content exposure.

## Memory Search Request

Scoped request to find approved memories.

**Fields**:

- `requesting_user_id`: caller identity.
- `scope_filters`: allowed scopes resolved by authorization.
- `query`: optional user/search text treated as data, not instructions.
- `tags`: optional filters.
- `limit`: bounded result count.
- `cursor`: optional pagination cursor.
- `purpose`: `run_context`, `user_list`, `user_search`, or `debug`.

**Rules**:

- Authorization resolves scopes before search.
- Empty query may return recent approved entries for user-visible list flows, but RunContext still applies bounded selection.
- Search never returns pending, denied, tombstoned, disabled, or out-of-scope entries.

## Memory Search Result

Safe preview returned to RunContext/API/UI.

**Fields**:

- `entry_id`
- `title`
- `summary`
- `scope_type`
- `source_thread_id`
- `source_run_id`
- `created_at`
- `updated_at`
- `rank_reason`: safe reason such as `recent`, `text_match`, or `thread_scope`.
- `redaction_applied`: boolean.

**Rules**:

- Result payload omits raw provenance payloads, raw provider traces, raw tool output, credentials, private paths, and deleted content.

## Memory Write Proposal

Agent-proposed memory candidate awaiting user decision.

**Fields**:

- `id`: proposal identifier.
- `scope_type`, `scope_id`: proposed visibility.
- `source_thread_id`, `source_run_id`, `source_event_id`: provenance.
- `proposed_title`: safe title preview.
- `proposed_summary`: redacted preview.
- `proposed_content`: candidate content after initial redaction.
- `status`: `pending`, `approved`, `denied`, `expired`, `superseded`.
- `idempotency_key`: prevents duplicate proposals.
- `safety_state`: `safe`, `redacted`, `blocked`.
- `created_at`, `decided_at`
- `decided_by_user_id`
- `decision_reason`: safe reason code.

**Rules**:

- Pending proposals are not searchable and are not included in RunContext.
- Approval creates or links exactly one Memory Entry.
- Denial prevents future eligibility unless a new distinct proposal is made.

## Memory Approval Decision

User action on a write proposal.

**Fields**:

- `proposal_id`
- `decision`: `approve` or `deny`
- `requesting_user_id`
- `idempotency_key`
- `created_entry_id`: set only on approval
- `audit_event_id`
- `decided_at`

**Rules**:

- Repeating the same idempotency key returns the same outcome.
- Approval is allowed only for pending, authorized proposals.
- Denial never creates a Memory Entry.

## Memory Snapshot

Bounded memory set attached to one RunContext.

**Fields**:

- `run_id`
- `thread_id`
- `loaded_at`
- `entries`: ordered Memory Search Results.
- `limit`
- `total_candidates`
- `exclusion_counts`: safe counts by reason such as deleted, pending, unsafe, out_of_scope.
- `load_status`: `loaded`, `empty`, `partial`, `unavailable`.
- `safe_debug_summary`: safe text for Timeline/debug.

**Rules**:

- Snapshot is an immutable view for one run.
- Snapshot never includes raw deleted/unsafe/proposal content.
- PG errors produce `unavailable` or `partial` safe events and should not expose query internals.

## Memory Tombstone

Deleted state for a Memory Entry.

**Fields**:

- `entry_id`
- `deleted_at`
- `deleted_by_user_id`
- `delete_reason`
- `audit_event_id`

**Rules**:

- Tombstoned entries disappear from search/list/RunContext immediately.
- Safe audit evidence remains available for debugging and idempotency.
- Normal product APIs do not return tombstoned content.

## Memory Audit Event

Safe event for memory operations.

**Fields**:

- `event_id`
- `event_type`: `memory.search`, `memory.snapshot.loaded`, `memory.write.proposed`, `memory.write.approved`, `memory.write.denied`, `memory.entry.deleted`, `memory.redaction.applied`, `memory.error`
- `actor_user_id`
- `entry_id` or `proposal_id`
- `run_id`, `thread_id` where applicable.
- `safe_metadata`: counts, statuses, redaction flags, reason codes.
- `created_at`

**Rules**:

- Audit events must not contain raw memory content, raw query text if unsafe, provider traces, tool output, credentials, private paths, or browser/activity data.

## Memory Provider Boundary

Future provider abstraction; v1 implementation is PG.

**Methods**:

- `SearchMemory(ctx, request) -> results`
- `CreateWriteProposal(ctx, proposal) -> proposal`
- `ApproveWrite(ctx, decision) -> entry`
- `DenyWrite(ctx, decision) -> proposal`
- `DeleteMemory(ctx, entry_id, actor) -> tombstone`
- `BuildSnapshot(ctx, run_context_seed) -> snapshot`

**Rules**:

- Provider outputs must be safe for callers by contract.
- Non-PG providers are design-only and not part of v1 tasks.
