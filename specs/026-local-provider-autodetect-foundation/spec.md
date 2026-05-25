# Feature Specification: Local Provider Autodetect Foundation

**Feature Branch**: `026-local-provider-autodetect-foundation`

**Created**: 2026-05-25

**Status**: Complete candidate

**Input**: User description: "M18.5 local provider autodetect foundation for Claude Code and Codex, safe detection only, explicit opt-in before use, Settings/provider surface evidence."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Detect local provider availability safely (Priority: P1)

A local Loomi user can ask the backend whether Claude Code or Codex appear usable from fixture-controlled local configuration without exposing secrets or reading uncontrolled real user directories during tests.

**Why this priority**: Detection is the foundation. Without a safe detector, UI and API surfaces would either guess or risk leaking sensitive data.

**Independent Test**: Run detector tests with temp HOME/CODEX_HOME/CLAUDE_CONFIG_DIR fixtures and verify each result contains only safe capability fields.

**Acceptance Scenarios**:

1. **Given** a temp HOME containing `.claude.json` with a `primaryApiKey`, **When** Loomi detects local providers, **Then** it reports Local Claude Code as `available` with `auth_mode = api_key` and no key content.
2. **Given** a temp Claude config containing `settings.json` env values, **When** Loomi detects local providers, **Then** it reports safe source/model/base-url summary only and no token content.
3. **Given** a temp Codex auth file containing OAuth tokens, **When** Loomi detects local providers, **Then** it reports Local Codex as `available` with `auth_mode = oauth` and no token content.
4. **Given** no supported local provider files or env vars in the fixture, **When** Loomi detects local providers, **Then** both local providers report stable unavailable status without errors.

---

### User Story 2 - Expose a safe read-only API surface (Priority: P2)

The Settings UI and local diagnostics can read local provider autodetect status through a backend API response that contains no secrets, no private paths, and no enablement side effects.

**Why this priority**: The browser cannot read local config safely. A backend endpoint provides one controlled redaction boundary and keeps Settings from inventing status.

**Independent Test**: Call the endpoint with injected fixture paths/env and assert the JSON contains safe providers while excluding secret markers and private absolute paths.

**Acceptance Scenarios**:

1. **Given** fixture local providers are detected, **When** `GET /v1/local-provider-detections` is called, **Then** it returns Local Claude Code and Local Codex safe capability objects.
2. **Given** unsupported or disabled detection inputs, **When** the endpoint is called, **Then** statuses remain `unsupported`, `disabled`, or `unavailable` without unstable error text.
3. **Given** fixture values resembling `sk-`, `Bearer`, `access_token`, `refresh_token`, or private filesystem paths, **When** the endpoint responds, **Then** none of those raw values appear in the response body.

---

### User Story 3 - Show safe Settings provider status (Priority: P3)

A user opening Settings > Providers can see whether Local Claude Code and Local Codex were detected, understand they require explicit opt-in before use, and confirm no token/key/path is shown.

**Why this priority**: The visible surface proves the local capability model without enabling hidden model use or changing the current provider route.

**Independent Test**: Run web tests that render Settings provider UI with detected and unavailable local providers and verify copy, status labels, and lack of secret/path content.

**Acceptance Scenarios**:

1. **Given** local provider detection data is available, **When** the user opens Settings > Providers, **Then** the UI shows Local Claude Code and Local Codex with detected/not detected status.
2. **Given** a provider is detected, **When** the user reads the provider card, **Then** it says explicit opt-in is required before use and no secrets are shown.
3. **Given** detected local providers exist, **When** the current provider route is displayed, **Then** Loomi does not automatically switch or enable those providers.
4. **Given** the backend detection endpoint is unavailable, **When** Settings renders Providers, **Then** the UI shows a clear unavailable/error state rather than fake detection.

### Edge Cases

- Claude config includes `apiKeyHelper`: report unsupported/unavailable for helper-based auth and do not execute the helper.
- `CODEX_API_KEY` is present while Codex auth file also exists: report env API key mode as the selected safe source.
- A file exists but contains invalid JSON: return unavailable with a stable safe message.
- OAuth structures exist without usable tokens: return `needs_login` rather than `available`.
- Detector fixtures point to temp directories only in tests; tests must fail if real HOME is required.
- Detection must not install, execute, or shell out to Claude Code, Codex, rtk, or any CLI.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a local provider detector for Claude Code and Codex that accepts explicit environment/path inputs for HOME, CODEX_HOME, and CLAUDE_CONFIG_DIR.
- **FR-002**: System MUST return only safe capability fields: `provider_id`, `display_name`, `provider_kind`, `auth_mode`, `status`, `model_candidates`, `source`, and `redaction_applied`.
- **FR-003**: System MUST support provider statuses `available`, `unavailable`, `needs_login`, `unsupported`, and `disabled`.
- **FR-004**: System MUST support auth modes `api_key`, `oauth`, and `unknown`.
- **FR-005**: System MUST support detection sources `local_config`, `env`, `keychain_unchecked`, and `unknown`.
- **FR-006**: Claude Code detection MUST recognize fixture `.claude.json` `primaryApiKey`, `.claude/settings.json` env entries for `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, and `ANTHROPIC_MODEL`, and `.claude/.credentials.json` OAuth-like structures.
- **FR-007**: Claude Code detection MUST NOT output token/key contents, execute `apiKeyHelper`, call third-party network refresh endpoints, or read keychain data.
- **FR-008**: Codex detection MUST recognize fixture `.codex/auth.json` or `CODEX_HOME/auth.json`, `OPENAI_API_KEY`, `CODEX_API_KEY`, `auth_mode`, and `tokens.access_token` presence.
- **FR-009**: Codex detection MUST prefer `CODEX_API_KEY` env over auth-file credentials.
- **FR-010**: Codex detection MUST NOT output token/key contents, refresh OAuth tokens, call OpenAI/ChatGPT endpoints, or modify auth files.
- **FR-011**: System MUST expose local provider autodetect through a read-only local API endpoint or equivalent provider capability endpoint extension.
- **FR-012**: API responses MUST NOT contain raw `sk-` values, `Bearer` tokens, `access_token`, `refresh_token`, private absolute paths, Authorization headers, or provider auth payloads.
- **FR-013**: Settings > Providers MUST show Local Claude Code and Local Codex detection status with copy for detected/not detected, explicit opt-in before use, and no secrets shown.
- **FR-014**: System MUST NOT automatically make detected local providers the default provider or start a real model call from detection.
- **FR-015**: First version MAY show "detected but not enabled" only; any enablement MUST be explicit, local/session-scoped or saved as a setting, and MUST NOT save tokens.
- **FR-016**: Documentation MUST describe architecture, API contract, runbook, devlog, roadmap status, and Spec Kit workflow status for this slice.

### Key Entities *(include if feature involves data)*

- **LocalProviderDetectionInput**: Explicit detector input, including fixture-safe environment variables and local config roots.
- **LocalProviderCapability**: Redacted capability result for one local provider.
- **LocalProviderStatus**: Stable status value for Settings/API presentation.
- **LocalProviderDetectionResponse**: Endpoint response containing only safe provider capabilities and request metadata.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Detector tests cover all requested Claude Code, Codex, env precedence, missing-file, and temp-HOME scenarios without reading real user auth paths.
- **SC-002**: HTTP API tests prove the detection response includes expected providers and excludes secret/path markers.
- **SC-003**: Web tests prove Settings > Providers displays Local Claude Code and Local Codex status, explicit opt-in copy, no token/key/path, and no automatic provider switch.
- **SC-004**: Existing model provider save/check behavior remains available and no detected local provider is used for model-gateway runs by default.
- **SC-005**: Documentation site builds after architecture/API/runbook/devlog/roadmap/spec-kit updates.

## Assumptions

- This slice runs inside the local Loomi API process and reuses existing Settings > Providers patterns.
- Local provider detection is a capability discovery surface, not an authentication broker.
- Model candidates are safe labels derived from explicit config fields or conservative defaults; no external provider catalog lookup is performed.
- Tests may inject fixture HOME/CODEX_HOME/CLAUDE_CONFIG_DIR and environment maps; production runtime may use process environment and user home only for safe existence/content-shape checks.
- No database schema change is required because detection state is read-only and computed on request.
