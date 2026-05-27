# Contract: Memory Events and Audit

Status: current implemented contract for `019-memory-foundation`, with optional future event categories noted explicitly.

## Event principles

- Events are safe for persistence, SSE/history replay, Timeline/debug, docs examples, and audit views.
- Events must not include raw secret-looking content, credentials, Authorization headers, private paths, file contents, raw provider traces, raw tool output, browser/activity recorder data, or raw deleted memory content.
- Event metadata may include ids, counts, scope labels, redaction flags, status, and reason codes.

## RunContext snapshot events

### `memory_snapshot_loaded`

Emitted when RunContext attaches memory.

**Safe metadata**:

- `status`: `loaded`, `empty`, `partial`, `unavailable`
- `entry_count`
- `limit`
- `redaction_applied`

### `memory.snapshot.unavailable`

Deferred. The current implementation records unavailable state through `memory_snapshot_loaded` metadata when snapshot loading fails safely.

**Safe metadata**:

- `run_id`
- `thread_id`
- `error_code`
- `entry_count`: `0`

## Search events

### `memory.search.performed`

Deferred optional audit/debug event for user-visible searches.

**Safe metadata**:

- `actor_user_id`
- `purpose`
- `result_count`
- `scope_count`
- `redaction_applied`

## Write proposal events

### `memory_write_proposed`

Emitted when an agent proposes memory.

**Safe metadata**:

- `memory_proposal_id`
- `memory_status`
- `memory_scope_type`
- `memory_safety`
- `source_event_id` when present

### `memory_write_approved`

Emitted when a user approves a proposal.

**Safe metadata**:

- `memory_proposal_id`
- `memory_entry_id`
- `memory_status`
- `memory_scope_type`
- `memory_safety`

### `memory_write_denied`

Emitted when a user denies a proposal.

**Safe metadata**:

- `memory_proposal_id`
- `memory_status`
- `memory_scope_type`
- `memory_safety`

## Delete events

### `memory_entry_deleted`

Emitted when a user tombstones an entry.

**Safe metadata**:

- `memory_entry_id`
- `memory_status`
- `memory_scope_type`
- `memory_safety`

## Redaction events

### `memory.redaction.applied`

Deferred optional event or audit metadata when content was altered before becoming visible.

**Safe metadata**:

- `target_type`: `entry`, `proposal`, `snapshot`, or `search_result`
- `target_id`
- `redaction_categories`: safe labels such as `credential_like`, `private_path`, `tool_output`, `provider_trace`, `oversized`

## Forbidden event fields

Events must not include:

- raw memory content when blocked, deleted, or unsafe
- raw query text if unsafe or secret-looking
- raw provider request/response bodies
- raw tool/MCP/stdout/stderr output
- local file paths that reveal private directories
- browser, desktop, or activity recorder captured state
- tokens, cookies, Authorization headers, API keys, private keys, or env values
