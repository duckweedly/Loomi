---
title: M6 Worker Job Pipeline API
description: Run creation, worker events, and queue configuration contracts for the M6 background execution MVP.
---

M6 reuses the M4/M5 run/event/SSE API envelope and changes run creation to an asynchronous acknowledgement.

## Run creation

### `POST /v1/threads/{thread_id}/runs`

The handler returns `202 Accepted` and a queued run:

```json
{
  "run": {
    "id": "run_...",
    "thread_id": "thread_...",
    "status": "queued",
    "source": "local_simulated",
    "title": "Local simulated run"
  },
  "request_id": "req_..."
}
```

The backend persists:

- `run_created`
- `run_queued`
- one active `background_jobs` row for the run

Only one active run is allowed per thread while a run is `pending`, `queued`, `running`, or `recovering`.

Background jobs store a redacted metadata snapshot for the queued execution input. Claims increment `ownership_version`; completion, failure, and lease renewal require the current worker id and ownership version.

## Run statuses

M6 frontend/backend run states include:

| Status | Meaning |
| --- | --- |
| `pending` | Legacy/pre-worker transient state. |
| `queued` | Run was accepted and is waiting for a worker claim. |
| `running` | A worker claimed the job and execution is active. |
| `recovering` | Work is being recovered after ownership interruption. |
| `stopping` | Frontend-visible transient state after a stop request is observed. |
| `completed` | Terminal success. |
| `failed` | Terminal failure with redacted explanation. |
| `stopped` | Terminal user-requested stop. |
| `cancelled` | Terminal cancellation-compatible frontend state. |
| `retrying` | Frontend-visible retry state used by existing runtime scripts and future US2 retries. |

## Worker and pipeline event types

M6 uses the existing run event object shape with new event `type` values:

| Category | Backend type | Frontend type | Status |
| --- | --- | --- | --- |
| `lifecycle` | `run_queued` | `run.queued` | `queued` |
| `progress` | `job_claimed` | `job.claimed` | `running` |
| `progress` | `lease_renewed` | `worker.lease_renewed` | `running` |
| `progress` | `pipeline_step_started` | `pipeline.step.started` | `running` |
| `progress` | `pipeline_step_completed` | `pipeline.step.completed` | `running` |
| `progress` | `job_recovering` | `job.recovering` | `recovering` |
| `progress` | `job_retry_scheduled` | `job.retry_scheduled` | `recovering` |
| `progress` | `stop_requested` | `run.stopping` | `stopping` |
| `error` | `job_attempt_failed` | `job.attempt_failed` | `failed` |
| `error` | `job_retry_exhausted` | `job.retry_exhausted` | `failed` |
| `final` | `run_completed` | `run.completed` | `completed` |
| `final` | `run_failed` | `run.failed` | `failed` |
| `final` | `run_stopped` | `run.stopped` | `stopped` |

SSE still sends history-first `run_event` frames and supports `after_sequence` replay. The frontend keeps queued and worker events visible in RunRail.

## Worker configuration

Local worker behavior is controlled by backend environment variables:

| Variable | Default | Meaning |
| --- | --- | --- |
| `LOOMI_WORKER_QUEUE_ENABLED` | `true` | Starts the in-process worker when enabled. |
| `LOOMI_WORKER_QUEUE_PAUSED` | `false` | Prevents worker startup when true. |
| `LOOMI_WORKER_LEASE_SECONDS` | `30` | Lease duration for claimed jobs. |
| `LOOMI_WORKER_MAX_ATTEMPTS` | `3` | Maximum attempts for future retry paths. |
| `LOOMI_WORKER_POLL_MILLIS` | `250` | Worker polling interval. |

## Diagnostics status

### `GET /v1/diagnostics/worker-queue`

Returns safe worker queue diagnostics:

```json
{
  "diagnostics": {
    "queue_status": "ready",
    "worker_status": "ready",
    "queued_count": 0,
    "leased_count": 0,
    "stale_count": 0,
    "retrying_count": 0,
    "dead_count": 0,
    "updated_at": "2026-05-24T10:00:00Z"
  },
  "request_id": "req_..."
}
```

`LOOMI_WORKER_QUEUE_PAUSED=true` reports paused queue/worker status. `LOOMI_WORKER_QUEUE_ENABLED=false` reports worker status `stopped`. Stale, retrying, or dead jobs report degraded queue/worker status.

## Redaction rules

Job errors and worker diagnostics must use stable, redacted error codes/messages. They must not expose secrets, raw provider request bodies, raw provider error bodies, or user-controlled data as executable instructions.
