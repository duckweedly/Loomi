# Data Model: M5 LLM Gateway

## Message

M5 extends the M3 message concept so completed assistant responses can become durable conversation history.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable message id, unique globally |
| `thread_id` | string | References one owned thread |
| `user_id` | string | Fixed local owner inherited from the current identity boundary |
| `role` | enum | `user` or `assistant` |
| `content` | string | Non-empty final message text |
| `metadata` | object | Safe metadata only; no credentials or raw provider secrets |
| `client_message_id` | string/null | User-message idempotency key; only required for client-created user messages |
| `created_at` | timestamp | Machine-readable creation time |

Rules:

- User messages continue to be created from the conversation composer.
- Assistant messages are created only after a model-backed response completes successfully.
- A completed model-gateway run appends at most one assistant message.
- Failed, stopped, refused, empty, or unavailable runs do not create a successful assistant message.
- Message metadata may reference `run_id`, provider family, or model label when safe, but must not include API keys or raw provider error payloads.

## Run

M5 extends M4 runs from deterministic local execution to model gateway execution.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable run id, unique globally |
| `thread_id` | string | References one owned thread |
| `user_id` | string | Fixed local owner |
| `status` | enum | `pending`, `running`, `completed`, `failed`, `stopped` |
| `source` | enum | `local_simulated` or `model_gateway` |
| `title` | string | Short user-visible run label |
| `created_at` | timestamp | Machine-readable creation time |
| `updated_at` | timestamp | Machine-readable latest lifecycle or event update time |
| `completed_at` | timestamp/null | Set when status is terminal |
| `error_code` | string/null | Stable redacted code for failed model runs |
| `error_message` | string/null | User-safe failure message |

Rules:

- A run belongs to exactly one thread and one local user.
- A single thread must not have more than one active run.
- Model gateway runs use `source = model_gateway`.
- Terminal runs remain readable as history.
- Stop is cooperative: no later model deltas are applied after the run reaches `stopped`.

State transitions:

```text
pending -> running -> completed
pending -> running -> failed
pending -> running -> stopped
pending -> stopped
```

Invalid transitions:

- Any terminal state to active.
- Creating a second active run for the same thread.
- Applying provider output after a terminal state.

## Run Event

M5 reuses M4 run events and adds provider-normalized event types within the existing categories.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable event id, unique globally |
| `run_id` | string | References one run |
| `thread_id` | string | Denormalized for scoped lookups and safety checks |
| `user_id` | string | Fixed local owner |
| `sequence` | integer | Monotonic per run, starts at 1 |
| `category` | enum | `lifecycle`, `progress`, `message`, `error`, `final` |
| `type` | string | Provider-normalized event type |
| `summary` | string | Short user-safe summary |
| `content` | string/null | Text delta, final text, or user-safe failure details when appropriate |
| `metadata` | object | Safe diagnostics such as provider family, model label, finish reason, and token usage |
| `created_at` | timestamp | Machine-readable event time |

M5 event types:

| Category | Event Type | Purpose |
|----------|------------|---------|
| `lifecycle` | `run_created` | Run record created |
| `lifecycle` | `run_started` | Gateway request started |
| `progress` | `model_request_started` | Provider request accepted locally |
| `message` | `model_output_delta` | Incremental assistant text |
| `message` | `model_output_completed` | Final assistant text is ready to persist |
| `progress` | `model_refusal` | Provider returned a refusal or blocked response without successful text |
| `progress` | `tool_call_blocked` | Provider requested tool/function use, but tool execution is out of scope |
| `error` | `provider_error` | Provider returned a redacted failure |
| `error` | `provider_timeout` | Provider did not respond within the configured timeout |
| `error` | `provider_rate_limited` | Provider rejected or throttled the request |
| `final` | `run_completed` | Run completed successfully |
| `final` | `run_failed` | Run failed terminally |
| `final` | `run_stopped` | Run stopped terminally |

Rules:

- Events are ordered by `(sequence, id)` within a run.
- Event payloads are data, not instructions.
- Provider-native event names are not product-facing contract names.
- Provider error metadata must be redacted before persistence.
- Duplicate stream chunks must not duplicate final assistant text.

## Provider Configuration

A local configuration record that selects and authenticates a model provider outside the product UI.

| Field | Type | Rules |
|-------|------|-------|
| `provider_id` | string | Stable local identifier such as `anthropic`, `openai`, `gemini`, or a custom id |
| `family` | enum | `anthropic`, `openai`, `gemini`, `openai_compatible` |
| `base_url` | string/null | Required for OpenAI-compatible custom providers; optional default for built-ins |
| `api_key_ref` | string | Reference to local configuration secret, never exposed to frontend or events |
| `model` | string | Provider model label used for validation and requests |
| `enabled` | boolean | Whether this provider can be selected for local execution |

Rules:

- Provider configuration is local development configuration, not product UI state.
- API keys are read by the backend only.
- Custom providers must use an OpenAI-compatible HTTP chat and streaming interface.
- User-visible behavior remains provider-neutral.

## Request Context

The bounded input sent to a provider for one model-backed response.

| Field | Type | Rules |
|-------|------|-------|
| `thread_id` | string | Selected conversation |
| `trigger_message_id` | string | User message that started the run |
| `messages` | ordered list | Current user message plus necessary recent messages from the same thread |
| `provider_id` | string | Selected provider configuration |
| `model` | string | Selected model label |

Rules:

- Request context never includes other threads.
- Attachments, RAG, memory, and pipeline context are out of scope for M5.
- Request context is constructed server-side so browser clients do not send arbitrary provider payloads.

## Gateway Stream State

The backend state while consuming provider output and converting it into Loomi events.

| State | Meaning |
|-------|---------|
| `configuring` | Provider configuration is being resolved |
| `connecting` | Provider request is being opened |
| `streaming` | Text deltas or provider events are being consumed |
| `finalizing` | Final assistant text and terminal run state are being persisted |
| `failed` | Provider or gateway failure was redacted and recorded |
| `stopped` | Stop was requested and no more deltas should apply |
| `completed` | Final assistant message and run completion are persisted |

Rules:

- `failed`, `stopped`, and `completed` are terminal for the gateway stream.
- Stop can happen before or during provider streaming.
- Provider stream disconnects map to user-safe failure or recovery behavior through run events.

## Tool Boundary Event

A non-executed record that model output requested or implied a tool call.

| Field | Type | Rules |
|-------|------|-------|
| `run_id` | string | Run where the tool-like output appeared |
| `tool_name` | string/null | Provider-reported tool/function name when available and safe |
| `summary` | string | User-safe explanation that tool execution is outside M5 scope |
| `created_at` | timestamp | Machine-readable event time |

Rules:

- No external action is performed.
- Tool arguments are not executed.
- Sensitive tool arguments are not persisted in user-visible content.
