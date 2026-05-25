# Research: Local Provider Autodetect Foundation

## Decision: implement read-only shape detection rather than credential reuse

**Rationale**: The user goal is to understand and surface whether local Claude Code/Codex providers may be usable. Reusing tokens for model calls would require deeper consent, auth ownership, refresh behavior, provider routing, and persistence decisions. Detection-only delivers a safe vertical slice.

**Alternatives considered**:

- Use detected tokens for model-gateway calls: rejected for this slice because it would silently broaden auth use.
- Save detected token references: rejected because the first slice must not persist tokens or hidden auth handles.

## Decision: use explicit detector input with fixture roots/env map

**Rationale**: Tests must never depend on real user HOME. A detector input object lets tests provide temp HOME, CODEX_HOME, CLAUDE_CONFIG_DIR, and env values while production can pass process environment.

**Alternatives considered**:

- Direct calls to `os.UserHomeDir()` inside every detector: rejected because it makes fixture isolation harder to prove.
- Browser-side detection: rejected because browser cannot safely read local config and would invite unsafe workarounds.

## Decision: expose a dedicated read-only endpoint

**Rationale**: Existing `/v1/model-providers` means configured providers that can be used by model gateway. Local autodetect results are capability candidates and must not be treated as configured/default providers. A separate endpoint avoids semantic drift.

**Alternatives considered**:

- Extend `/v1/model-providers`: rejected because it risks UI or runtime treating detected local providers as configured providers.
- Hide detection behind Settings mock state: rejected because it would not provide backend evidence or safe redaction tests.

## Decision: safe status and candidate model labels only

**Rationale**: Returning only provider id, display name, provider kind, auth mode, status, model candidates, source, and redaction flag lets Settings explain capability without leaking paths or auth payloads.

**Alternatives considered**:

- Return config file path or raw environment key names with values: rejected due to private path and secret leakage risk.
- Return detailed OAuth account metadata: rejected because it may expose user identity and provider internals.

## Decision: helper/keychain/refresh paths are unsupported or unchecked

**Rationale**: `apiKeyHelper`, keychain reads, OAuth refresh, and third-party network checks are active auth operations. They are outside detection-only scope and require explicit future design.

**Alternatives considered**:

- Execute helpers to determine availability: rejected because it executes user-local commands and may disclose credentials.
- Read OS keychain: rejected because it is a sensitive store requiring a separate permission design.
