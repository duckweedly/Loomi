# Feature Specification: M11 Tool Catalog Visibility

**Feature Branch**: `013-tool-catalog-visibility`

**Created**: 2026-05-26

**Status**: Draft

**Input**: Make implemented runtime/workspace tools visible through a read-only catalog API and Settings Tools panel.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Inspect available tools (Priority: P1)

A user opens Settings > Tools and sees the actual allowlisted tools Loomi can request through the approval flow, including runtime, workspace read, workspace write, and exec command tools.

**Why this priority**: Implemented tools need an auditable catalog. Otherwise the user cannot inspect the active code-agent surface without reading source code.

**Independent Test**: Call the tool catalog API and render Settings > Tools. Verify all current allowlisted tool names appear with approval, risk, capability, and safety information.

**Acceptance Scenarios**:

1. **Given** the local API is available, **When** the frontend loads Settings > Tools, **Then** it displays the backend tool catalog instead of the placeholder panel.
2. **Given** a tool requires approval, **When** it is listed, **Then** the row shows approval required and read/write/exec risk classification.
3. **Given** provider secrets or raw schemas exist internally, **When** the catalog is returned, **Then** no secret values or raw provider payloads appear.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST expose a read-only `GET /v1/tools/catalog` API.
- **FR-002**: The catalog MUST list `runtime.get_current_time`, `workspace.glob`, `workspace.grep`, `workspace.read_file`, `workspace.write_file`, `workspace.edit`, and `workspace.exec_command`.
- **FR-003**: Each tool entry MUST include name, label, group, capability, approval policy, safety class, risk level, side effect type, enabled state, and a short description.
- **FR-004**: The catalog MUST NOT include secrets, provider credentials, raw provider payloads, file content, command output, or executable examples that imply execution.
- **FR-005**: Settings > Tools MUST render the catalog as read-only state, not a mock placeholder.
- **FR-006**: The frontend MUST support real API catalog loading and mock catalog fallback for local mock mode.
- **FR-007**: The feature MUST NOT add permission editing, auto-approval, tool execution controls, MCP, browser automation, or multi-agent behavior.

## Success Criteria *(mandatory)*

- **SC-001**: `GET /v1/tools/catalog` returns all seven current allowlisted tools in deterministic order.
- **SC-002**: Settings > Tools displays workspace read/write/exec risk and approval state without placeholder copy.
- **SC-003**: Automated tests verify catalog mapping and no secret-looking fields.
- **SC-004**: Local validation passes `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run build` from `docs-site/`, and `git diff --check`.

## Assumptions

- M7-M10 tool implementations exist.
- This is a visibility slice only; it does not change runtime execution or approval behavior.
