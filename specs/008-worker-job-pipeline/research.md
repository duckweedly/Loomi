# Research: M6 Worker Job Pipeline

## Decision: Use a durable database-backed job queue for the M6 vertical slice

**Decision**: M6 will store background jobs durably alongside existing product run data and process them through local worker loops instead of keeping jobs in memory or adding an external queue service.

**Rationale**: The constitution requires runnable vertical slices and observable execution before platform breadth. Existing M4/M5 execution data is already persisted and history-first; a durable local job queue preserves reconnect/recovery semantics, supports crash recovery, and avoids introducing infrastructure that would obscure the product contract.

**Alternatives considered**:

- In-memory worker queue: rejected because browser disconnects or process interruption would lose work and fail the recovery requirements.
- External queue service: rejected because M6's first slice is local-development scoped and should not add operational dependencies before the product job semantics are proven.
- Reusing only run status as the queue: rejected because worker ownership, retry attempts, scheduling, stale ownership, and recovery need a dedicated work record separate from the user-visible run.

## Decision: Keep the initial worker model inside the local API process, with clear worker boundaries

**Decision**: M6 will introduce worker execution as a local runtime boundary that can run in the API process for the first slice while keeping job claim, lease renewal, cancellation, and processing logic separated from HTTP handlers.

**Rationale**: The current app is local-first and already runs request-scoped runtime work in the backend. Starting workers in-process gives a demonstrable vertical slice with fewer moving parts, while separate worker/job interfaces leave room for a future standalone worker process without rewriting run/event or API contracts.

**Alternatives considered**:

- Separate worker binary immediately: rejected because it increases setup, readiness, and lifecycle complexity before the queue semantics are validated.
- Running job work directly from HTTP handlers: rejected because it would preserve the request-scoped behavior M6 is meant to replace.
- Browser-driven background work: rejected because it cannot satisfy durable recovery or credential/data-boundary requirements.

## Decision: Claim jobs through time-bounded worker leases

**Decision**: Workers will claim jobs by acquiring a time-bounded lease that records the worker owner and expiry. Active workers renew leases while processing; stale leases make unfinished jobs eligible for recovery.

**Rationale**: Lease ownership makes concurrent workers safe, allows interrupted workers to be detected without hard process coordination, and maps directly to the spec's recovery and duplicate-prevention requirements.

**Alternatives considered**:

- Static worker ownership until completion: rejected because crashed workers would leave jobs stuck indefinitely.
- Global single worker lock: rejected because it prevents validating multi-worker claim safety and future parallelism.
- No ownership record: rejected because duplicate processing could produce conflicting terminal events and duplicate assistant messages.

## Decision: Enforce idempotency at job, run, event, and final-message boundaries

**Decision**: M6 will treat job execution as at-least-once and protect user-visible effects with idempotency rules: one active job per active run, monotonic run event sequencing, one terminal event per run, and at most one final assistant message for a completed run.

**Rationale**: Recovery and duplicate delivery are normal queue behaviors. The user-visible product contract must remain exactly-once for terminal outcomes even if processing is retried or a worker loses ownership mid-step.

**Alternatives considered**:

- Assume exactly-once worker processing: rejected because crashes, stale leases, and concurrent claim attempts are explicit M6 edge cases.
- Store partial assistant drafts as durable messages during each attempt: rejected because retries could duplicate or reorder visible conversation history.
- Hide duplicate events in the frontend only: rejected because persisted execution history must be the source of truth.

## Decision: Model cancellation as a persisted stop request observed at safe boundaries

**Decision**: M6 cancellation will persist a stop request for the run/job. Queued jobs with stop requested will not start normal execution; running jobs will observe the request at safe processing boundaries and write a stopped terminal outcome.

**Rationale**: M4 already established cooperative stop. M6 extends that truthfully to background work without pretending to provide hard interruption for provider calls or future tool execution that cannot always be stopped immediately.

**Alternatives considered**:

- Hard-kill running work: rejected because the first worker slice should not depend on process termination, unsafe interruption, or external provider cancellation guarantees.
- Mark stopped without worker acknowledgement: rejected because it can conflict with an already-running worker and produce later output after a terminal state.
- Ignore stop for queued jobs until a worker picks them up: rejected because queued cancellation should prevent unnecessary execution.

## Decision: Add minimal pipeline steps as observable execution stages, not a full orchestration engine

**Decision**: M6 will represent pipeline execution as a small set of named stages needed for background run processing: enqueue, claim, prepare context, invoke runtime work, finalize, and recover or fail. It will not introduce arbitrary DAGs, plugins, desktop actions, or multi-agent orchestration.

**Rationale**: The staged roadmap calls for Worker, Job Queue, and Pipeline, but the constitution requires core flow before platform complexity. A linear stage model makes background execution observable and testable without pulling M7+ orchestration concerns forward.

**Alternatives considered**:

- Full configurable pipeline DAG: rejected because tool execution, memory, sandbox, channels, and multi-agent dependencies are not ready.
- No pipeline concept: rejected because users and developers need timeline/debug visibility into why a background run is pending, running, recovering, failed, or complete.
- Frontend-only pipeline labels: rejected because execution observability must come from persisted backend events.

## Decision: Extend existing run/event APIs and frontend runtime adapter instead of adding a parallel job UI

**Decision**: M6 will expose background execution primarily through existing run/event/history-first stream behavior, adding queue and worker states to run events and diagnostics. The web shell should continue to render through the existing Chat Canvas, Run Timeline, Run Rail, and runtime adapter seams.

**Rationale**: M4/M5 made runs and persisted events the user-visible execution contract. A separate job UI would split the mental model and duplicate state. Existing adapter boundaries can show queued/recovering/cancelled/failed states without replacing the product shell.

**Alternatives considered**:

- New worker dashboard as the main M6 UI: rejected because it would prioritize operational chrome over the user's thread/run experience.
- Hidden queue with no timeline events: rejected because it violates observable agent execution.
- Replace the run/event contract with job status polling: rejected because history-first streaming and persisted timelines are existing product contracts.

## Decision: Add local diagnostics for queue and worker readiness

**Decision**: M6 will extend local readiness/diagnostics so developers can distinguish ready, paused, unhealthy, and degraded worker/queue states without exposing secrets.

**Rationale**: Background execution introduces failure modes that are hard to understand from run status alone. Minimal diagnostics keep the slice operable while avoiding a broader admin console.

**Alternatives considered**:

- No diagnostics beyond run failures: rejected because stalled workers and paused queues need clear validation feedback.
- Full admin console: rejected because it belongs to later platform capabilities.
- Logging raw job/provider payloads: rejected because credentials and user data must remain bounded and redacted.

## Decision: Use migration 000005 for background job persistence

**Decision**: M6 will add a new migration after M5, tentatively `000005_m6_worker_job_pipeline`, for job, worker lease, and any run/status extensions required by the background execution model.

**Rationale**: M5 already used migration 000004 for assistant messages and model-gateway runs. M6 needs durable queue data and must remain rollback/testable through the existing explicit migration workflow.

**Alternatives considered**:

- Modifying older migrations: rejected because existing milestones rely on stable migration history.
- No migration: rejected because recovery, lease, retry, and idempotency state must survive process interruption.
- Combining M6 with settings placeholder data: rejected because the user explicitly requested M6 not be mixed with the current 007 settings work.
