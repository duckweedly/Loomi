---
title: M20 Local Codex Execution API
description: Supported Local Codex provider capability and model gateway execution contract.
---

## `POST /v1/local-provider-detections/local_codex/enable`

Enables Local Codex for the current API process session after explicit user detection and enable actions.

When the execution bridge and credential snapshot are available:

```json
{
  "provider": {
    "id": "local_codex",
    "family": "openai_compatible",
    "model": "gpt-5.5",
    "status": "available",
    "message": "Local Codex is enabled for this session.",
    "local_provider": true,
    "session_local": true,
    "credential_reference": "redacted",
    "execution_state": "supported"
  },
  "request_id": "req_..."
}
```

If the bridge or login snapshot is unavailable, the endpoint returns a provider unavailable error and does not enable execution.

## `GET /v1/model-providers`

Returns configured providers plus explicitly enabled local providers. Enabled executable Local Codex returns `status=available` and `execution_state=supported`.

This endpoint must not run local detection, read auth files, refresh OAuth, call keychain, invoke CLI, or validate external login.

## `POST /v1/threads/{thread_id}/runs`

For `source=model_gateway` and `provider_id=local_codex`, the HTTP handler accepts supported Local Codex runs and still rejects unsupported local providers before run creation.

Successful runs emit the existing Gateway event contract:

- `model_request_started`
- `model_output_delta`
- `model_output_completed`
- `run_completed`

Provider failures emit existing Gateway failure events and do not create assistant output.

## Redaction Contract

Responses, run events, assistant metadata, and frontend state must not include API keys, bearer tokens, `access_token`, `refresh_token`, Authorization headers, auth file paths, private home paths, raw auth JSON, or CLI paths.
