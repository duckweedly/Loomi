# Data Model: M4 Run, Event, and SSE

## Run

A durable execution attempt for one thread and the fixed local owner.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable run id, unique globally, prefixed for readability |
| `thread_id` | string | References one owned thread |
| `user_id` | string | Fixed local owner inherited from the current identity boundary |
| `status` | enum | `pending`, `running`, `completed`, `failed`, `stopped` |
| `source` | enum | `local_simulated` for M4 |
| `title` | string | Short user-visible run label, derived from local simulation context |
| `created_at` | timestamp | Machine-readable creation time |
| `updated_at` | timestamp | Machine-readable latest lifecycle or event update time |
| `completed_at` | timestamp/null | Set when status is terminal |
| `error_code` | string/null | Stable code for failed runs |
| `error_message` | string/null | Redacted human-readable failure message |

Rules:

- A run belongs to exactly one thread and one local user.
- Active statuses are `pending` and `running`.
- Terminal statuses are `completed`, `failed`, and `stopped`.
- Different threads may have active runs at the same time.
- A single thread must not have more than one active run.
- Terminal runs remain readable as history.

State transitions:

```text
pending -> running -> completed
pending -> running -> failed
pending -> running -> stopped
pending -> stopped
```

Invalid transitions:

- Any terminal state to active.
- Creating a second active run for the same thread.
- Hard-interrupting a run after it already reached a terminal state.

## Run Event

A durable timeline entry for one run.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable event id, unique globally |
| `run_id` | string | References one run |
| `thread_id` | string | Denormalized for scoped lookups and safety checks |
| `user_id` | string | Fixed local owner |
| `sequence` | integer | Monotonic per run, starts at 1 |
| `category` | enum | `lifecycle`, `progress`, `message`, `error`, `final` |
| `type` | string | More specific event type within the category |
| `summary` | string | Short user-visible summary |
| `content` | string/null | User-visible text payload when appropriate |
| `metadata` | object | Machine-readable details safe for local diagnostics |
| `created_at` | timestamp | Machine-readable event time |

Rules:

- Events are ordered by `(sequence, id)` within a run.
- `(run_id, sequence)` is unique.
- Events from one run must never appear in another run's timeline.
- Event payloads are data, not instructions.
- Events must not expose secrets, full database URLs, tokens, or sensitive local file contents.
- Duplicate or retried event creation must not duplicate timeline entries for the same event identity.

Initial categories:

| Category | Purpose | Example types |
|----------|---------|---------------|
| `lifecycle` | Run status transitions | `run_created`, `run_started`, `run_stopping`, `run_stopped` |
| `progress` | User-visible simulated work step | `context_loaded`, `thinking`, `drafting` |
| `message` | Simulated output text | `assistant_message` |
| `error` | Failure details | `run_failed`, `stream_error` |
| `final` | Terminal outcome | `run_completed`, `run_stopped`, `run_failed` |

## Event Stream Cursor

A client-visible position in a run's event timeline.

| Field | Type | Rules |
|-------|------|-------|
| `run_id` | string | The streamed run |
| `last_event_id` | string/null | Last delivered event id when reconnecting |
| `last_sequence` | integer/null | Last delivered sequence when available |

Rules:

- Opening a stream without a cursor delivers persisted history from the beginning of the run.
- Reconnecting with a cursor delivers persisted events after the cursor, then live events.
- Clients may dedupe by event id.

## Stop Request

A user-visible cooperative request to end an active run.

| Field | Type | Rules |
|-------|------|-------|
| `run_id` | string | Run being stopped |
| `user_id` | string | Fixed local owner |
| `requested_at` | timestamp | Machine-readable request time |
| `result_status` | enum | `stopped`, `already_terminal`, `not_found` |
| `message` | string | Human-readable result |

Rules:

- Stop is best-effort cooperative in M4.
- Stop for active deterministic local runs should produce a `lifecycle` event and a terminal `final` event.
- Stop for terminal runs must not change the existing terminal outcome.

## Deterministic Local Simulation

The first M4 execution source.

| Field | Type | Rules |
|-------|------|-------|
| `source` | enum | Always `local_simulated` in M4 |
| `script_name` | string | Deterministic script label |
| `steps` | ordered list | Produces lifecycle/progress/message/final events |

Rules:

- The same start input produces a predictable sequence of event categories and statuses.
- The simulation may pause between steps for live-stream visibility.
- The simulation must check for cooperative stop at step boundaries.
- The simulation must be clearly labeled as local simulated execution.

## Stream State

The web shell's user-visible stream condition.

| State | Meaning |
|-------|---------|
| `connecting` | Opening or reconnecting to a run stream |
| `live` | Receiving history or live events successfully |
| `recoverable_error` | Stream failed but persisted history can be reloaded |
| `closed` | Stream ended after terminal run state |

Rules:

- Stream errors are visible and recoverable.
- Stream state must not imply the run is still live after the stream fails.
- Persisted event history remains the recovery source.

## M4 Schema Revision

The migration state required for run/event readiness.

Rules:

- M4 readiness requires clean migration version `3` or later.
- M4 readiness does not require any seeded runs or events.
- Rollback to M3 removes run/event tables and makes M4 readiness fail while preserving M3 readiness expectations.
