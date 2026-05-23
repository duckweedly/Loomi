# Quickstart: M2 API and Database Base

This quickstart validates the M2 local service boundary, readiness behavior, schema workflow, and M1 UI preservation.

## Prerequisites

- Go 1.23+
- Docker with compose support
- Bun 1.3+
- `migrate` CLI available, or use the documented Docker image equivalent during implementation

## 1. Configure local environment

Create and load a local env file from the example once implementation adds it:

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

## 3. Apply the M2 schema baseline

```bash
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected: version `1` with no dirty state.

## 4. Run backend tests

```bash
go test ./...
```

Expected: all Go tests pass.

## 5. Start the API service

```bash
go run ./cmd/loomi-api
```

Expected: the service starts on `127.0.0.1:8080` and emits structured diagnostics with an operation id.

## 6. Check liveness

```bash
curl -i http://127.0.0.1:8080/healthz
```

Expected: HTTP 200 and JSON body with `status: "alive"`, `service: "loomi-api"`, and `request_id`.

## 7. Check readiness

```bash
curl -i http://127.0.0.1:8080/readyz
```

Expected: HTTP 200 and JSON body with `status: "ready"` after PostgreSQL is reachable and the schema baseline is applied.

## 8. Validate not-ready behavior

Stop local PostgreSQL:

```bash
docker compose stop postgres
```

Then run:

```bash
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
```

Expected:

- `/healthz` still returns HTTP 200.
- `/readyz` returns HTTP 503 with `status: "not_ready"` and a non-secret database failure reason within 10 seconds.

Restart PostgreSQL before continuing:

```bash
docker compose start postgres
```

## 9. Validate rollback and re-apply

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected: migration can roll back and re-apply without manual cleanup; final version is `1` and clean.

## 10. Validate existing M1 UI shell still builds

```bash
cd web
bun run build
```

Expected: build passes. Existing Vite chunk-size warnings are acceptable if there are no errors.

## 11. Validate documentation site after docs updates

```bash
cd docs-site
bun run build
```

Expected: docs site builds successfully.

## Explicitly Deferred Beyond M2

- Authentication
- Users, threads, and messages
- Runs, events, and SSE
- Worker and job queue
- LLM gateway
- Tool calling
- Desktop runtime, SQLite adapter, bridge, tray, and auto-update
- Production deployment and release packaging
