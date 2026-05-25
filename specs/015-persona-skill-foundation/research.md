# Research: Persona Skill Foundation

## Decision: Use durable persona records plus versioned persona snapshots

**Rationale**: M10 requires persona DB, file-to-DB sync, versioning, and old-run attribution. A separate version/snapshot concept lets built-in config update for new runs while preserving what older runs used.

**Alternatives considered**:

- Single mutable persona row only: rejected because old runs could not identify the exact prompt/model/tools version used.
- Store persona only in config files: rejected because RunContext and history/debug need durable references.

## Decision: Sync built-ins idempotently from repository-local config

**Rationale**: Built-in personas are enough for the foundation slice and avoid marketplace/admin complexity. Idempotent sync gives deterministic local setup and tests.

**Alternatives considered**:

- Admin UI editing first: rejected because it expands scope beyond the thin slice.
- Plugin/marketplace source of personas: rejected by explicit non-goal.

## Decision: Resolve persona by run override, then thread selection, then default built-in

**Rationale**: This supports both explicit run choice and natural thread inheritance while keeping a deterministic fallback for smoke tests.

**Alternatives considered**:

- Thread-only selection: rejected because the target requires thread/run can select or inherit persona.
- Run-only selection: rejected because it would not support durable thread inheritance.

## Decision: Store prompt in runtime snapshot but expose only safe summary

**Rationale**: The runtime needs system prompt to affect model behavior, but normal Timeline/debug must not expose persona prompt text. Safe summary is enough for observability.

**Alternatives considered**:

- Put prompt in run events for debugging: rejected because it violates the stated safety boundary.
- Omit persona from Timeline/debug entirely: rejected because observable agent execution requires explaining which persona influenced a run.

## Decision: Persona allowed tools narrow the existing runtime allowlist

**Rationale**: Persona should control behavior without introducing new executable tool families or a permission framework. Unknown tools should fail sync or be ignored with a safe validation error before runtime.

**Alternatives considered**:

- Persona can define new tools: rejected because Skill marketplace/plugin/MCP are non-goals.
- Ignore allowed tools until later: rejected because M10 acceptance requires different tools by persona.

## Decision: Persona model route selects within existing provider/model routing

**Rationale**: The feature can prove different persona routing without building provider administration. Existing provider availability checks and redaction rules remain authoritative.

**Alternatives considered**:

- New route registry or admin console: rejected as P3/P4 platform scope.
- Hardcode model route in runtime only: rejected because durable persona versions must record route information.

## Decision: Minimal frontend may be selector or read-only display, but smoke must prove real run path

**Rationale**: The user explicitly allows minimal selector or read-only display, with run-chain verification taking priority. If adding a selector is cheap in existing composer/thread flow, prefer it; otherwise display resolved default persona and prove default run resolution.

**Alternatives considered**:

- Full persona management UI: rejected as outside foundation.
- No frontend surface: rejected because browser smoke requires visible Timeline/debug summary.
