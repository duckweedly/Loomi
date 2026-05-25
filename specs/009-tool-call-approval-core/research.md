# Research: M7 Tool Call Approval Core

## Decision: Use a minimal allowlisted internal tool registry

**Rationale**: M7 needs to cross from non-executed tool boundary to safe execution without introducing a general plugin system. A fixed allowlist lets Loomi validate tool names, schemas, approval policy, and redaction policy before any execution path can run.

**Alternatives considered**:

- Provider-native tool execution passthrough: rejected because provider payloads and arguments are untrusted and would blur Loomi's safety boundary.
- Generic plugin/MCP registry: rejected because MCP and plugin execution are explicit non-goals for M7.
- Shell/file/network tools: rejected because they introduce side effects, user-file exposure, arbitrary network egress, and sandbox requirements that are outside the first approval-core slice.

## Decision: Pick `runtime.get_current_time` as the MVP tool

**Rationale**: `runtime.get_current_time` is the safest useful executable tool for proving the lifecycle. It has no side effects, requires no user data, does not access files, does not call arbitrary URLs, does not depend on worker diagnostics, and can return a tiny redacted result such as an ISO timestamp plus timezone/source metadata.

**Alternatives considered**:

- `thread.list_recent_messages`: useful for context inspection, but it exposes user conversation content and requires pagination, ownership checks, content redaction rules, and UI disclosure before execution.
- `diagnostics.get_worker_queue`: useful for developers, but it exposes operational state and could leak implementation details; it belongs after the approval core can handle diagnostic redaction confidently.

## Decision: Persist a minimal `tool_calls` projection in addition to run events

**Rationale**: Run events remain the source of user-visible audit and history-first SSE. However, approve/deny idempotency, worker block/resume lookup, and execution ownership need a current-state projection that can be atomically constrained. A minimal `tool_calls` table or equivalent durable projection avoids deriving mutable state by scanning event history for every decision.

**Alternatives considered**:

- Run events only: simpler storage, but idempotent approve/deny and worker resume would require fragile event replay and duplicate-protection logic under concurrency.
- Full invocation/job table hierarchy: rejected as too broad for MVP because M6 already has job/lease state and M7 only needs one current-state projection per tool call.

## Decision: Model approval wait as a worker-blocked run state, not a new queue system

**Rationale**: M6 already provides durable background jobs, lease, recovery, cancellation, and diagnostics. M7 should add the minimal interface for a worker to record `approval_required`, release or pause execution, and resume after an idempotent approval decision schedules or wakes the existing job pipeline.

**Alternatives considered**:

- Busy-wait worker lease while waiting for approval: rejected because it wastes leases and makes approval delays look like worker stalls.
- Separate tool execution queue: rejected because it duplicates M6 background jobs for the first slice.
- Synchronous HTTP execution after approve: rejected because tool execution should remain observable through the same run/event/worker pipeline.

## Decision: Require schema validation before approval and execution

**Rationale**: Model-generated arguments are untrusted input. Validation before approval gives users a stable safe summary to review and prevents malformed payloads from entering the execution path.

**Alternatives considered**:

- Validate only during execution: rejected because users would approve untrusted raw data and UI summaries could leak invalid or sensitive content.
- Trust provider schemas: rejected because provider output is external data and may be malformed, adversarial, or mismatched to Loomi's internal tool contract.

## Decision: Keep full multi-step tool-result model continuation out of MVP

**Rationale**: M7's purpose is the approval and audited execution boundary. Feeding tool results into another provider request requires loop limits, context budgeting, repeated approval handling, and terminal-state semantics. M7 should define the boundary and may stop after recording the tool result.

**Alternatives considered**:

- Implement full tool-result continuation immediately: rejected as too broad and likely to pull in multi-agent/loop complexity.
- Ignore continuation entirely: rejected because result events should be shaped so a later milestone can feed safe tool results into model context without changing the audit contract.

## Decision: Use distinct tool event types and UI grouping

**Rationale**: Tool lifecycle transitions are safety-significant and must be distinguishable from model text streaming. Distinct event types make history replay, Timeline grouping, RunRail summaries, and audits clear.

**Alternatives considered**:

- Reuse `model_output_delta` or generic progress events: rejected because it would mix user-visible generated text with approval and execution state.
- Store tool status only in UI state: rejected because reconnect/history-first SSE would lose the audit trail.

## Decision: Redact all persisted arguments, results, and errors

**Rationale**: Even the safe MVP tool should establish the invariant that events and UI summaries are redacted by default. This prevents future tools from accidentally leaking API keys, provider payloads, user file paths, or raw operational details.

**Alternatives considered**:

- Redact only high-risk tools: rejected because it makes safety depend on per-tool discipline.
- Store raw payloads for debugging: rejected because the user's safety requirements explicitly forbid raw provider payloads and secret leakage in events.
