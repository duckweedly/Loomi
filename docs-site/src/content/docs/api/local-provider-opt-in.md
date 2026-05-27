---
title: M19 Local Provider Opt-in API
description: Explicit session-local enable and disable endpoints for local provider detections.
---

## `POST /v1/local-provider-detections/{provider_id}/enable`

Enables a local provider for the current API process session after explicit user action.

M19 supports `local_codex` only. Claude Code returns unsupported.

```json
{
  "provider": {
    "id": "local_codex",
    "family": "openai_compatible",
    "model": "gpt-5",
    "status": "unavailable",
    "message": "Local Codex is enabled for this session, but execution is unsupported until the local provider execution bridge is implemented.",
    "local_provider": true,
    "session_local": true,
    "credential_reference": "redacted",
    "execution_state": "unsupported"
  },
  "request_id": "req_..."
}
```

## `DELETE /v1/local-provider-detections/{provider_id}/enable`

Disables the session-local enablement and removes the local provider from `GET /v1/model-providers`.

## `GET /v1/model-providers`

Returns configured OpenAI-compatible providers plus explicitly enabled local provider route candidates.

This endpoint must not run local provider detection, read local auth files, refresh tokens, call CLIs, or call external provider APIs.

Configured providers use `status: "configured"` when local configuration is complete but no live completion smoke has been run. A provider should only be treated as completion-verified after `POST /v1/model-providers/check` returns `status: "completion-ok"`.

## `POST /v1/model-providers/check`

Runs a minimal non-streaming completion smoke for the selected configured provider and returns the same safe provider shape with diagnostic fields:

```json
{
  "provider": {
    "id": "custom",
    "family": "openai_compatible",
    "model": "gpt-5.5",
    "status": "completion-failed",
    "check_stage": "completion",
    "check_code": "completion-failed-503",
    "http_status": 503,
    "message": "Provider completion check failed with HTTP 503."
  },
  "request_id": "req_..."
}
```

The check path must not return API keys, Authorization headers, raw provider bodies, prompts, provider trace payloads, or local auth paths. HTTP provider bodies are discarded after bounded reading; only the normalized category, HTTP status, and safe message are exposed.

## Redaction contract

Local provider opt-in responses must not include API keys, bearer tokens, `access_token`, `refresh_token`, Authorization headers, private filesystem paths, raw auth JSON, or CLI paths.
