# Feature Specification: M15 Chat Real Integrated Smoke Closeout

**Feature Branch**: `[017-mcp-approval-gated-execution]`

**Created**: 2026-05-25

**Status**: Complete candidate

**Input**: User description: "Create and complete M15 / 022-chat-real-integrated-smoke-closeout. Make Chat's real main path repeatably smokeable across real API path, deterministic provider fixture, approved memory snapshot, persona-allowed discovered local stdio MCP tool approval, HTTP approve, worker MCP tools/call execution, redacted result, provider continuation, final assistant message, and timeline/history replay. This is a closeout/evidence slice, not a new platform feature."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Prove the real chat path reaches approval (Priority: P1)

As a Loomi maintainer closing M15, I want one deterministic chat smoke to start from the real API/service/worker path, load approved memory into RunContext, and stop at a persona-allowed discovered MCP tool approval, so the critical pre-execution boundaries are repeatable without external model spend.

**Why this priority**: The run is only meaningful if it proves real Chat orchestration, RunContext memory, discovery authorization, and M7 approval blocking before any tool execution.

**Independent Test**: Run the gated M15 smoke against the real in-process API/service/worker stack using deterministic fixtures and verify the run reaches `blocked_on_tool_approval` with safe memory and MCP candidate metadata.

**Acceptance Scenarios**:

1. **Given** an approved memory entry, a selected persona allowed-tools snapshot, and a discovered local stdio MCP candidate, **When** a chat run is created through the real API path, **Then** RunContext contains only the safe memory snapshot and the provider receives a deterministic fixture context.
2. **Given** the deterministic provider requests the discovered persona-allowed MCP tool, **When** the pipeline records the request, **Then** the run creates one approval-required projection and emits safe discovery, candidate hash, and approval-required events.
3. **Given** the same smoke is repeated, **When** the fixture inputs are unchanged, **Then** event types, terminal states, and redaction assertions remain stable.

---

### User Story 2 - Approve, execute, continue, and complete (Priority: P2)

As a maintainer, I want the smoke to approve the pending tool through HTTP, let the worker execute one MCP `tools/call`, persist a redacted result, continue the provider once, and write a final assistant message, so M15 proves the full real happy path.

**Why this priority**: M12 proved MCP approval and execution in narrower slices; M15 must prove the pieces work together from Chat through final assistant output.

**Independent Test**: Continue the same smoke by approving through the HTTP handler and verify exactly one MCP execution, one continuation, one final assistant message, and one completed run.

**Acceptance Scenarios**:

1. **Given** a run blocked on MCP tool approval, **When** the smoke calls the real approve endpoint, **Then** the run records approval and the worker executes exactly one MCP `tools/call`.
2. **Given** the MCP tool returns data containing secret-looking material, **When** the worker records the result, **Then** only redacted summaries are persisted and exposed.
3. **Given** the redacted tool result is available, **When** provider continuation runs, **Then** the final assistant message is written and the run completes without a second tool loop.

---

### User Story 3 - Replay evidence without leaking secrets (Priority: P3)

As a maintainer or reviewer, I want history replay and run events to show the whole M15 path while proving sensitive fixture values do not appear in API responses, RunContext safe summaries, run events, tool result summaries, documentation examples, or browser-visible timeline state.

**Why this priority**: Closeout evidence is incomplete unless it is observable and safe to share.

**Independent Test**: Fetch the run and event history after completion and assert required event milestones are present while forbidden sensitive markers are absent from every shareable surface.

**Acceptance Scenarios**:

1. **Given** the integrated smoke completes, **When** history replay is loaded, **Then** it contains queued/worker/pipeline context, `memory_snapshot_loaded`, MCP discovery/candidate hash, approval required/approved/executing/succeeded, continuation, and run completed events.
2. **Given** the fixture contains sensitive canary strings, **When** API responses, RunContext safe summaries, run events, tool result summaries, and docs examples are inspected, **Then** none of those sensitive strings appear.
3. **Given** a browser smoke can run, **When** Chat is opened in real API mode, **Then** the timeline visibly renders the integrated path; if blocked, equivalent backend smoke evidence and the blocker are documented.

### Edge Cases

- The deterministic provider fixture requests a tool that is not discovered, not persona-allowed, disabled, or missing the MCP namespace.
- The approved memory entry contains raw sensitive content that must never leave the memory boundary except as a safe summary.
- Approval is attempted when the run is not blocked on approval or when the tool-call id does not match the pending projection.
- The local stdio fixture returns result fields, stderr-like text, or secret-looking values that must be redacted.
- Provider continuation asks for a second tool instead of a final assistant message.
- Browser smoke cannot run because the local UI/API startup path is unavailable in the current environment.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: M15 MUST add a repeatable smoke under `specs/022-chat-real-integrated-smoke-closeout/` and wire it to the real API/service/worker path, not UI-only mocks.
- **FR-002**: The smoke MUST use a deterministic provider fixture and MUST NOT depend on external paid model calls.
- **FR-003**: The smoke MUST prepare approved memory and prove `RunContext.MemorySnapshot` is attached to the run.
- **FR-004**: The smoke MUST use a local stdio MCP tool that has been discovered and is allowed by the selected persona snapshot.
- **FR-005**: The provider fixture MUST request the discovered namespaced MCP tool and the run MUST enter approval required before execution.
- **FR-006**: The smoke MUST approve the tool through the HTTP approve path.
- **FR-007**: The worker MUST execute exactly one MCP `tools/call` after approval.
- **FR-008**: The smoke MUST record and expose only redacted tool result summaries.
- **FR-009**: The provider continuation MUST produce one final assistant message after the redacted tool result.
- **FR-010**: History replay and run events MUST show queued/worker/pipeline context, `memory_snapshot_loaded`, MCP discovery/candidate hash, tool approval required/approved/executing/succeeded, continuation, and run completed.
- **FR-011**: Sensitive fixture content MUST NOT appear in API responses, RunContext safe summaries, run events, tool result summaries, docs examples, or browser-visible evidence.
- **FR-012**: M15 MUST update docs-site devlog, local runbook, roadmap/current-status, spec-kit/workflow, and any API/architecture pages whose behavior differs.
- **FR-013**: M15 MUST NOT introduce sandboxing, filesystem/shell/browser automation tools, activity recorder, OpenViking/vector/RAG/distill, marketplace/plugin install, multi-agent, or worker queue rewrites.
- **FR-014**: If browser smoke cannot run, the closeout MUST state the blocker and include equivalent backend smoke evidence.

### Key Entities *(include if feature involves data)*

- **M15 Smoke Scenario**: Repeatable test fixture tying together user, thread, run, provider fixture, memory entry, persona, MCP candidate, approval, worker execution, continuation, and replay evidence.
- **Approved Memory Snapshot**: Safe bounded memory data attached to RunContext for this run.
- **Deterministic Provider Fixture**: Local provider behavior that first requests the MCP tool and then emits a final assistant message after tool result continuation.
- **Local MCP Tool Fixture**: Discovered persona-allowed stdio tool candidate used only for one safe `tools/call`.
- **Closeout Evidence Surface**: API response, run events, RunContext summaries, tool result summary, docs examples, and optional browser timeline.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: One gated M15 smoke command completes deterministically and verifies every required milestone from run creation through completed replay.
- **SC-002**: The smoke observes exactly one approval-required MCP projection, one HTTP approval, one MCP `tools/call`, one provider continuation, one final assistant message, and one completed run.
- **SC-003**: Required event milestones are all present in persisted event replay with stable event types and safe metadata.
- **SC-004**: Redaction assertions check every shareable surface and find zero occurrences of the configured sensitive canary values.
- **SC-005**: Required backend, frontend, docs, diff, and gated smoke validations run before M15 is reported as a complete candidate.

## Assumptions

- M15 reuses M7 approval, M9 RunContext, M11 MCP discovery, M12 execution/continuation, M13 memory, and M14 memory audit UX boundaries.
- A Go integration test with an explicit environment gate is acceptable as the repeatable backend smoke command.
- The browser smoke is opportunistic because this slice is primarily real backend/API evidence; any blocker must be documented.
- Documentation examples use synthetic redacted values only.
