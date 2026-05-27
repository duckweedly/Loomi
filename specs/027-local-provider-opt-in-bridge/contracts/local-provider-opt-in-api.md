# Contract: Local Provider Opt-in API

## `POST /v1/local-provider-detections/{provider_id}/enable`

Explicitly enables a detected local provider for this API process session.

### Success response

```json
{
  "provider": {
    "id": "local_codex",
    "family": "openai_compatible",
    "model": "gpt-5",
    "status": "unavailable",
    "message": "Local Codex is enabled for this session, but execution is unsupported until the Local Codex bridge is implemented.",
    "local_provider": true,
    "session_local": true,
    "credential_reference": "redacted",
    "execution_state": "unsupported"
  },
  "request_id": "req_..."
}
```

### Rejection response

Unavailable, needs-login, disabled, or unsupported local providers return a provider error and are not enabled.

## `DELETE /v1/local-provider-detections/{provider_id}/enable`

Disables the session-local local provider enablement. The response returns the removed safe capability when one existed.

## `GET /v1/model-providers`

Returns configured model providers plus explicitly enabled local provider capabilities. It must not run local detection or read local auth files.
