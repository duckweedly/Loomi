# Contract: Frontend Runtime for M5 Model Gateway

M5 keeps the frontend runtime adapter boundary and adds real model capability states returned by the backend.

## Data-source modes

| Mode | Trigger | User-visible behavior |
|------|---------|-----------------------|
| Mock | API base URL absent | Existing mock thread/runtime behavior remains clearly mock-only. |
| Real local simulated | API base URL set and run source is `local_simulated` | Existing M4 run/event/SSE behavior remains available. |
| Real model gateway | API base URL set and provider capability is available | User messages can start `model_gateway` runs and stream provider-normalized events. |
| Provider unavailable | API base URL set but selected provider cannot be used | UI shows model capability unavailable and does not simulate success. |

## Runtime adapter expectations

### sendMessage

Creates or returns the user message in the selected thread.

**M5 rule**: A model-backed run starts only after the user message is durable or otherwise has a stable message id.

### createRun

Starts a run for an existing message.

**Input**

- `threadId`
- `messageId`
- `source`: `local_simulated` or `model_gateway`
- `providerId` when `source` is `model_gateway`

**Output**

- Runtime run with status `pending` or `running`.
- First visible event `run_created`.
- Explicit provider unavailable state when provider capability is missing.

### subscribeRunEvents

Streams M4/M5 history-first run events.

**M5 rule**: `model_output_delta` events update assistant draft text; `model_output_completed` finalizes the assistant response; terminal final events close the active run.

### stopRun

Requests cooperative stop for an active run.

**M5 rule**: Later model deltas for the stopped run are ignored by the stale-event guard.

## UI mapping

| Event Type | Chat Canvas | Timeline / Debug | Agent Motion |
|------------|-------------|------------------|--------------|
| `model_request_started` | pending assistant bubble | model request row | thinking |
| `model_output_delta` | append assistant draft | model stream row | speaking |
| `model_output_completed` | final assistant content ready | completion row | speaking/done |
| `model_refusal` | failed/refused assistant state | provider refusal row | error |
| `tool_call_blocked` | capability-boundary note | tool boundary row | thinking/error depending run outcome |
| `provider_error` | failed assistant state | redacted provider error row | error |
| `provider_timeout` | failed assistant state | timeout row | error |
| `provider_rate_limited` | failed assistant state | rate-limit row | error |

## Capability display rules

- Provider status must distinguish unavailable, misconfigured, rate-limited, timed out, and failed model generation when the backend can classify them.
- No frontend surface displays provider credentials.
- Real API mode never falls back to mock output for a model-gateway run.
- Mock and local simulated modes remain labeled as non-model execution.
