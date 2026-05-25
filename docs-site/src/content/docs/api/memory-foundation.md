---
title: M13 Memory Foundation API and Events
description: Memory entry, search, write proposal, approval, delete, and durable audit contracts.
---

M13 exposes the first memory API under `/v1/memory`. All responses use safe summaries and omit raw memory content from list/search surfaces. M14 prepares the management/audit contract and unifies list/search/audit filters so the Settings > Memory UX can be implemented without fake history.

M14 blocker foundation adds durable `memory_audit_events` and scoped read/delete authorization for thread memories. Local smokes must apply migrations through version `10`; the API process still does not auto-apply migrations.

## Entry APIs

| Method | Path | Meaning |
| --- | --- | --- |
| `GET` | `/v1/memory` | List approved, safe memory summaries. |
| `POST` | `/v1/memory` | Search approved, safe memory summaries by text. |
| `POST` | `/v1/memory/search` | Search approved, safe memory summaries by text. |
| `GET` | `/v1/memory/entries/{entry_id}` | Read one approved or tombstoned safe memory detail. |
| `DELETE` | `/v1/memory/entries/{entry_id}` | Tombstone a memory entry and remove it from future search/snapshots. |
| `GET` | `/v1/memory/audit` | List scoped safe memory audit/history items. |

Search request:

```json
{
  "query": "postgres memory",
  "scope_type": "thread",
  "scope_id": "thr_123",
  "source_thread_id": "thr_123",
  "source_run_id": "run_123",
  "source_type": "run",
  "include_tombstoned": false,
  "limit": 5
}
```

Search/list item:

```json
{
  "id": "mem_abc",
  "title": "Preference",
  "summary": "Prefers compact implementation slices",
  "scope_type": "user",
  "scope_id": "usr_local_dev",
  "status": "approved",
  "safety_state": "safe",
  "source_thread_id": "thr_123",
  "source_run_id": "run_123",
  "source_type": "run",
  "rank_reason": "text_match",
  "redaction_applied": true,
  "updated_at": "2026-05-25T00:00:00Z"
}
```

Empty list/search responses return `items: []`, not `null`, so the Settings > Memory surface can render empty state without special casing.

List/search requests with `scope_type=thread` require a non-empty `scope_id`. Missing `scope_id` returns `invalid_request` so callers do not confuse a malformed thread-scoped request with a genuine empty memory list.

Detail and delete for thread-scoped entries require a matching scope boundary. A request may use `scope_type=thread&scope_id={thread_id}`, `source_thread_id`, or `source_run_id`. User-scoped entries remain visible to the current user. Out-of-scope detail/delete returns generic `memory_not_found` and never echoes the target entry id.

Detail responses use the same safe projection and must not include raw `content`, `content_hash`, provider trace, tool output, local path, `.env`, `Authorization`, credential, token, or secret-like fields.

## Write Proposal APIs

| Method | Path | Meaning |
| --- | --- | --- |
| `POST` | `/v1/memory/write-proposals` | Propose a memory write without making it searchable. |
| `POST` | `/v1/memory/write-proposals/{proposal_id}/approve` | Approve a pending proposal and create an entry. |
| `POST` | `/v1/memory/write-proposals/{proposal_id}/deny` | Deny a pending proposal without creating an entry. |

Proposal request:

```json
{
  "scope_type": "user",
  "title": "Preference",
  "content": "Prefers PostgreSQL-first memory",
  "source_thread_id": "thr_123",
  "source_run_id": "run_123",
  "idempotency_key": "agent-write-1"
}
```

Approve/deny request:

```json
{
  "reason": "user approved",
  "idempotency_key": "decision-1"
}
```

Pending and denied proposals do not appear in memory search. Approval responses include the proposal and the created safe entry summary.

The implemented public create path for agent memory is the write-proposal API. Direct approved-entry creation remains a service/repository boundary, not a public HTTP create endpoint.

## Events

Memory audit is durably stored in `memory_audit_events`. When the source run can accept events, Loomi may also write a related `run_events` row for timeline context, but run events are no longer the only audit store. Memory audit remains available after the source run has completed or failed.

| Event type | Meaning |
| --- | --- |
| `memory_snapshot_loaded` | Worker loaded a safe memory snapshot into `RunContext`. |
| `memory_write_proposed` | A memory write proposal was created. |
| `memory_write_approved` | A proposal was approved and an entry was created. |
| `memory_write_denied` | A proposal was denied. |
| `memory_entry_deleted` | A memory entry was tombstoned. |

Event metadata is restricted to ids, counts, scope, status, safety state, limits, and redaction flags. It must not include raw memory content or external tool/provider payloads.

For user-facing audit history, `memory_entry_deleted` is projected as `memory_deleted`.

Audit query filters accept the same thread/source shape used by list/search: `scope_type=thread&scope_id={thread_id}`, `source_thread_id={thread_id}`, `source_run_id={run_id}`, and `limit`. `thread_id` remains accepted as a direct audit filter. Thread-scoped audit requests must return only matching thread history and must not mix in same-user history from other threads.

Audit item:

```json
{
  "id": "evt_123",
  "event_type": "memory_write_approved",
  "summary": "Memory write approved",
  "thread_id": "thr_123",
  "run_id": "run_123",
  "memory_entry_id": "mem_123",
  "memory_proposal_id": "memprop_123",
  "status": "approved",
  "scope_type": "thread",
  "source_type": "run",
  "redaction_applied": true,
  "occurred_at": "2026-05-25T00:00:00Z"
}
```

Out-of-scope entry/proposal/run/thread ids return generic not found or empty scoped results. They must not reveal whether another user's memory exists.

## Current Filter Shape

Implemented now:

- `query` in JSON search requests and `q` in list query strings.
- `limit`.
- `scope_type` and `scope_id`.
- `source_thread_id`; for audit this is applied as the thread history boundary.
- `source_run_id`.
- `source_type` with `run`, `thread`, `manual`, or `any`.
- `include_tombstoned`.

Deferred:

- `workspace_id` and workspace-scoped memory. Docs and frontend must not send `workspace_id` until a workspace memory scope exists.
- Cursor pagination. Current list/search/audit responses are bounded by `limit`.
