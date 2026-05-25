# Research: M8 Worker Job Closeout

## Decision: Use M6 Worker Job Pipeline as the M8 evidence base

**Rationale**: The original M8 roadmap items map directly to `specs/008-worker-job-pipeline`, migration `000005_m6_worker_job_pipeline`, productdata repository/service methods, runtime worker/job coordinator, and M6 docs.

**Alternatives considered**:

- Create a new M8 worker implementation: rejected because it would duplicate the existing durable queue.
- Defer closeout to M9: rejected because M9 should start from a clean worker/job baseline.

## Decision: Treat retry/backoff as the only behavior gap

**Rationale**: Current code already has retry events, attempt counting, max attempts, terminal failure, and stale-owner protection. Recovery previously made jobs immediately claimable again, so "backoff" was not represented as a delay.

**Alternatives considered**:

- Accept event-only retry scheduling: rejected because the original M8 wording includes "retry/backoff".
- Add a full configurable retry policy: rejected as platform complexity beyond closeout.

## Decision: Use existing `scheduled_at` for local retry backoff

**Rationale**: `background_jobs` already has `scheduled_at`, and claim already requires `scheduled_at <= now()`. Updating recovery to push `scheduled_at` forward is the smallest behavior change.

**Alternatives considered**:

- Add a new retry table: rejected as unnecessary.
- Change worker loop timing: rejected because claim scheduling is already the durable boundary.

## Decision: No frontend or browser smoke needed

**Rationale**: The patch changes backend scheduling only. Frontend states and event mapping already support queued/recovering/retrying worker events.

**Alternatives considered**:

- Run browser smoke anyway: not required by the user unless frontend state changes are involved.
