---
title: M8 Worker Job Closeout
description: Audit result and validation for closing the original M8 Worker + Job Queue roadmap scope.
---

## Result

M8 closeout passed. The original M8 Worker + Job Queue scope is covered by `specs/008-worker-job-pipeline` plus the 013 closeout patch.

## Audit evidence

| Original M8 target | Result |
| --- | --- |
| `jobs` table | Covered by `background_jobs` in migration `000005_m6_worker_job_pipeline`. |
| API creates run and enqueues work in one transaction | Covered by `StartRun`, which creates queued run, `run_queued`, and `background_jobs` inside one transaction. |
| worker claim pending job | Covered by `ClaimBackgroundJob` with `for update skip locked` and `scheduled_at <= now()`. |
| lease heartbeat / renew | Covered by `RenewBackgroundJobLease` and runner-side renewal before worker writes. |
| retry/backoff | Patched in 013 by scheduling recovered jobs in the future and emitting that `scheduled_at` in `job_retry_scheduled`. |
| failed terminal | Covered by retry exhaustion to `dead` job and failed run with `job_retry_exhausted` + `run_failed`. |
| lost-lock ownership guard | Covered by complete/fail/renew requiring current worker id and ownership version. |
| API create run immediately returns | Covered by `POST /v1/threads/{thread_id}/runs` returning `202 Accepted`. |
| worker crash recovery | Covered by stale lease recovery and retry exhaustion paths. |
| old worker cannot write terminal after losing lock | Covered by stale-owner tests and ownership-guarded terminal updates. |

## Patch

The only behavior-level gap was retry backoff. Recovery already emitted retry scheduling and enforced max attempts, but the recovered job stayed immediately claimable. The 013 patch moves recovered jobs' `scheduled_at` into the future with a short bounded backoff in both the PostgreSQL repository and in-memory service.

No M9 RunContext/Pipeline work was added. No Redis, external queue, multi-worker platform layer, MCP, Memory, Desktop Runtime, or frontend behavior was introduced.

## Validation

Run from the repository root:

```bash
go test ./internal/productdata ./internal/runtime
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
bun run --cwd docs-site build
```

Result on 2026-05-25: all three commands passed.

Browser smoke was not required because this closeout changed backend retry scheduling and docs only; frontend state and UI flow were unchanged.

Final closeout validation on 2026-05-25 also ran:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
bun run --cwd web build
bun run --cwd docs-site build
```

Result: all three commands passed. Browser smoke remains not required for M8 because 013 changed retry scheduling and documentation only.
