# Contract: Memory API

Status: planned/design-only for `019-memory-foundation`.

## Shared rules

- All endpoints require the existing Loomi user/session identity.
- Authorization resolves user/workspace/thread scope before data access.
- Out-of-scope requests return a generic not-found or forbidden response without confirming existence.
- Responses include safe summaries and metadata only.
- Memory content, search queries, and proposal text are untrusted data, not instructions.
- Idempotent write/decision/delete requests use an idempotency key where the existing API pattern supports it.

## List memory

```text
GET /v1/memory?scope_type={scope_type}&scope_id={scope_id}&cursor={cursor}&limit={limit}
```

Returns approved, non-deleted memory entries visible to the caller.

**Response fields**:

- `items[].id`
- `items[].title`
- `items[].summary`
- `items[].scope_type`
- `items[].source_thread_id`
- `items[].source_run_id`
- `items[].created_at`
- `items[].updated_at`
- `items[].redaction_applied`
- `next_cursor`

## Search memory

```text
POST /v1/memory/search
```

**Request fields**:

- `query`: optional text query.
- `scope_filters`: optional caller-visible scopes.
- `tags`: optional labels.
- `limit`: bounded count.
- `purpose`: `user_search` or `run_context_debug`.

**Response fields**:

- `items[]`: Memory Search Result.
- `safe_query_summary`: redacted/debug-safe query summary.
- `excluded_count`: safe count of filtered entries.

## Read memory entry

```text
GET /v1/memory/{entry_id}
```

Returns safe entry metadata and redacted content for an approved, visible, non-deleted entry.

## Delete memory entry

```text
DELETE /v1/memory/{entry_id}
```

Tombstones an approved visible entry.

**Response fields**:

- `entry_id`
- `status`: `tombstoned`
- `deleted_at`
- `audit_event_id`

Repeated deletes return the existing tombstone state without exposing deleted content.

## Propose memory write

```text
POST /v1/memory/write-proposals
```

Agent/runtime-facing boundary for proposing memory. This may be internal-only in the first implementation, but the contract defines required safety behavior.

**Request fields**:

- `scope_type`
- `scope_id`
- `source_thread_id`
- `source_run_id`
- `source_event_id`
- `title`
- `content`
- `idempotency_key`

**Response fields**:

- `proposal_id`
- `status`: `pending`, `blocked`, or existing terminal state for duplicate idempotency key.
- `safe_preview`
- `redaction_applied`
- `audit_event_id`

Pending proposals are not searchable and are not included in RunContext.

## Approve memory write

```text
POST /v1/memory/write-proposals/{proposal_id}/approve
```

**Response fields**:

- `proposal_id`
- `status`: `approved`
- `entry_id`
- `audit_event_id`

Approval creates or links exactly one approved Memory Entry.

## Deny memory write

```text
POST /v1/memory/write-proposals/{proposal_id}/deny
```

**Response fields**:

- `proposal_id`
- `status`: `denied`
- `audit_event_id`

Denied proposals never become searchable memory.
