---
title: M6 Worker Job Pipeline Architecture
description: Durable background run execution, worker-owned jobs, and pipeline events for Loomi M6.
---

M6 moves run execution out of the request lifecycle and into a durable worker/job boundary. The current implemented slice is the US1 MVP: a run can be acknowledged as queued, persisted as a background job, claimed by a local worker, executed, and replayed through persisted events.

## Boundary

The product data service owns durable run and job state. Runtime code owns worker polling, job claiming, pipeline event recording, and delegation to the local simulated runner or model gateway runner.

`POST /v1/threads/{thread_id}/runs` now creates a queued run and a `background_jobs` row, then returns `202 Accepted`. The HTTP request no longer performs synchronous execution.

`cmd/loomi-api` starts an in-process worker when `LOOMI_WORKER_QUEUE_ENABLED` is not false and `LOOMI_WORKER_QUEUE_PAUSED` is not true.

## Durable job model

M6 adds `background_jobs` with one active job per run and extends active run status to include `queued` and `recovering`.

Current job state fields include:

- run, thread, user, and job kind identifiers
- status, priority, attempt count, and max attempts
- scheduled time
- lease owner, lease expiry, and ownership version
- redacted metadata needed to resume queued execution
- redacted last error code/message

The implemented slice supports queued claim, lease renewal, stale lease recovery, retry scheduling, retry exhaustion, ownership-guarded completion/failure, and queue diagnostic aggregation.

## Worker flow

The worker loop performs:

1. recover expired leases before claiming new work
2. claim the next queued job and increment its ownership version
3. mark the run as running
4. record `job_claimed`
5. renew the lease before invoking the runtime boundary
6. record pipeline step events around run execution
7. verify ownership before each local simulated runtime step writes events
8. complete or fail the job only when the worker still owns the current lease version

Local simulated runs are executed through `LocalRunner`, which persists the simulated assistant message before recording final completion. Model-gateway runs route through the same worker boundary and hydrate `message_id`, `provider_id`, and model override from the durable job metadata snapshot.

## Observable pipeline events

M6 keeps worker execution visible through persisted run events:

- `run_queued`
- `job_claimed`
- `pipeline_step_started`
- `pipeline_step_completed`
- `lease_renewed`
- `job_recovering`
- `job_retry_scheduled`
- `job_attempt_failed`
- `job_retry_exhausted`
- `run_completed`
- `run_failed`
- `run_stopped`

The frontend maps these to dotted runtime event names such as `run.queued`, `job.claimed`, `pipeline.step.started`, and `run.completed`. RunRail groups worker and pipeline events under the worker/job timeline section.

## Cancellation and recovery status

The frontend understands queued, running, recovering, stopping, completed, failed, stopped, and cancelled run states. Backend cancellation now persists `stop_requested_at`, cancels active jobs, records `stop_requested` and terminal `run_stopped`, and makes local runner writes stop at safe boundaries once a run is stopped.

## Safety boundaries

M6 job, worker, retry, and diagnostic output must not expose provider credentials, Authorization headers, raw provider request payloads, or raw provider failure bodies. Worker events should explain state transitions without turning provider/user content into instructions.

## Deferred capabilities

Desktop runtime, tool execution, RAG/memory, plugins, and multi-agent orchestration remain outside the current slice.
