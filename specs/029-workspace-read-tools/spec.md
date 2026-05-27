# Feature Specification: M21 Workspace Read Tools

**Feature Branch**: `029-workspace-read-tools`

**Created**: 2026-05-25

**Status**: Ready for planning

**Input**: User description: "Complete M21 with bounded read-only workspace tools: workspace.glob, workspace.grep, and workspace.read. Integrate through existing tool catalog, broker, run context, work persona, settings catalog, timeline events, docs, and smoke tests. Study Arkloop read-only mechanisms only; do not copy brand, copy, private APIs, or large sandbox architecture."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Approved Workspace Reading (Priority: P1)

A Work mode run can request `workspace.glob`, `workspace.grep`, or `workspace.read`, execute the bounded read automatically inside the selected workspace root, and continue with a safe result.

**Why this priority**: This is the smallest useful code-agent read loop and proves the existing tool-call path works for local workspace read access without making the user approve every read-only glob/grep/read.

**Independent Test**: Start a backend smoke run with a fixture workspace root, request each read-only tool, and verify `tool_call_succeeded` events plus provider continuation without sensitive content or manual approval prompts.

**Acceptance Scenarios**:

1. **Given** a Work mode run requests `workspace.glob`, **When** a workspace root is available, **Then** the worker executes through ToolBroker and emits a bounded result followed by continuation without a per-call approval prompt.
2. **Given** a Work mode run requests `workspace.read`, **When** the path is inside the workspace root and not sensitive, **Then** the worker reads only the bounded slice and records safe result metadata.
3. **Given** a Chat mode run, **When** the provider requests workspace tools, **Then** persona tool policy prevents unbounded workspace access.

---

### User Story 2 - Workspace Boundary Protection (Priority: P2)

Local workspace tools enforce a configured workspace root/scope and refuse absolute path escape, `..` traversal, symlink escape, and sensitive file access.

**Why this priority**: Local file access is only acceptable if workspace and secret boundaries are explicit, testable, and observable.

**Independent Test**: Run backend smoke cases against fixture files for valid paths, outside-root traversal, sensitive patterns, and symlink escape.

**Acceptance Scenarios**:

1. **Given** `LOOMI_WORKSPACE_ROOT` is set, **When** a tool receives a path inside the root, **Then** it resolves and executes inside that root only.
2. **Given** a path targets `.env`, `secrets/**`, `credentials/**`, private keys, or sensitive home directories, **When** any read tool is executed, **Then** it fails safely without leaking file content.
3. **Given** a symlink inside the workspace points outside the root, **When** a read tool follows the path, **Then** the tool refuses it as outside scope.

---

### User Story 3 - Operator Visibility (Priority: P3)

Settings and timelines show workspace tools as read-only executable capabilities with safe metadata and clear request/auto-approved/execution/success/failure states.

**Why this priority**: Users need to understand what local workspace capability a run used and what happened after execution.

**Independent Test**: Load Settings > Tools and Work/Chat timeline in the web app, verify workspace group/risk labels and event rendering without exposing host absolute paths.

**Acceptance Scenarios**:

1. **Given** Settings > Tools catalog is open, **When** workspace tools are available, **Then** they appear under a workspace group with read-only/executable risk metadata.
2. **Given** a workspace tool call progresses, **When** the timeline renders events, **Then** requested, auto-approved, executing, succeeded, and failed states are visible.
3. **Given** tool metadata includes workspace scope, **When** the UI renders it, **Then** local absolute paths are not exposed.

### Edge Cases

- Empty glob/grep results return success with zero matches, not an execution failure.
- Large read results are byte-bounded, UTF-8 safe, and explicitly marked `truncated`.
- Grep/glob results stop at configured limits and return safe metadata about truncation/counts.
- Binary or invalid text files are summarized as unsupported text content without returning raw bytes.
- Tool arguments with absolute paths, `..`, home expansion, or path separators that escape scope are rejected.
- `.git` internals and known secret/key file patterns are denied even when they live under the workspace root.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST register `workspace.glob`, `workspace.grep`, and `workspace.read` in the existing tool catalog with workspace group metadata and read-only/executable risk labels.
- **FR-002**: System MUST execute workspace read tools only through the existing ToolBroker and RunContext tool-call path.
- **FR-003**: System MUST auto-approve `workspace.glob`, `workspace.grep`, and `workspace.read` as bounded read-only tools once a workspace root is available; workspace mutation tools remain explicit-approval gated.
- **FR-004**: System MUST bind workspace tools to a single workspace root resolved from persisted local user config, `LOOMI_WORKSPACE_ROOT`, or the local desktop/dev Home fallback.
- **FR-005**: System MUST reject path traversal, absolute path escape, symlink escape, and scope violations before reading file contents.
- **FR-006**: System MUST reject sensitive file patterns including `.env*`, `secrets/**`, `credentials/**`, `id_rsa*`, `id_ed25519*`, `*.pem`, `.git` internals, `~/.ssh`, `~/.aws`, and `~/.gnupg`.
- **FR-007**: `workspace.read` MUST support max bytes plus offset/limit style bounded reads and return UTF-8 safe text summaries with `truncated` metadata.
- **FR-008**: `workspace.grep` and `workspace.glob` MUST support caller limits, path filtering, result count caps, and safe metadata.
- **FR-009**: Production workspace tool implementations MUST use Go filesystem and scanner APIs, not shelling out to `rg` or other command-line search tools.
- **FR-010**: Work mode persona MUST allow these read-only workspace tools; Chat mode MUST not broaden local workspace access beyond existing policy.
- **FR-011**: Settings > Tools catalog MUST display workspace tools without exposing host absolute paths.
- **FR-012**: Work/Chat timeline MUST render workspace tool requested, auto-approved, executing, succeeded, and failed states.
- **FR-013**: System MUST NOT implement shell execution, file writing, file editing, browser automation, web search, artifact creation, or sandbox architecture in this milestone.
- **FR-014**: Work mode MUST scope enabled tools to the latest user intent so casual chat does not expose workspace, sandbox, agent, artifact, browser, or web tools, and folder listing/classification starts with a single broad workspace glob before summarizing.

### Key Entities *(include if feature involves data)*

- **Workspace Tool Definition**: Catalog metadata for a read-only workspace capability, including name, group, risk, approval policy, and safe display fields.
- **Workspace Scope**: Resolved root and validation rules used by tool execution; the root itself is not exposed in UI metadata.
- **Workspace Tool Result**: Safe structured result with relative paths, matches or content summary, counts, truncation flags, and error metadata.
- **Tool Call Event**: Existing run event sequence showing request, auto-approval, execution, success, failure, and continuation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Backend smoke demonstrates successful auto-approved glob, read, and grep against a fixture workspace.
- **SC-002**: Backend smoke demonstrates rejection of traversal, sensitive files, and symlink escape without leaking sensitive content.
- **SC-003**: Workspace read-only tool calls do not produce `approval_required`; sensitive or out-of-scope paths fail safely without leaking content.
- **SC-004**: Settings catalog shows all three workspace tools with read-only/workspace metadata and no local absolute path.
- **SC-005**: Timeline displays all required workspace tool states for Work/Chat views.
- **SC-006**: Required validation commands complete or the exact blocker is reported: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.

## Assumptions

- The existing M18 ToolCatalog, ToolBroker, RunContext, worker, and tool events are the integration path and remain the source of truth.
- The default workspace root for local development is the persisted local folder, then `LOOMI_WORKSPACE_ROOT`, then the user's Home directory when unset.
- Workspace read results should favor safe, small summaries over full fidelity when content is large, binary, or invalid UTF-8.
- Arkloop is used only as read-only mechanism reference for registry, allowlist, filesystem tools, truncation, boundaries, sensitive file protection, and approval semantics.
- Arkloop's mechanism reference reinforces a narrow callable tool surface: Loomi should expose only the tool families that match the latest user request rather than every Work capability at once.
