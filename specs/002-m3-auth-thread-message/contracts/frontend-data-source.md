# Frontend Data Source Contract: M3 Real/Mock Switching

The web shell must keep M1 mock behavior when no backend is configured and use M3 real thread/message APIs when a backend base URL is configured.

## Environment Variable

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080
```

Rules:

- Empty or missing value means `mock` mode.
- Non-empty value means `real_api` mode.
- The value must be treated as the base URL for M3 API requests.

## Mock Mode

Trigger:

```text
VITE_LOOMI_API_BASE_URL is absent or empty
```

Behavior:

- Existing mock thread, message, run timeline, and debug rail behavior remains usable.
- Sending a mock message may continue to create mock assistant/run timeline content because it is clearly mock-only behavior.
- No backend request is attempted for thread/message data.

## Real API Mode

Trigger:

```text
VITE_LOOMI_API_BASE_URL is set
```

Behavior:

- Thread list loads from `GET /v1/threads`.
- Opening a thread loads messages from `GET /v1/threads/{thread_id}/messages`.
- Creating, renaming, and archiving threads use the M3 API.
- Sending a message uses `POST /v1/threads/{thread_id}/messages` and includes a client-generated `client_message_id` with timestamp plus random suffix.
- API timestamps remain machine-readable at the client boundary; display formatting happens in UI components.

## Real API Error Mode

Trigger:

```text
VITE_LOOMI_API_BASE_URL is set and the API is unavailable or returns an error
```

Behavior:

- The UI shows a recoverable error state.
- The UI must not silently fall back to mock thread/message data.
- Existing run timeline/debug surfaces remain mock, empty, or explicitly deferred; they must not imply real backend execution.

## React Loading Rules

When Effects load real API data:

- Declare all dependencies used by the Effect.
- Use cleanup or an equivalent stale-response guard so old responses do not overwrite the current selected thread state; the implemented hook compares the requested thread id with the latest selected thread id before applying results.
- Keep request errors visible in state so the UI can render a recoverable error.

## Deferred Frontend Behavior

M3 does not add:

- Real run timeline data.
- Real SSE/event streaming.
- LLM-generated assistant messages.
- Tool call cards backed by the API.
- Desktop runtime or Electron bridge behavior.
- Attachment upload, RAG, or catalog UI.
