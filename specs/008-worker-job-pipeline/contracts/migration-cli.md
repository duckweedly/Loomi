# Contract: M6 Migration and Local Validation CLI

## Migration

M6 adds explicit schema migration `000005_m6_worker_job_pipeline`.

Required apply behavior:

- Adds durable background job state required for queueing, leasing, retry, recovery, cancellation, and diagnostics.
- Preserves existing M3 thread/message, M4 run/event, and M5 model-gateway data.
- Does not insert demo jobs or runs.
- Can be applied after migration `000004`.

Required rollback behavior:

- Rolls back M6 job/worker-queue additions cleanly.
- Does not rewrite older migrations.
- Leaves earlier M3/M4/M5 data in the state expected by migration `000004`.
- Must be safe for local development rollback/reapply validation.

## Readiness

M6 readiness should distinguish:

- Schema unavailable or below required version.
- Queue storage unavailable.
- Worker disabled intentionally.
- Worker unhealthy unexpectedly.
- Queue degraded because stale or dead jobs need attention.
- Queue ready and worker ready.

Readiness output must be redacted.

## Local Validation Commands

The implementation quickstart must include commands or smoke steps for:

1. Applying migrations through `000005`.
2. Verifying readiness succeeds after M6 migration.
3. Starting the local API with worker processing enabled.
4. Starting the local API with worker processing paused or disabled.
5. Creating a run and observing queued acknowledgement.
6. Reading queue diagnostics.
7. Simulating or forcing worker interruption for recovery validation.
8. Running rollback/reapply validation for migration `000005`.
9. Running backend, frontend, and docs validation commands.

## Exit Expectations

Smoke or validation scripts should fail clearly when:

- A run start blocks until execution completion instead of returning queued/running acknowledgement.
- A thread can create two active runs.
- Two workers can both complete the same job.
- A retry creates duplicate terminal events or duplicate assistant messages.
- A queued stop request allows normal execution to start.
- Diagnostics expose secrets or raw provider credentials.
