# Contract: Memory Events and Audit

Status: planned/design-only for `019-memory-foundation`.

## Event principles

- Events are safe for persistence, SSE/history replay, Timeline/debug, docs examples, and audit views.
- Events must not include raw secret-looking content, credentials, Authorization headers, private paths, file contents, raw provider traces, raw tool output, browser/activity recorder data, or raw deleted memory content.
- Event metadata may include ids, counts, scope labels, redaction flags, status, and reason codes.

## RunContext snapshot events

### `memory.snapshot.loaded`

Emitted when RunContext attaches memory.

**Safe metadata**:

- `run_id`
- `thread_id`
- `status`: `loaded`, `empty`, `partial`, `unavailable`
- `entry_count`
- `limit`
- `redaction_applied`
- `exclusion_counts`

### `memory.snapshot.unavailable`

Emitted when PG search/snapshot load fails safely.

**Safe metadata**:

- `run_id`
- `thread_id`
- `error_code`
- `entry_count`: `0`

## Search events

### `memory.search.performed`

Optional audit/debug event for user-visible or RunContext searches.

**Safe metadata**:

- `actor_user_id`
- `purpose`
- `result_count`
- `scope_count`
- `redaction_applied`

## Write proposal events

### `memory.write.proposed`

Emitted when an agent proposes memory.

**Safe metadata**:

- `proposal_id`
- `source_run_id`
- `source_thread_id`
- `scope_type`
- `status`: `pending` or `blocked`
- `redaction_applied`
- `blocked_reason`

### `memory.write.approved`

Emitted when a user approves a proposal.

**Safe metadata**:

- `proposal_id`
- `entry_id`
- `actor_user_id`
- `status`: `approved`

### `memory.write.denied`

Emitted when a user denies a proposal.

**Safe metadata**:

- `proposal_id`
- `actor_user_id`
- `status`: `denied`
- `reason_code`

## Delete events

### `memory.entry.deleted`

Emitted when a user tombstones an entry.

**Safe metadata**:

- `entry_id`
- `actor_user_id`
- `status`: `tombstoned`
- `deleted_at`
- `reason_code`

## Redaction events

### `memory.redaction.applied`

Optional event or audit metadata when content was altered before becoming visible.

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
