# Quickstart: M5 LLM Gateway

This quickstart defines the local validation path for M5 planning. Exact command names may be adjusted during implementation, but the user-visible outcomes must remain the same.

## 1. Start local database

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
```

## 2. Apply migrations through M5

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version:

```text
4
```

## 3. Configure providers locally

M5 provider configuration is local-only and outside the product UI.

Required provider families for validation:

- Anthropic
- OpenAI
- Gemini
- OpenAI-compatible custom provider

Example custom provider values for local development:

```bash
LOOMI_MODEL_PROVIDER=custom
LOOMI_CUSTOM_MODEL_BASE_URL=https://apikey.tgjqr.com/v1
LOOMI_CUSTOM_MODEL_NAME=gpt-5.5
# LOOMI_CUSTOM_MODEL_API_KEY is supplied locally and must not be logged.
```

## 4. Start local API

```bash
APP_ENV=local \
HTTP_ADDR=127.0.0.1:8080 \
DATABASE_URL="$DATABASE_URL" \
LOG_LEVEL=info \
READINESS_TIMEOUT_SECONDS=5 \
go run ./cmd/loomi-api
```

Expected:

- `/readyz` is ready when the database is available and migration version is clean and at least `4`.
- Provider capability endpoint reports available, unavailable, or misconfigured without exposing secrets.

## 5. Seed or create a thread and user message

```bash
go run ./cmd/loomi-seed
curl -s http://127.0.0.1:8080/v1/threads
curl -s -X POST http://127.0.0.1:8080/v1/threads/$THREAD_ID/messages \
  -H 'Content-Type: application/json' \
  -d '{"content":"Say hello from the configured model provider."}'
```

Use the returned message id for model-gateway run smoke.

## 6. Check provider capability

```bash
curl -s http://127.0.0.1:8080/v1/model-providers
curl -s -X POST http://127.0.0.1:8080/v1/model-providers/check \
  -H 'Content-Type: application/json' \
  -d '{"provider_id":"custom"}'
```

Expected:

- The provider list includes configured provider families.
- No API key or Authorization header appears in the response.
- Misconfigured providers return user-safe status text.

## 7. Start a model gateway run

```bash
curl -s -X POST http://127.0.0.1:8080/v1/threads/$THREAD_ID/runs \
  -H 'Content-Type: application/json' \
  -d '{"message_id":"'$MESSAGE_ID'","source":"model_gateway","provider_id":"custom"}'
```

Expected:

- HTTP 201.
- Response includes `source: model_gateway`.
- Starting a second active run for the same thread returns a conflict.

## 8. Stream provider-normalized events

```bash
curl -N http://127.0.0.1:8080/v1/runs/$RUN_ID/events/stream
```

Expected:

- Existing persisted events are delivered first.
- Model output appears as `model_output_delta` events.
- Successful completion emits `model_output_completed` and `run_completed`.
- The final assistant response appears in message history exactly once.

Reconnect smoke:

```bash
curl -N "http://127.0.0.1:8080/v1/runs/$RUN_ID/events/stream?after_sequence=$LAST_SEQUENCE"
```

Expected: events after `$LAST_SEQUENCE` only, followed by live events if the run is still active.

## 9. Stop an active model run

```bash
curl -s -X POST http://127.0.0.1:8080/v1/runs/$RUN_ID/stop
```

Expected:

- Active model-gateway runs cooperatively reach `stopped`.
- Later provider deltas are ignored.
- Timeline contains stop-related final events.

## 10. Failure smoke

Run validation cases for:

- Provider unavailable.
- Provider misconfigured.
- Provider timeout.
- Provider rate limit.
- Empty response.
- Refusal or blocked response.
- Tool/function-call request.

Expected:

- Each case leaves the conversation usable.
- Each case emits a user-safe execution state.
- Secrets and raw provider payloads do not appear in events, logs intended for users, or message history.
- Tool/function-call requests create a non-executed boundary event.

## 11. Frontend smoke

Real API mode:

```bash
cd web
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run dev
```

Expected:

- Real thread/message behavior still works.
- Submitting a user message can start a model-gateway run when provider capability is available.
- Chat Canvas shows streaming assistant output from `model_output_delta` events.
- Run Timeline and debug surfaces show model/provider states with redacted errors.
- Mock success is not used as a hidden fallback when provider capability is unavailable.

Mock mode:

```bash
cd web
bun run dev
```

Expected: mock behavior remains available and is clearly mock-only.

## 12. Validation commands

```bash
go test ./...
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

If browser interaction is blocked by a local profile lock, record the exact blocker and use API/SSE smoke as fallback evidence.

## 13. Rollback and reapply smoke

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
migrate -path migrations -database "$DATABASE_URL" up 1
migrate -path migrations -database "$DATABASE_URL" version
```

Expected:

- Version goes from `4` to `3`, then back to `4`.
- M5 model gateway readiness fails after rollback and passes again after reapply when provider configuration is valid.
