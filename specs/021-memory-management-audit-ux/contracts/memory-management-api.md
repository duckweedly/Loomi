# Contract: M14 Memory Management API

## List/Search Memory

`GET /v1/memory/entries`

Query:

- `q`: optional search query over safe title/summary/metadata
- `scope_type`: optional `user` or `thread`
- `scope_id`: required when `scope_type=thread`
- `source_thread_id`: implemented source thread filter
- `source_run_id`: optional visible run filter
- `source_type`: optional `manual`, `thread`, `run`, or `any`
- `include_tombstoned`: optional safe deleted-state flag
- `limit`: optional bounded page size
- `workspace_id`: deferred until workspace-scoped memory exists

Response:

- `items`: safe memory management items only
- `filters`: applied grounded filters using the same shape as search
- `next_cursor`: optional cursor when paging exists

Security:

- Responses include only current-user scoped safe summaries and metadata.
- Out-of-scope filters return empty or unauthorized responses without existence leaks.

## Memory Detail

`GET /v1/memory/entries/{memory_id}`

Query:

- `scope_type`, `scope_id`, `source_thread_id`, or `source_run_id`: implemented authorization context for thread-scoped entries

Response:

- Safe summary, scope, source run/thread/source type metadata, created/updated/deleted state, and redaction markers.

Security:

- No raw memory body, secret, provider trace, tool output, local path, or credential is returned.
- Out-of-scope ids do not reveal whether the memory exists.
- Thread-scoped memories require matching thread/source context; wrong or missing context returns generic not found.

## Delete Memory

`DELETE /v1/memory/entries/{memory_id}`

Request:

- `reason`: optional redacted user reason
- `scope_type`, `scope_id`, `source_thread_id`, or `source_run_id`: implemented authorization context for thread-scoped entries

Behavior:

- Requires explicit UI confirmation before the request is sent.
- Tombstones the memory, records safe audit metadata, and excludes the memory from active list/search/snapshot.
- Duplicate delete remains idempotent.

## Memory Audit History

`GET /v1/memory/audit`

Query:

- `thread_id`: optional visible thread filter
- `source_run_id`: optional visible run filter when available
- `event_type`: optional memory event type filter
- `limit`: implemented bounded page size
- `workspace_id`: deferred until workspace-scoped memory exists

Response:

- `items`: safe memory audit items for `memory_write_proposed`, `memory_write_approved`, `memory_write_denied`, `memory_deleted`, and `memory_snapshot_loaded`
- `next_cursor`: optional cursor when paging exists

Security:

- History is backed by durable productdata `memory_audit_events` rows. Run timeline events may also exist, but are not the only audit store.
- It never exposes raw memory, secrets, provider traces, tool output, local paths, or credentials.
- Terminal-run memory audit is retained and still readable by this endpoint.

## Forbidden Fields

The API must never return raw `content`, full provider trace, stdout/stderr payload, tool output, local file paths, `.env` values, `Authorization`, credentials, API keys, tokens, or secret-like values in list, detail, or audit responses.

## Out-of-Scope Behavior

Out-of-scope ids and filters return generic not found/empty scoped responses. They must not reveal whether another user's memory, proposal, run, or thread exists.
