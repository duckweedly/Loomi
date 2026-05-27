# Feature Specification: MCP Approval-Gated Execution

**Feature Branch**: `[017-mcp-approval-gated-execution]`

**Created**: 2026-05-25

**Status**: Implemented

**Input**: User description: "M12 MCP Approval-Gated Execution. M12 covers how already-discovered local stdio MCP tools enter M7 approval, tool-call projection, audit, worker, run-event, and provider continuation boundaries, then execute one minimal safe loop. It is not remote MCP, marketplace/plugin install, shell/filesystem/browser automation, automatic execution, or complex sandbox."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Gate discovered MCP tool execution through M7 approval (Priority: P1)

As a Loomi user supervising an agent run, I want any model-requested MCP tool call to stop at the existing M7 approval state before execution, so that a discovered local stdio tool cannot run automatically or bypass audit.

**Why this priority**: Approval is the safety boundary. Without this story, execution would undercut M7 and the M11 non-executable foundation.

**Independent Test**: Run a model/tool-call fixture that requests a namespaced MCP candidate and verify the run records a scoped approval-required tool call, blocks execution, redacts arguments, and executes nothing until approval.

**Acceptance Scenarios**:

1. **Given** a local stdio MCP tool was discovered by M11 and is referenced by the selected persona allowed-tools, **When** the provider requests that namespaced MCP tool, **Then** Loomi creates or reuses one scoped tool-call projection with approval required and execution blocked.
2. **Given** the provider requests an undiscovered, disabled, unlisted, unnamespaced, or persona-disallowed MCP tool, **When** tool resolution runs, **Then** Loomi rejects the request with a redacted run event and no approval action is offered.
3. **Given** the same provider tool request is replayed or retried for the same run and tool-call id, **When** projection is recorded, **Then** Loomi returns the existing projection and does not duplicate approval/audit events.
4. **Given** approval is denied, **When** the run resumes or is replayed, **Then** no MCP stdio process is started and the denial is visible through live SSE and history replay.

---

### User Story 2 - Execute one approved local stdio MCP tool safely in the worker (Priority: P2)

As a Loomi developer validating M12, I want the worker to execute exactly one approved local stdio MCP tool under ownership, lease, cancellation, and recovery guards, so that the minimal execution path is demonstrable without duplicate side effects.

**Why this priority**: After approval, the worker must prove a safe, observable execution handoff. This is the smallest runnable M12 slice.

**Independent Test**: Approve one valid MCP tool call, run the worker, and verify it starts the configured local stdio server only for that call, records execution start/success/failure once, redacts arguments/result/error, and does not re-execute after lost lease, retry, cancellation, or replay.

**Acceptance Scenarios**:

1. **Given** an approved MCP tool call with execution not started, **When** the owning worker leases the run/job, **Then** it marks execution started once before invoking the MCP stdio call.
2. **Given** the worker loses ownership, the run is stopped, or cancellation is requested before MCP invocation, **When** the worker checks execution readiness, **Then** it does not start the stdio process and records only safe cancellation/ownership events.
3. **Given** an MCP stdio process times out, exits early, writes stderr, returns invalid JSON, or returns an error response, **When** the worker records the outcome, **Then** persisted events include only redacted error codes/messages and no raw env, args, stdout, stderr, command path, tokens, secret paths, file contents, shell output, or browser/desktop captured state.
4. **Given** a worker retries after crash/recovery and the tool-call projection already has started/succeeded/failed/cancelled terminal state, **When** retry handling runs, **Then** Loomi does not invoke the MCP tool again.

---

### User Story 3 - Continue the provider once with the redacted MCP tool result (Priority: P3)

As a Loomi user watching the Timeline, I want an approved MCP tool result to be passed back to the provider for one final continuation, so that the run can produce an assistant answer without opening a multi-tool loop.

**Why this priority**: The visible value of execution is the continuation answer, but it must stay bounded to one continuation and reuse M7's existing tool-result model.

**Independent Test**: Complete one approved MCP execution and verify provider continuation receives only the matching redacted MCP tool result, emits continuation events, writes one final assistant message, and fails safely if the continuation asks for another tool.

**Acceptance Scenarios**:

1. **Given** an approved MCP execution succeeds, **When** continuation context is built, **Then** it includes the original assistant tool request and the matching redacted MCP tool result for the same scoped tool-call id.
2. **Given** the continuation provider produces a final assistant response, **When** the run finalizes, **Then** Timeline/debug shows the MCP execution and continuation phases from both live SSE and history replay.
3. **Given** the continuation provider asks for another tool, **When** Loomi detects the request, **Then** it records an `unsupported_tool_loop` style safe failure and executes no additional tools.
4. **Given** the MCP tool fails, is denied, cancelled, or stopped, **When** continuation would otherwise be considered, **Then** Loomi skips continuation and records the terminal state.

### Edge Cases

- Provider requests a namespaced MCP tool that was discovered earlier but is no longer available, no longer persona-allowed, or belongs to a disabled server.
- Provider requests raw MCP tool name without `mcp.<server_slug>.<tool_name>` namespace.
- Persona allowed-tools references a discovered MCP tool, but the active run snapshot predates the latest discovery.
- Approval arrives after the run is stopped, cancelled, terminal, or owned by a worker that has already moved past the approval boundary.
- Approve/deny endpoints receive repeated requests, wrong thread/run/tool-call scope, or incompatible projection states.
- Worker crashes after approval but before execution start, after execution start but before result persistence, or after result persistence but before continuation.
- Worker lease expires while an MCP stdio process is running.
- A cancellation/stop request races with approval or with stdio process startup.
- MCP stdio process hangs, exceeds timeout, exits before response, writes large stderr, returns invalid JSON, returns unsafe result payload, or includes secret-looking data.
- MCP server command, args, env, stderr, result, or error includes tokens, credentials, absolute private paths, file contents, shell output, browser state, or desktop captured data.
- Result is too large, schema-incompatible, binary-looking, or cannot be summarized safely.
- History replay sees older M7 or M11 events without M12 MCP execution metadata.
- Provider continuation fails, streams partial content, or asks for another tool.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST execute only already-discovered local stdio MCP tool candidates from M11 that use stable namespaced names.
- **FR-002**: Loomi MUST require the selected persona's allowed-tools snapshot to reference the namespaced MCP candidate before offering approval or execution.
- **FR-003**: Loomi MUST reject undiscovered, disabled, unnamespaced, ambiguous, internal-tool-conflicting, or persona-disallowed MCP tool requests before approval is offered.
- **FR-004**: MCP tool execution MUST enter through the existing M7 tool-call projection and approval lifecycle; no MCP executor path may bypass approve/deny.
- **FR-005**: Each MCP tool-call projection MUST be scoped to thread, run, user/workspace identity where available, provider tool-call id, namespaced tool name, discovered candidate version or schema hash, and redacted argument summary/hash.
- **FR-006**: Projection recording MUST be idempotent for the same run and provider tool-call id and MUST NOT duplicate approval, denial, execution, or audit events during retry/replay.
- **FR-007**: Approval MUST be valid only while the run and tool-call projection are in an approval-blocked state; denial MUST prevent MCP process startup.
- **FR-008**: Worker execution MUST require current run/job ownership or lease before marking execution started and before invoking the stdio process.
- **FR-009**: Stop, cancel, lost ownership, terminal run state, or incompatible projection state MUST prevent new MCP stdio process startup.
- **FR-010**: Retry/recovery MUST NOT re-execute an MCP tool once the projection is started, succeeded, failed, denied, or cancelled unless a later spec defines explicit idempotency keys and safe retry classes.
- **FR-011**: M12 MUST support only one MCP tool execution and one provider continuation per run; if another tool is requested, Loomi MUST fail safely without executing it.
- **FR-012**: The MCP stdio process lifecycle MUST be bounded by timeout, explicit cleanup, redacted stderr handling, safe exit/error classification, and no raw env/args exposure.
- **FR-013**: Run events MUST cover approval required, approved, denied, execution started, succeeded, failed, cancelled/stopped, continuation started, continuation succeeded, continuation failed, and unsupported additional tool request states using safe metadata.
- **FR-014**: Persisted tool-call projection, run events, Timeline/debug metadata, worker diagnostics, and continuation payloads MUST use redacted arguments, result summaries, and error summaries.
- **FR-015**: Raw command paths, args, env values, stdout/stderr, tokens, credentials, Authorization headers, secret-looking paths, file contents, shell output, browser state, and desktop captured data MUST NOT appear in persisted events, API responses, UI replay, docs examples, or provider continuation.
- **FR-016**: Provider continuation MUST receive only the redacted MCP result for the matching scoped tool-call id and MUST NOT receive raw executor output.
- **FR-017**: Live SSE and history replay MUST render equivalent MCP approval/execution/continuation states and tolerate older events without MCP metadata.
- **FR-018**: M12 MUST reuse M7 approval/continuation, M6 worker/job lease and cancellation, M9 RunContext/pipeline, M10 persona allowed-tool resolution, M11 discovery/candidate mapping, existing run events, and existing Timeline/debug grouping.
- **FR-019**: M12 MUST NOT implement remote MCP, MCP HTTP/SSE/OAuth, marketplace/plugin install, DB-managed MCP server administration, shell/filesystem/browser automation, automatic execution, complex sandbox, admin UI, or multi-step tool loops.
- **FR-020**: Documentation and tasks MUST include backend tests, worker tests, frontend replay tests, docs-site updates, and validation commands for Go, web, and docs-site.

### Key Entities *(include if feature involves data)*

- **MCP Tool Execution Request**: Provider-requested namespaced MCP tool call after M11 discovery and persona allowlist resolution, before approval.
- **Scoped MCP Tool-Call Projection**: Existing M7 projection extended or reused for an MCP source, scoped to run/thread/provider tool-call id, safe candidate identity, redacted arguments, approval status, and execution status.
- **MCP Execution Attempt**: One worker-owned attempt to invoke the local stdio MCP tool after approval, with start/end timestamps, ownership evidence, timeout status, and redacted result/error.
- **MCP Stdio Process Invocation**: Bounded local process lifecycle for a single approved tool call, with no public raw command/args/env/stderr exposure.
- **MCP Tool Result Summary**: Redacted result/error object safe for persistence, Timeline/debug, and provider continuation.
- **MCP Continuation Context**: Provider-neutral single-continuation input containing the original assistant tool request and the matching redacted MCP tool result.
- **MCP Audit Event**: Persisted run-event/audit metadata for approval, denial, execution, cancellation, failure, and continuation states.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of approval-flow tests confirm MCP tool requests create approval-blocked projections and do not execute before approval.
- **SC-002**: 100% of authorization tests reject undiscovered, disabled, unnamespaced, conflicting, or persona-disallowed MCP tool requests before approval.
- **SC-003**: 100% of idempotency tests confirm repeated projection, approve, deny, worker retry, and history replay do not duplicate execution or audit events.
- **SC-004**: 100% of worker ownership tests confirm lost lease, stop/cancel, terminal run state, and incompatible projection state prevent process startup or duplicate execution.
- **SC-005**: 100% of stdio lifecycle tests cover success, timeout, early exit, stderr, invalid JSON, unsafe result, cleanup, and redacted error classification.
- **SC-006**: 100% of redaction tests confirm raw env, args, command paths, stdout/stderr, tokens, credentials, secret-looking paths, file contents, shell output, browser state, and desktop captured data are absent from persistence, UI replay, docs examples, and continuation payloads.
- **SC-007**: 100% of continuation tests confirm exactly one provider continuation uses the matching redacted MCP result and rejects additional tool loops safely.
- **SC-008**: Frontend replay tests show MCP approval/execution/continuation states from live-style and persisted events without breaking older event history.
- **SC-009**: The implementation quickstart can be validated with documented Go tests, web tests, `bun run --cwd web build`, and `bun run --cwd docs-site build`.

## Assumptions

- M11 discovery already produces safe, namespaced, non-executable MCP ToolSpec candidates and RunContext availability summaries.
- M7 already supports one approval-blocked tool call per run and one tool-result continuation; M12 extends that path to MCP rather than designing a new permission system.
- M12 is local-development first and uses explicit local stdio config already accepted by M11.
- Persona allowed-tools snapshots are the authorization input for whether a discovered MCP candidate can be considered for approval.
- MCP server claims, tool descriptions, schemas, arguments, results, and errors are untrusted data.
- This design intentionally stops before remote MCP, OAuth, sandbox policy, admin UI, DB-managed servers, and multi-tool loops.
