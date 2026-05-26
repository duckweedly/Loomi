---
description: "M25 MCP Management + LSP Read-only Foundation feature specification"
---

# Feature Specification: M25 MCP Management + LSP Read-only Foundation

**Feature Branch**: `[033-mcp-management-lsp-readonly]`

**Created**: 2026-05-26

**Status**: Draft

**Input**: User description: "Continue Arkloop-level code-agent coverage after M24 sandbox exec. Productize local MCP visibility in Settings and add a first read-only LSP tool foundation. Keep the slice observable, Work-mode bounded, approval-gated where tools execute, and do not add remote MCP/OAuth, marketplace install, writable MCP config UI, browser/web/artifact runtime, or multi-agent orchestration yet."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Inspect Local MCP Servers (Priority: P1)

As a user configuring Loomi for local tool use, I want Settings > MCP to show safe local stdio MCP server status and discovered tools, so I can understand what is available without reading raw env or logs.

**Why this priority**: Loomi already has local MCP discovery and approval-gated execution, but MCP remains hidden behind env/test paths. Arkloop parity needs MCP management visibility before adding richer tool settings.

**Independent Test**: Configure one local stdio MCP server through the existing local config path, list MCP status through the API, and render Settings > MCP with server slug, display name, transport, enabled state, discovery status, candidate tool names, execution mode, and redacted error codes without secrets, commands, args, env values, or absolute private paths.

**Acceptance Scenarios**:

1. **Given** local MCP server config exists, **When** Settings > MCP opens, **Then** Loomi shows a read-only server row with safe status and discovered tool names.
2. **Given** discovery failed or was rejected, **When** MCP status is shown, **Then** Loomi shows a redacted error code/message class without raw command, args, env, or secret values.
3. **Given** no MCP servers are configured, **When** Settings > MCP opens, **Then** Loomi shows an explicit empty state instead of placeholder controls.

---

### User Story 2 - Expose Read-only LSP Tools (Priority: P2)

As a Work mode user, I want the agent to request read-only code intelligence tools for diagnostics, symbols, and references, so Loomi can inspect code structure without broad shell access.

**Why this priority**: Arkloop's minimum code-agent layer includes LSP after file read/write/exec. A read-only, workspace-scoped, approval-gated LSP slice adds code intelligence without opening browser/web/artifact or remote services.

**Independent Test**: Start a Work mode run whose provider requests one LSP read-only tool, approve the call, and verify Loomi returns a bounded, workspace-scoped result summary through ToolBroker, worker events, and provider continuation.

**Acceptance Scenarios**:

1. **Given** a Work mode run and an `lsp.symbols` request for a workspace file, **When** the user approves the tool call, **Then** Loomi returns bounded symbol summaries for that file.
2. **Given** an `lsp.references` request with a file position, **When** the user approves it, **Then** Loomi returns bounded workspace-relative reference locations for the symbol at that position.
3. **Given** an `lsp.diagnostics` request, **When** the user approves it, **Then** Loomi returns bounded, safe diagnostics or an explicit empty diagnostic result without running shell commands.

---

### User Story 3 - Keep LSP and MCP Safe and Visible (Priority: P3)

As an operator, I want MCP and LSP capabilities to be clearly separated from workspace mutation and sandbox exec tools, so I can audit risk, mode gating, and failure states.

**Why this priority**: MCP and LSP add new capability surfaces. They need visible, redacted, and mode-scoped behavior before user-editable management or richer code intelligence.

**Independent Test**: Render Settings Tools, Settings MCP, and RunRail with MCP/LSP metadata and verify Chat-mode rejection, safe argument previews, no secret/path leakage, and clear read-only labels.

**Acceptance Scenarios**:

1. **Given** the tool catalog is shown, **When** LSP tools are present, **Then** they are marked builtin, LSP-scoped, read-only, approval-required, executable, and low risk.
2. **Given** a Chat mode run asks for an LSP tool, **When** the gateway validates the request, **Then** Loomi rejects it before approval/execution.
3. **Given** Settings or RunRail renders MCP/LSP data, **When** the page is inspected, **Then** it shows safe server/tool summaries without raw env values, provider raw payloads, absolute host roots, or secret-looking content.

### Edge Cases

- MCP server config can be absent, disabled, failed, rejected, or succeeded.
- MCP management is read-only in this slice; users cannot create, edit, delete, enable, disable, or restart MCP servers through UI.
- LSP tool paths must stay under the workspace root and reuse sensitive path protection.
- LSP tools must not call shell commands, package managers, language servers, network services, or write files in this slice.
- LSP result payloads must be bounded and workspace-relative.
- Chat mode, missing persona allowlist, unsupported tool names, duplicate tool-call IDs, terminal runs, denied calls, and stopped runs must not execute LSP tools.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST expose a read-only MCP status API that returns safe local stdio MCP server summaries and discovered candidate tool names.
- **FR-002**: MCP status MUST be derived from existing local MCP config and discovery/run-event metadata without exposing command strings, args, env values, secrets, raw payloads, or absolute private paths.
- **FR-003**: Settings > MCP MUST render a real read-only management surface using the MCP status API or mock equivalent, not a placeholder panel.
- **FR-004**: Loomi MUST add builtin read-only LSP tool catalog entries for `lsp.diagnostics`, `lsp.symbols`, and `lsp.references`.
- **FR-005**: LSP tools MUST reuse the existing tool catalog, RunContext enabled-tool snapshot, ToolBroker, approval, worker resume, and run event lifecycle.
- **FR-006**: LSP tools MUST be available only for Work mode runs and only when the selected persona/run context enables them.
- **FR-007**: LSP tools MUST validate workspace-relative paths and reuse existing workspace root, traversal, symlink escape, and sensitive path boundaries.
- **FR-008**: LSP tool results MUST be bounded, UTF-8 safe, workspace-relative, and marked read-only.
- **FR-009**: Settings Tools and RunRail MUST show LSP tool risk/scope/read-only lifecycle metadata distinctly from workspace, sandbox, runtime, and MCP tools.
- **FR-010**: Normal API/UI events MUST NOT persist raw provider payloads, absolute host roots, hidden local state, credentials, env values, or secret-looking content.
- **FR-011**: This feature MUST NOT add remote MCP/OAuth, marketplace install, writable MCP config UI, language-server process management, shell/package-manager diagnostics, browser automation, web fetch/search, artifact runtime, activity recording, or multi-agent orchestration.

### Key Entities *(include if feature involves data)*

- **MCP Server Status**: A safe read-only summary of one configured local stdio MCP server, including slug, display name, transport, enabled state, config source, discovery status, candidate names, execution mode, and redacted error code.
- **LSP Tool**: A builtin read-only code-intelligence tool entry with name, scope, approval policy, execution state, and safe argument metadata.
- **LSP Request Summary**: Safe metadata derived from provider tool arguments: workspace-relative path, query text or file position, language hint, include-declaration flag, limit, and redaction status.
- **LSP Result Summary**: Safe bounded output: tool name, operation, workspace-relative path, symbol/reference/diagnostic rows, counts, truncation flags, and redaction status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Settings > MCP can show configured, empty, failed, and succeeded local MCP status states without placeholder copy or secret leakage.
- **SC-002**: A Work mode smoke can approve one LSP tool call and observe exactly one safe read-only execution with provider continuation.
- **SC-003**: Safety tests prove Chat-mode, denied, stopped, terminal, traversal path, symlink escape, sensitive path, and unsupported LSP requests do not execute.
- **SC-004**: Settings Tools and RunRail expose LSP scope/read-only/approval lifecycle metadata without host absolute paths or secret-looking content.
- **SC-005**: The feature passes backend tests, web tests/build, docs build, and browser smoke for visible MCP management and LSP tool states.

## Assumptions

- Existing `LOOMI_MCP_SERVERS_JSON` remains the source of local MCP server config for this read-only management slice.
- MCP config mutation is intentionally deferred; users can change local config outside Loomi and restart/refresh through later features.
- The first LSP foundation may use deterministic workspace text analysis for safe symbols/references/diagnostics rather than spawning a full language server.
- LSP tools are approval-gated read-only tools, not background watchers.
