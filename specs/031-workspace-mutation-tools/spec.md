---
description: "M23 Workspace Mutation Tools feature specification"
---

# Feature Specification: M23 Workspace Mutation Tools

**Feature Branch**: `[031-workspace-mutation-tools]`

**Created**: 2026-05-26

**Status**: Draft

**Input**: User description: "Continue Arkloop-level code-agent coverage after M21/M22 by adding approval-gated workspace mutation tools. Implement the next safe vertical slice toward glob/read/edit/write/exec_command coverage: workspace.write_file and workspace.edit, with diff preview, scope guard, sensitive denylist, audit events, bounded writes, and timeline visibility. Do not add shell/exec/browser/web tools in this slice."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Write New Workspace Files (Priority: P1)

As a Work mode user, I want the agent to propose and, after explicit approval, write a bounded new text file inside the configured workspace, so Loomi can create small code or artifact files without leaving the audited run/tool path.

**Why this priority**: Arkloop-style code-agent coverage needs a safe "hand" after read tools. Creating a new file is the narrowest useful mutation and can reuse the existing approval-gated worker/tool lifecycle.

**Independent Test**: Start a Work mode model-gateway run whose provider requests `workspace.write_file`, approve the tool call, and verify the target file is created under the workspace root, tool events show requested/approval/executing/succeeded, and no write occurs before approval.

**Acceptance Scenarios**:

1. **Given** a Work mode run and a `workspace.write_file` request for a new relative text path, **When** the user approves the tool call, **Then** Loomi writes the file within the workspace root and records a redacted result summary.
2. **Given** the same request before approval, **When** the run is observed through events or UI, **Then** Loomi shows the mutation request and preview metadata without writing the file.
3. **Given** Chat mode or a persona that has not enabled workspace mutation tools, **When** the provider requests `workspace.write_file`, **Then** Loomi rejects it before execution.

---

### User Story 2 - Edit Existing Workspace Files (Priority: P2)

As a Work mode user, I want the agent to propose a bounded text replacement in an existing workspace file, so I can approve small code edits with visible before/after context.

**Why this priority**: Editing is the next required code-agent primitive after file creation, but it carries more risk because it changes existing user files.

**Independent Test**: Approve a `workspace.edit` request with an exact expected text replacement and verify the file changes once, the result summary contains safe diff metadata, and repeated execution does not duplicate the edit.

**Acceptance Scenarios**:

1. **Given** an existing text file containing the expected old text once, **When** the approved `workspace.edit` executes, **Then** Loomi replaces it once and records line/count metadata.
2. **Given** the old text is missing or appears multiple times without an explicit expected occurrence, **When** the edit executes, **Then** Loomi fails safely without modifying the file.
3. **Given** an edit request that would exceed configured size or path limits, **When** Loomi validates it, **Then** the write is denied and file contents remain unchanged.

---

### User Story 3 - Show Mutation Risk and Audit Trail (Priority: P3)

As an operator reviewing a run, I want workspace mutation tools to be visible as higher-risk, approval-required actions with safe preview and result metadata, so I can understand what changed without exposing secrets or absolute local paths.

**Why this priority**: Mutation tools must remain observable and auditable before expanding to command execution.

**Independent Test**: Render Settings Tools, ToolCallCard, RunRail, and replayed run history for workspace mutation events and verify risk, approval, preview, result, failure, and stopped states are visible without raw secrets or host absolute roots.

**Acceptance Scenarios**:

1. **Given** the tool catalog is shown, **When** workspace mutation entries are present, **Then** they are marked workspace-scoped, write-capable, approval-required, and higher risk than read tools.
2. **Given** a write/edit request is pending, **When** the UI renders the tool call, **Then** it shows path and diff/size preview metadata but no raw secret-looking content.
3. **Given** a mutation succeeds or fails, **When** events are replayed, **Then** timeline rows preserve the final mutation outcome.

### Edge Cases

- Relative paths that escape via `..`, absolute paths outside the root, symlink escapes, `.git`, `.env*`, private keys, `secrets/**`, and credential-looking targets are rejected.
- Binary or invalid UTF-8 content is rejected for mutation tools.
- Existing file writes through `workspace.write_file` are rejected unless an explicit overwrite flag is introduced in a later feature.
- Empty file content is allowed only for `workspace.write_file`; empty old text for `workspace.edit` is rejected.
- Writes are bounded by maximum bytes and line count.
- Stopped, denied, failed, or terminal runs must not mutate files.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST add a `workspace.write_file` tool that writes bounded UTF-8 text to a new relative path inside the configured workspace root only after explicit approval.
- **FR-002**: Loomi MUST add a `workspace.edit` tool that applies a bounded exact text replacement to an existing UTF-8 workspace file only after explicit approval.
- **FR-003**: Mutation tools MUST reuse the existing tool catalog, RunContext enabled-tool snapshot, ToolBroker, approval, worker resume, and run event lifecycle.
- **FR-004**: Mutation tools MUST be available only for Work mode runs and only when the selected persona/run context enables them.
- **FR-005**: Loomi MUST reject path traversal, absolute root escape, symlink escape, sensitive filenames/directories, and secret-looking targets before modifying files.
- **FR-006**: Loomi MUST record safe request/result metadata including relative path, operation, bytes written or changed, line counts, truncation/diff-preview flags, and redaction status.
- **FR-007**: Loomi MUST NOT persist raw file contents, host absolute workspace root paths, credentials, shell output, provider raw payloads, or hidden local state in normal API/UI event metadata.
- **FR-008**: Mutation execution MUST be idempotent with existing tool-call terminal guards and MUST NOT reapply a terminal edit on retry/replay.
- **FR-009**: Denied, stopped, failed, terminal, or out-of-scope tool calls MUST NOT write or edit files.
- **FR-010**: Settings Tools and runtime timeline surfaces MUST show workspace mutation tools distinctly from read-only tools.
- **FR-011**: This feature MUST NOT add shell command execution, browser automation, web fetch/search, activity recording, multi-agent orchestration, or plugin marketplace behavior.

### Key Entities *(include if feature involves data)*

- **Workspace Mutation Tool**: A builtin workspace tool entry with name, risk level, approval policy, write-capable metadata, and enabled state.
- **Mutation Request Summary**: Redacted metadata derived from provider tool arguments: operation, relative path, size/count limits, expected replacement shape, and safe preview flags.
- **Mutation Result Summary**: Redacted execution result metadata: operation, relative path, bytes written/changed, line counts, changed flag, and truncation/redaction flags.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A Work mode smoke can approve `workspace.write_file` and observe exactly one new file created under the workspace root with one succeeded tool call.
- **SC-002**: A Work mode smoke can approve `workspace.edit` and observe exactly one bounded replacement in an existing file with one succeeded tool call.
- **SC-003**: Safety tests cover traversal, absolute escape, symlink escape, sensitive paths, denied approval, stopped run, and terminal retry with no file mutation.
- **SC-004**: Timeline and Settings surfaces expose mutation tool risk/approval/write-capable metadata without host absolute paths or secret-looking content.
- **SC-005**: The feature passes backend tests, web tests/build, docs build, and a browser smoke for visible mutation catalog/timeline state.

## Assumptions

- `LOOMI_WORKSPACE_ROOT` or the repo root remains the single workspace root for this slice.
- Mutation tools operate on UTF-8 text only.
- `workspace.write_file` creates new files only; overwrite and delete are separate future features.
- `workspace.edit` uses exact old-text/new-text replacement in this slice; patch hunks, AST edits, and merge/conflict tools are future features.
- Existing approval UI is enough for approval/deny actions; this slice adds safer preview metadata rather than a full diff editor.
