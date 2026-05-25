# Data Model: MCP Approval-Gated Execution

## MCP Tool Execution Request

Provider-requested tool call that names an M11 discovered local stdio MCP candidate.

Fields:

- `threadID`
- `runID`
- `providerToolCallID`
- `toolSpecName`: `mcp.<server_slug>.<tool_name>`
- `candidateSchemaHash`
- `personaSnapshotID`
- `argumentsSummary`
- `argumentsHash`
- `requestedAt`

Validation rules:

- Tool name must be namespaced and resolved to an enabled, discovered M11 candidate.
- Active persona allowed-tools snapshot must include the namespaced tool.
- Request is rejected if the candidate is unavailable, disabled, ambiguous, unnamespaced, internal-tool-conflicting, or schema-incompatible.
- Arguments are untrusted data and must be validated and redacted before projection persistence.

## Scoped MCP Tool-Call Projection

Existing M7 tool-call projection extended or reused for MCP source.

Fields:

- `id`
- `threadID`
- `runID`
- `userID` or local identity scope where available
- `providerToolCallID`
- `toolName`
- `toolSource`: `mcp`
- `serverSlug`
- `candidateSchemaHash`
- `argumentsSummary`
- `argumentsHash`
- `approvalStatus`: required, approved, denied
- `executionStatus`: blocked, not_started, started, succeeded, failed, cancelled
- `requestedAt`, `approvedAt`, `deniedAt`, `startedAt`, `completedAt`, `updatedAt`

State transitions:

```text
requested -> approval_required/blocked
approval_required/blocked -> approved/not_started
approval_required/blocked -> denied/cancelled
approved/not_started -> started
started -> succeeded
started -> failed
approved/not_started -> cancelled
```

Rules:

- Projection identity is stable for `(runID, providerToolCallID)`.
- Repeated request, approve, deny, or worker resume must reuse the current projection.
- Terminal execution states must not return to executable states in M12.

## MCP Execution Attempt

One worker-owned attempt to execute an approved MCP tool call.

Fields:

- `projectionID`
- `runID`
- `jobID`
- `workerID`
- `leaseToken` or equivalent ownership evidence
- `attemptNumber`
- `startedAt`
- `completedAt`
- `timeoutMs`
- `status`: started, succeeded, failed, cancelled, ownership_lost
- `resultSummary`
- `errorSummary`

Rules:

- The worker must prove ownership before marking started and before process startup.
- `started` is persisted before invoking the stdio process.
- Retry/recovery must not invoke again after `started`, `succeeded`, `failed`, or `cancelled`.
- If recovery sees `started` without terminal outcome, M12 treats the call as unsafe to re-execute and records a redacted recovery failure or cancellation state.

## MCP Stdio Process Invocation

Bounded local process lifecycle for a single approved call.

Fields:

- `serverSlug`
- `toolName`
- `timeoutMs`
- `startedAt`
- `exitStatus`
- `durationMs`
- `stderrClass`
- `cleanupStatus`

Forbidden persisted fields:

- raw command path
- raw args
- env values
- raw stdout/stderr
- tokens, credentials, Authorization headers
- secret-looking paths
- file contents
- shell output
- browser or desktop captured state

Lifecycle:

```text
validate safe candidate and approval
-> verify worker ownership and run state
-> mark execution started
-> start local stdio process for one tool call
-> send validated MCP tool-call request
-> wait for response or timeout/cancel/exit
-> cleanup process
-> persist redacted success/failure
```

## MCP Tool Result Summary

Redacted result or error safe for persistence, UI replay, and provider continuation.

Fields:

- `toolCallID`
- `toolName`
- `serverSlug`
- `status`: succeeded, failed, cancelled
- `resultSummary`
- `resultHash`
- `errorCode`
- `safeMessage`
- `retryable`: false by default for post-start MCP execution in M12
- `completedAt`

Rules:

- Raw MCP result payload is never persisted as normal event metadata.
- Result summary must be size-bounded and schema-safe.
- Error summary must classify timeout, early exit, invalid JSON, unsafe result, cancelled, ownership lost, and process cleanup failures without raw stderr.

## MCP Continuation Context

Provider-neutral input for one continuation after an approved MCP result.

Fields:

- `runID`
- `threadID`
- `providerToolCallID`
- `assistantToolCall`
- `toolResult`
- `modelPhase`: `continuation`
- `continuationAttempt`: 1

Rules:

- The tool result must match the scoped projection.
- Only redacted `result_for_model` or safe result summary is eligible.
- Denied, failed, cancelled, or stopped tool calls do not produce continuation.
- Any continuation tool request fails with an unsupported tool-loop event and no execution.

## MCP Audit Event

Persisted run-event metadata for the approval/execution lifecycle.

Event types reuse existing M7 tool-call and run lifecycle events with `tool_source = mcp` metadata where needed:

- `tool_call_requested`
- `tool_call_approval_required`
- `tool_call_approved`
- `tool_call_denied`
- `tool_call_executing`
- `tool_call_succeeded`
- `tool_call_failed`
- `tool_call_cancelled`
- `run_failed` with `error_code = unsupported_tool_loop` for additional continuation tool requests
- existing continuation model-output events with `metadata.model_phase = continuation`

Rules:

- Events must be replayable from live SSE and history.
- Metadata must be redacted before persistence.
- Events should include stable IDs, `tool_source`, and statuses sufficient for Timeline/debug without raw process data.
