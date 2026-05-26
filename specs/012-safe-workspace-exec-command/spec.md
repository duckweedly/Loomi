# Feature Specification: M10 Safe Workspace Exec Command

**Feature Branch**: `012-safe-workspace-exec-command`

**Created**: 2026-05-26

**Status**: Draft

**Input**: Continue the Arkloop-level code-agent path after workspace read/write tools by adding the first approval-gated command execution tool.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Approve bounded workspace commands (Priority: P1)

A user sees when the agent asks to run `workspace.exec_command`. Loomi validates command arguments, cwd, timeout, and output bounds, requires approval, runs the command without a shell, and records stdout/stderr summaries through the existing tool-call lifecycle.

**Why this priority**: Code agents need test/build commands after read/write. Command execution is high risk, so the first slice must be explicit, bounded, and auditable.

**Independent Test**: Trigger a safe command such as `printf hello` in a temporary workspace. Verify approval is required, denial does not execute, approval executes once, and history replay shows exit status and bounded output.

**Acceptance Scenarios**:

1. **Given** a safe command request, **When** Loomi records it, **Then** the UI shows approval-required state with a redacted command summary.
2. **Given** the user approves the request, **When** the worker resumes, **Then** Loomi executes the command once without a shell and records terminal events.
3. **Given** the user denies the request, **When** the worker resumes or retries, **Then** no command executes.

---

### User Story 2 - Keep execution inside safety bounds (Priority: P2)

A user can trust that command execution starts in the workspace, cannot use shell wrappers, has a timeout, returns bounded output, and rejects obviously destructive commands.

**Why this priority**: Command execution can damage local state. Guardrails must be enforced before broader automation.

**Independent Test**: Attempt cwd escape, shell wrappers, missing command argv, dangerous first tokens, excessive timeout, and large output. Verify unsafe requests fail before execution and large output is truncated.

**Acceptance Scenarios**:

1. **Given** cwd escapes the workspace root, **When** the tool validates it, **Then** Loomi rejects it before execution.
2. **Given** the command starts with `sh`, `bash`, `zsh`, `rm`, `dd`, `mkfs`, `chmod`, `kill`, `shutdown`, `reboot`, `git push`, or `git reset`, **When** the tool validates it, **Then** Loomi rejects it before execution.
3. **Given** a command writes more than the output limit, **When** it completes, **Then** Loomi returns bounded stdout/stderr previews and truncation flags.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST add allowlisted `workspace.exec_command` behind the existing tool-call lifecycle.
- **FR-002**: `workspace.exec_command` MUST require explicit approval before execution.
- **FR-003**: The tool MUST accept argv-style commands only; it MUST NOT execute through a shell.
- **FR-004**: The tool MUST resolve cwd inside the workspace root before execution.
- **FR-005**: The tool MUST enforce a bounded timeout.
- **FR-006**: The tool MUST bound stdout and stderr result payloads.
- **FR-007**: The tool MUST reject shell wrappers and obvious destructive commands before execution.
- **FR-008**: The result summary MUST include relative cwd, exit code, stdout preview, stderr preview, timeout flag, and truncation flags.
- **FR-009**: Results, errors, event metadata, and UI display MUST redact secrets.
- **FR-010**: M10 MUST NOT add persistent terminal sessions, PTY, background daemons, MCP, browser automation, multi-agent delegation, external upload, or approval bypass.

### Key Entities

- **Exec Command Tool**: An approval-gated internal tool that runs a bounded argv command inside the workspace root.
- **Exec Result Summary**: A redacted, bounded payload containing exit code, stdout/stderr previews, timeout state, and truncation flags.
- **Command Safety Policy**: Validation rules for cwd, argv, timeout, shell wrapper rejection, and destructive command rejection.

## Success Criteria *(mandatory)*

- **SC-001**: Approved `workspace.exec_command` calls execute through worker execution and history-first replay.
- **SC-002**: Denied exec calls do not execute commands.
- **SC-003**: cwd escape, shell wrappers, and dangerous first tokens fail before execution in automated tests.
- **SC-004**: Timeout and output truncation behavior is covered by automated tests.
- **SC-005**: Local validation passes `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run build` from `docs-site/`, and `git diff --check`.

## Assumptions

- M7 approval and worker resume exist.
- M8/M9 workspace root validation and tool UI exist.
- The first M10 slice executes a single bounded command per tool call.
- Persistent terminals, process management, shell interactivity, MCP, browser automation, and multi-agent behavior remain later milestones.
