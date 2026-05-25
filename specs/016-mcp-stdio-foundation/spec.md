# Feature Specification: MCP Stdio Foundation

**Feature Branch**: `[016-mcp-stdio-foundation]`

**Created**: 2026-05-25

**Status**: Implemented

**Input**: User description: "M11 MCP stdio foundation minimal slice: define MCP server configuration model; support only local stdio MCP servers; support discover/list tools; map MCP tool schema into Loomi ToolSpec/ToolRegistry as read-only candidates; Persona allowed tools can reference discovered MCP tools but they are disabled by default; RunContext records a safe summary of MCP tool availability; Timeline/debug shows discovery success/failure and safety errors; do not execute MCP tools yet, or only design the approval/execution boundary. Non-goals: MCP HTTP/SSE/OAuth, remote network MCP, marketplace/plugin install, shell/filesystem/browser automation, bypassing M7 approval, automatic MCP tool execution, leaking stderr/env/tokens/secret paths, complex sandbox, or redoing Persona/Skill, RunContext, or Worker queue."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure and discover a local MCP server (Priority: P1)

As a Loomi developer validating M11, I want a local explicitly configured stdio MCP server to be discoverable without executing any tool, so that Loomi can see external tool candidates while keeping the foundation safe.

**Why this priority**: M11 begins with local stdio configuration and tool discovery. Without this, tool mapping, persona references, and RunContext availability have no real input.

**Independent Test**: Provide a local stdio MCP config fixture, run discovery, and verify Loomi records discovered tool names/schemas or a redacted failure without executing any tool.

**Acceptance Scenarios**:

1. **Given** an explicitly enabled local stdio MCP server config, **When** discovery runs, **Then** Loomi starts the server only for discovery/list-tools and records the discovered tool candidates.
2. **Given** the MCP server returns tool schemas, **When** Loomi stores or projects candidates, **Then** each candidate is namespaced by server identity and collision-safe.
3. **Given** discovery fails because the command is missing, times out, exits, or returns invalid schema, **When** the failure is observed, **Then** Loomi records a redacted retryable or non-retryable discovery error without exposing env, args, tokens, raw stderr, or secret-looking paths.

---

### User Story 2 - Expose discovered MCP tools as read-only ToolRegistry candidates (Priority: P2)

As a Loomi maintainer preparing future tool execution, I want discovered MCP tool schemas mapped into Loomi ToolSpec/ToolRegistry as read-only candidates, so that persona and runtime surfaces can reason about availability without enabling execution.

**Why this priority**: Mapping is the core bridge between MCP discovery and Loomi's existing M7/M10 tool and persona boundaries, while still stopping before execution.

**Independent Test**: Feed valid and conflicting MCP tool schemas into the mapper, then verify read-only namespaced ToolSpecs, conflict handling, schema validation, and disabled-by-default execution state.

**Acceptance Scenarios**:

1. **Given** a discovered MCP tool named `search`, **When** Loomi maps it into ToolSpec, **Then** the registry candidate uses a stable namespaced name such as `mcp.<server_slug>.search`.
2. **Given** two servers expose the same tool name, **When** Loomi maps candidates, **Then** both remain distinguishable and no unqualified name overrides an internal runtime tool.
3. **Given** a persona allowed-tools list references a discovered MCP tool, **When** runtime tool resolution runs, **Then** Loomi can show it as available or discovered but non-executable until a future approval/execution implementation enables it.

---

### User Story 3 - Observe MCP availability and safety errors in RunContext and Timeline (Priority: P3)

As a user or developer reading Timeline/debug, I want MCP discovery outcomes and safe availability summaries to be visible, so that I can understand whether MCP tools were discovered, unavailable, disabled, or blocked by safety rules.

**Why this priority**: The constitution requires observable agent execution, and MCP introduces untrusted external data and process failures that must be visible without leaking secrets.

**Independent Test**: Create or replay a run after MCP discovery and verify RunContext safe summary plus Timeline/debug labels show discovered counts, disabled execution boundary, and redacted errors from live SSE and history replay.

**Acceptance Scenarios**:

1. **Given** discovery succeeds, **When** RunContext is prepared, **Then** it records a safe MCP availability summary with server id/slug, tool count, namespaced safe tool names, discovery timestamp/status, and execution disabled state.
2. **Given** discovery fails or is blocked by validation, **When** Timeline/debug renders, **Then** it shows a safe discovery failure or safety error without raw stderr, env, args, tokens, credentials, or secret paths.
3. **Given** a run would otherwise invoke model/runtime, **When** MCP candidates exist, **Then** no MCP tool is automatically executed and any future execution path is documented as approval-gated and audited.

### Edge Cases

- Config references a non-local or network MCP endpoint.
- Config contains env keys, args, or command paths that look sensitive.
- MCP server emits long stderr, secrets, absolute private paths, or invalid JSON.
- Discovery hangs, returns too many tools, returns duplicate names, or returns schema fields Loomi cannot represent.
- A discovered tool name conflicts with `runtime.get_current_time` or another internal Loomi tool.
- A persona references a discovered MCP tool before discovery has run, after discovery failure, or after the server is disabled.
- RunContext preparation happens while discovery cache is stale or unavailable.
- A worker loses ownership while recording MCP availability.
- Timeline/history replay receives older events without MCP metadata.
- Model output or discovery output tries to instruct Loomi to run shell/filesystem/browser automation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST define an explicit local MCP server configuration model with server slug, display name, stdio command reference, arguments, environment reference, enabled flag, timeout, and safe summary fields; this slice may persist or project only discovery safe status, not admin-managed server configuration.
- **FR-002**: MCP configuration MUST come only from local explicit configuration for this slice; Loomi MUST NOT discover, install, download, or enable MCP servers automatically.
- **FR-003**: The feature MUST support only local stdio MCP servers and MUST reject MCP HTTP, SSE, OAuth, remote network endpoints, marketplace/plugin installs, and shell/filesystem/browser automation capabilities.
- **FR-004**: Env values, args, command paths, stderr, tokens, credentials, and secret-looking paths MUST be treated as sensitive and MUST NOT appear in normal Timeline/debug, run events, docs examples, or safe summaries.
- **FR-005**: Loomi MUST support MCP discovery/list-tools for enabled local stdio configs without invoking any discovered tool.
- **FR-006**: Discovery output MUST be treated as external untrusted data, never as instructions for Loomi, and MUST be schema-validated before becoming a candidate.
- **FR-007**: Discovered MCP tools MUST map to Loomi ToolSpec/ToolRegistry as read-only candidates with stable namespaced names and collision handling.
- **FR-008**: MCP candidates MUST be disabled for execution by default and MUST NOT bypass existing M7 approval, audit, tool-call projection, worker ownership, or run-event boundaries.
- **FR-009**: Persona allowed tools MAY reference discovered namespaced MCP tools, but resolution MUST show them as discovered/non-executable until a future execution slice explicitly enables approval-gated execution.
- **FR-010**: RunContext MUST record a safe MCP tool availability summary for the current run, including discovery status, server safe id/slug, candidate count, safe namespaced tool names, disabled execution state, and redacted error codes when applicable.
- **FR-011**: Timeline/debug and Background task surfaces MUST show MCP discovery success, failure, retryability, stale/unavailable state, and safety errors using redacted metadata from both live SSE and history replay.
- **FR-012**: Discovery failures MUST be observable, redacted, and retryable when appropriate, without failing unrelated runs unless the run or persona explicitly requires MCP availability.
- **FR-013**: The plan MUST document the future execution boundary: every MCP tool execution requires user approval, persisted audit events, scoped tool-call state, redacted arguments/results, and no automatic execution.
- **FR-014**: The feature MUST reuse existing M7 approval concepts, M9 RunContext/pipeline, M10 persona allowed-tool resolution, M6 worker/job, run events, SSE/history replay, and existing frontend Timeline/debug grouping without redoing those systems.
- **FR-015**: Documentation MUST plan docs-site updates for architecture, API/event or approval extension, local M11 runbook, roadmap/current-status, devlog, and Spec Kit workflow.

### Key Entities *(include if feature involves data)*

- **MCP Server Config**: Explicit local stdio server definition containing command/args/env references, enabled state, timeout, and safe display metadata.
- **MCP Discovery Session**: One bounded list-tools attempt for a configured server, with redacted status, timestamps, counts, and failure code.
- **MCP Tool Candidate**: Validated discovered tool metadata mapped into a namespaced read-only Loomi ToolSpec.
- **MCP Tool Availability Summary**: RunContext-safe summary of discovered, unavailable, disabled, or failed MCP tool candidates for a run.
- **MCP Execution Boundary**: Future approval-gated contract documenting how execution would enter M7 approval/audit before any tool invocation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of config validation tests accept explicit local stdio configs and reject HTTP/SSE/OAuth/remote or sensitive unsafe configs with redacted errors.
- **SC-002**: 100% of discovery parser tests map valid `list_tools` output into namespaced read-only ToolSpecs and reject invalid or conflicting schemas safely.
- **SC-003**: 100% of redaction tests confirm env values, args, raw stderr, tokens, credentials, and secret-looking paths are absent from persisted events, Timeline/debug metadata, and safe summaries.
- **SC-004**: 100% of persona/tool-resolution tests allow persona references to discovered MCP tools while keeping them non-executable by default.
- **SC-005**: 100% of RunContext tests include MCP availability safe summaries without invoking MCP tools.
- **SC-006**: Timeline/debug tests show MCP discovery success/failure/safety labels from live or replayed events when UI surfaces are touched.
- **SC-007**: The quickstart and tasks include Go tests for config validation, discovery parser, redaction, and tool mapping; web tests for Timeline/debug labels if UI is touched; and `bun run --cwd docs-site build`.

## Assumptions

- M7 tool-call approval and tool-result continuation remain the only execution/audit baseline; M11 does not execute MCP tools.
- M9 RunContext and pipeline can carry safe MCP availability metadata without a new queue or runtime platform.
- M10 Persona allowed tools already accepts tool names and can be extended to reference namespaced MCP candidates as non-executable availability.
- Local explicit MCP config is repository/local-dev configuration in this slice; admin UI, DB-managed server configuration, or marketplace-driven configuration is out of scope.
- Discovery may be run on startup, explicit local command, or worker preparation, but it must be bounded, redacted, and observable.
- Implementation started after user confirmation; docs-site updates and validation evidence are tracked in tasks and quickstart.
