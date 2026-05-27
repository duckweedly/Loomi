# Feature Specification: M18 Tool Runtime + Tool Catalog Foundation

**Feature Branch**: `025-tool-runtime-catalog-foundation`  
**Created**: 2026-05-25  
**Status**: Complete candidate  
**Input**: Make tools first-class Loomi runtime entities by unifying builtin and local stdio MCP discovery, authorization, execution, event, audit, API, and Settings visibility boundaries.

## User Stories & Tests

### User Story 1 - Inspect the safe tool catalog (Priority: P1)

As a Loomi user or developer, I want Settings > Tools and a read-only API to show the safe tool catalog, so I can understand which tools exist, where they came from, and whether they are executable without seeing secrets, raw args, or raw results.

**Independent Test**: Start the local API and web shell, open Settings > Tools, and verify the builtin current-time tool plus discovered MCP candidates render with name, source, group, risk, approval policy, enabled state, execution state, and schema hash when available.

**Acceptance Scenarios**:

1. **Given** Loomi has builtin runtime tools, **When** the tools catalog is listed, **Then** `runtime.get_current_time` appears as source `builtin`, group `runtime`, risk `low`, approval policy `always_required`, enabled, executable, and safe metadata only.
2. **Given** a local MCP discovery event contains namespaced candidates, **When** the catalog is listed, **Then** `mcp.<server>.<tool>` candidates appear as source `mcp`, group `mcp`, with candidate schema hashes and no raw command/env/args/result data.
3. **Given** Settings > Tools renders the catalog, **When** the user views it, **Then** the page has no install, edit, enable, disable, or policy override write controls.

### User Story 2 - Execute approved tools only through the broker (Priority: P1)

As Loomi runtime, I want a single Tool broker/executor path for builtin and MCP tools, so provider and worker code cannot bypass catalog, persona, discovery, approval, schema hash, scope, event, and redaction checks.

**Independent Test**: Use deterministic provider tests for both `runtime.get_current_time` and a local stdio MCP tool. Verify both requests enter the same broker entrypoint after approval and produce the same event chain.

**Acceptance Scenarios**:

1. **Given** a provider requests a builtin tool, **When** approval is recorded, **Then** the worker executes it through the broker and persists `tool_call_executing` then `tool_call_succeeded` with redacted result summary.
2. **Given** a provider requests a discovered MCP tool, **When** approval is recorded, **Then** the worker executes it through the broker and persists the same lifecycle events with namespaced tool name and candidate schema hash.
3. **Given** a provider or worker attempts an unsupported, disabled, undiscovered, or not persona-allowed tool, **When** execution is attempted, **Then** broker rejects it with a safe failure and no direct executor invocation.

### User Story 3 - Resolve RunContext tools from catalog, persona, and discovery (Priority: P2)

As Loomi runtime, I want RunContext enabled tools to be generated from catalog entries, persona allowed tools, and MCP discovery metadata, so persona allowlists keep working while discovered MCP candidates are represented consistently.

**Independent Test**: Prepare RunContext for personas with builtin-only, MCP-allowed, and non-allowed tool names. Verify enabled tools reflect catalog/discovery/persona intersection and unsafe candidates cannot execute.

**Acceptance Scenarios**:

1. **Given** a persona allows `runtime.get_current_time`, **When** RunContext is prepared, **Then** the builtin appears in enabled tools and tool catalog summary.
2. **Given** a persona allows `mcp.local-smoke.echo` and discovery produced the same candidate with schema hash, **When** RunContext is prepared, **Then** that namespaced tool is enabled with MCP source metadata.
3. **Given** a tool is absent from persona allowlist, disabled, or not discovered, **When** provider requests or worker executes it, **Then** it is rejected before tool execution.

## Requirements

### Functional Requirements

- **FR-001**: Loomi MUST define a unified safe Tool catalog entry containing tool name, display name, description, source, group, input schema hash, risk level, approval policy, enabled state, execution state, and safe metadata.
- **FR-002**: Tool source MUST distinguish `builtin` and `mcp`; tool group MUST support `runtime` and `mcp`, with reserved values for `workspace`, `artifact`, `sandbox`, `web`, and `browser`.
- **FR-003**: Tool risk level MUST support `low`, `medium`, and `high`; approval policy MUST support at least `always_required`, `read_only`, and `disabled`.
- **FR-004**: `runtime.get_current_time` MUST be registered in the catalog as the first builtin runtime tool and MUST execute through the broker.
- **FR-005**: Local stdio MCP candidates MUST remain namespaced as `mcp.<server_slug>.<tool_name>` and MUST be represented in the catalog when discovered.
- **FR-006**: Loomi MUST define `ToolExecutor`, `ToolInvocation`, and `ToolResult` envelope types shared by builtin and MCP execution.
- **FR-007**: Worker/provider runtime code MUST NOT directly execute builtin or MCP tools outside the broker path.
- **FR-008**: RunContext enabled tools MUST be resolved from catalog entries intersected with persona allowed tools and MCP discovery metadata.
- **FR-009**: Persona allowed tools MUST remain effective; tools outside the allowlist MUST NOT execute.
- **FR-010**: Undiscovered, disabled, non-executable, or schema-hash-mismatched tools MUST fail safely before execution.
- **FR-011**: Broker execution MUST validate approval state, tool name, schema hash or candidate hash, thread/run/tool-call scope, enabled state, and execution state before invoking a concrete executor.
- **FR-012**: Broker execution MUST continue to write/reuse the M7/M12 lifecycle events: `tool_call_requested`, `tool_call_approval_required`, `tool_call_approved`, `tool_call_executing`, `tool_call_succeeded`, `tool_call_failed`.
- **FR-013**: Tool result summaries, API responses, run event metadata, continuation context, and Settings UI MUST be redacted and MUST NOT expose secrets, raw args, raw result, process env, command, stderr, or credentials.
- **FR-014**: Loomi MUST expose a minimal read-only Tools API returning catalog safe summaries.
- **FR-015**: Settings > Tools MUST render the safe catalog and MUST NOT provide install/edit/enable/disable write actions in M18.

### Non-Goals

- workspace glob/read/grep
- host shell
- sandbox exec/python
- browser automation
- web search/fetch
- artifact create runtime
- plugin marketplace
- remote MCP or OAuth
- automatic CLI installation
- reading real `~/.claude` or `~/.codex` credentials
- Local Provider autodetect
- multi-agent/spawn_agent
- worker queue rewrite
- multi-tool loop or automatic execution

## Key Entities

- **Tool Catalog Entry**: Safe, user-visible and runtime-usable description of a tool.
- **Tool Broker**: The only runtime entrypoint that validates scope, approval, catalog, persona, discovery, and schema hash before execution.
- **Tool Invocation**: Scoped request envelope for one tool call.
- **Tool Result**: Redacted result/error envelope suitable for events, API, and continuation.
- **Tool Runtime Summary**: RunContext-safe projection of enabled tools and MCP availability.

## Success Criteria

- **SC-001**: Productdata/runtime unit tests cover builtin catalog entry, MCP catalog entry, persona filtering, disabled/non-executable rejection, unified broker entrypoint, and result/event redaction.
- **SC-002**: HTTP/worker smoke tests cover deterministic provider builtin and MCP approval-to-broker execution with replayable lifecycle events.
- **SC-003**: A sensitive canary is absent from Tools API responses, run event metadata, continuation context, and Settings > Tools rendering.
- **SC-004**: Web tests verify Settings > Tools renders safe catalog fields and no write controls.
- **SC-005**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check` pass or report an exact blocker.
