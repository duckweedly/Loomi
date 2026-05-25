---
title: M13 Memory Foundation API and Events
description: Memory entry, search, write proposal, approval, delete, and run-event contracts.
---

M13 exposes the first memory API under `/v1/memory`. All responses use safe summaries and omit raw memory content from list/search surfaces.

## Entry APIs

| Method | Path | Meaning |
| --- | --- | --- |
| `GET` | `/v1/memory` | List approved, safe memory summaries. |
| `POST` | `/v1/memory/search` | Search approved, safe memory summaries by text. |
| `POST` | `/v1/memory` | Create an approved memory entry inside the explicit safety boundary. |
| `GET` | `/v1/memory/{entry_id}` | Read one approved, safe memory entry. |
| `DELETE` | `/v1/memory/{entry_id}` | Tombstone a memory entry and remove it from future search/snapshots. |

Search request:

```json
{
  "query": "postgres memory",
  "scope_type": "thread",
  "scope_id": "thr_123",
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
  "source_thread_id": "thr_123",
  "source_run_id": "run_123",
  "rank_reason": "text_match",
  "redaction_applied": true,
  "updated_at": "2026-05-25T00:00:00Z"
}
```

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

## Events

Run events are emitted only when the memory operation references a source run and the run can accept events.

| Event type | Meaning |
| --- | --- |
| `memory_snapshot_loaded` | Worker loaded a safe memory snapshot into `RunContext`. |
| `memory_write_proposed` | A memory write proposal was created. |
| `memory_write_approved` | A proposal was approved and an entry was created. |
| `memory_write_denied` | A proposal was denied. |
| `memory_entry_deleted` | A memory entry was tombstoned. |

Event metadata is restricted to ids, counts, scope, status, safety state, limits, and redaction flags. It must not include raw memory content or external tool/provider payloads.
