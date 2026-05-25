# Contract: M8 Closeout Audit

## Audit Matrix

| Original M8 item | Closeout result | Evidence |
| --- | --- | --- |
| `jobs` table | Covered | `migrations/000005_m6_worker_job_pipeline.up.sql` creates `background_jobs` with job lifecycle, lease, attempt, schedule, and ownership fields. |
| API creates run and enqueues `run.execute` in one transaction | Covered | `internal/productdata/repository.go` `StartRun` creates queued run, `run_queued` event, and `background_jobs` row in one transaction. |
| Worker claims pending job | Covered | `ClaimBackgroundJob` uses `for update skip locked`, `status='queued'`, and `scheduled_at<=now()`. |
| Lease heartbeat / renew | Covered | `RenewBackgroundJobLease` updates lease expiry only for current `worker_id` and `ownership_version`; `LocalRunner` renews before writes. |
| Retry/backoff | Patched | Recovery now moves `scheduled_at` into the future and exposes that retry time in `job_retry_scheduled` metadata. |
| Failed terminal | Covered | Recovery exhaustion marks job `dead`, run `failed`, and records `job_retry_exhausted` plus `run_failed`. |
| Lost-lock ownership guard | Covered | Complete/fail/renew SQL conditions require `leased_by`, `ownership_version`, and `status='leased'`; stale tests assert no terminal write. |
| API create run immediately returns | Covered | `internal/httpapi/runtime.go` returns `202 Accepted` after `StartRun`. |
| Worker crash can recover job | Covered | `RecoverBackgroundJobs` detects expired leases, requeues retryable work, and fails exhausted work. |
| Old worker cannot write completed/failed/retry after losing lock | Covered | Existing stale-owner tests plus ownership-guarded complete/fail paths prevent conflicting terminal writes. |

## Closeout Rules

- Do not implement another worker queue.
- Do not change the M7 tool-call continuation boundary.
- Do not start M9 RunContext/Pipeline work.
- Documentation must state that original M8 passed after this closeout patch.
