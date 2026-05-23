# Quickstart: M3 Auth, Thread, and Message

This quickstart validates the M3 local identity, durable thread/message API, idempotent message creation, migration readiness, explicit seed behavior, frontend real/mock switching, and docs validation.

## Prerequisites

- Go 1.23+
- Docker with compose support
- Bun 1.3+
- `migrate` CLI available, or the equivalent documented Docker image command
- Local repository dependencies installed for `web/` and `docs-site/`

## 1. Configure local environment

```bash
cp .env.example .env
set -a
source .env
set +a
```

Expected local values:

```bash
APP_ENV=local
HTTP_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
LOG_LEVEL=info
READINESS_TIMEOUT_SECONDS=5
```

## 2. Start local PostgreSQL

```bash
docker compose up -d postgres
```

Expected: the local `postgres` service starts on `127.0.0.1:55433`.

## 3. Apply only the M2 baseline and verify M3 is not ready

From an empty local database, apply one migration:

```bash
migrate -path migrations -database "$DATABASE_URL" up 1
migrate -path migrations -database "$DATABASE_URL" version
```

Expected: version `1` with no dirty state.

Start the API in another terminal:

```bash
go run ./cmd/loomi-api
```

Check readiness:

```bash
curl -i http://127.0.0.1:8080/readyz
```

Expected: HTTP 503 with `status: "not_ready"` and a schema check failure because M3 requires schema version `2`.

## 4. Apply the M3 schema

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected: version `2` with no dirty state.

Check readiness again:

```bash
curl -i http://127.0.0.1:8080/readyz
```

Expected: HTTP 200 with `status: "ready"` when PostgreSQL is reachable.

## 5. Read the fixed local identity

```bash
curl -s http://127.0.0.1:8080/v1/me
```

Expected response shape:

```json
{
  "user": {
    "id": "user_local_dev",
    "display_name": "Local Developer",
    "created_at": "2026-05-23T00:00:00Z",
    "updated_at": "2026-05-23T00:00:00Z"
  },
  "request_id": "req_000000000000000000000000"
}
```

Timestamps and request id vary.

## 6. Create and list a real thread

```bash
curl -s -X POST http://127.0.0.1:8080/v1/threads \
  -H 'Content-Type: application/json' \
  -d '{"title":"M3 smoke thread","mode":"chat"}'
```

Expected: HTTP 201 response with a thread whose `lifecycle_status` is `active`.

List threads:

```bash
curl -s http://127.0.0.1:8080/v1/threads
```

Expected: the new thread appears in active-thread order by most recent update.

## 7. Rename the thread

Replace `$THREAD_ID` with the created thread id:

```bash
curl -s -X PATCH "http://127.0.0.1:8080/v1/threads/$THREAD_ID" \
  -H 'Content-Type: application/json' \
  -d '{"title":"M3 renamed smoke thread","mode":"work"}'
```

Expected: response contains the updated title and mode.

## 8. Create an idempotent user message

```bash
curl -s -X POST "http://127.0.0.1:8080/v1/threads/$THREAD_ID/messages" \
  -H 'Content-Type: application/json' \
  -d '{"content":"Persist this local user message.","client_message_id":"smoke-message-001"}'
```

Expected: response contains one `role: "user"` message with complete text content and no assistant placeholder.

Repeat the same command.

Expected: response returns the same message id, and the thread does not get a duplicate message.

List messages:

```bash
curl -s "http://127.0.0.1:8080/v1/threads/$THREAD_ID/messages"
```

Expected: messages appear in stable creation order.

## 9. Archive the thread

```bash
curl -s -X POST "http://127.0.0.1:8080/v1/threads/$THREAD_ID/archive"
```

Expected: response contains `lifecycle_status: "archived"`.

Default active list:

```bash
curl -s http://127.0.0.1:8080/v1/threads
```

Expected: archived thread is excluded.

Direct retrieval:

```bash
curl -s "http://127.0.0.1:8080/v1/threads/$THREAD_ID"
```

Expected: archived thread remains retrievable as durable state.

## 10. Validate invalid input errors

Empty message:

```bash
curl -i -X POST "http://127.0.0.1:8080/v1/threads/$THREAD_ID/messages" \
  -H 'Content-Type: application/json' \
  -d '{"content":"   "}'
```

Expected: HTTP 400 with a structured error containing `code`, `message`, and `request_id`.

Unknown thread:

```bash
curl -i http://127.0.0.1:8080/v1/threads/thr_missing/messages
```

Expected: HTTP 404 with `code: "thread_not_found"` and no ownership details.

## 11. Run the explicit seed command

```bash
go run ./cmd/loomi-seed
```

Expected: structured success diagnostic with an operation id, demo thread id, and demo message id.

Run it again.

Expected: no duplicate demo message is created.

## 12. Validate rollback and reapply

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
curl -i http://127.0.0.1:8080/readyz
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
curl -i http://127.0.0.1:8080/readyz
```

Expected:

- After rollback, version is `1` and readiness is HTTP 503.
- After reapply, version is `2` and readiness is HTTP 200 when PostgreSQL is reachable.
- No manual database cleanup is required.

## 13. Validate backend tests

```bash
go test ./...
```

Expected: all Go tests pass.

## 14. Validate web mock mode

With `VITE_LOOMI_API_BASE_URL` unset:

```bash
cd web
bun run build
bun run dev
```

Expected: the web shell opens with existing mock thread/message/run behavior.

## 15. Validate web real API mode

With the API running and M3 ready:

```bash
cd web
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run dev
```

Expected:

- Thread list and messages load from the real API.
- Sending a user message persists through the backend and survives refresh.
- Run timeline/debug surfaces remain mock, empty, or explicitly deferred.
- If the API is stopped while the env var is set, the UI shows a recoverable error and does not fall back to mock data.

## 16. Validate documentation site

```bash
cd docs-site
bun run build
```

Expected: docs site builds successfully.

## Explicitly Deferred Beyond M3

- run/event/SSE execution timeline
- LLM gateway and assistant message generation
- Tool calling
- Worker and job queue
- Desktop runtime, SQLite adapter, bridge, tray, and auto-update
- Attachments and file upload
- RAG/context ingestion
- Catalog, marketplace, plugin, or extension runtime
- Production authentication, organizations, hosted deployment, and release packaging
