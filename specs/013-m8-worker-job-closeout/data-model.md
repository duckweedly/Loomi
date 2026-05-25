# Data Model: M8 Worker Job Closeout

## M8 Audit Item

| Field | Type | Rules |
| --- | --- | --- |
| `name` | string | One original M8 roadmap requirement |
| `expected_behavior` | string | What closeout must verify |
| `status` | enum | `covered`, `patched`, `deferred` |
| `evidence` | list | Spec, code, test, or docs references |

Rules:

- Every original M8 item must have one closeout status.
- `deferred` is allowed only for explicitly non-M8 work.

## Background Job

Existing durable job row created by M6 for run execution.

Closeout-relevant fields:

- `status`
- `attempt_count`
- `max_attempts`
- `scheduled_at`
- `leased_by`
- `lease_expires_at`
- `ownership_version`

Rules:

- Claimable jobs require `status = queued` and `scheduled_at <= now`.
- Recovery must not clear ownership without updating retry scheduling evidence.

## Worker Lease

Existing job ownership guard.

Rules:

- Renew, complete, and fail operations require matching `worker_id` and `ownership_version`.
- A stale worker must receive no state change after recovery reassigns ownership.

## Retry Schedule

The next eligible claim time after an expired lease is recovered.

Rules:

- The first recovery retry is scheduled in the future.
- Later recovery retries may use a short bounded backoff.
- Retry scheduling must keep max-attempt terminal failure behavior intact.

## Closeout Record

Documentation evidence that original M8 is complete.

Rules:

- Must include audit result, minimal patch, validation commands, and non-goals.
- Must not claim M9 RunContext/Pipeline completion.
