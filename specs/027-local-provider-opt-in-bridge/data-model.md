# Data Model: Local Provider Opt-in Bridge

## LocalProviderEnablement

- `provider_id`: stable local provider id, initially `local_codex`.
- `display_name`: safe display name from detection.
- `provider_kind`: `codex` or `claude_code`.
- `auth_mode`: safe mode from detection (`api_key`, `oauth`, or `unknown`).
- `model_candidates`: safe model labels from detection.
- `source`: coarse source (`local_config`, `env`, `unknown`); no paths.
- `enabled`: true while session-local enablement is active.
- `scope`: `session`.
- `credential_reference`: always `redacted` in this slice.
- `execution_state`: `unsupported`.
- `message`: safe user-facing status.

## ProviderCapability Extension

- `local_provider`: optional boolean marker.
- `session_local`: optional boolean marker.
- `credential_reference`: optional safe string, `redacted` for local providers.
- `execution_state`: optional string, `unsupported` for M19 local providers.

## State Transitions

1. `detected_available` -> `enabled_unsupported` after explicit enable.
2. `enabled_unsupported` -> `disabled` after explicit disable.
3. `detected_unavailable`, `needs_login`, `unsupported`, or `disabled` -> enable rejected.

## Redaction Rules

No entity may contain token, refresh token, API key, CLI path, private path, raw auth JSON, or keychain material.
