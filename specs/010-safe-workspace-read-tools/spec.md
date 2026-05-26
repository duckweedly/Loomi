# Feature Specification: M8 Safe Workspace Read Tools

**Feature Branch**: `010-safe-workspace-read-tools`

**Created**: 2026-05-26

**Status**: Draft

**Input**: Continue the Arkloop-level code-agent path after M7 by adding the first safe workspace read tools.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Approve workspace read requests (Priority: P1)

A user sees when the agent asks to inspect the local workspace with `workspace.glob`, `workspace.grep`, or `workspace.read_file`. Loomi validates the request, shows a redacted summary, requires approval, and records the lifecycle through the existing M7 tool-call events.

**Why this priority**: Workspace read tools are the next useful step after approval-gated tools, but local file access must remain explicit and observable.

**Independent Test**: Trigger each read tool through a controlled tool-call request. Verify the tool remains blocked until approval, rejected arguments never execute, and approved calls record executing plus terminal events.

**Acceptance Scenarios**:

1. **Given** a supported workspace read tool request, **When** the request is recorded, **Then** Loomi persists requested and approval-required tool events with a safe summary.
2. **Given** the user approves the request, **When** the worker resumes, **Then** Loomi executes the read-only tool once and records a redacted result event.
3. **Given** the user denies the request, **When** the worker resumes or retries, **Then** no filesystem read occurs and the run reaches a denied or stopped visible state.

---

### User Story 2 - Keep reads inside the workspace (Priority: P2)

A user can trust that read tools only inspect files under the configured workspace root and do not read secrets, hidden sensitive directories, or paths outside the project.

**Why this priority**: File reads expose user data. The boundary must be enforced before the tools are useful.

**Independent Test**: Attempt absolute paths, `..`, symlink escapes, `.env`, `.ssh`, `secrets/`, binary files, oversized files, and unsupported globs. Verify all unsafe requests fail before reading content and emit redacted errors.

**Acceptance Scenarios**:

1. **Given** a path escapes the workspace root, **When** the tool validates it, **Then** Loomi rejects it without reading.
2. **Given** a sensitive file pattern is requested, **When** the tool validates it, **Then** Loomi rejects it with a safe error summary.
3. **Given** a file is too large or binary, **When** `workspace.read_file` runs, **Then** Loomi returns a bounded safe failure or preview without dumping raw content.

---

### User Story 3 - Surface readable results in the UI (Priority: P3)

A user reviewing a run can distinguish workspace read results from model output and see bounded summaries in ToolCallCard, RunRail, and Timeline.

**Why this priority**: Code-agent behavior must be auditable without opening backend logs.

**Independent Test**: Replay runs containing approved `glob`, `grep`, and `read_file` results. Verify the UI labels the tool name, approval state, and bounded result summary without secrets or raw provider payloads.

**Acceptance Scenarios**:

1. **Given** an approved glob result, **When** it appears in history replay, **Then** the UI shows a bounded list/count summary.
2. **Given** an approved grep result, **When** it appears in history replay, **Then** the UI shows bounded matches with file paths relative to the workspace.
3. **Given** an approved file read result, **When** it appears in history replay, **Then** the UI shows a bounded text preview and metadata.

## Edge Cases

- Path traversal through `..`, absolute paths, repeated separators, URL-encoded path text, or symlink escape.
- Sensitive targets such as `.env*`, `.ssh/`, `.aws/`, `secrets/`, `credentials/`, private keys, lockbox files, and generated credential dumps.
- Very broad glob or grep patterns.
- Binary, huge, unreadable, missing, or permission-denied files.
- Concurrent approvals for the same tool call.
- Cancellation while a read tool is waiting or executing.
- History-first SSE reconnect after approval, execution, failure, or cancellation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST add allowlisted `workspace.glob`, `workspace.grep`, and `workspace.read_file` internal tools behind the existing M7 tool-call lifecycle.
- **FR-002**: Every workspace read tool MUST require explicit approval before execution in M8.
- **FR-003**: Tool arguments MUST be schema-validated before approval state is shown.
- **FR-004**: Tool execution MUST be read-only and MUST NOT write files, execute shell commands, call the network, invoke MCP, automate browsers, or modify process state.
- **FR-005**: Every requested path or glob root MUST resolve inside the configured workspace root before any filesystem read.
- **FR-006**: Loomi MUST reject sensitive path patterns including `.env*`, private key names, `.ssh/`, `.aws/`, `secrets/`, and `credentials/`.
- **FR-007**: `workspace.glob` MUST return bounded relative paths and counts.
- **FR-008**: `workspace.grep` MUST return bounded relative path, line number, and line preview matches.
- **FR-009**: `workspace.read_file` MUST return bounded UTF-8 text previews with metadata and reject binary or oversized files.
- **FR-010**: Results, errors, event metadata, and UI display MUST redact secrets and avoid raw provider payloads.
- **FR-011**: Existing approval idempotency, cancellation precedence, terminal guards, and history-first SSE replay MUST continue to apply.
- **FR-012**: The frontend MUST render workspace read tools through existing ToolCallCard, RunRail, and Timeline pathways without adding a separate file browser.

### Key Entities

- **Workspace Read Tool**: An allowlisted internal tool that reads bounded project data under the workspace root after approval.
- **Workspace Root**: The configured local root directory used to resolve and constrain all read requests.
- **Read Result Summary**: A redacted, bounded result payload suitable for run events and UI replay.
- **Sensitive Path Policy**: The denylist and validation behavior that prevents accidental secret reads.

## Success Criteria *(mandatory)*

- **SC-001**: All three workspace read tools can complete approved local smoke runs and replay through history-first SSE.
- **SC-002**: Attempts to read outside the workspace root or sensitive files fail before content is read in automated tests.
- **SC-003**: Repeated approve requests execute each read tool at most once.
- **SC-004**: UI smoke shows requested, approval-required, executing, and terminal read states with no console errors.
- **SC-005**: Local validation passes `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.

## Assumptions

- M7 approval, tool-call projection, worker resume, and ToolCallCard/RunRail/Timeline display states exist.
- M8 uses the local repository root as the default workspace root for development.
- The first slice executes one workspace read tool per run through the existing worker flow.
- Write tools, `edit`, `exec_command`, MCP, browser automation, multi-agent delegation, and activity recording are separate later milestones.
