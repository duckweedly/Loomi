# Data Model: MCP Stdio Foundation

## MCP Server Config

Explicit local stdio server definition from repository/local-dev configuration.

Fields:

- `id`
- `slug`
- `displayName`
- `enabled`
- `transport`: `stdio` only in this slice
- `commandRef`: local command reference or configured executable label
- `argsRef`: sensitive argument list reference
- `envRef`: sensitive environment reference
- `timeoutMs`
- `createdAt`, `updatedAt`

Validation rules:

- `slug` is stable and unique.
- `transport` must be `stdio`.
- HTTP, SSE, OAuth, remote URLs, marketplace identifiers, and auto-install instructions are rejected.
- Env values, raw args, raw command paths, tokens, credentials, and secret-looking paths are sensitive and excluded from safe summaries.
- Disabled servers are not discovered and appear as disabled in availability summaries.
- Server configuration itself is not admin-managed or DB-managed in this slice; only safe discovery status may be persisted or projected.

## MCP Discovery Session

One bounded attempt to list tools from a configured local stdio server.

Fields:

- `serverID`
- `serverSlug`
- `status`: pending, succeeded, failed, disabled, stale, rejected
- `startedAt`, `completedAt`
- `toolCount`
- `candidateNames`
- `errorCode`
- `redactedMessage`
- `retryable`

Lifecycle:

```text
configured
-> validation passed
-> stdio process opened for discovery
-> list-tools parsed
-> safe candidates recorded
-> process closed/cleaned up
```

Failure lifecycle:

```text
configured
-> validation failed or process/discovery failed
-> redacted failure recorded
-> process closed/cleaned up when applicable
```

## MCP Tool Candidate

Validated discovered MCP tool metadata before execution support.

Fields:

- `serverID`
- `serverSlug`
- `mcpToolName`
- `toolSpecName`: `mcp.<server_slug>.<tool_name>`
- `descriptionSummary`
- `inputSchemaSummary`
- `inputSchemaHash`
- `executionEnabled`: false for this slice
- `approvalRequired`: true for future execution
- `discoveredAt`

Validation rules:

- Tool names are normalized only when safe and stable.
- Namespaced names must not conflict with internal runtime tools.
- Invalid, oversized, or unsupported schemas are rejected or represented as redacted unavailable candidates.
- Description and schema text are untrusted data and must not be treated as instructions.

## ToolSpec Mapping

Projection from MCP candidate to Loomi ToolSpec/ToolRegistry.

Rules:

- `name` uses the namespaced candidate name.
- `source` is `mcp`.
- `serverSlug` identifies the origin.
- `executionPolicy` is disabled/not-executable in this slice.
- `approvalPolicy` is future `always_required`.
- Raw MCP schema is validated and summarized; raw dangerous metadata is not exposed in normal Timeline/debug.

## MCP Tool Availability Summary

RunContext-safe summary of MCP availability.

Fields:

- `serversConfigured`
- `serversEnabled`
- `serversSucceeded`
- `serversFailed`
- `candidateCount`
- `candidateNames`
- `nonExecutableCandidateNames`
- `discoveryStatusByServer`
- `redactedErrorCodes`
- `lastDiscoveredAt`

Forbidden fields:

- env values
- raw args
- raw command paths
- tokens, credentials, Authorization headers
- raw stderr/stdout
- private absolute paths or secret-looking file names
- raw MCP result payloads
- shell output, file contents, browser or desktop captured state

## MCP Safety Error

Redacted error returned by config validation, discovery, mapping, or RunContext availability.

Fields:

- `code`
- `safeMessage`
- `retryable`
- `serverSlug`
- `stage`: config_validation, discovery, tool_mapping, run_context

Rules:

- Failure must be observable without leaking sensitive fields.
- Discovery failures are retryable when caused by timeout/process unavailable and non-retryable when caused by unsupported transport or invalid config.

## MCP Execution Boundary

Future contract for executing discovered MCP tools.

Required properties before any execution:

- user approval through M7-style approval flow
- persisted tool-call projection and audit events
- scoped run/thread/user ownership checks
- redacted arguments before persistence
- redacted result/error summaries before persistence
- worker ownership and cancellation guards
- no automatic execution from persona, model output, or discovery metadata

This entity is design-only in M11 foundation and has no real executor.
