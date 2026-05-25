# Data Model: Local Provider Autodetect Foundation

## LocalProviderDetectionInput

Fields:

- `home`: optional root used to find `.claude.json` and `.codex/auth.json`
- `codex_home`: optional root used to find `auth.json`
- `claude_config_dir`: optional root used to find `settings.json` and `.credentials.json`
- `env`: key/value map supplied by tests or process environment
- `detection_enabled`: boolean, default true for local endpoint

Validation:

- Empty roots are allowed and result in unavailable providers.
- Tests must pass temp roots and must not require real HOME.
- Sensitive env values must be treated as presence-only.

## LocalProviderCapability

Fields:

- `provider_id`: stable id, `local_claude_code` or `local_codex`
- `display_name`: user-facing safe name
- `provider_kind`: `claude_code` or `codex`
- `auth_mode`: `api_key`, `oauth`, or `unknown`
- `status`: `available`, `unavailable`, `needs_login`, `unsupported`, or `disabled`
- `model_candidates`: safe list of model labels
- `source`: `local_config`, `env`, `keychain_unchecked`, or `unknown`
- `redaction_applied`: true for every detector result
- `message`: optional stable safe explanation

Validation:

- Must not contain raw token/key, Authorization header, OAuth refresh/access token field values, or private absolute path.
- Must not include file paths; source is categorical only.
- Model candidates are labels only.

## LocalProviderDetectionResponse

Fields:

- `providers`: array of `LocalProviderCapability`
- `request_id`: existing request metadata pattern

Relationships:

- Response is read-only and independent of configured `ProviderCapability`.
- Response does not alter gateway provider list, current provider route, or Settings provider draft fields.

## State Transitions

Claude/Codex detection status:

- missing files/env -> `unavailable`
- API key present -> `available` with `api_key`
- OAuth token structure with usable token presence -> `available` with `oauth`
- OAuth account/config exists but no usable token presence -> `needs_login`
- helper-only auth -> `unsupported`
- endpoint disabled in future config -> `disabled`
