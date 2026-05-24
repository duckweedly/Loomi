# Quickstart: M6 Worker Job Pipeline

## Purpose

Validate the M6 vertical slice locally: background run acknowledgement, durable queue processing, worker ownership, recovery, cancellation, timeline replay, diagnostics, and documentation updates.

## Prerequisites

- Local database available through the existing Loomi development setup.
- Existing M3/M4/M5 migrations applied or ready to apply.
- Local provider configuration from M5 available when validating model-backed completion.
- M6 worker processing can be enabled, paused, or interrupted for smoke tests.

## 1. Apply Migrations

```sh
go run ./cmd/loomi-api --migrate up
```

Expected outcome:

- Migration `000005_m6_worker_job_pipeline` is applied after `000004`.
- Readiness reports the schema version required for M6.
- No demo jobs or runs are inserted by migration.

## 2. Start API With Worker Processing Enabled

```sh
go run ./cmd/loomi-api
```

Expected outcome:

- Local API starts successfully.
- Worker queue diagnostics report `ready` or an expected local degraded state.
- Provider credentials are not printed in logs or diagnostics.

## 3. Start a Background Run

Create or reuse an existing thread with a user message, then start a run through the existing run start contract.

Expected outcome:

- The start response returns within 2 seconds with `queued` or `running` status.
- A durable job exists for the run.
- The run timeline includes `run_queued`.
- The initial request does not perform the full run synchronously.

## 4. Validate History-First Timeline Recovery

After the run is acknowledged, close the browser or stop consuming the event stream, wait for worker progress, then reconnect to the run event stream.

Expected outcome:

- Persisted queue and worker events are replayed before live events.
- Event ordering remains stable.
- Final assistant output appears once if the run completes.

## 5. Validate Multi-Worker Claim Safety

Run local validation with two workers or two claim attempts against queued work.

Expected outcome:

- Only one worker claims a given job.
- Losing workers do not record progress events for the job.
- Across repeated attempts, the run receives no duplicate terminal events and no duplicate assistant messages.

## 6. Validate Worker Interruption and Recovery

Start a run, interrupt the worker after it claims the job, wait for the ownership window to expire, then resume worker processing.

Expected outcome:

- Diagnostics show stale or degraded work before recovery.
- Timeline records `job_recovering` and retry scheduling.
- The run either completes safely or reaches a redacted failed terminal state after attempts are exhausted.
- Partial prior output does not duplicate final assistant history.

## 7. Validate Cancellation

Run both cancellation paths:

1. Request stop while a run is queued.
2. Request stop while a run is already owned by a worker.

Expected outcome:

- Queued stop prevents normal execution from starting and records `run_stopped`.
- Running stop is observed at a safe boundary and records `run_stopped`, unless the run is already terminal.
- A terminal run does not change to a conflicting terminal state.

## 8. Validate Diagnostics

Read local worker queue diagnostics.

Expected outcome:

- Ready, paused, unhealthy, degraded, stale, retrying, and dead states are distinguishable where applicable.
- Counts are consistent with the jobs created during smoke tests.
- Diagnostics do not expose provider credentials or raw secret values.

## 9. Browser Smoke

Use the web shell against the real local API.

Expected outcome:

- Starting a message-driven run shows queued/running progress quickly.
- Timeline shows queue, worker, recovery/cancellation, and terminal events.
- Refreshing the browser recovers the timeline from persisted history.
- Real API mode does not silently fall back to mock output when queue processing fails.

## 10. Validation Commands

Run the relevant validation before claiming implementation complete:

```sh
go test ./...
bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

If any command cannot run, record the exact reason in the implementation report.

## 11. Rollback/Reapply Smoke

```sh
go run ./cmd/loomi-api --migrate down --to 000004
go run ./cmd/loomi-api --migrate up
```

Expected outcome:

- M6 queue additions roll back and reapply cleanly in local development.
- Earlier M3/M4/M5 thread, message, run, event, and provider data remain in the state expected by migration `000004`.
