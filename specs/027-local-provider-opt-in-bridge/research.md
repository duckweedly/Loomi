# Research: Local Provider Opt-in Bridge

## Decision: Store enablement in process memory only

**Rationale**: The slice must not persist tokens and does not yet have a durable local-provider credential reference model. In-process enablement satisfies explicit opt-in while resetting on API restart.

**Alternatives considered**:
- Persist enablement in DB: rejected because it needs a broader credential-reference and revocation model.
- Store frontend-only state: rejected because provider routing must be represented by backend provider capability APIs.

## Decision: Enable action re-runs safe detection; model list does not detect

**Rationale**: Detection reads local auth files and must only happen behind explicit user actions. Enable is explicit and may safely reuse the same redacted detection path. `GET /v1/model-providers` must be read-only over already-enabled state.

**Alternatives considered**:
- Have `GET /v1/model-providers` discover local providers: rejected because page mount and Chat readiness calls could scan local auth implicitly.

## Decision: Local Codex is enabled as unsupported

**Rationale**: M19 proves opt-in routing surface without executing Codex CLI, refreshing OAuth, calling external APIs, or pretending OAuth can be used through the existing OpenAI-compatible HTTP provider.

**Alternatives considered**:
- Real execution in M19: rejected for this slice because it needs a dedicated Local Codex execution bridge with fixture-backed tests.
- Treat enabled Local Codex as available: rejected because Chat would send into a route that cannot execute honestly.
