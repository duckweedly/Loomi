# Data Model: Memory Agent Tools

## Tool Arguments

- `memory.search`: `query`, `limit`, optional `scope_type`, `scope_id`.
- `memory.read`: `entry_id`, optional `scope_type`, `scope_id`, `source_thread_id`, `source_run_id`.
- `memory.write`: `title`, `content`, optional `scope_type`, `scope_id`, source ids, `idempotency_key`.
- `memory.forget`: `entry_id`, optional `reason`, scope/source context.
- `memory.status`: no required arguments.

## Result Summaries

All results are bounded safe summaries:

- Search: count and safe item summaries.
- Read: safe memory summary and metadata only.
- Write: proposal id, status, scope, safety state.
- Forget: entry id, tombstone status, deleted timestamp.
- Status: provider, state, configured, diagnostic code/message.
