---
title: Local M5 LLM Gateway Runbook
description: Local validation path for model provider configuration, model-gateway runs, and provider-normalized events.
---

M5 validates model-backed execution in local development while keeping provider secrets outside the product UI.

## Provider setup

Provider configuration is supplied through local backend environment variables.

```bash
# Anthropic
export LOOMI_ANTHROPIC_API_KEY=...
export LOOMI_ANTHROPIC_MODEL=claude-opus-4-7

# OpenAI
export LOOMI_OPENAI_API_KEY=...
export LOOMI_OPENAI_MODEL=gpt-4.1

# Gemini
export LOOMI_GEMINI_API_KEY=...
export LOOMI_GEMINI_MODEL=gemini-3.5-flash

# OpenAI-compatible custom provider
export LOOMI_CUSTOM_MODEL_PROVIDER_ID=custom
export LOOMI_CUSTOM_MODEL_BASE_URL=https://gateway.example.test/v1
export LOOMI_CUSTOM_MODEL_NAME=gpt-5.5
export LOOMI_CUSTOM_MODEL_API_KEY=...
```

Do not commit or paste real provider keys. The API exposes only redacted provider capability.

## Start local services

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
APP_ENV=local HTTP_ADDR=127.0.0.1:8080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
```

`/readyz` should report ready after schema version `4` is applied.

## Capability smoke

```bash
curl -s http://127.0.0.1:8080/v1/model-providers
curl -s -X POST http://127.0.0.1:8080/v1/model-providers/check \
  -H 'Content-Type: application/json' \
  -d '{"provider_id":"custom"}'
```

For browser smoke from the Vite dev server, start the API with `APP_ENV=local` and the web app with:

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev
```

The API allows CORS only for `http://127.0.0.1:5173` and `http://localhost:5173` in local/development mode. Preflight supports `GET`, `POST`, `PATCH`, `OPTIONS`, and `Content-Type` without wildcard credentials.

Expected:

- available providers return redacted id/family/model/base URL
- disabled providers return `provider_unavailable`
- incomplete custom providers return `provider_misconfigured`
- no API key, Authorization header, or raw provider error body appears in the response

## Model-gateway run smoke

Create or seed a thread, then create a durable user message:

```bash
curl -s -X POST http://127.0.0.1:8080/v1/threads/$THREAD_ID/messages \
  -H 'Content-Type: application/json' \
  -d '{"content":"Say hello from the configured model provider."}'
```

Start a model-gateway run with the returned message id:

```bash
curl -s -X POST http://127.0.0.1:8080/v1/threads/$THREAD_ID/runs \
  -H 'Content-Type: application/json' \
  -d '{"message_id":"'$MESSAGE_ID'","source":"model_gateway","provider_id":"custom"}'
```

Stream events:

```bash
curl -N http://127.0.0.1:8080/v1/runs/$RUN_ID/events/stream
```

Expected successful path:

- `model_request_started`
- one or more `model_output_delta` events
- `model_output_completed`
- `run_completed`
- exactly one assistant message in thread history for the completed run

## Failure and boundary smoke

Use local provider configuration or test providers to exercise:

- disabled or misconfigured provider
- provider timeout
- rate limit
- generic provider error
- refusal or blocked response
- empty response
- stop while running
- provider tool/function-call output

Expected:

- failures use stable redacted event codes
- conversations remain usable
- stopped runs do not append further assistant text
- tool/function-call output is recorded as `tool_call_blocked` and is not executed

## Automated validation

```bash
go test ./...
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

If browser interaction or real provider credentials are unavailable, record the blocker in the devlog and use API/unit coverage as fallback evidence.
