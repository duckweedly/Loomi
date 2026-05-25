# Research: MCP Stdio Foundation

## Decision: Support only explicit local stdio MCP configs

**Rationale**: M11's first useful slice is local MCP stdio configuration and discovery. Explicit config keeps user intent clear and avoids marketplace, plugin install, remote network, OAuth, and hidden auto-discovery risk.

**Alternatives considered**:

- HTTP/SSE/OAuth MCP support: rejected as later M11+ scope and explicitly out of bounds.
- Marketplace or plugin install: rejected because it introduces external writes and trust policy before discovery is proven.
- Auto-scan local tools: rejected because MCP config must be explicit and auditable.

## Decision: Treat config process details and stderr as sensitive

**Rationale**: Commands, args, env, stderr, tokens, credentials, and absolute paths can reveal local secrets or private filesystem layout. Normal Timeline/debug only needs safe server identity, status, counts, and redacted error codes.

**Alternatives considered**:

- Persist raw stderr for debugging: rejected due to secret leakage risk.
- Redact only in frontend: rejected because persisted events and history replay must already be safe.

## Decision: Run discovery/list-tools without tool execution

**Rationale**: Discovery proves MCP connectivity and schema ingestion while avoiding the main safety risk: arbitrary external tool execution. Execution needs separate approval, audit, argument redaction, result redaction, and possibly sandbox decisions.

**Alternatives considered**:

- Execute read-only-looking tools immediately: rejected because MCP server claims are untrusted and M7 approval cannot be bypassed.
- Skip discovery and model config only: rejected because tool mapping and availability require real list-tools output.

## Decision: Map MCP tools to namespaced read-only ToolSpec candidates

**Rationale**: Namespacing prevents collisions with internal tools such as `runtime.get_current_time` and keeps multiple servers distinguishable. Read-only candidates let persona and RunContext reason about availability without enabling execution.

**Alternatives considered**:

- Use raw MCP tool names directly: rejected because collisions and spoofing would be ambiguous.
- Register MCP tools as executable ToolSpecs: rejected until execution boundary is implemented.

## Decision: Persona allowed tools may reference MCP candidates but execution stays disabled

**Rationale**: M10 already gives personas an allowed-tool list. Allowing references to namespaced MCP candidates lets Loomi test policy wiring, while disabled execution preserves the M7 approval boundary.

**Alternatives considered**:

- Reject all MCP names in personas until execution exists: rejected because the target requires persona references.
- Treat persona allow as approval: rejected because persona config cannot replace user approval and audit.

## Decision: Attach safe MCP availability to RunContext and Timeline/debug

**Rationale**: Observable execution requires users/developers to see whether MCP discovery succeeded, failed, is stale, or is disabled. RunContext is the right narrow place to summarize current availability without passing raw discovery output.

**Alternatives considered**:

- Keep MCP discovery only in logs: rejected because logs are not replayable in Loomi's Timeline/debug.
- Add a separate MCP dashboard first: rejected as broader admin-console/platform scope.

## Decision: Define future execution as approval-gated and audited

**Rationale**: MCP tools may have side effects regardless of schema labels. Future execution must enter the existing tool-call projection, approval, worker ownership, run events, redacted args/results, and audit semantics.

**Alternatives considered**:

- Execute MCP tools through discovery session directly: rejected because it bypasses M7.
- Design a separate MCP permission system now: rejected as complex sandbox/permission framework scope.
