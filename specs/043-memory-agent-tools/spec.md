# Feature Specification: Memory Agent Tools

**Feature Branch**: `[043-memory-agent-tools]`

**Created**: 2026-05-26

**Status**: Draft

**Input**: User description: "Continue the Arkloop-class memory parity plan after provider foundation by adding agent-facing memory tools in Loomi's existing ToolCatalog/ToolBroker/approval runtime."

## User Scenarios & Testing

### User Story 1 - Agent can safely inspect memory (Priority: P1)

As a Loomi user, I want the model to search, read, and check memory status through explicit tools, so memory recall is observable in the run timeline instead of hidden in prompts.

**Independent Test**: A run with memory tools allowlisted can request `memory.search`, `memory.read`, and `memory.status`; after approval/execution, tool results contain only safe summaries/status and no raw content or secrets.

### User Story 2 - Agent can propose or delete memory through approval (Priority: P1)

As a Loomi user, I want memory writes/deletes to stay approval-gated, so the agent can help maintain memory without silently changing long-term state.

**Independent Test**: `memory.write` creates a pending proposal, and `memory.forget` tombstones an approved entry only after the tool call is approved.

### User Story 3 - Existing memory management still works (Priority: P2)

As a user, I want Settings > Memory list/search/audit/delete to keep working while agent tools are added.

**Independent Test**: Existing M13/M14 memory API and UI tests still pass, and tool result summaries appear as normal RunRail tool events.

## Requirements

- **FR-001**: Add catalog entries for `memory.search`, `memory.read`, `memory.write`, `memory.forget`, and `memory.status`.
- **FR-002**: Memory tools MUST use the existing ToolBroker approval/execution path and persisted tool call events.
- **FR-003**: Read-only memory tools MUST still be approval-gated in this slice.
- **FR-004**: Tool arguments MUST be bounded and validated.
- **FR-005**: `memory.search` and `memory.read` MUST return safe summaries only, never raw memory content.
- **FR-006**: `memory.write` MUST create a write proposal and not directly approve a memory entry.
- **FR-007**: `memory.forget` MUST tombstone an approved memory entry through existing scoped delete behavior.
- **FR-008**: `memory.status` MUST return the provider readiness status from the provider foundation.
- **FR-009**: Tool catalog, run context enabled tools, tool request validation, and provider schema serialization MUST include the new tools.
- **FR-010**: Documentation MUST state that automatic distillation and semantic external read/write remain later slices.

## Success Criteria

- **SC-001**: Tool catalog and Settings > Tools show the five memory tools as builtin memory tools.
- **SC-002**: Worker continuation can execute approved memory tools and provider continuation receives the tool result.
- **SC-003**: Tool results and run events contain no raw memory content, API keys, provider traces, local paths, or secrets.
- **SC-004**: Existing memory management API/UI behavior remains unchanged.

## Assumptions

- The default built-in persona may allow the memory tools.
- This slice does not add automatic memory distillation.
- This slice does not implement a semantic provider adapter beyond status.
