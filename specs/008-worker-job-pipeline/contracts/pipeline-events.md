# Contract: M6 Pipeline and Runtime Events

## Purpose

Define how background execution stages appear in Loomi's persisted run timeline and frontend runtime state.

## Event Ordering

- Events remain ordered by per-run sequence.
- History-first stream delivery remains the reconnect contract.
- Background events are persisted before or at the same boundary as the state transition they explain.
- Terminal events are final; later worker output cannot change the run state.

## New M6 Event Types

| Type | Category | Required Metadata | User Meaning |
|------|----------|-------------------|--------------|
| `run_queued` | `lifecycle` | `job_id` | Work has been accepted for background execution |
| `job_claimed` | `progress` | `job_id`, `worker_id`, `attempt` | A worker started handling the job |
| `lease_renewed` | `progress` | `job_id`, `worker_id` | Worker ownership remains active |
| `pipeline_step_started` | `progress` | `job_id`, `attempt`, `step` | A background stage started |
| `pipeline_step_completed` | `progress` | `job_id`, `attempt`, `step` | A background stage completed |
| `job_recovering` | `progress` | `job_id`, `previous_worker_id`, `attempt` | Loomi detected stale ownership and is recovering work |
| `job_retry_scheduled` | `progress` | `job_id`, `next_attempt`, `scheduled_at` | Work will be tried again |
| `stop_requested` | `progress` | `requested_at` | User cancellation was recorded |
| `job_attempt_failed` | `error` | `job_id`, `attempt`, `error_code` | One attempt failed with safe details |
| `job_retry_exhausted` | `error` | `job_id`, `attempt_count`, `error_code` | No more recovery attempts remain |
| `run_completed` | `final` | `job_id` | Run completed successfully |
| `run_failed` | `final` | `job_id`, `error_code` | Run failed terminally |
| `run_stopped` | `final` | `job_id` when available | Run stopped terminally |

Metadata rules:

- `worker_id`, `job_id`, `attempt`, and `step` are safe diagnostic metadata.
- Provider credentials, raw secret values, raw provider request bodies, and unredacted provider errors are never event metadata.
- User-controlled or provider-controlled content is data, not instructions.

## Pipeline Step Names

The first M6 slice uses these named steps:

1. `enqueue`: work accepted and made durable.
2. `claim`: worker ownership acquired.
3. `prepare_context`: bounded run context prepared.
4. `invoke_runtime`: model/runtime work invoked through existing runtime boundaries.
5. `finalize`: terminal run state and final message handled.
6. `recover`: stale or failed attempt prepared for retry.
7. `fail`: exhausted or unrecoverable work made terminal.

Rules:

- Step names are product diagnostics, not plugin extension points.
- A job attempt may skip steps that are not relevant after cancellation or recovery.
- The first slice is linear; no arbitrary graph or user-authored workflow is introduced.

## Frontend Runtime Mapping

| Run/Event Condition | Runtime UI State |
|---------------------|------------------|
| `run_queued` before worker claim | Queued / waiting |
| `job_claimed` or `pipeline_step_started` | Running |
| `job_recovering` or `job_retry_scheduled` | Recovering |
| `stop_requested` without terminal event | Stopping |
| `run_stopped` | Stopped |
| `run_failed` | Failed |
| `run_completed` | Completed |

Required frontend behavior:

- Existing mock mode remains available only when mock mode is explicitly active.
- Real API mode must not silently fall back to mock execution if background queue execution fails.
- Reconnecting to a run stream must replay persisted queue and worker events before live updates.
- Stale stream updates must not overwrite a newer terminal state.

## Failure Visibility

Attempt failures should be visible without overwhelming the user:

- Recoverable failures use `job_attempt_failed` followed by `job_retry_scheduled`.
- Exhausted failures use `job_retry_exhausted` followed by `run_failed`.
- Cancellation uses `stop_requested` followed by `run_stopped` or an already-terminal response.
- Missing run/thread references fail safely and become visible diagnostics rather than invisible worker crashes.
