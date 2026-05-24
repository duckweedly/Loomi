# Feature Specification: M5.5 Settings Placeholder

**Feature Branch**: `[007-settings-placeholder]`

**Created**: 2026-05-24

**Status**: Draft

**Input**: User description: "M5 已经完成，临时加一个M5.5，先做一个设置的占位界面，就照着这个这种设计，我的也这样设计。看目前为止有哪些可以设置的，就作为实际的开发内容，其他的部分先做占位，Mock。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open a desktop-style settings surface (Priority: P1)

As a Loomi user, I want to open a Settings area that visually matches the provided desktop-style reference, so I can understand where future configuration will live without leaving the current workspace context.

**Why this priority**: This provides the visible M5.5 slice: a recognizable settings shell with Loomi's own copy, spacing, categories, and current product boundaries.

**Independent Test**: Can be tested by selecting Settings from the existing app shell and verifying a two-column settings layout appears with category navigation on the left and setting cards on the right.

**Acceptance Scenarios**:

1. **Given** the user is in the Loomi workspace, **When** they select Settings, **Then** a settings view opens with a back affordance, category list, and a General settings panel.
2. **Given** the settings view is open, **When** the user scans the category list, **Then** actual supported settings are visually distinguished from placeholder/mock categories.
3. **Given** the General category is selected, **When** the user views the right panel, **Then** settings are grouped into desktop-style cards with labels, helper text, and control affordances consistent with the provided reference image.

---

### User Story 2 - Adjust currently supported local settings (Priority: P2)

As a Loomi user, I want the settings screen to expose controls that already correspond to real current app behavior, so the page is not only decorative.

**Why this priority**: The user explicitly requested that currently configurable capabilities become actual development content while unsupported areas remain mock placeholders.

**Independent Test**: Can be tested by changing each supported setting and verifying the visible workspace state changes or persists as expected during the current local session.

**Acceptance Scenarios**:

1. **Given** the Settings General panel is open, **When** the user changes the default workspace mode, **Then** new conversations or visible workspace entry points reflect the selected mode.
2. **Given** mock runtime scenarios are currently available, **When** the user changes the default mock run scenario, **Then** subsequent mock sends use the selected scenario.
3. **Given** real API mode depends on a configured backend URL, **When** the user views backend/model gateway settings, **Then** the UI shows current mode/status and separates it from current-session provider draft fields.
4. **Given** the user enters provider draft values, **When** they fill Base URL, model ID, or API key, **Then** those values remain local to the browser session, the API key is masked, and only key presence is retained.

---

### User Story 3 - Preview future settings areas safely (Priority: P3)

As a Loomi user, I want to see placeholder categories for planned Agent Core, connectors, tools, safety, routes, and advanced settings, so the product direction is understandable without implying those features are usable yet.

**Why this priority**: Placeholder/mock sections help shape the product IA, but they must not mislead users or pull deferred platform capabilities into M5.5.

**Independent Test**: Can be tested by selecting each placeholder category and confirming it shows clear mock/coming-soon content without changing runtime, provider, tool, worker, or safety behavior.

**Acceptance Scenarios**:

1. **Given** the settings category list includes future areas, **When** the user selects a placeholder category, **Then** the detail panel clearly states that the section is a preview or not yet connected.
2. **Given** a placeholder section contains controls for future concepts, **When** the user interacts with them, **Then** no external actions, provider calls, tool execution, or persisted configuration changes occur.
3. **Given** a placeholder setting is not implemented, **When** the user sees it, **Then** the UI uses disabled or mock styling so it cannot be confused with a working setting.

---

### Edge Cases

- Settings opens while a run is active: the user can inspect settings without stopping or mutating the active run.
- Real API mode is unavailable or backend is not configured: settings shows a clear unavailable state and never echoes provider secret values.
- Provider capability endpoint returns unavailable or misconfigured providers: settings displays redacted status text only.
- User changes mock scenario while a run is already active: the change applies only to future mock runs.
- User navigates between placeholder categories: mock controls do not persist as real settings or trigger external effects.
- Small desktop window widths: settings remains usable with scrollable category/detail areas.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to open a Settings view from the existing Loomi shell.
- **FR-002**: The Settings view MUST use a desktop-style layout inspired by the provided reference: left category navigation, right content panel, grouped cards, compact rows, and toggle/select controls.
- **FR-003**: The Settings view MUST use Loomi's own labels and product language, not copy the reference app's brand, proprietary names, or private wording.
- **FR-004**: The Settings view MUST include a General category as the default selected category.
- **FR-005**: The Settings view MUST expose actual controls for currently supported local behavior: default workspace mode, interface language, and mock runtime scenario.
- **FR-006**: The Settings view MUST show read-only current backend/model gateway state when available, including data source mode and provider capability status, without exposing provider keys or raw secret configuration values.
- **FR-007**: The Providers category MUST allow current-session draft entry for Base URL, model ID, and masked API key presence without sending those values to the backend, calling a provider, persisting the draft, or echoing the API key string.
- **FR-008**: The Settings view MUST provide placeholder/mock categories for Appearance, Connectors, Plugins, Skill, MCP, Notebook, Memory, Activity Recorder, Context, Safety, Tools, Routes, and Advanced, plus mixed/status panels for Providers and About.
- **FR-009**: Placeholder/mock categories MUST be visually marked as preview, mock, disabled, or not connected.
- **FR-010**: Placeholder/mock controls MUST NOT perform external actions, execute tools, call providers, change safety boundaries, or persist real configuration.
- **FR-011**: The Settings view MUST include an About/support area with local version/status information when known and mock placeholders when not known.
- **FR-012**: Settings navigation MUST preserve the user's current workspace/thread context so returning from Settings does not reset the conversation.
- **FR-013**: The Settings view MUST avoid retaining, displaying, documenting, or logging model provider secret strings in the browser.
- **FR-014**: Changes to actual supported settings MUST be reflected in visible local behavior during the current session.
- **FR-015**: The UI MUST clearly distinguish actual working settings from mock placeholders in both copy and interaction states.

### Key Entities *(include if feature involves data)*

- **Settings Category**: A navigation item in the settings sidebar. Key attributes include label, icon meaning, group, active state, and whether the category is working or placeholder.
- **Setting Row**: A single setting displayed inside a grouped card. Key attributes include label, helper text, current value, control type, and working/mock state.
- **Local Settings State**: Current-session user preferences that can affect existing app behavior, such as workspace mode, interface language, and mock runtime scenario.
- **Runtime Capability Summary**: Read-only view of current data source mode, backend availability, provider capability, and model gateway readiness without secrets.
- **Provider Draft Settings**: Current-session provider notes for Base URL, model ID, and whether an API key was entered, without retaining or echoing the key string.
- **Placeholder Setting**: A non-functional preview item for deferred features such as provider management, connectors, tools, safety policies, routes, memory, notebooks, or activity recording.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can open Settings and identify the active General category in under 5 seconds.
- **SC-002**: 100% of working settings in M5.5 produce a visible local-session effect or status change when changed, including Chinese/English language switching.
- **SC-003**: 100% of placeholder sections are labeled as mock, preview, disabled, or not connected.
- **SC-004**: No provider API key, Authorization header, or secret value is visible, logged, documented, persisted, or echoed anywhere in the Settings UI; only key presence may be shown.
- **SC-005**: A reviewer can navigate all settings categories without triggering provider calls, tool execution, external writes, or destructive actions from placeholder controls.
- **SC-006**: The settings layout remains usable in a desktop window equivalent to the current Loomi shell without horizontal scrolling of the main content.
- **SC-007**: Returning from Settings preserves the previously selected workspace/thread context in at least 95% of manual smoke attempts.

## Assumptions

- M5.5 is a temporary vertical slice between M5 LLM Gateway and later platform settings work.
- The provided screenshot is visual direction only; Loomi will use original labels, hierarchy, and copy.
- Settings are local-session settings unless later specs define persistence.
- Provider draft entry is browser-session local in M5.5; the UI may accept a masked key input but must retain only whether a key was entered.
- Real provider management, settings persistence, backend secret storage, account/team settings, tool permissions, worker routes, memory/RAG controls, and activity recorder controls remain deferred unless separately specified.
- The current real configurable behavior includes workspace mode selection, Chinese/English language selection, mock runtime scenario selection, data source/backend status visibility, and read-only provider capability visibility.
