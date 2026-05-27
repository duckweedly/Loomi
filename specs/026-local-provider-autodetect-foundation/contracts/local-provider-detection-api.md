# Contract: Local Provider Detection API

## `GET /v1/local-provider-detections`

Returns safe, read-only local provider candidate detection.

Response:

```json
{
  "providers": [
    {
      "provider_id": "local_claude_code",
      "display_name": "Local Claude Code",
      "provider_kind": "claude_code",
      "auth_mode": "api_key",
      "status": "available",
      "model_candidates": ["claude-sonnet-4-5"],
      "source": "local_config",
      "redaction_applied": true,
      "message": "Detected but not enabled. Explicit opt-in is required before use."
    },
    {
      "provider_id": "local_codex",
      "display_name": "Local Codex",
      "provider_kind": "codex",
      "auth_mode": "unknown",
      "status": "unavailable",
      "model_candidates": ["gpt-5"],
      "source": "unknown",
      "redaction_applied": true,
      "message": "Not detected."
    }
  ],
  "request_id": "req_..."
}
```

Rules:

- Endpoint is read-only.
- Endpoint does not save, enable, or select a provider.
- Endpoint does not execute CLI helpers, OAuth refresh, keychain reads, or network calls.
- Response must not include `sk-`, `Bearer`, `access_token`, `refresh_token`, raw auth values, Authorization headers, or private absolute paths.

Status values:

- `available`
- `unavailable`
- `needs_login`
- `unsupported`
- `disabled`

Auth mode values:

- `api_key`
- `oauth`
- `unknown`

Source values:

- `local_config`
- `env`
- `keychain_unchecked`
- `unknown`
