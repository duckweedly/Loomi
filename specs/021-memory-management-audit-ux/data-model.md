# Data Model: M14 Memory Management Audit UX

## Memory Management Item

- `id`: stable memory entry id
- `safe_summary`: redacted user-readable summary
- `scope`: visible scope metadata for user/workspace/thread where available
- `source`: safe source metadata, including source run/thread/source type where available
- `status`: active approved or tombstoned/deleted projection
- `created_at`, `updated_at`, `deleted_at`: lifecycle timestamps
- `redaction`: marker that indicates whether sensitive data was removed

## Memory Detail

- Includes all Memory Management Item fields
- Adds source run/thread labels/ids when visible to the current user
- Does not include raw memory body, provider trace, tool output, local path, or credentials

## Memory Filter

- `query`: optional search text over safe summary/metadata
- `thread_id`: optional scoped thread filter
- `workspace_id`: optional scoped workspace filter when available
- `source_run_id`: optional scoped run filter when available
- `source_type`: optional source type filter when available
- `include_tombstoned`: optional flag for deleted-state inspection if the backend supports it safely

Supported v1 fields are `q`, `scope_type`, `scope_id`, `source_run_id`, `source_type`, `include_tombstoned`, and `limit`. `source_type` values are `manual`, `thread`, `run`, or `any`.

## Memory Audit Item

- `id`: stable event id
- `event_type`: one of `memory_write_proposed`, `memory_write_approved`, `memory_write_denied`, `memory_deleted`, `memory_snapshot_loaded`
- `occurred_at`: event time
- `scope`: safe user/workspace/thread metadata
- `source`: safe source run/thread/source type metadata
- `memory_entry_id` or `proposal_id`: included only when visible and authorized
- `status`: proposal/entry/snapshot state
- `redaction`: marker that indicates safe projection was applied

## State Rules

- Approved active memories are visible in normal list/search and eligible for detail/delete.
- Tombstoned memories are excluded from active list/search and snapshots immediately.
- Tombstoned/deleted state may be shown only through safe detail/history when authorized.
- Audit items never expose raw memory content, secrets, provider traces, tool output, or local paths.
- Out-of-scope reads/deletes/history filters return denial or empty safe results without confirming existence.
- Terminal-run memory audit items remain readable through history even when normal run-event append APIs reject terminal runs.
- Redaction must classify `/Users/...`, `/home/...`, Windows paths, stdout/stderr dumps, provider traces, Authorization/env-like strings, tokens, and secret-like values as unsafe.
