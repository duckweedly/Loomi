# Contract: Worker Queue Semantics

## Purpose

Define the observable behavior for M6 background job creation, claim, lease renewal, recovery, retry, completion, cancellation, and diagnostics.

## Job Creation

When a user starts a run:

1. Loomi validates that the thread has no active run.
2. Loomi creates a run in an active state.
3. Loomi creates one durable `run_execution` job for that run.
4. Loomi records a `run_queued` event.
5. Loomi returns the run with `queued` or `running` status within the acknowledgement target.

Required guarantees:

- Run creation and job creation are atomic from the user's perspective.
- A failed job creation must not leave a user-visible active run without work.
- A retry of the same accepted start operation must not create multiple active runs for one thread.

## Claim

A worker may claim a job only when:

- Job status is claimable.
- The job is scheduled now or earlier.
- The job is not terminal.
- The associated run is not terminal.
- The associated run has no stop request that should prevent execution.

On successful claim:

- `leased_by` is set to the worker id.
- `lease_expires_at` is set to a future time.
- `attempt_count` is incremented for a new processing attempt.
- A `job_claimed` event is recorded.

Concurrent claim behavior:

- Exactly one worker wins a claim for a job.
- Losing workers receive no ownership and must not record run events for that job.
- Stale workers must not complete a job after losing ownership unless the run is already terminal in a compatible state.

## Lease Renewal

While processing a job, the owner renews the lease before expiry.

Renewal succeeds only when:

- The renewing worker is the current owner.
- The ownership version still matches.
- The job is not terminal.
- The run is not terminal.

Renewal failure means the worker must stop normal processing and avoid writing additional non-terminal progress for the job.

## Recovery

A job is recoverable when:

- It is non-terminal.
- Its lease has expired.
- Its attempt count is below the configured maximum.
- The associated run is not terminal.

Recovery behavior:

1. Record `job_recovering` for the run.
2. Clear stale ownership or mark the job retrying.
3. Schedule the job for another claim.
4. Record `job_retry_scheduled` when retrying.

If attempts are exhausted:

1. Mark the job `dead`.
2. Mark the run `failed`.
3. Record `job_retry_exhausted` and `run_failed`.

## Cancellation

A stop request is persisted on the run.

Queued job behavior:

- If a stop request exists before claim, the job transitions to `cancelled`.
- The run transitions to `stopped`.
- `stop_requested` and `run_stopped` events are recorded.
- Normal execution does not begin.

Running job behavior:

- The worker observes stop requests at safe boundaries.
- The worker stops producing normal output after accepting the stop.
- The run transitions to `stopped` unless it is already terminal.
- A terminal run must not be overwritten by cancellation.

## Completion

Successful completion requires current ownership and a non-terminal run.

Completion behavior:

1. Final output, if any, is persisted once.
2. The run transitions to `completed`.
3. The job transitions to `completed`.
4. A single `run_completed` terminal event is recorded.

Duplicate completion attempts:

- Must not create duplicate final assistant messages.
- Must not create duplicate terminal events.
- Must return or observe the existing compatible terminal state.

## Diagnostics

Diagnostics expose redacted local queue and worker health.

Required diagnostic distinctions:

- `ready`: queue and worker can accept or process work.
- `paused`: processing is intentionally disabled.
- `degraded`: work exists but needs recovery or retry attention.
- `unhealthy`: required queue/worker dependencies cannot support execution.
- `stopped`: worker is not running.

Diagnostics must not expose provider credentials, raw secret values, or unredacted provider failures.
