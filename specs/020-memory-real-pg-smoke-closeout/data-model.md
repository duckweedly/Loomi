# Data Model: M13.5 Memory Real PG Smoke Closeout

This closeout introduces no new runtime data model. It verifies existing M13 tables and artifacts.

## Existing Runtime Entities Verified

### memory_entries

- Stores approved reusable memory.
- Uses `status = approved | tombstoned | disabled`.
- Tombstoned entries clear content, use `[deleted]` summary, and disappear from search/snapshot immediately.

### memory_write_proposals

- Stores agent-proposed memory writes before reuse.
- Uses `status = pending | approved | denied`.
- Approval links one `created_entry_id`; repeated approval returns the same entry.
- Denial is idempotent and never creates a memory entry.

### run_events

- Stores safe memory audit metadata for proposal, approval, denial, deletion, and snapshot loading.
- Must not store raw sensitive memory content in metadata.

### RunContext.MemorySnapshot

- Contains bounded safe `MemorySearchResult` entries.
- Excludes pending, denied, tombstoned, disabled, blocked, and out-of-scope memory.

## Closeout Artifact Entities

### Real PG Smoke

- Test file: `internal/httpapi/memory_real_pg_smoke_test.go`.
- Requires `LOOMI_TEST_DATABASE_URL`.
- Assumes migrations through `000009_m13_memory_foundation` have already been applied.

### Documentation Evidence

- Devlog entry records exact smoke coverage and validation outcome.
- Runbook records commands and environment variables for rerunning the smoke.
