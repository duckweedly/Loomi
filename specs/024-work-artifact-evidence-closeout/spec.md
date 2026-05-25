# Feature Specification: M17 Work Artifact Evidence Closeout

**Feature Branch**: `[017-mcp-approval-gated-execution]`

**Created**: 2026-05-25

**Status**: Complete candidate

**Input**: User description: "Create and complete M17 / 024-work-artifact-evidence-closeout. Move M16 Work mode from mock seed Work Plan View to repeatable Work artifact evidence closeout using real thread/message/run/event replay or a minimal local-dev/test seed path. Do not build artifact execution."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Replay evidence into Work Plan View (Priority: P1)

As a Loomi user opening a seeded Work mode thread, I want the Work Plan View to render goal, steps, current status, artifact metadata, and recent progress from real thread/message/run/event data, so M16 can be validated repeatably without relying on mock-only state.

**Why this priority**: M17 is not complete until Work mode is demonstrable from the same API/runtime data path used by local development.

**Independent Test**: Run the local M17 seed path, open the Work thread in real API mode, and verify the Work Plan View displays the seeded goal, steps, artifacts, and recent progress from the current run events.

**Acceptance Scenarios**:

1. **Given** a local API with a seeded `Thread.mode = work`, message, run, and work event metadata, **When** the Work thread opens, **Then** Work Plan View shows the seeded goal, ordered steps, run status, artifact references, and recent progress.
2. **Given** event replay fetches run events after page load, **When** the selected work run is restored, **Then** Work Plan View derives its projection from replayed events rather than mock data.
3. **Given** the seed is run more than once, **When** the Work thread is opened again, **Then** the evidence path remains repeatable and identifies the same seeded thread/message while avoiding duplicate visible evidence.

---

### User Story 2 - Show safe artifact evidence only (Priority: P2)

As a Loomi user reviewing Work artifact evidence, I want artifact cards to show only safe metadata and explicit redaction markers, so I can inspect outputs without seeing executable fields, secrets, or action controls.

**Why this priority**: Work artifact evidence is useful only if it does not imply shell, filesystem, browser, URL, or tool execution.

**Independent Test**: Seed or render artifact metadata containing unsafe fields and verify the UI includes safe id/title/type/source/summary/timestamps/redaction marker while omitting executable fields and controls.

**Acceptance Scenarios**:

1. **Given** artifact metadata includes `id`, `title`, `type`, `source_thread_id`, `source_run_id`, `summary`, `created_at`, `updated_at`, and redacted fields, **When** artifact cards render, **Then** only those safe fields plus a redaction marker appear.
2. **Given** artifact metadata includes command, path, file, shell, browser, filesystem, execute, URL, secret, token, Authorization, private path, provider trace, or tool output fields, **When** projected, **Then** those values are redacted or omitted and never rendered as clickable/executable actions.
3. **Given** artifact cards are visible, **When** the user inspects available controls, **Then** no artifact-specific execute/open/run/download button exists.

---

### User Story 3 - Preserve mode isolation during real smoke (Priority: P3)

As a Loomi user switching between Chat and Work threads during local smoke, I want Chat mode to stay free of Work Plan View, so Work-specific evidence does not leak into normal chat.

**Why this priority**: M17 reuses shared ChatCanvas/run/event boundaries and must prove the Work projection is gated only by `Thread.mode = work`.

**Independent Test**: Open a real API Chat thread after the Work smoke and verify no Work Plan View appears even if work-like metadata exists elsewhere.

**Acceptance Scenarios**:

1. **Given** a Chat mode thread is selected, **When** the main canvas renders, **Then** no Work Plan View appears.
2. **Given** a Work mode thread is selected, **When** the main canvas renders, **Then** Work Plan View appears without replacing message history or RunRail.

### Edge Cases

- The real API has no current run for a Work thread.
- The local seed runs against an existing seeded thread.
- Work artifact metadata contains unsafe executable fields or private paths.
- Secret-like values are nested inside artifact metadata.
- Chat mode is opened after Work mode in the same browser session.
- Event replay returns lifecycle/job events before the work-specific progress event.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: M17 MUST create `specs/024-work-artifact-evidence-closeout/` with spec, plan, research, data model, contract, quickstart, tasks, and checklist artifacts.
- **FR-002**: Work mode MUST continue to reuse existing `Thread.mode = work`, message, run, event, ChatCanvas, and RunRail boundaries.
- **FR-003**: M17 MUST design and implement one repeatable evidence path using existing real API data when possible.
- **FR-004**: If existing public API cannot safely seed Work event metadata, M17 MAY add a minimal local-dev/test-only seed path, but it MUST NOT become a general production write API.
- **FR-005**: Work event metadata MUST cover `work_goal`, `work_steps`, `work_artifacts`, and recent progress events.
- **FR-006**: Artifact evidence MUST display only safe metadata: id, title, type, source thread/run, summary, created/updated, and redaction marker when unsafe fields were removed or redacted.
- **FR-007**: UI MUST NOT render command/path/file/shell/browser/filesystem/execute/url fields as artifact actions.
- **FR-008**: Secret/token/Authorization/private path/provider trace/tool output values MUST be redacted or omitted before rendering.
- **FR-009**: Artifact cards MUST NOT include executable buttons or controls.
- **FR-010**: Chat mode MUST NOT display Work Plan View.
- **FR-011**: M17 MUST include frontend projection/rendering tests, real API or local seed evidence test, no-executable-controls test, and mode isolation test.
- **FR-012**: M17 MUST include browser smoke evidence with ports, seed/thread/run identifiers, screenshot path or equivalent evidence, and console error status.
- **FR-013**: M17 MUST update docs-site devlog, runbook, architecture/API notes, roadmap current status, and Spec Kit workflow.
- **FR-014**: M17 MUST NOT implement artifact execution/runtime, sandbox, shell/filesystem/browser automation tools, activity recorder, multi-agent, plugin marketplace, new task system, or worker queue rewrite.

### Key Entities *(include if feature involves data)*

- **Work Evidence Seed**: Local-dev/test-only data creation flow that creates or reuses a Work thread, seed message, current run, and work progress event metadata.
- **Work Event Metadata**: Existing run event metadata containing work goal, steps, artifact references, and progress evidence.
- **Artifact Evidence Reference**: Metadata-only artifact card with safe fields and redaction status.
- **Browser Smoke Evidence**: Recorded local ports, thread/run identifiers, UI assertions, console status, and screenshot path.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running the local evidence path produces a Work mode thread with a current run and at least one work metadata event.
- **SC-002**: Browser smoke in real API mode shows goal, steps, status, artifact references, and recent progress from the seeded Work thread.
- **SC-003**: Artifact evidence tests prove unsafe metadata is redacted or omitted and no executable artifact controls are rendered.
- **SC-004**: Mode isolation tests prove Chat mode does not render Work Plan View.
- **SC-005**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, `git diff --check`, and browser smoke are completed or exact blockers are reported.

## Assumptions

- Existing real API can create work threads/messages and start runs, but does not expose a general public endpoint for arbitrary run event metadata.
- A local-dev/test seed path is acceptable for M17 evidence because it is limited to the seed command and uses existing product service methods.
- Artifact evidence remains metadata-only; opening, executing, downloading, or reading files is out of scope.
