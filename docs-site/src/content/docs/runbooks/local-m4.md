---
title: Local M4 Run/Event/SSE Runbook
description: Commands for validating the local M4 run, event history, SSE stream, stop, and rollback path.
---

## Start dependencies

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
```

## Apply migration 000003

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version is `3`. `/readyz` should fail before version 3 and pass after the database is reachable and the schema version is clean.

## Run API locally

```bash
APP_ENV=local \
HTTP_ADDR=127.0.0.1:8080 \
DATABASE_URL="$DATABASE_URL" \
LOG_LEVEL=info \
READINESS_TIMEOUT_SECONDS=5 \
go run ./cmd/loomi-api
```

In another shell:

```bash
curl -s http://127.0.0.1:8080/readyz
```

## Seed a thread

```bash
go run ./cmd/loomi-seed
curl -s http://127.0.0.1:8080/v1/threads
export THREAD_ID=thr_local_demo
```

## Start a local simulated run

```bash
curl -s -X POST "http://127.0.0.1:8080/v1/threads/$THREAD_ID/runs" \
  -H 'Content-Type: application/json' \
  -d '{"script_name":"m4_smoke"}'
```

Save the returned run id:

```bash
export RUN_ID=run_...
```

Expected: `status` is `running` or terminal if the deterministic runner already finished, and `source` is `local_simulated`.

## Read history

```bash
curl -s "http://127.0.0.1:8080/v1/runs/$RUN_ID/events"
```

Expected: events are ordered by `sequence`, use only M4 categories, and do not include database URLs or secret-looking values in summary/content.

## Stream history then live

```bash
curl -N "http://127.0.0.1:8080/v1/runs/$RUN_ID/events/stream"
```

Reconnect from a known cursor:

```bash
curl -N "http://127.0.0.1:8080/v1/runs/$RUN_ID/events/stream?after_sequence=3"
```

If the stream drops, reload persisted history with `GET /v1/runs/{run_id}/events` and reconnect with the last observed sequence.

## Stop a run

```bash
curl -s -X POST "http://127.0.0.1:8080/v1/runs/$RUN_ID/stop"
```

Expected active runs become `stopped`; terminal runs return `already_terminal`.

## Frontend real API mode

```bash
bun install --cwd web
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev
```

The shell should keep M3 thread/message behavior, start a local simulated run when sending a message, show persisted events in the run rail, and show stop state consistently.

## Validation commands

```bash
go test ./...
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

## Rollback and reapply

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
migrate -path migrations -database "$DATABASE_URL" up 1
migrate -path migrations -database "$DATABASE_URL" version
```

Version should go from `3` to `2` and back to `3`. Rollback removes local M4 run/event data.
