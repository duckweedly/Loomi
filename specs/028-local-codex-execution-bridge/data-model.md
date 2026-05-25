# Data Model: Local Codex Execution Bridge

## LocalCodexCredentialSnapshot

- `auth_mode`: `api_key` or `oauth`.
- `bearer_token`: secret string held only in memory for request construction.
- `model`: safe model label, default `gpt-5.5`.
- `base_url`: execution endpoint used by the provider; not returned to frontend for local providers.
- `available`: true only when token/key and endpoint are present.

## LocalCodexProvider

- Implements `runtime.Provider`.
- `Config()` returns safe `ProviderConfig` with `ID=local_codex`, `Family=openai_compatible`, `Model`, `Enabled=true`.
- `Stream()` re-reads the credential snapshot for the enabled provider, builds an OpenAI-compatible request, and emits existing `ProviderEvent` values.
- Failure states emit provider errors; they do not produce assistant text.

## LocalProviderEnablement

- `provider_id`: `local_codex`.
- `display_name`: `Local Codex`.
- `provider_kind`: `codex`.
- `auth_mode`: safe detection auth mode.
- `model_candidates`: safe model labels.
- `source`: coarse source only (`local_config`, `env`, `unknown`).
- `execution_state`: `supported` when bridge can execute, `unsupported` for non-executable local providers.
- `status`: `available` when executable, `unavailable` when auth/bridge is unavailable.
- `credential_reference`: `redacted`.

## ProviderCapability Extension

- `local_provider`: true for Local Codex.
- `session_local`: true for process-local enablement.
- `credential_reference`: `redacted`.
- `execution_state`: `supported` or `unsupported`.
- `status`: `available` only when Chat can send.

## State Transitions

1. `detected_available` -> `enabled_supported` after explicit enable and executable snapshot validation.
2. `enabled_supported` -> `disabled` after explicit disable.
3. `enabled_supported` -> `run_failed_provider_error` if execution-time auth or provider request fails.
4. `detected_unavailable`, `needs_login`, `unsupported`, or malformed auth -> `blocked_unavailable`.

## Redaction Rules

No persisted entity, API response, frontend state, run event metadata, assistant message metadata, doc, or log may contain `access_token`, `refresh_token`, API key, Authorization header, auth file path, private home path, or raw auth JSON.
