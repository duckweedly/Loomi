# Feature Specification: Real Usage Readiness

**Feature Branch**: `040-real-usage-readiness`

**Created**: 2026-05-26

**Status**: Candidate

**Input**: User description: "UI-02 Real Usage Readiness: after UI-01 light shell, close the gap between demo-looking UI and real-use states. Keep provider failures actionable, remove fake or confusing chat controls, make Work task panels and tool events readable, support sidebar basics, and add basic Markdown rendering without changing backend/runtime/provider/tool/database behavior."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Understand Current Runtime Reality (Priority: P1)

As a Loomi user, I see provider failures only when they block sending, and the main chat canvas is not cluttered with runtime/debug chips.

**Why this priority**: The UI must be honest before it can be trusted for real work.

**Independent Test**: Render ChatCanvas in mock, stream-disconnected, and real-provider-unavailable states; verify the old runtime/header chips are absent and provider-unavailable guidance still links to Provider Settings.

**Acceptance Scenarios**:

1. **Given** the frontend is in Mock mode, **When** the canvas and composer render, **Then** the chat canvas does not show the old Mock/demo runtime header.
2. **Given** Real API is selected and no supported provider is available, **When** the user opens the workspace, **Then** the UI shows provider unavailable state and a Provider Settings entry.

---

### User Story 2 - Use Work Mode Without Fake Controls (Priority: P1)

As a Work mode user, I see an honest task panel rather than disabled controls that pretend a folder picker exists.

**Why this priority**: Work mode is where Loomi resembles an agent workspace; demo affordances make it misleading.

**Independent Test**: Render Work mode with no plan metadata and no selected folder. Verify the panel says plan metadata is missing and the composer does not show fake folder controls.

**Acceptance Scenarios**:

1. **Given** a Work thread has user messages but no plan metadata, **When** WorkPlanView renders, **Then** it shows goal/status/recent progress honestly and does not invent plan steps from messages or tool events.
2. **Given** no safe directory picker exists, **When** the Work composer renders, **Then** it does not show a disabled fake folder button or folder limitation control.

---

### User Story 3 - Read Runs Like User Progress (Priority: P2)

As a Loomi user, I can understand tool events and approval waits without reading raw tool names first.

**Why this priority**: Real runs need observable execution, but the first line should be user-facing progress.

**Independent Test**: Render RunRail and ToolCallCard with workspace/web/lsp/artifact/agent events and approval-required tool calls. Verify human labels appear first and raw tool names/arguments remain in details.

**Acceptance Scenarios**:

1. **Given** a run includes tool events, **When** RunRail renders, **Then** it shows human labels such as "Read project files", "Visit web page", "Analyze code", "Handle artifact", and "Coordinate subtasks" before raw technical details.
2. **Given** a run is blocked on tool approval, **When** ChatCanvas renders, **Then** the top/Work surface says "Waiting for your confirmation" and Approve/Deny/Stop are visible.

---

### User Story 4 - Basic Sidebar and Composer Are Usable (Priority: P3)

As a Loomi user, I can filter, rename, delete threads in the current mode, create the right kind of thread, see mode-specific input guidance, and read basic Markdown replies.

**Why this priority**: Sidebar and composer basics are low-risk UI readiness work that prevents the shell from feeling like a static prototype.

**Independent Test**: Render ThreadSidebar, Composer, and ChatCanvas for Chat/Work. Verify search filters visible threads, rename/delete menu is present, create copy says New Chat/New Work, run dots have readable labels, placeholders differ by mode, unused bottom controls are absent, and basic Markdown renders.

**Acceptance Scenarios**:

1. **Given** Chat or Work mode is selected, **When** the user searches in the sidebar, **Then** only threads in the current mode and matching title remain visible.
2. **Given** Chat mode is selected, **When** the create action renders, **Then** it says New Chat.
3. **Given** Work mode is selected, **When** the create action renders, **Then** it says New Work.
4. **Given** Chat or Work composer renders, **When** the user sees the placeholder, **Then** Chat says to message Loomi and Work says to describe the task.
5. **Given** a thread row renders, **When** the user opens the row action menu, **Then** rename and delete actions are visible.
6. **Given** an assistant reply contains Markdown, **When** the message renders, **Then** headings, lists, bold, inline code, links, paragraphs, and fenced code render as formatted content.

### Edge Cases

- Approval-blocked runs must keep Stop visible even when the composer is disabled.
- Work Plan View must not use generic tool events as plan steps when explicit plan metadata is missing.
- Real API provider unavailable state must not silently fall back to mock output.
- Search must not mix Chat and Work threads.
- Raw tool names, arguments, and result summaries may be visible as details, but not as the primary progress label.
- Unimplemented attachment, voice, persona/provider selector, and Work folder controls must not appear in the Composer.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The chat canvas MUST NOT show the old Mock/Real/runtime debug header.
- **FR-002**: Real API provider unavailable state MUST show provider unavailable copy and a Provider Settings entry.
- **FR-003**: Composer MUST NOT render unimplemented attachment, voice, persona/provider selector, or Work folder controls.
- **FR-004**: WorkPlanView MUST render goal, status, plan/todos, artifacts, and recent progress when metadata exists.
- **FR-005**: WorkPlanView MUST NOT invent plan steps from regular user messages or raw tool events when plan metadata is missing.
- **FR-006**: RunRail and ToolCallCard MUST present common tool events with human-first labels while keeping raw tool name/arguments/result as secondary detail.
- **FR-007**: Approval-blocked runs MUST show waiting-for-confirmation state and visible Approve, Deny, and Stop actions.
- **FR-008**: Sidebar MUST NOT show the duplicate thread search field or bottom new/search action cluster.
- **FR-008a**: The titlebar compose button MUST create a new thread for the current Chat/Work mode only when the sidebar is collapsed.
- **FR-009**: Sidebar create copy MUST distinguish New Chat and New Work.
- **FR-010**: Thread run status dots MUST have readable status labels.
- **FR-011**: Composer placeholders MUST differ between Chat and Work modes.
- **FR-012**: Sidebar thread rows MUST expose rename and delete actions.
- **FR-013**: Chat messages MUST render basic safe Markdown.
- **FR-014**: This feature MUST NOT change backend, runtime, provider, tool execution, database, M38/activity recorder, or add new tool capability.

### Key Entities

- **Provider Blocking Warning**: UI-visible warning shown only when the configured provider state blocks sending.
- **Work Task Projection**: Existing work metadata projection into goal, steps/todos, artifacts, and recent progress.
- **Human Tool Label**: User-facing label for a technical tool call or event.
- **Sidebar Search State**: Session-local text filter applied to current-mode threads.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Automated tests cover removal of the old runtime/debug header and provider-unavailable warning.
- **SC-002**: Automated tests cover removal of fake Composer controls.
- **SC-003**: Automated tests cover WorkPlanView empty, metadata-backed, and approval-blocked states.
- **SC-004**: Automated tests cover RunRail human tool labels.
- **SC-005**: Automated tests cover removed Sidebar search/footer entries and readable thread status labels.
- **SC-006**: Automated tests cover Chat/Work composer placeholder behavior.
- **SC-007**: Required validation commands complete or report exact blockers: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.
- **SC-008**: Browser smoke records screenshot path and verifies no chat runtime header, Chat/Work input, no fake Composer controls, simplified Sidebar actions, basic Markdown, human tool labels, approval blocked controls, Settings Providers/Tools, and console error count 0.

## Assumptions

- No safe OS directory picker API is available in this slice, so the correct behavior is to avoid showing a fake folder control.
- Existing run/tool metadata remains the source of truth; this feature changes presentation only.
- Raw technical details remain useful for debugging and may stay visible in secondary detail text.
