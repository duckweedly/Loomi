# Contract: Local Codex Execution API

## `POST /v1/local-provider-detections/local_codex/enable`

Explicitly enables Local Codex for the current API process session after manual detection.

### Supported success response

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

### Unavailable response

Unavailable, malformed, or missing auth returns a provider error. Responses must not include token, key, Authorization header, auth file path, private home path, or raw auth JSON.

## `GET /v1/model-providers`

Returns configured providers plus explicitly enabled local provider capabilities. It must not run local detection or read local auth files. Enabled executable Local Codex returns `status=available` and `execution_state=supported`.

## `POST /v1/threads/{thread_id}/runs`

For `source=model_gateway` and `provider_id=local_codex`, the handler accepts only supported Local Codex capabilities and rejects unsupported/unavailable local provider states before run creation.

## Run Event Contract

Successful execution emits the existing event types:

- `model_request_started`
- `model_output_delta`
- `model_output_completed`
- `run_completed`

Provider failure emits existing gateway failure events and must not append assistant output.
