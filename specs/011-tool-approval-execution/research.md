# Research: M7 Tool Approval Execution Closure

## Decision: Keep approve/deny state transitions in `internal/productdata`

**Rationale**: Idempotency, scoped reads, event writes, and terminal guards need the same transaction boundary as the `tool_calls` projection and run events.

**Alternatives considered**: HTTP-only state handling was rejected because route retries could duplicate events. Worker-owned approval handling was rejected because user decisions must be recorded immediately without waiting for a worker lease.

## Decision: Denial finalizes MVP run as stopped

**Rationale**: M7 does not implement a multi-step model continuation after a denied tool. A visible stopped run with `tool.call.denied` gives the user a clear terminal state and avoids hidden retries.

**Alternatives considered**: Continuing model execution with synthetic tool denial is deferred to a later agent loop phase.

## Decision: Approval resumes the existing M6 worker/job pipeline

**Rationale**: The foundation already models `blocked_on_tool_approval`. Approval should make the run claimable/resumable once, not create a second queue.

**Alternatives considered**: A dedicated tool executor queue was rejected as platform complexity for one no-side-effect MVP tool.

## Decision: Execute only `runtime.get_current_time`

**Rationale**: It has no shell, file, network, MCP, browser, user-data, or secret access and exercises the approval/execution/result path.

**Alternatives considered**: Diagnostics and thread-message tools expose operational or user data and need stricter disclosure policy.

## Decision: Persist redacted summaries only

**Rationale**: Run events and UI are audit surfaces. They should show enough to understand the execution without storing raw provider payloads or secrets.

**Alternatives considered**: Persisting full JSON result was rejected because future tools may carry sensitive data.
