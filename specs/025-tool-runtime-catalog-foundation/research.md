# Research: M18 Tool Runtime + Tool Catalog Foundation

## Decision: Computed catalog for M18

**Decision**: Build the M18 catalog from builtin definitions plus safe MCP discovery/run metadata instead of adding a persistent `tools` table.

**Rationale**: Current tool surface is tiny and already represented by builtin code plus MCP discovery events. A computed catalog keeps the slice small and avoids premature settings/policy management.

**Alternatives rejected**:

- Persistent tool registry table: useful later for user overrides, too large for M18.
- Settings-write catalog management: explicitly out of scope.

## Decision: Single broker entrypoint

**Decision**: Worker approved-tool resume calls a `ToolBroker.Execute` envelope for builtin and MCP tools.

**Rationale**: M7 builtin execution and M12 MCP execution previously had separate branches. Broker centralizes approval, scope, schema hash, allowlist, enabled/executable, executor dispatch, and redaction checks.

**Alternatives rejected**:

- Keep worker branches and add more checks: preserves bypass risk.
- Move approval decisions into broker: M7 approve/deny API already owns projection state and job resume.

## Decision: Reuse M7/M12 event model

**Decision**: Preserve existing tool lifecycle events and projection names.

**Rationale**: Timeline, RunRail, replay API, continuation, and browser smoke already depend on these events. M18 is a boundary unification slice, not an event rewrite.

## Decision: MCP remains local stdio only

**Decision**: M18 only catalogs and brokers already configured local stdio MCP candidates.

**Rationale**: Remote MCP/OAuth/server management would require new auth, network, and settings safety decisions.

## Decision: Settings > Tools is read-only

**Decision**: Add a safe summary panel only.

**Rationale**: Enable/disable/policy overrides need persistence, UX confirmation, and audit. M18 only proves visibility and runtime boundary.
