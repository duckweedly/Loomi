# Research: M29 Multi-agent Runtime Foundation

## Decision: Start With Coordination Records, Not Autonomous Execution

M29 creates child-agent task records through `agent.spawn`, reads them through `agent.list`, and marks them complete through `agent.complete`.

**Rationale**: This proves delegation, status visibility, and continuation contracts without introducing uncontrolled model calls, worker pools, or external agent processes.

**Alternatives rejected**:

- Real parallel sub-agent execution first: needs ownership, cancellation, budgeting, logs, and scheduling.
- Chat-room transport first: useful later, but it does not prove runtime tool integration.
- File ownership locks first: valuable after child tasks can exist.

## Decision: Reuse ToolBroker and Worker Approval

Agent tools are normal builtin tools.

**Rationale**: Keeps approval, run events, loop limits, provider continuation, and Settings/RunRail visibility consistent with other tool families.

## Decision: Thread-scoped Tasks

Child-agent tasks are scoped to the current thread for M29.

**Rationale**: Prevents cross-thread leakage and avoids workspace-wide coordination policy until real autonomous execution exists.
