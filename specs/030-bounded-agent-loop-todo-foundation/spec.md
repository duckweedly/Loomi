# Feature Specification: M22 Bounded Agent Loop + Todo Foundation

**Feature Branch**: `030-bounded-agent-loop-todo-foundation`

**Created**: 2026-05-25

**Status**: Draft

**Input**: User description: "After M21 workspace read tools, continue toward Arkloop-level code-agent coverage. Add the next minimal slice: a bounded single-run agent loop that can continue from tool results more than once, plus observable todo state for Work mode. Keep every tool call approval-gated, sequential, and persisted through existing run events. Do not add workspace write/edit, sandbox shell, browser automation, web search, multi-agent, marketplace, or unbounded autonomous loops yet."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Bounded Multi-Step Tool Continuation (Priority: P1)

A Work mode run can request an allowed tool, pause for approval, execute it, continue with the provider, and repeat this cycle for a small configured number of sequential tool calls before producing a final assistant response.

**Why this priority**: M21 proves one read-only tool call. A useful code-agent loop needs at least `glob -> read -> final` or `grep -> read -> final` without ending after the first tool result.

**Independent Test**: Run a backend smoke with a provider fixture that requests `workspace.glob`, receives its result, then requests `workspace.read`, receives its result, and then emits a final assistant message. Verify both tool calls require approval and the final run completes.

**Acceptance Scenarios**:

1. **Given** a Work mode run with approved `workspace.glob`, **When** the provider requests a second allowed workspace read tool during continuation, **Then** the run pauses for a second approval instead of failing with an unsupported tool loop.
2. **Given** each requested tool is approved in order, **When** the last continuation returns final text, **Then** the run completes with one assistant message and a complete persisted event sequence.
3. **Given** the provider requests more tool calls than the configured loop limit, **When** the limit is reached, **Then** the run fails safely with an observable loop-limit error and no further tool executes.

---

### User Story 2 - Observable Todo State (Priority: P2)

A Work mode run can maintain a small visible todo list that represents the agent's current plan, updates, and completed steps without depending on file mutation or shell execution.

**Why this priority**: Arkloop-style code-agent work is understandable when the operator can see plan state. Todo state gives the loop an inspectable planning surface before write/shell tools arrive.

**Independent Test**: Run a fixture provider that proposes a todo list, updates step statuses after each approved tool result, and verify the Work Plan View and timeline reflect the current todo state.

**Acceptance Scenarios**:

1. **Given** a Work mode run emits todo updates, **When** the UI renders the Work Plan View, **Then** the current todo items, statuses, and latest update are visible.
2. **Given** todo updates include secret-looking values or executable hints, **When** they are persisted and rendered, **Then** unsafe values are redacted or omitted.
3. **Given** a Chat mode run emits todo-shaped metadata, **When** the UI renders Chat mode, **Then** it does not broaden Work mode planning surfaces into Chat.

---

### User Story 3 - Operator Control and Failure Visibility (Priority: P3)

An operator can understand why a bounded loop is waiting, continuing, completed, stopped, failed, or loop-limited from the timeline and run detail surfaces.

**Why this priority**: Multi-step loops are only safe if each transition is visible and stoppable. The user should never confuse waiting-for-approval with autonomous execution.

**Independent Test**: Render timeline fixtures for approval-required, tool-executing, continuation, loop-limit, stopped, and completed states, and verify the labels are clear and safe.

**Acceptance Scenarios**:

1. **Given** a run is waiting for a later tool approval, **When** the timeline renders, **Then** it clearly shows the tool index and approval-required state.
2. **Given** the user stops a run between tool calls, **When** the worker resumes, **Then** no additional provider continuation or tool execution occurs.
3. **Given** a continuation fails, **When** the timeline renders, **Then** the failure reason is visible without raw provider payloads or tool result contents.

### Edge Cases

- Provider requests two tools in one response: reject safely; this milestone remains sequential one-tool-at-a-time.
- Provider requests a disallowed tool during a later continuation: fail safely without recording an approval-required tool call.
- Provider repeats the same tool call id: reject or deduplicate without executing twice.
- User denies a later tool call: stop continuation and preserve previous successful tool results.
- Run is cancelled or stopped while blocked on approval: no later continuation executes.
- Todo update contains raw file contents, secrets, absolute host paths, shell commands, or browser URLs: persist/render only safe summaries.
- Event replay starts mid-run: UI reconstructs the current loop/todo state from persisted events.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support a bounded sequential tool continuation loop for Work mode runs using the existing provider, ToolBroker, approval, worker, run event, and continuation boundaries.
- **FR-002**: System MUST require explicit approval for every tool call in the loop; prior approval MUST NOT authorize a later tool call.
- **FR-003**: System MUST allow at most one pending or executing tool call per run at a time in this milestone.
- **FR-004**: System MUST enforce a configurable small maximum tool-call count per run and fail safely when the limit is reached.
- **FR-005**: System MUST reject disallowed or unavailable tools during initial and later continuations before entering approval-required state.
- **FR-006**: System MUST persist enough safe run-event metadata to replay each loop step, including tool index, tool name, approval state, execution state, continuation phase, and loop-limit failures.
- **FR-007**: System MUST preserve stop/cancel behavior across blocked-on-approval and continuation states so no extra tool or provider call runs after terminal stop/cancel.
- **FR-008**: System MUST expose todo state for Work mode as safe run metadata that can be replayed into the Work Plan View.
- **FR-009**: System MUST render todo items with stable statuses such as pending, running, completed, blocked, and failed.
- **FR-010**: System MUST redact or omit secrets, raw provider payloads, raw tool results, raw file contents, absolute private paths, shell commands, executable hints, browser state, and hidden local state from todo metadata, run events, API responses, UI replay, and docs examples.
- **FR-011**: System MUST keep Chat mode isolated from Work mode todo/planning surfaces and MUST NOT broaden Chat mode workspace tool access.
- **FR-012**: System MUST NOT add workspace write/edit tools, sandbox shell or code execution, browser automation, web search/fetch, artifact creation, multi-agent orchestration, remote MCP/OAuth, marketplace/plugin install, or unbounded autonomous loops in this milestone.

### Key Entities *(include if feature involves data)*

- **Agent Loop State**: Safe per-run state describing the current loop count, maximum loop count, current phase, and whether the run is waiting for approval, continuing, completed, stopped, failed, or loop-limited.
- **Loop Tool Step**: One requested tool call in the sequence, including tool call id, tool name, tool index, approval status, execution status, and safe summaries.
- **Todo List Snapshot**: Safe Work mode planning state with todo items, statuses, order, and update source.
- **Todo Item**: A single planned step with id, title, status, optional safe summary, and redaction marker.
- **Loop Event**: Existing run event plus safe metadata that allows replaying loop/todo state without raw provider/tool payloads.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Backend smoke demonstrates a Work mode run completing after at least two sequential approved tool calls and one final provider response.
- **SC-002**: Backend smoke demonstrates that a second tool call pauses for a new approval and does not execute before approval.
- **SC-003**: Backend smoke demonstrates loop-limit rejection without executing the over-limit tool call.
- **SC-004**: UI replay shows current loop/todo state from persisted events after refresh.
- **SC-005**: Chat mode tests prove todo/planning surfaces and workspace tool access are not broadened.
- **SC-006**: Required validation commands complete or the exact blocker is reported: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.

## Assumptions

- The M21 workspace read tools remain the only local workspace tools available during this milestone.
- Existing ToolCatalog, ToolBroker, RunContext, provider continuation, worker, approval, and run-event APIs remain the integration path.
- Todo state is product-visible planning metadata, not a separate project-management system.
- The first loop limit should be intentionally small and testable; exact default belongs in the implementation plan.
- Arkloop is used only as a mechanism reference for bounded tool loops and todo visibility, not for brand, copy, private APIs, or full platform architecture.
