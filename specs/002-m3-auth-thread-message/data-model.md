# Data Model: M3 Auth, Thread, and Message

M3 introduces Loomi's first durable product records. It owns local identity, users, threads, and complete user-authored messages. It does not model runs, events, streaming deltas, tool calls, model outputs, workers, attachments, RAG, desktop runtime, or catalog extensions.

## Local Identity

A fixed local development identity used to scope every M3 product-data request.

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| `user_id` | string | yes | Stable local id, e.g. `user_local_dev`; not supplied by request headers |
| `display_name` | string | yes | Human-readable local developer name |
| `source` | enum | yes | `local_dev` |

Rules:

- The API resolves identity internally; clients do not choose a user.
- Resolving identity must ensure the durable user row exists.
- The identity boundary is not a session system and does not imply production authentication.

## User

A durable owner record for threads and messages.

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| `id` | string | yes | Primary key; fixed local id for M3 |
| `display_name` | string | yes | Non-empty display name |
| `created_at` | timestamp | yes | Machine-readable timestamp |
| `updated_at` | timestamp | yes | Machine-readable timestamp |

Relationships:

- One `User` owns many `Thread` records.
- One `User` owns many `Message` records through owned threads.

Validation:

- The fixed local user may be upserted by API startup flow, `/v1/me`, product-data requests, or the seed command.
- Migrations create the table but do not need to insert demo users, threads, or messages.

## Thread

A durable conversation container owned by one user.

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| `id` | string | yes | Primary key; app-generated text id with `thr_` prefix |
| `user_id` | string | yes | Foreign key to `users.id` |
| `title` | string | yes | Trimmed; 1-120 characters |
| `mode` | enum | yes | `chat` or `work` |
| `lifecycle_status` | enum | yes | `active` or `archived`; separate from run status |
| `created_at` | timestamp | yes | Machine-readable timestamp |
| `updated_at` | timestamp | yes | Changes when thread fields change or a new non-duplicate message is created |
| `archived_at` | nullable timestamp | no | Set when archived; null while active |

State transitions:

```text
absent -> active       when a thread is created
active -> active       when title or mode changes
active -> archived     when archived
archived -> archived   durable recovery state; no unarchive behavior in M3
```

List behavior:

- Default list returns active threads only.
- Active list is ordered by `updated_at` descending.
- Archived threads remain durable and retrievable by direct id for future recovery behavior.

Access rules:

- All thread queries include the current local `user_id`.
- A missing thread and a thread owned by another user both produce the same `thread_not_found` API error.

## Message

A durable complete-text entry inside one owned thread.

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| `id` | string | yes | Primary key; app-generated text id with `msg_` prefix |
| `thread_id` | string | yes | Foreign key to `threads.id` |
| `user_id` | string | yes | Foreign key to `users.id`; must match the owning thread's user |
| `role` | enum | yes | M3 persists `user` only |
| `content` | string | yes | Trimmed; must not be empty or whitespace-only |
| `metadata` | JSON object | no | Optional; defaults to empty object; no run/tool/model semantics in M3 |
| `client_message_id` | nullable string | no | Optional idempotency key; non-empty when provided |
| `created_at` | timestamp | yes | Machine-readable timestamp |

State:

```text
absent -> persisted final user message
```

Rules:

- Messages are listed in stable creation order.
- Message creation does not create assistant placeholders.
- Message creation does not create run events, streaming deltas, tool calls, model outputs, worker jobs, or LLM requests.
- M3 has no message update/delete behavior.

## Client Message Identifier

An optional client-provided idempotency key for user message creation.

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| `client_message_id` | string | no | Trimmed; if present, must be non-empty and no longer than 120 characters |
| `thread_id` | string | yes for uniqueness | Idempotency scope |
| `user_id` | string | yes for uniqueness | Idempotency scope |

Invariant:

```text
unique(thread_id, user_id, client_message_id) where client_message_id is not null
```

Duplicate behavior:

- If a duplicate key is reused for the same thread and local identity, the API returns the existing message.
- Returning an existing message must not create a new row and must not update the thread timestamp.

## API Error

A stable client-facing error response.

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| `code` | string | yes | Stable machine-readable code |
| `message` | string | yes | Human-readable, non-secret message |
| `request_id` | string | yes | Request id generated by diagnostics boundary |

Initial codes:

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `invalid_request` | 400 | Invalid JSON, empty message, invalid title, invalid mode, or invalid client message id |
| `thread_not_found` | 404 | Thread is absent or not owned by the current identity |
| `method_not_allowed` | 405 | HTTP method does not match a supported route |
| `internal_error` | 500 | Unexpected server failure without leaking implementation details |

## Schema Revision

M3's migration state.

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| `version` | integer | yes | `2` after `000002_m3_auth_thread_message.up.sql` |
| `dirty` | boolean | yes | Must be `false` for readiness to pass |
| `tables` | set | yes | `users`, `threads`, `messages` |

Lifecycle:

```text
M2 baseline version 1 -> M3 applied version 2 -> rollback to version 1 -> reapply version 2
```

Readiness:

- M3 readiness fails when schema version is absent, dirty, or less than `2`.
- M3 readiness does not require seed data.

## Seed Data Set

An explicit local demo dataset created by command, never by migration.

| Field | Type | Rules |
|-------|------|-------|
| `user_id` | string | Fixed local user id |
| `thread_id` | string | Deterministic demo thread id |
| `message_id` | string | Deterministic demo user message id |
| `client_message_id` | string | Deterministic idempotency key for the demo message |

Rules:

- Running the seed command multiple times is idempotent.
- Seed creates no assistant messages, run events, tool calls, or worker records.
- Seed output is structured and safe to paste into docs or bug reports.

## Frontend Data Source Mode

The web shell's active source for thread/message data.

| Mode | Trigger | Behavior |
|------|---------|----------|
| `mock` | `VITE_LOOMI_API_BASE_URL` absent or empty | Existing in-memory mock thread/message/run demonstration remains usable |
| `real_api` | `VITE_LOOMI_API_BASE_URL` set | Thread list, thread mutations, and messages use M3 HTTP API |
| `real_api_error` | Real API configured but unavailable or returns API error | UI shows recoverable error and does not fall back to mock data |

Rules:

- Run timeline/debug rail remains mock, empty, or explicitly deferred in M3.
- Real API timestamps remain machine-readable in the client boundary; display formatting happens in UI components.
