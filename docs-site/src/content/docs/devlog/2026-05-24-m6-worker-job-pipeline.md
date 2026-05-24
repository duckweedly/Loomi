---
title: 2026-05-24 M6 Worker Job Pipeline Devlog
description: Implementation notes, validation results, limitations, and next steps for the M6 worker/job pipeline MVP.
---

## Completed scope

M6 US1-US4 now moves run creation to a durable queued background execution path, recovers interrupted worker ownership, cancels queued or running background work safely, and exposes safe worker queue diagnostics.

Implemented slice:

- migration `000005_m6_worker_job_pipeline` for queued/recovering run states and `background_jobs`
- worker queue configuration defaults and validation
- readiness target updated to schema version `5`
- backend domain constants for jobs, worker/queue status, pipeline steps, and M6 events
- run start service creates a queued run, background job, `run_created`, and `run_queued`
- `POST /v1/threads/{thread_id}/runs` returns `202 Accepted` without synchronous execution
- local in-process worker claims queued jobs and invokes the runtime boundary
- local simulated run execution works through the worker path and persists the assistant message before final completion
- pipeline step recorder persists worker-visible execution events before terminal completion
- background jobs persist redacted queued metadata plus ownership versions
- queued model-gateway jobs hydrate `message_id`, `provider_id`, and model override from durable metadata
- worker/coordinator paths recover expired leases before new claims
- stale workers cannot complete/fail a job or write local simulated terminal events after ownership has moved
- retry exhaustion records redacted `job_retry_exhausted` and terminal `run_failed` events
- queued/running stop requests cancel active jobs and record terminal stopped outcomes
- `/v1/diagnostics/worker-queue` exposes safe ready/paused/degraded/stopped queue health
- frontend maps `queued`, `recovering`, `stopping`, worker, and pipeline events
- Chat Canvas and RunRail render queued/background-running state and worker/pipeline timeline rows

## Safety notes

Worker/job errors and diagnostics remain redacted. Gateway job metadata is limited to safe routing identifiers and model selection, and M6 does not add tool execution, desktop runtime, activity capture, plugins, RAG/memory, or multi-agent orchestration.

## Validation log

Validated during implementation:

```bash
go test ./...
bun test ./web/src/runtime/executionAdapter.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/realApiClient.test.ts ./web/src/state.runtime.test.ts ./web/src/runtime/chatCanvasState.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/RunTimeline.runtime.test.ts
zsh -o null_glob -c 'bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts'
bun run --cwd web build
bun run --cwd docs-site build
```

Latest recorded results:

- `go test ./...` passed after US2 recovery/ownership updates.
- Targeted M6 frontend tests passed with 52 tests and 177 expectations.
- Frontend quickstart-equivalent suite passed with 149 tests and 434 expectations using zsh `null_glob` expansion.
- `bun run --cwd web build` passed.
- `bun run --cwd docs-site build` passed and generated the new M6 architecture/API/runbook/devlog pages.
- Browser smoke passed in mock mode: RunRail displayed `run.queued`, `job.claimed`, `pipeline.step.started`, and `pipeline.step.completed` rows after a regenerated run, with no warning/error console messages.
- Real OpenAI-compatible provider smoke passed through the queued model-gateway path with `model_output_delta`, `model_output_completed`, and `run_completed` events after fixing worker lease interval SQL.

## Final review fixes

- Queued model-gateway runs now hydrate `message_id`, `provider_id`, and model override from durable job metadata.
- Local simulated worker runs persist the assistant message before final completion and record `pipeline_step_completed` before the terminal event.
- Worker-created job and recovery events are published to live SSE subscribers as well as persisted for replay.
- Gateway worker failures return an error to the worker boundary so failed runs are not followed by a completed job path.

## Remaining follow-up

- Improve live frontend recovery when an old SSE connection fails during local API restarts.
