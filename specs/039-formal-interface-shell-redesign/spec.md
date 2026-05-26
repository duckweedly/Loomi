# Feature Specification: Formal Interface Shell Redesign

**Feature Branch**: `039-formal-interface-shell-redesign`

**Created**: 2026-05-26

**Status**: Draft

**Input**: User description: "UI-01 Formal Interface Shell Redesign: adjust Loomi toward a light desktop application shell with a narrow workspace/session sidebar, large white chat canvas, centered content, and fixed bottom composer while preserving Loomi identity and existing Chat/Work capabilities."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open a Formal Desktop Shell (Priority: P1)

As a Loomi user, I can open the app and immediately see a light, desktop-like workspace with a narrow left sidebar, a large white canvas, centered chat content, and a fixed bottom composer.

**Why this priority**: This is the primary first-round visual direction and must be visible before detailed refinement.

**Independent Test**: Open the local web app and verify the overall proportions: soft outer background, rounded main window, light sidebar, white main canvas, centered content, and bottom composer.

**Acceptance Scenarios**:

1. **Given** the app is loaded on a desktop viewport, **When** the workspace renders, **Then** the shell presents a rounded light desktop window with a narrow sidebar and white main canvas.
2. **Given** an existing chat thread, **When** messages render, **Then** message content stays centered with a bounded maximum width rather than stretching across the screen.
3. **Given** a narrow viewport, **When** the shell renders, **Then** primary regions do not visibly overlap or overflow.

---

### User Story 2 - Preserve Existing Conversation Controls (Priority: P2)

As a Loomi user, I can continue using Chat and Work mode from the redesigned shell without losing existing send, stop, provider, persona, and folder-entry affordances.

**Why this priority**: The redesign must not break the runnable vertical slice or hide execution controls.

**Independent Test**: In the browser, type in Chat mode, switch to Work mode, submit a run, and verify Stop appears while a run is active.

**Acceptance Scenarios**:

1. **Given** Chat mode is selected, **When** I type into the composer, **Then** the composer accepts input and exposes the existing send action.
2. **Given** Work mode is selected, **When** I type into the composer and submit, **Then** a Work run can start using the existing runtime flow.
3. **Given** a run is active, **When** the canvas renders, **Then** a Stop control remains visible.
4. **Given** provider setup is unavailable, **When** the canvas renders, **Then** the existing provider warning remains clear and actionable.

---

### User Story 3 - Keep Secondary Panels Usable (Priority: P3)

As a Loomi user, I can still open Settings, Tools, and RunRail after the shell redesign.

**Why this priority**: Secondary panels are not redesigned in this round, but they must not be broken by the new layout.

**Independent Test**: Open Settings > Tools and RunRail from the redesigned shell and verify they render without layout breakage.

**Acceptance Scenarios**:

1. **Given** the redesigned shell, **When** I open Settings, **Then** the existing Settings view occupies the main workspace region and remains navigable.
2. **Given** Settings is open, **When** I select Tools, **Then** the existing Tools catalog remains readable.
3. **Given** a run exists, **When** I open run details, **Then** RunRail remains visible and usable.

### Edge Cases

- Active run states must keep Stop visible in both canvas context and composer actions.
- Provider-unavailable state must not be hidden behind the fixed composer.
- Sidebar content with many threads must scroll inside the sidebar without shifting the main canvas.
- Long thread titles and compact mobile widths must truncate or wrap without overlapping controls.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The app MUST default this first-round shell to a light desktop-window visual treatment with a soft blue-purple outer background, rounded main window, subtle border, and shadow.
- **FR-002**: The sidebar MUST be narrower and lighter than the previous shell, with a workspace area, compact mode icon row, scrollable conversation list, and bottom new conversation/search controls.
- **FR-003**: The main region MUST use a large white canvas with a minimal top bar and centered chat content constrained by a maximum readable width.
- **FR-004**: The composer MUST stay fixed to the bottom center of the chat canvas and preserve existing send, Stop, provider, persona, and Work in Folder entry points.
- **FR-005**: The redesign MUST preserve existing Chat mode send behavior, Work mode run behavior, active-run Stop visibility, and provider-unavailable warnings.
- **FR-006**: Settings, Tools, and RunRail MUST remain accessible and not be redesigned beyond layout compatibility.
- **FR-007**: The implementation MUST avoid copying the reference product's brand, private icons, application names, copywriting, or proprietary expression; Loomi identity and wording remain Loomi-specific.
- **FR-008**: The implementation MUST not change backend, runtime, tool, provider, memory, work execution, database, or M38/activity-recorder behavior except for frontend compile fixes if required.
- **FR-009**: The first round MUST explicitly leave pixel-level polishing and detailed visual refinements for later review.

### Key Entities

- **Workspace Shell**: The outer desktop-like container that frames sidebar, main canvas, and secondary panels.
- **Thread Sidebar**: The narrow left navigation and conversation list for the current Chat or Work mode.
- **Chat Canvas**: The primary white message surface, state notices, provider warning, and execution context.
- **Composer**: The bottom fixed input area and existing run/provider/persona/folder controls.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Desktop browser smoke confirms the visible layout direction matches the requested proportions: narrow sidebar, large white canvas, centered content, fixed bottom composer.
- **SC-002**: Chat mode accepts typed input in the composer during browser smoke.
- **SC-003**: Work mode accepts typed input and can start a run during browser smoke.
- **SC-004**: Active run Stop is visible during browser smoke.
- **SC-005**: Settings > Tools opens successfully during browser smoke.
- **SC-006**: Browser console has 0 errors after the smoke path.
- **SC-007**: Required validation commands complete or report exact blockers: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.

## Assumptions

- The reference image is used only for layout mechanism and visual proportion, not for brand, icon, text, or private expression.
- This round uses the existing React/CSS structure and existing dependencies.
- No new backend capability, database migration, runtime behavior, provider behavior, memory behavior, or tool capability is part of this feature.
- Detailed spacing, iconography, density, and copy refinements will happen in later user-directed iterations after this shell can be opened and tested.
