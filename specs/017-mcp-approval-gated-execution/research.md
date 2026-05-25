# Research: MCP Approval-Gated Execution

## Decision: Reuse M7 approval as the only execution entry

**Rationale**: M7 already provides scoped tool-call projection, approve/deny, idempotent decisions, run status, run events, worker resume, and single continuation semantics. MCP execution must be a new tool source inside that lifecycle, not a separate permission path.

**Alternatives considered**:

- Direct MCP execution from ToolRegistry: rejected because it bypasses approval and audit.
- Separate MCP approval system: rejected because it duplicates M7 and increases safety surface before the minimal slice proves value.
- Persona allowlist as approval: rejected because persona configuration authorizes availability, not per-run execution.

## Decision: Require both discovered candidate and persona allowed-tools reference

**Rationale**: M11 discovery proves a local stdio candidate exists and maps to a stable namespaced ToolSpec. M10 persona allowed-tools constrains which candidates the active persona may use. M12 needs both before presenting approval.

**Alternatives considered**:

- Allow any discovered MCP candidate to request approval: rejected because persona scope would be ignored.
- Allow persona entries before discovery to create pending approvals: rejected because unknown tools cannot be safely resolved to executable process config.

## Decision: Keep M12 at one MCP tool execution per run

**Rationale**: Existing M7 continuation is one-tool-call oriented, and the requested slice explicitly asks for a minimal single continuation rather than a multi-tool loop. One execution per run keeps idempotency, worker recovery, and provider serialization tractable.

**Alternatives considered**:

- Multi-step loop: rejected as deferred platform behavior.
- Multiple approved MCP calls in one run: rejected until ordering, partial failure, and per-call idempotency semantics are designed.

## Decision: Treat execution as at-most-once after start in this slice

**Rationale**: Local stdio MCP tools may have side effects even if described as read-only. If a worker crashes after process startup, M12 should surface recovery as failed/unknown rather than blindly re-executing.

**Alternatives considered**:

- Retry after crash automatically: rejected because MCP side effects are unknown.
- Require tool-declared idempotency keys: deferred until a richer safety class model exists.

## Decision: Mark execution started before stdio invocation

**Rationale**: Persisting `execution_started` before process startup lets recovery detect that a side effect may already have occurred. This favors safety over automatic completion after ambiguous crashes.

**Alternatives considered**:

- Mark started after process response: rejected because crash during invocation would look safe to retry.
- Avoid a started state: rejected because worker recovery needs a durable boundary.

## Decision: Redact arguments, results, errors, and process details before persistence and continuation

**Rationale**: MCP inputs/outputs can include secrets, local paths, file contents, shell output, or arbitrary server text. Redaction must occur before run-event persistence, UI replay, diagnostics, and provider continuation.

**Alternatives considered**:

- Store raw output in DB and redact in UI: rejected because history replay and provider continuation would have unsafe source data.
- Send raw result to provider but redact UI: rejected because provider continuation is also an external boundary.

## Decision: Bound stdio process lifecycle per approved call

**Rationale**: Local stdio MCP execution needs timeout, cleanup, exit classification, stderr redaction, and cancellation handling. The process must not expose raw command, args, env, stdout, or stderr through normal product surfaces.

**Alternatives considered**:

- Keep a long-lived MCP server process pool: rejected as lifecycle complexity beyond the minimal slice.
- Execute through shell wrappers: rejected because shell/filesystem automation and command leakage are out of scope.

## Decision: Continue provider exactly once with redacted MCP result

**Rationale**: M7 already defines provider-neutral tool-result continuation. M12 should serialize the MCP result through that channel and fail safely if the provider requests another tool.

**Alternatives considered**:

- No continuation after MCP result: rejected because the user-facing loop would not produce the final answer.
- Arbitrary loop until provider stops: rejected as deferred multi-tool behavior.

## Decision: Defer remote/server administration/sandbox features

**Rationale**: Remote MCP, OAuth, marketplace/plugin install, DB-managed MCP servers, admin UI, complex sandbox, and shell/filesystem/browser automation all require separate trust and permission models. M12 is only the execution bridge for already-discovered local stdio candidates.

**Alternatives considered**:

- Combine execution with admin UI or DB-managed config: rejected because it expands storage, permissions, and UX beyond the approval-gated loop.
- Add sandbox now: rejected because the current requirement is to define no complex sandbox and to keep the slice small.
