---
description: "M24 Sandbox Exec Command Tools feature specification"
---

# Feature Specification: M24 Sandbox Exec Command Tools

**Feature Branch**: `[032-sandbox-exec-command-tools]`

**Created**: 2026-05-26

**Status**: Draft

**Input**: User description: "Continue Arkloop-level code-agent coverage after workspace read/mutation tools by adding the next safe vertical slice: approval-gated sandbox.exec_command. Cover command execution through the existing ToolCatalog, ToolBroker, approval, worker, event, HTTP smoke, Settings, RunRail, and docs path. Keep it bounded, Work-mode-only, auditable, and do not add browser/web/artifact/plugin/multi-agent behavior in this slice."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run One Approved Command (Priority: P1)

As a Work mode user, I want the agent to request a bounded command execution and run it only after explicit approval, so Loomi can perform the smallest useful code-agent execution step through the audited tool lifecycle.

**Why this priority**: Arkloop-style code-agent coverage needs `exec_command` after read/write/edit. A single approved, bounded command is the narrowest executable slice.

**Independent Test**: Start a Work mode model-gateway run whose provider requests `sandbox.exec_command`, approve the tool call, and verify one non-destructive command runs in the workspace sandbox with safe stdout/stderr/exit metadata and provider continuation.

**Acceptance Scenarios**:

1. **Given** a Work mode run and a `sandbox.exec_command` request for a safe argv command, **When** the user approves the tool call, **Then** Loomi runs the command with bounded timeout/output and records a safe result summary.
2. **Given** the same request before approval, **When** run events are inspected, **Then** no command has executed and only safe argument preview metadata is visible.
3. **Given** the command exits non-zero, **When** execution finishes, **Then** Loomi records the exit code and bounded output without crashing the worker.

---

### User Story 2 - Enforce Command Safety Boundaries (Priority: P2)

As an operator, I want sandbox command execution to reject unsafe or out-of-scope requests before execution, so the first exec slice cannot mutate arbitrary host state or leak secrets.

**Why this priority**: Command execution is higher risk than workspace file mutation and must have stronger boundaries before broadening the tool runtime.

**Independent Test**: Attempt Chat-mode, unapproved, denied, stopped, absolute-cwd, traversal-cwd, shell-form, file-reading, path-bearing `ls`, oversized-output, and destructive command requests and verify unsafe requests do not execute.

**Acceptance Scenarios**:

1. **Given** a Chat mode run or a persona without sandbox exec enabled, **When** the provider requests `sandbox.exec_command`, **Then** Loomi rejects the tool before approval/execution.
2. **Given** a command request with `cwd` outside the workspace root, **When** Loomi validates it, **Then** the command is rejected before execution.
3. **Given** a command request outside the tiny allowlist, including file-reading commands, **When** Loomi validates it, **Then** the command is rejected before execution.

---

### User Story 3 - Show Exec Risk and Audit Trail (Priority: P3)

As a user reviewing the run, I want command execution to appear as a high-risk, approval-required sandbox action with bounded output, so I can understand what ran without exposing local secrets.

**Why this priority**: Exec tools must be observable before Loomi expands into richer sandbox sessions or streaming terminals.

**Independent Test**: Render Settings Tools and RunRail for sandbox exec events and verify high-risk/exec-capable metadata, command preview, timeout, exit code, output truncation, and failure states are visible without secrets or host absolute paths.

**Acceptance Scenarios**:

1. **Given** the tool catalog is shown, **When** `sandbox.exec_command` is present, **Then** it is marked bounded read-only command scoped, exec-capable, approval-required, non-isolated, and high risk.
2. **Given** an exec request is pending, **When** the UI renders the tool lifecycle, **Then** it shows command/cwd/timeout preview metadata without raw environment values or absolute host root.
3. **Given** command output is longer than the configured limit, **When** the result is recorded, **Then** output is truncated and marked as truncated.

### Edge Cases

- Empty argv, shell strings, command paths with separators, file-reading commands, path-bearing `ls`, and unknown cwd values are rejected.
- `cwd` must be relative and stay inside the configured workspace root.
- Destructive commands such as `rm`, `chmod 777`, `dd`, `mkfs`, `shutdown`, `reboot`, `kill -9`, and force git operations are rejected in this slice.
- Commands use bounded execution context and output capture; long-running command categories are not currently exposed.
- Output is bounded and sanitized; normal UI/API events must not contain environment variables, absolute workspace roots, credentials, provider raw payloads, or hidden local state.
- Denied, stopped, failed, terminal, or duplicate tool-call paths must not execute commands.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST add a `sandbox.exec_command` builtin tool that runs argv-form commands only after explicit approval.
- **FR-002**: The tool MUST reuse the existing tool catalog, RunContext enabled-tool snapshot, ToolBroker, approval, worker resume, and run event lifecycle.
- **FR-003**: The tool MUST be available only for Work mode runs and only when the selected persona/run context enables it.
- **FR-004**: Command execution MUST run with a relative workspace-scoped cwd and MUST reject absolute or traversal cwd values.
- **FR-005**: Command execution MUST reject shell-form commands and destructive command patterns before spawning a process.
- **FR-006**: Command execution MUST bound timeout and captured stdout/stderr bytes.
- **FR-007**: Result metadata MUST include command preview, relative cwd, exit code, timed-out flag, stdout/stderr previews, byte counts, truncation flags, and redaction status.
- **FR-008**: Normal API/UI events MUST NOT persist raw environment values, host absolute roots, credentials, provider raw payloads, or hidden local state.
- **FR-009**: Denied, stopped, failed, terminal, duplicate, or out-of-scope tool calls MUST NOT execute commands.
- **FR-010**: Settings Tools and runtime timeline surfaces MUST show sandbox exec tools distinctly from read/write workspace tools.
- **FR-011**: This feature MUST NOT add browser automation, web fetch/search, artifact runtime, activity recording, multi-agent orchestration, shell sessions, streaming terminal UI, or plugin marketplace behavior.

### Key Entities *(include if feature involves data)*

- **Sandbox Exec Tool**: A builtin sandbox tool entry with name, risk level, approval policy, exec-capable metadata, and enabled state.
- **Exec Request Summary**: Safe metadata derived from provider tool arguments: argv preview, relative cwd, timeout, output limit, and redaction flags.
- **Exec Result Summary**: Safe execution result metadata: command preview, cwd, exit code, timed-out flag, stdout/stderr previews, byte counts, truncation flags, and redaction flags.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A Work mode smoke can approve `sandbox.exec_command` and observe exactly one safe command execution with one succeeded tool call and provider continuation.
- **SC-002**: Safety tests prove Chat-mode, denied, stopped, terminal, traversal cwd, shell-form, and destructive command requests do not execute.
- **SC-003**: Timeout and output-limit tests produce bounded result summaries without worker crashes or unbounded output.
- **SC-004**: Timeline and Settings surfaces expose sandbox exec risk/approval/exec-capable metadata without host absolute paths or secret-looking content.
- **SC-005**: The feature passes backend tests, web tests/build, docs build, and browser smoke for visible sandbox exec catalog/timeline state.

## Assumptions

- `LOOMI_WORKSPACE_ROOT` or the repo root remains the only command cwd root for this slice.
- Commands are argv arrays, not shell strings.
- The first slice captures bounded output after command completion rather than streaming stdout/stderr live.
- Environment is inherited minimally by Go defaults for local development, but no env map is accepted from the model in this slice.
- Destructive commands are denied even if the model requests them; broader sandbox policies require a later feature.
