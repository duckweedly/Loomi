# Data Model: RunContext Pipeline Foundation

## RunContext

Prepared execution context for one run.

Fields:

- `run`: run id, thread id, user id, status, source, created/updated timestamps, stop state
- `thread`: thread id, user id, title/summary fields needed for ownership checks
- `messages`: ordered conversation messages needed for provider/runtime context
- `job`: background job id, kind, attempt, ownership version, safe metadata
- `providerRoute`: provider id, model id/override, runtime source, availability summary
- `enabledTools`: allowlisted MVP tool names and approval/execution availability
- `continuationProjection`: optional tool-result continuation facts derived from persisted events when resuming after approved tool execution

Validation rules:

- The run, thread, and job must refer to the same user/thread/run boundary.
- Messages must be ordered by persisted chronology.
- Provider/model route must be present for model-gateway runs and may be absent only for supported local simulated paths.
- Enabled tools may be empty, but when non-empty must be allowlisted.
- Raw provider payloads, credentials, raw tool results, and unredacted user-controlled blobs are not RunContext summary fields.

## ContextSource

The durable records used to build RunContext.

Sources:

- run row
- thread row
- message history
- background job row and safe metadata
- provider/model run metadata
- tool-call projection and run events needed for current-time continuation

## Pipeline Stage

A named linear execution step.

MVP stages:

- `prepare_context`
- `resolve_tools`
- `invoke_runtime`
- `finalize`

Stage contract:

- receives pipeline state
- may add safe summarized state
- records started/completed/failed trace events
- returns success or redacted failure

## Pipeline State

In-memory state passed between stages during one worker-owned job attempt.

Fields:

- `runID`, `threadID`, `jobID`, `workerID`, `attempt`, `ownershipVersion`
- optional `RunContext`
- optional `ToolResolutionSummary`
- optional `RuntimeInvocationSummary`
- terminal outcome/finalization summary

Lifecycle:

```text
empty
-> context prepared
-> tools resolved
-> runtime invoked
-> finalized
```

Cancellation, ownership loss, or stage failure can terminate the flow at safe boundaries.

## Pipeline Trace Event

Persisted run event that explains a stage transition.

Fields:

- event type: `pipeline_stage_started`, `pipeline_stage_completed`, or `pipeline_stage_failed`
- stage name
- run id, thread id, job id, attempt
- safe summary
- redacted error code/message when failed

Metadata restrictions:

- Allowed: ids, stage name, counts, booleans, safe status labels, allowlisted tool names
- Forbidden: secrets, raw Authorization headers, raw provider requests/responses, raw tool result payloads, shell output, file contents, hidden local state

## Tool Resolution Summary

Safe stage output describing available MVP tools for this run.

Fields:

- enabled tool names, currently limited to `runtime.get_current_time`
- approval policy summary
- whether a resumable approved tool call is present
- unsupported tool names only as redacted counts or safe names when already allowlisted

## Runtime Invocation Summary

Safe stage output indicating the existing runtime/model boundary was invoked.

Fields:

- runtime source
- provider id/model id when safe
- model phase when applicable
- continuation flag when applicable
- terminal outcome reference

## Stage Failure

Redacted failure returned by a stage.

Fields:

- stage name
- stable error code
- safe user-facing message
- retryable flag only when existing worker retry behavior applies

Rules:

- Missing required context fails before runtime invocation.
- Provider/tool/runtime failures preserve existing M7/M8 event semantics.
- A stage failure must not create duplicate terminal events.
