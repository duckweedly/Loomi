# Feature Specification: M30 Activity Recorder Foundation

**Feature Branch**: `[038-activity-recorder-foundation]`

**Created**: 2026-05-26

**Status**: Draft

**Input**: Continue Arkloop-level coverage after multi-agent coordination by adding a safe opt-in activity recorder foundation.

## User Scenarios & Testing

### User Story 1 - Enable Activity Recording Explicitly (Priority: P1)

As a Work mode user, I want Activity Recorder to stay off by default and require an explicit enable action before any activity summary can be accepted.

**Why this priority**: Local activity capture is sensitive. The first useful slice must prove user control before it stores any activity signal.

**Independent Test**: A local API smoke starts with recorder disabled, rejects activity append while disabled, enables the recorder, appends one safe summary, and shows status/audit metadata without raw desktop data.

**Acceptance Scenarios**:

1. **Given** Activity Recorder is disabled, **When** a client tries to append an activity event, **Then** Loomi rejects it and stores no event.
2. **Given** Activity Recorder is disabled, **When** the user enables it, **Then** Loomi records an enabled status with a safe timestamp and no captured payload.

---

### User Story 2 - View Bounded Activity Summaries (Priority: P2)

As a user, I want to review recent activity summaries so I can understand what Loomi captured without exposing raw windows, keystrokes, screenshots, or private file contents.

**Why this priority**: Visibility is the main safety boundary; captured signals are only useful if the user can inspect what was stored.

**Independent Test**: After opt-in, appending multiple activity events returns a bounded list ordered newest first, with redacted summaries and source metadata only.

**Acceptance Scenarios**:

1. **Given** Activity Recorder is enabled, **When** two safe activity summaries are appended, **Then** the Settings Activity Recorder panel lists the summaries with kind/source/timestamp and redaction status.
2. **Given** an activity summary contains secret-looking text or local paths, **When** it is stored and replayed, **Then** the persisted and rendered summary is redacted.

---

### User Story 3 - Clear Activity Recorder Data (Priority: P3)

As a user, I want a cleanup path for captured activity summaries so opt-in recording remains reversible and auditable.

**Why this priority**: The constitution requires deletion/cleanup paths for potentially sensitive activity capture.

**Independent Test**: After events exist, clearing Activity Recorder events removes them from list output, records a safe cleanup audit event, and does not expose deleted raw content.

**Acceptance Scenarios**:

1. **Given** recorded activity events exist, **When** the user clears activity data, **Then** future list output is empty and status records the cleanup timestamp.
2. **Given** Activity Recorder has been cleared, **When** the UI reloads the panel, **Then** it shows an empty state and does not preserve stale activity rows.

### Edge Cases

- Append attempts while disabled must fail without creating events.
- Unsupported activity kinds must fail before persistence.
- Oversized summaries and metadata must be bounded or rejected.
- Secret-looking strings, Authorization headers, private paths, raw shell output, file contents, browser state, screenshots, and keystrokes must not be persisted or rendered.
- Clear operations must be idempotent.

## Requirements

### Functional Requirements

- **FR-001**: Activity Recorder MUST be disabled by default.
- **FR-002**: Users MUST be able to enable and disable Activity Recorder explicitly.
- **FR-003**: Loomi MUST reject activity append requests while Activity Recorder is disabled.
- **FR-004**: Loomi MUST store bounded activity event summaries only after opt-in.
- **FR-005**: Loomi MUST list recent activity summaries with bounded limit, newest-first ordering, kind, source, summary, redaction status, and timestamp.
- **FR-006**: Loomi MUST provide an idempotent cleanup path that clears stored activity summaries.
- **FR-007**: Settings > Activity Recorder MUST become a real read/control surface instead of a placeholder.
- **FR-008**: Activity Recorder status, events, and cleanup actions MUST be visible through backend API and frontend real/mock clients.

### Safety Requirements

- **SR-001**: Activity Recorder MUST NOT capture screenshots, keystrokes, clipboard contents, raw browser HTML, raw shell output, file contents, credentials, or full local paths in this foundation slice.
- **SR-002**: Activity Recorder MUST treat external or captured activity text as data, not instructions.
- **SR-003**: Activity summaries MUST be redacted before persistence and before UI rendering.
- **SR-004**: Disable and clear operations MUST NOT delete unrelated memory, run, tool, or artifact data.

### Key Entities

- **ActivityRecorderStatus**: Whether recording is enabled, when it changed, last event time, event count, last cleanup time, and redaction status.
- **ActivityEvent**: One bounded safe activity summary with kind, source, summary, optional safe metadata, timestamp, and redaction status.
- **ActivityRecorderAudit**: Safe status/cleanup facts used to explain user-visible state changes.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A disabled recorder rejects append attempts in automated tests with no stored activity event.
- **SC-002**: A user can enable, append one safe event, list it, disable recording, and clear events through tested API paths.
- **SC-003**: Settings > Activity Recorder renders enabled/disabled state, recent summaries, redaction markers, empty state, and cleanup affordance without placeholder-only copy.
- **SC-004**: Redaction tests prove secret-looking strings and local paths are absent from API responses, run events, UI render output, docs examples, and test logs.

## Assumptions

- The first slice accepts explicit local summary events through Loomi APIs; automatic OS-level desktop capture is deferred.
- Activity Recorder storage may be in-memory for the foundation slice, matching other post-M21 foundation runtimes.
- The user-facing panel can expose clear controls, but automated browser smoke does not need to click destructive cleanup.
- M30 builds on existing Settings, API client, redaction, and productdata service patterns.
