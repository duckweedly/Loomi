# Feature Specification: M9 Safe Workspace Write Tools

**Feature Branch**: `011-safe-workspace-write-tools`

**Created**: 2026-05-26

**Status**: Draft

**Input**: Continue the Arkloop-level code-agent path after M8 read tools by adding the first approval-gated workspace write tools.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Approve workspace file creation or replacement (Priority: P1)

A user sees when the agent asks to write a file under the workspace with `workspace.write_file`. Loomi validates the path and content boundaries, requires approval, executes the write once after approval, and records the result through the existing tool-call lifecycle.

**Why this priority**: A useful code-agent loop needs safe file writes after safe reads. Write access must be explicit, auditable, scoped, and testable before adding shell execution or broader automation.

**Independent Test**: Trigger `workspace.write_file` with a safe relative path and content. Verify approval is required, denial does not write, approval writes once, and history replay shows terminal tool events.

**Acceptance Scenarios**:

1. **Given** a safe write request, **When** Loomi records it, **Then** the UI shows approval-required state with a redacted content summary.
2. **Given** the user approves the request, **When** the worker resumes, **Then** Loomi writes the file once and records `tool_call_succeeded`.
3. **Given** the user denies the request, **When** the worker resumes or retries, **Then** no file is written and the run remains auditable.

---

### User Story 2 - Apply exact text edits safely (Priority: P2)

A user sees when the agent asks to replace an exact text span in an existing workspace file with `workspace.edit`. Loomi validates the target path, rejects ambiguous or missing matches, requires approval, applies one deterministic edit, and records a bounded result summary.

**Why this priority**: Exact replacement is the smallest useful edit primitive. It avoids broad patch parsers while still supporting real code changes.

**Independent Test**: Create a file, request a single exact replacement, approve it, and verify exactly one replacement happened. Repeat with missing or duplicate old text and verify safe failure without mutation.

**Acceptance Scenarios**:

1. **Given** `old_text` occurs exactly once, **When** the approved edit executes, **Then** Loomi replaces it with `new_text` and reports changed byte counts.
2. **Given** `old_text` is missing, **When** the approved edit executes, **Then** Loomi fails the tool call and leaves the file unchanged.
3. **Given** `old_text` occurs more than once, **When** the approved edit executes, **Then** Loomi fails the tool call and leaves the file unchanged.

---

### User Story 3 - Keep write access inside safety boundaries (Priority: P3)

A user can trust that write tools cannot modify files outside the workspace root, cannot touch sensitive files, and cannot silently create broad directory trees.

**Why this priority**: Write tools are higher risk than read tools. Safety failures must be prevented before introducing shell execution.

**Independent Test**: Attempt absolute paths, `..`, symlink escapes, `.env`, `.ssh`, `secrets/`, binary-looking content, huge payloads, and missing parent directories. Verify unsafe requests fail before mutation.

**Acceptance Scenarios**:

1. **Given** a target escapes the workspace root, **When** the tool validates it, **Then** Loomi rejects it before writing.
2. **Given** a target is sensitive, **When** the tool validates it, **Then** Loomi rejects it before writing.
3. **Given** the parent directory does not exist, **When** `workspace.write_file` runs, **Then** Loomi rejects it instead of creating unexpected directories.

## Edge Cases

- Absolute paths, `..` traversal, symlink escape, repeated separators, and Windows-style separators.
- Sensitive targets such as `.env*`, `.ssh/`, `.aws/`, `secrets/`, `credentials/`, private key basenames, and `.pem`.
- Oversized write or edit content.
- Missing parent directory for write.
- Editing a missing file, directory, binary file, or file containing invalid UTF-8.
- Missing or duplicate `old_text`.
- Concurrent approve retries for the same write call.
- Cancellation while a write tool is pending or executing.
- History-first SSE reconnect after approval, execution, failure, or cancellation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST add allowlisted `workspace.write_file` and `workspace.edit` internal tools behind the existing M7/M8 tool-call lifecycle.
- **FR-002**: Every workspace write tool MUST require explicit approval before execution.
- **FR-003**: Tool arguments MUST be schema-validated before approval state is shown.
- **FR-004**: Tool execution MUST be scoped to the configured workspace root.
- **FR-005**: Loomi MUST reject path traversal, absolute paths, symlink escapes, sensitive path patterns, directories, and missing parent directories before mutation.
- **FR-006**: Loomi MUST bound write and edit content size.
- **FR-007**: `workspace.write_file` MUST write UTF-8 text content to an existing parent directory and return a bounded summary with relative path and bytes written.
- **FR-008**: `workspace.edit` MUST replace `old_text` with `new_text` only when `old_text` appears exactly once in the UTF-8 target file.
- **FR-009**: Failed validation or failed exact-match edits MUST leave files unchanged.
- **FR-010**: Results, errors, event metadata, and UI display MUST redact secrets and avoid raw provider payloads.
- **FR-011**: Existing approval idempotency, cancellation precedence, terminal guards, and history-first SSE replay MUST continue to apply.
- **FR-012**: M9 MUST NOT add shell execution, command execution, arbitrary network, MCP, browser automation, multi-agent delegation, or external upload.

### Key Entities

- **Workspace Write Tool**: An allowlisted internal tool that mutates bounded text files under the workspace root after approval.
- **Write Target**: A relative path under the workspace root whose parent directory already exists and is not sensitive.
- **Exact Edit**: A deterministic single replacement from `old_text` to `new_text`.
- **Write Result Summary**: A redacted, bounded payload suitable for run events and UI replay.

## Success Criteria *(mandatory)*

- **SC-001**: Approved `workspace.write_file` and `workspace.edit` calls complete through worker execution and history-first replay.
- **SC-002**: Denied write calls do not mutate files.
- **SC-003**: Attempts to write outside the workspace root, through symlinks, or into sensitive paths fail before mutation in automated tests.
- **SC-004**: Exact edit missing-match and duplicate-match cases fail without changing file content.
- **SC-005**: Local validation passes `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run build` from `docs-site/`, and `git diff --check`.

## Assumptions

- M7 approval, tool-call projection, worker resume, and tool UI exist.
- M8 workspace path validation patterns are reused and hardened for mutation.
- The first M9 slice uses text-only write/edit primitives.
- Shell/exec, patch application, MCP, browser automation, sandbox, multi-agent delegation, and activity recording remain later milestones.
