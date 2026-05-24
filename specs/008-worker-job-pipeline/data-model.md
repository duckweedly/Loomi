# Data Model: M6 Worker Job Pipeline

## Run

M6 extends the existing run concept so user-visible execution can be queued, recovered, cancelled, and completed by background workers.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable run id, unique globally |
| `thread_id` | string | References one owned thread |
| `user_id` | string | Fixed local owner inherited from the current identity boundary |
| `status` | enum | `pending`, `queued`, `running`, `recovering`, `completed`, `failed`, `stopped` |
| `source` | enum | Existing sources plus background-capable model execution |
| `title` | string | Short user-visible run label |
| `stop_requested_at` | timestamp/null | Set when a user requests cancellation |
| `created_at` | timestamp | Machine-readable creation time |
| `updated_at` | timestamp | Latest lifecycle, queue, or event update time |
| `completed_at` | timestamp/null | Set when status is terminal |
| `error_code` | string/null | Stable redacted code for failed runs |
| `error_message` | string/null | User-safe failure message |

Rules:

- A run belongs to exactly one thread and one local user.
- A single thread must not have more than one active run.
- Active states are `pending`, `queued`, `running`, and `recovering`.
- Terminal states are `completed`, `failed`, and `stopped`.
- A terminal run must not return to an active state.
- A run may have at most one terminal event and at most one final assistant message.
- Stop requests are persisted so queued and running background work can observe them.

State transitions:

```text
pending -> queued -> running -> completed
pending -> queued -> running -> failed
pending -> queued -> running -> stopped
pending -> queued -> stopped
queued -> recovering -> queued
running -> recovering -> queued
recovering -> failed
```

Invalid transitions:

- Any terminal state to active.
- Creating a second active run for the same thread.
- Completing a run after a stop request has already produced `stopped`.
- Creating duplicate final assistant messages for the same run.

## Background Job

A durable unit of background execution associated with one run.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable job id, unique globally |
| `run_id` | string | References one run |
| `thread_id` | string | Denormalized for scoped lookup and active-run validation |
| `user_id` | string | Fixed local owner |
| `kind` | enum | Initial value: `run_execution` |
| `status` | enum | `queued`, `leased`, `retrying`, `completed`, `failed`, `cancelled`, `dead` |
| `priority` | integer | Lower scope first; defaults to normal local priority |
| `attempt_count` | integer | Incremented for each processing attempt |
| `max_attempts` | integer | Upper bound before terminal failure |
| `scheduled_at` | timestamp | Earliest time the job may be claimed |
| `leased_by` | string/null | Worker id that currently owns the job |
| `lease_expires_at` | timestamp/null | Time after which ownership is stale |
| `last_error_code` | string/null | User-safe failure class from the latest failed attempt |
| `last_error_message` | string/null | Redacted latest failure summary |
| `created_at` | timestamp | Machine-readable creation time |
| `updated_at` | timestamp | Latest state or ownership update time |

Rules:

- A non-terminal run has at most one active job for execution.
- Only `queued` or stale `leased`/`retrying` jobs are claimable.
- Claiming a job sets `leased_by`, `lease_expires_at`, and an owned status atomically.
- A job cannot be claimed by more than one active worker at a time.
- A job with a stopped run must not start normal execution.
- Exhausted attempts transition the job to `dead` and the run to `failed`.
- Completed, cancelled, failed, and dead jobs are terminal and remain readable for diagnostics.

State transitions:

```text
queued -> leased -> completed
queued -> cancelled
leased -> retrying -> queued
leased -> completed
leased -> failed -> retrying
leased -> cancelled
retrying -> dead
```

## Worker

A local execution owner that claims and advances background jobs.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable worker identity for the process lifetime |
| `status` | enum | `ready`, `paused`, `draining`, `unhealthy`, `stopped` |
| `started_at` | timestamp | Worker start time |
| `last_heartbeat_at` | timestamp/null | Latest visible worker heartbeat |
| `current_job_id` | string/null | Job currently owned, if any |
| `processed_count` | integer | Count of completed or terminally handled jobs |
| `last_error_code` | string/null | Redacted latest worker error class |

Rules:

- A worker may own zero or one job in the first M6 slice.
- A worker must renew an active lease before expiry while processing.
- A worker must stop claiming new work when paused or draining.
- Worker diagnostics must not include provider credentials or raw secret values.

## Worker Lease

The time-bounded ownership record embedded in or associated with a background job.

| Field | Type | Rules |
|-------|------|-------|
| `job_id` | string | Job whose ownership is controlled |
| `worker_id` | string | Current owner |
| `lease_started_at` | timestamp | When ownership began |
| `lease_expires_at` | timestamp | When ownership becomes stale |
| `renewed_at` | timestamp | Latest successful renewal time |
| `version` | integer | Monotonic ownership version for stale update protection |

Rules:

- A lease is valid only before `lease_expires_at`.
- Renewals must only succeed for the current owner and current version.
- Completion must only succeed for the current owner or when the run is already terminal in a compatible state.
- Expired leases allow another worker to recover the job.

## Run Event

M6 reuses existing persisted run events and adds queue, worker, recovery, and pipeline event types within existing event categories.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable event id, unique globally |
| `run_id` | string | References one run |
| `thread_id` | string | Denormalized for scoped lookups and safety checks |
| `user_id` | string | Fixed local owner |
| `sequence` | integer | Monotonic per run, starts at 1 |
| `category` | enum | `lifecycle`, `progress`, `message`, `error`, `final` |
| `type` | string | Loomi event type |
| `summary` | string | Short user-safe summary |
| `content` | string/null | Text delta, final text, or safe failure details when appropriate |
| `metadata` | object | Safe diagnostics only |
| `created_at` | timestamp | Machine-readable event time |

M6 event types:

| Category | Event Type | Purpose |
|----------|------------|---------|
| `lifecycle` | `run_queued` | Background job was created for the run |
| `progress` | `job_claimed` | Worker acquired ownership |
| `progress` | `lease_renewed` | Worker confirmed active ownership |
| `progress` | `pipeline_step_started` | A named background stage began |
| `progress` | `pipeline_step_completed` | A named background stage completed |
| `progress` | `job_recovering` | Stale ownership was detected |
| `progress` | `job_retry_scheduled` | Work will be retried later |
| `progress` | `stop_requested` | User cancellation was recorded |
| `error` | `job_attempt_failed` | One attempt failed with redacted details |
| `error` | `job_retry_exhausted` | Retries are exhausted |
| `final` | `run_completed` | Run completed successfully |
| `final` | `run_failed` | Run failed terminally |
| `final` | `run_stopped` | Run stopped terminally |

Rules:

- Events are ordered by `(sequence, id)` within a run.
- Event payloads are data, not instructions.
- Worker ids and job ids may appear in safe metadata; credentials and raw provider payloads must not.
- Recovery must not duplicate prior terminal events.

## Pipeline Step

A named stage within one background job attempt.

| Field | Type | Rules |
|-------|------|-------|
| `name` | enum | `enqueue`, `claim`, `prepare_context`, `invoke_runtime`, `finalize`, `recover`, `fail` |
| `status` | enum | `pending`, `running`, `completed`, `failed`, `skipped` |
| `attempt` | integer | Job attempt number |
| `started_at` | timestamp/null | Set when the stage begins |
| `completed_at` | timestamp/null | Set when the stage reaches a terminal stage state |
| `summary` | string | User-safe explanation for timeline/debug output |

Rules:

- The first M6 slice uses a linear stage sequence, not a configurable graph.
- Stages exist to explain execution and validation; they do not execute arbitrary plugins.
- Stage failure either schedules retry, stops the run, or fails the run with a redacted explanation.

## Queue Diagnostics

A local diagnostic view of queue and worker health.

| Field | Type | Rules |
|-------|------|-------|
| `queue_status` | enum | `ready`, `paused`, `unhealthy`, `degraded` |
| `worker_status` | enum | `ready`, `paused`, `unhealthy`, `degraded`, `stopped` |
| `queued_count` | integer | Number of claimable queued jobs |
| `leased_count` | integer | Number of currently leased jobs |
| `stale_count` | integer | Number of expired non-terminal leases |
| `retrying_count` | integer | Number of jobs waiting for retry |
| `dead_count` | integer | Number of jobs exhausted without success |
| `updated_at` | timestamp | Diagnostic snapshot time |

Rules:

- Diagnostics are for local validation and troubleshooting.
- Diagnostics must be redacted and must not expose secrets.
- Paused execution is distinct from unhealthy execution.
