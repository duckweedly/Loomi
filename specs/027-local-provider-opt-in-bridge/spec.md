# Feature Specification: Local Provider Opt-in Bridge

**Feature Branch**: `027-local-provider-opt-in-bridge`
**Created**: 2026-05-25
**Status**: Complete candidate
**Input**: User description: "M19 Local Provider Opt-in Bridge: turn M18.5 detection-only Local Codex into explicit session-local provider route candidate, without automatic enablement or secret leakage."

## User Scenarios & Testing

### User Story 1 - Explicitly enable detected Local Codex (Priority: P1)

A user runs Settings > Providers local autodetect, sees Local Codex detected, and explicitly enables it for the current session.

**Independent Test**: Backend enable/disable tests prove detected-but-not-enabled local providers stay out of `GET /v1/model-providers`, while enabled Local Codex appears as a safe session-local local provider capability.

**Acceptance Scenarios**:

1. **Given** Local Codex is detected as available, **When** the user chooses enable, **Then** Loomi stores only a session-local redacted enablement and returns Local Codex in configured provider candidates.
2. **Given** Local Codex has not been explicitly enabled, **When** `GET /v1/model-providers` is called, **Then** Local Codex is not listed.
3. **Given** Local Codex is disabled after enablement, **When** `GET /v1/model-providers` is called, **Then** Local Codex is removed from configured candidates.

### User Story 2 - Preserve unsupported execution honesty (Priority: P1)

A user can see that Local Codex is enabled as a route candidate, but real chat execution is still blocked until the Local Codex execution bridge exists.

**Independent Test**: Web and backend tests prove enabled Local Codex is marked `local_provider`, `session/local only`, `credential_reference=redacted`, and `execution_state=unsupported`, and Chat submit remains blocked.

**Acceptance Scenarios**:

1. **Given** enabled Local Codex is returned by `GET /v1/model-providers`, **When** Settings renders configured providers, **Then** it labels the provider as local, session-local, redacted, and execution unsupported.
2. **Given** only enabled-but-unsupported Local Codex exists, **When** Chat Composer renders, **Then** it does not claim a ready provider and prevents sending.
3. **Given** enabled Local Codex is unsupported, **When** a run is started against that provider, **Then** backend rejects it rather than faking execution.

### User Story 3 - Keep existing provider save/check behavior (Priority: P2)

A user who configures an OpenAI-compatible provider can still save and test that provider without regression from local provider enablement.

**Independent Test**: Existing OpenAI-compatible save/check tests still pass and local provider response fields never include token, API key, refresh token, or private path values.

## Requirements

- **FR-001**: System MUST keep Local Provider Autodetect behind the existing manual Settings > Providers detect action.
- **FR-002**: System MUST NOT automatically enable detected local providers during startup, page mount, Chat send, or model provider list reads.
- **FR-003**: System MUST provide explicit enable and disable endpoints for local provider detections.
- **FR-004**: System MUST support Local Codex opt-in when detection status is `available`.
- **FR-005**: System MUST reject or keep non-ready local providers unsupported when detection status is `unavailable`, `needs_login`, `unsupported`, or `disabled`.
- **FR-006**: Enabled local provider state MUST be process-local/session-local unless a future spec introduces safe persistence.
- **FR-007**: `GET /v1/model-providers` MUST include enabled local providers as safe capabilities and MUST NOT perform local auth detection itself.
- **FR-008**: Local provider capabilities MUST include safe markers for local provider, session-local scope, redacted credential reference, and execution support state.
- **FR-009**: System MUST NOT save or expose access tokens, refresh tokens, API keys, private paths, CLI paths, or raw auth file content in DB, run events, docs, frontend state, logs, or API responses.
- **FR-010**: System MUST NOT execute Codex or Claude CLI, install CLIs, read keychain, refresh OAuth, or call external APIs for login validation.
- **FR-011**: Chat MUST remain blocked for enabled-but-unsupported local providers and clearly communicate that execution is unsupported.
- **FR-012**: Existing OpenAI-compatible provider save/check behavior MUST continue to work.

## Key Entities

- **LocalProviderEnablement**: Session-local redacted opt-in record derived from an explicit detection/enable action.
- **ProviderCapability**: Safe provider route candidate returned by `GET /v1/model-providers`.
- **LocalProviderExecutionState**: Whether a local provider route can execute chat (`unsupported` in this slice).

## Success Criteria

- **SC-001**: Backend tests prove Local Codex detected-only state does not appear in model provider list until enabled.
- **SC-002**: Backend/API tests prove enabled Local Codex capability contains no token/key/path canaries.
- **SC-003**: Web tests prove Settings exposes detect, enable, disable, and unsupported states without secret/path leakage.
- **SC-004**: Chat tests prove enabled-but-unsupported Local Codex still blocks send.
- **SC-005**: Existing OpenAI-compatible save/check tests pass unchanged or with compatible expectations.

## Assumptions

- Real Local Codex execution is deferred because safely using local OAuth material requires a dedicated execution bridge not present in M19.
- Claude Code remains detection/contract-only in this slice unless a later safe execution contract is designed.
