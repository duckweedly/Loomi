# Quickstart: M4 Run, Event, and SSE

This quickstart defines the local validation path for M4 planning. Exact commands may be adjusted during implementation, but the user-visible outcomes must remain the same.

## 1. Start local database

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
```

## 2. Apply migrations through M4

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected version:

```text
3
```

## 3. Start local API

```bash
APP_ENV=local \
HTTP_ADDR=127.0.0.1:8080 \
DATABASE_URL="$DATABASE_URL" \
LOG_LEVEL=info \
READINESS_TIMEOUT_SECONDS=5 \
go run ./cmd/loomi-api
```

## 4. Check readiness

```bash
curl -s http://127.0.0.1:8080/readyz
```

Expected: ready when database is available and migration version is clean and at least `3`.

## 5. Seed or create a thread

```bash
go run ./cmd/loomi-seed
curl -s http://127.0.0.1:8080/v1/threads
```

Use `thr_local_demo` or another active thread id for run smoke.

## 6. Start a deterministic local run

```bash
curl -s -X POST http://127.0.0.1:8080/v1/threads/$THREAD_ID/runs \
  -H 'Content-Type: application/json' \
  -d '{"script_name":"m4_smoke"}'
```

Expected:

- HTTP 201.
- Response includes run id, thread id, `source: local_simulated`, and active or terminal lifecycle state.
- Starting a second active run for the same thread returns a conflict instead of creating a second active run.

## 7. Read persisted event history

```bash
curl -s http://127.0.0.1:8080/v1/runs/$RUN_ID/events
```

Expected:

- Events are ordered by sequence.
- Events use only `lifecycle`, `progress`, `message`, `error`, or `final` categories.
- No secrets or local database URLs appear in event content.

## 8. Stream history then live events

```bash
curl -N http://127.0.0.1:8080/v1/runs/$RUN_ID/events/stream
```

Expected:

- Existing persisted events are delivered first.
- New events appear as they are persisted while the run is active.
- Terminal runs close the stream after the terminal event or stream close marker.

Reconnect smoke:

```bash
curl -N "http://127.0.0.1:8080/v1/runs/$RUN_ID/events/stream?after_sequence=$LAST_SEQUENCE"
```

Expected: events after `$LAST_SEQUENCE` only, followed by live events if the run is still active.

## 9. Stop an active run

```bash
curl -s -X POST http://127.0.0.1:8080/v1/runs/$RUN_ID/stop
```

Expected:

- Active deterministic local runs cooperatively reach `stopped` within the local smoke target.
- Terminal runs return `already_terminal` without changing the terminal outcome.
- Timeline contains stop-related lifecycle/final events.

## 10. Frontend smoke

Real API mode:

```bash
cd web
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run dev
```

Expected:

- Real thread/message behavior from M3 still works.
- Starting a run shows real run status and timeline events.
- Refresh restores persisted timeline history.
- Stream interruption is visible as recoverable state.
- Stop updates Chat Canvas, Run Timeline, and Agent state motion consistently.

Mock mode:

```bash
cd web
bun run dev
```

Expected: mock behavior remains available and is clearly mock-only.

## 11. Validation commands

```bash
go test ./...
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

If Playwright browser interaction is blocked by a local profile lock, record the exact blocker and use API/SSE smoke as fallback evidence.

## 12. Rollback and reapply smoke

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
migrate -path migrations -database "$DATABASE_URL" up 1
migrate -path migrations -database "$DATABASE_URL" version
```

Expected:

- Version goes from `3` to `2`, then back to `3`.
- M4 readiness fails after rollback and passes again after reapply.
