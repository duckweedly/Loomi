# Research: M21 Workspace Read Tools

## Decision: Reuse M18 ToolCatalog, ToolBroker, RunContext, and approval events

**Rationale**: These are already the Loomi boundaries for tool discoverability, per-run allowlist resolution, approval, worker execution, and provider continuation. Workspace tools should not introduce a second execution lane.

**Alternatives considered**:
- Direct API endpoint execution: rejected because it bypasses provider-requested approval semantics.
- Separate filesystem service with its own event model: rejected because it would fragment timeline/debug visibility.

## Decision: Implement workspace filesystem tools with Go stdlib

**Rationale**: The milestone requires no production shell/rg usage. Go `filepath.WalkDir`, `os.Open`, `bufio.Scanner`, bounded readers, and `regexp` are enough for a small read-only tool slice.

**Alternatives considered**:
- Calling `rg`/shell: rejected by milestone scope and harder to bound safely.
- Importing a large search library: rejected because current needs are small and stdlib is sufficient.

## Decision: Resolve a single workspace root from `LOOMI_WORKSPACE_ROOT` or repo root

**Rationale**: A single root keeps approval prompts and safety checks understandable. Tests can override it with fixture roots.

**Alternatives considered**:
- Multiple roots per run: rejected for M21 because it complicates allowlists and UI scope display.
- User-supplied absolute roots per tool call: rejected because it creates path escape risk.

## Decision: Deny sensitive paths before content access

**Rationale**: Secret-bearing files should not be opened or summarized. Denylist checks apply to each path component and relative path before read/search.

**Alternatives considered**:
- Redact after reading: rejected because content access itself violates the boundary.
- Only deny `.env`: rejected because private keys, credentials, and VCS internals are common local risks.

## Decision: Keep Arkloop learning limited to mechanisms

**Rationale**: Arkloop confirms useful mechanism categories: registry metadata, tool allowlists, static filesystem tools, bounded/truncated results, and approval/run event visibility. Loomi should not copy expression, private interfaces, or large sandbox architecture.

**Alternatives considered**:
- Porting tool code or sandbox layers: rejected by constitution and milestone non-goals.
