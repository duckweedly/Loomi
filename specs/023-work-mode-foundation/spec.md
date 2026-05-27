# Feature Specification: M16 Work Mode Foundation

**Feature Branch**: `[017-mcp-approval-gated-execution]`

**Created**: 2026-05-25

**Status**: Complete candidate

**Input**: User description: "Create and complete M16 / 023-work-mode-foundation. Move Work mode from a mode shell to a minimal usable slice where a work thread can show plan, artifact references, and progress/event view. Reuse existing thread/message/run/event/right drawer boundaries and do not add a new execution environment."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See a work plan on a work thread (Priority: P1)

As a Loomi user opening a Work mode thread, I want to see the work goal, steps, current progress/status, linked artifacts, and recent run/event progress in one read-only surface, so I can understand what the thread is trying to accomplish without leaving the existing conversation/run context.

**Why this priority**: Work mode is not usable until a work thread exposes more structure than ordinary chat history while still reusing the existing thread and run model.

**Independent Test**: Open a seeded Work mode thread and verify the Work Plan View renders goal, ordered steps, current status, artifact references, and recent progress from existing messages/run events.

**Acceptance Scenarios**:

1. **Given** a thread with `mode = work` and existing messages/run events, **When** the thread opens, **Then** the main area or right-side panel shows a Work Plan View with goal, steps, progress/status, artifacts, and recent progress.
2. **Given** no active run exists for the work thread, **When** the thread opens, **Then** the Work Plan View still shows a clear empty state instead of inventing an execution state.
3. **Given** run events arrive during event replay, **When** the selected work run updates, **Then** the progress/status and recent event list update from those events.

---

### User Story 2 - Keep artifact references safe (Priority: P2)

As a Loomi user reviewing Work mode artifacts, I want linked artifacts to show only safe metadata and markdown-like summaries, so I can inspect references without executing files, tools, browsers, or shell commands.

**Why this priority**: Work mode needs artifact visibility, but M16 must not pull sandboxing or real artifact execution forward.

**Independent Test**: Seed artifact metadata containing secret-looking fields and verify the rendered artifact references contain title, type, source run/thread, summary, timestamps, and redacted safe metadata only.

**Acceptance Scenarios**:

1. **Given** work event metadata includes artifact references, **When** the Work Plan View renders them, **Then** each reference shows title, type, source run/thread, summary, created/updated timestamps, and no executable controls.
2. **Given** artifact metadata contains secret-looking strings or unsupported fields, **When** projected for display, **Then** unsafe values are redacted or omitted.

---

### User Story 3 - Preserve Chat and Work isolation (Priority: P3)

As a Loomi user switching between Chat and Work threads, I want Chat mode to keep its existing conversation behavior and Work mode to add only the work-specific surface, so the new slice does not regress normal chat.

**Why this priority**: Work mode reuses core thread/message/run/event surfaces and must not fork or degrade the chat path.

**Independent Test**: Render a Chat mode thread and a Work mode thread with similar messages/runs; verify only Work mode shows the Work Plan View and Chat mode still renders normal chat history/composer behavior.

**Acceptance Scenarios**:

1. **Given** a Chat mode thread is selected, **When** the main area renders, **Then** no Work Plan View appears and existing chat state handling remains unchanged.
2. **Given** a Work mode thread is selected, **When** the main area renders, **Then** the Work Plan View appears without replacing message history, run events, or tool approval controls.

### Edge Cases

- A work thread has messages but no explicit work-plan metadata.
- A work thread has malformed or partial work-plan metadata.
- A run has no recent events or only terminal events.
- A Chat mode thread receives work-like metadata in events.
- Artifact metadata includes secret-looking values, raw file paths, shell/browser hints, or unknown object fields.
- The selected thread changes while event replay is still applying events.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: M16 MUST create `specs/023-work-mode-foundation/` with spec, plan, tasks, and validation artifacts.
- **FR-002**: Work mode MUST continue to use existing `Thread.mode = work`; the feature MUST NOT create a second task system.
- **FR-003**: Work Plan View MUST expose goal, steps, current progress/status, linked artifact references, and recent run/event progress for Work mode threads.
- **FR-004**: Work Plan View MUST project progress from existing run events, messages, and safe metadata rather than from a new worker, task queue, or execution environment.
- **FR-005**: Artifact references MUST render only safe metadata: title, type, source run/thread, summary, created/updated timestamps, and markdown-like preview text.
- **FR-006**: Artifact references MUST NOT execute files and MUST NOT expose filesystem, browser, shell, or automation tool controls.
- **FR-007**: Unsupported or secret-looking metadata MUST be redacted or omitted before rendering.
- **FR-008**: Work mode UI MUST include clear empty, loading, and error states.
- **FR-009**: Chat mode MUST NOT show Work Plan View and MUST preserve existing chat rendering.
- **FR-010**: M16 MUST prefer existing thread/message/run/event APIs. New API or event payloads are allowed only if existing data cannot project the minimum Work view.
- **FR-011**: M16 MUST NOT introduce sandboxing, filesystem execution, shell/browser tools, activity recorder, multi-agent behavior, marketplace/plugin install, or worker queue rewrites.
- **FR-012**: M16 MUST update docs-site architecture, API/payload notes if applicable, local runbook, devlog, roadmap/current-status, and spec-kit workflow.

### Key Entities *(include if feature involves data)*

- **Work Thread**: Existing thread with `mode = work`; owns messages and current run like Chat threads.
- **Work Plan Projection**: Read-only UI projection derived from messages, run events, and safe metadata.
- **Work Step**: Ordered plan item with title and status derived from metadata or event progression.
- **Artifact Reference**: Safe metadata-only reference to an output or note associated with a source thread/run.
- **Recent Progress Event**: Existing run event rendered as recent work progress.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A seeded Work mode thread renders goal, steps, current status, artifact references, and recent progress in the browser.
- **SC-002**: Chat mode render tests confirm the Work Plan View is absent from Chat threads.
- **SC-003**: Event replay/projection tests prove incoming run events update Work progress without a separate queue.
- **SC-004**: Safe metadata tests prove secret-looking values and unsupported executable fields are not rendered.
- **SC-005**: Required web tests, web build, docs-site build, diff check, and browser smoke complete before M16 is reported as a complete candidate.

## Assumptions

- Existing thread/message/run/event APIs carry enough event metadata to seed a minimal Work Plan View in M16.
- M16 can use deterministic mock data and frontend projection tests for the first work-thread seed/projection coverage.
- Artifact "preview" means markdown-like safe text and metadata only, not opening, executing, or reading local files.
- Browser smoke uses the existing local web shell and mock data unless a real API seed exists in the current environment.
