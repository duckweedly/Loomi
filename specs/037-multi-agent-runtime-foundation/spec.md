# Feature Specification: M29 Multi-agent Runtime Foundation

**Feature Branch**: `[037-multi-agent-runtime-foundation]`  
**Created**: 2026-05-26  
**Status**: Draft  
**Input**: Continue Arkloop-level coverage after artifact runtime by adding a safe multi-agent coordination foundation.

## User Scenarios & Testing

### User Story 1 - Spawn a Bounded Agent Task (Priority: P1)

As a Work mode user, I want an approved tool call to spawn a bounded child agent task so a run can delegate a clearly scoped subtask without starting uncontrolled background execution.

**Independent Test**: A Work mode run requests `agent.spawn`, requires approval, creates one child task record with role/goal/status/source run metadata, records safe events, and continues the provider with the task summary.

### User Story 2 - List and Complete Agent Tasks (Priority: P2)

As a Work mode user, I want the agent to list active child tasks and mark one completed with a bounded result summary.

**Independent Test**: A run spawns a child task, then approved `agent.list` and `agent.complete` return safe summaries and update the task status without external side effects.

### User Story 3 - Keep Multi-agent Coordination Observable and Safe (Priority: P3)

As a user, I want multi-agent coordination to be visible and non-autonomous until explicit execution support exists.

**Independent Test**: Chat mode omits agent tools; unsupported roles, oversized goal/result content, denied/stopped/terminal/out-of-scope calls, duplicate tool calls, and unsafe metadata fail before records are created or changed.

## Requirements

### Functional Requirements

- **FR-001**: Tool catalog MUST include builtin `agent.spawn`, `agent.list`, and `agent.complete`.
- **FR-002**: Agent tools MUST be Work-mode only and always approval required.
- **FR-003**: `agent.spawn` MUST create exactly one bounded child task record with role, goal, status, source thread, and source run.
- **FR-004**: `agent.list` MUST return bounded child task summaries for the current thread.
- **FR-005**: `agent.complete` MUST mark one existing child task completed with a bounded result summary.
- **FR-006**: Agent tools MUST route through ToolBroker, worker approved-tool resume, provider continuation, and run events.
- **FR-007**: Settings > Tools and RunRail MUST label agent tools separately from workspace, sandbox, LSP, web, browser, and artifact tools.

### Safety Requirements

- **SR-001**: Agent tools MUST NOT launch external processes, call models, call network, call filesystem, or create worker jobs beyond the approved tool execution itself.
- **SR-002**: Agent role, goal, and result fields MUST be bounded and safe for run-event summaries.
- **SR-003**: Chat mode MUST NOT enable agent tools.
- **SR-004**: Child tasks MUST be scoped to the current thread.

## Non-goals

- Autonomous sub-agent execution, parallel worker pools, external agent processes, chat-room transport, code ownership locking, automatic task scheduling, cross-thread delegation, marketplace agents, remote orchestration, and UI task assignment editing.
