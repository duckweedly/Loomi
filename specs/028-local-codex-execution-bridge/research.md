# Research: Local Codex Execution Bridge

## Decision: Use auth.json direct bridge, not Codex CLI

**Rationale**: Direct auth material access can be limited to explicit enable and execution boundaries, tested with temporary `CODEX_HOME`, and kept inside the existing runtime `Provider` interface. It avoids interactive CLI hangs, CLI installation assumptions, uncontrolled stdout/stderr, shell execution, and filesystem side effects. The provider can use the existing OpenAI-compatible streaming parser and return normal gateway events.

**Alternatives considered**:
- Call installed `codex` CLI: rejected for M20 because the repo has no proven non-interactive contract that guarantees no prompt hang, no uncontrolled writes, and no prompt/token leakage to stdout/stderr or logs.
- Fixture-only provider: rejected as the primary implementation because it would not satisfy the user-facing executable provider goal, but fixtures remain mandatory for tests.
- Auto-refresh OAuth: rejected because refresh design, storage, and failure semantics are outside this spec.

## Decision: Read credentials only at explicit enable and execution boundaries

**Rationale**: M18.5/M19 already require manual detection/enable before reading local auth. M20 keeps that boundary and does not read auth during provider list, page mount, or Chat send preflight. Execution reads only for an already enabled supported local provider because the user explicitly initiated the run.

**Alternatives considered**:
- Refresh capability on every `GET /v1/model-providers`: rejected because provider list is called by UI readiness paths and would become implicit scanning.
- Persist secret snapshot in DB: rejected because secrets must not be stored.

## Decision: Capability support requires an executable credential snapshot

**Rationale**: Enabled Local Codex must be available/supported only when the auth snapshot contains a bearer token or API key and the bridge has a target endpoint. Missing or malformed auth becomes unavailable and blocks Chat. Provider execution failure records provider failure events through the gateway.

**Alternatives considered**:
- Treat token presence as unsupported until real endpoint is known: rejected because M20 must make Local Codex executable for Chat.
- Mark OAuth available without execution check: rejected because Chat would start a route that may fail before gateway events.

## Decision: Redact provider metadata to routing identifiers only

**Rationale**: Existing gateway metadata already uses provider id, family, and model. M20 keeps Local Codex metadata to those safe values plus model phase. Token, Authorization header, base auth path, and raw auth JSON never enter provider events or assistant metadata.

**Alternatives considered**:
- Include credential source/path for debugging: rejected because local private paths are explicitly disallowed in frontend, events, docs, and logs.
