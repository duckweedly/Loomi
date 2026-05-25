---
title: M18.5 Local Provider Autodetect API
description: Read-only safe capability endpoint for local Claude Code and Codex detection.
---

## `GET /v1/local-provider-detections`

Returns safe local provider candidates.

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

## Field enums

`auth_mode`:

- `api_key`
- `oauth`
- `unknown`

`status`:

- `available`
- `unavailable`
- `needs_login`
- `unsupported`
- `disabled`

`source`:

- `local_config`
- `env`
- `keychain_unchecked`
- `unknown`

## Redaction contract

Responses must not include:

- raw API keys
- bearer tokens
- `access_token`
- `refresh_token`
- Authorization headers
- private absolute filesystem paths
- raw provider request or auth payloads

The endpoint is read-only. It does not save settings, enable providers, switch defaults, refresh OAuth credentials, execute helpers, read keychain data, call third-party endpoints, or install local CLIs.
