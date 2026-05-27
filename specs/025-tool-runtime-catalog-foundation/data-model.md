# Data Model: M18 Tool Runtime + Tool Catalog Foundation

## Tool Catalog Entry

- `name`: stable runtime name, e.g. `runtime.get_current_time` or `mcp.local-smoke.echo`
- `displayName`: safe human label
- `description`: safe summary
- `source`: `builtin` or `mcp`
- `group`: `runtime`, `mcp`, reserved `workspace`, `artifact`, `sandbox`, `web`, `browser`
- `inputSchemaHash`: optional `sha256:*`
- `riskLevel`: `low`, `medium`, `high`
- `approvalPolicy`: `always_required`, `read_only`, `disabled`
- `enabled`: boolean
- `executionState`: `executable`, `disabled`, `not_discovered`, `not_allowed`, `non_executable`
- `safeMetadata`: redacted map for server slug, discovery status, candidate hash, and other safe details

## Tool Invocation

- `threadID`
- `runID`
- `toolCallID`
- `toolName`
- `candidateSchemaHash`
- `argumentsSummary`
- `argumentsHash`
- `approvalStatus`
- `executionStatus`
- `catalogEntry`
- `scopeSource`: worker approved tool resume

## Tool Result

- `toolName`
- `toolCallID`
- `status`: `succeeded` or `failed`
- `resultSummary`
- `resultForModelRedacted`
- `errorCode`
- `errorMessage`

## RunContext Tool Runtime Summary

- `enabledTools`: catalog-backed tool resolutions allowed for this run
- `mcpAvailability`: existing MCP safe summary
- `toolCatalog`: safe catalog entries relevant to the run

## State Rules

- Builtin current-time is enabled and executable when persona allows it.
- MCP candidate is visible when discovered, but executable only when discovered, enabled by MCP execution config, persona-allowed, and schema hash matches.
- Broker rejects every tool outside catalog/persona/discovery/scope/approval checks before calling a concrete executor.
