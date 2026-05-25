---
title: Local M6 Worker Job Pipeline Runbook
description: Local validation path for queued run acknowledgement, worker processing, frontend replay, and M6 MVP limitations.
---

M6 validates the first background execution slice: queued acknowledgement, durable job creation, worker claim, persisted worker/pipeline events, and frontend replay.

## Start local services

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
APP_ENV=local HTTP_ADDR=127.0.0.1:8080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
```

`/readyz` should report ready after schema version `5` is applied.

Worker defaults:

```bash
export LOOMI_WORKER_QUEUE_ENABLED=true
export LOOMI_WORKER_QUEUE_PAUSED=false
export LOOMI_WORKER_LEASE_SECONDS=30
export LOOMI_WORKER_MAX_ATTEMPTS=3
export LOOMI_WORKER_POLL_MILLIS=250
```

## Queued acknowledgement smoke

Start a run from an existing thread:

```bash
curl -s -X POST http://127.0.0.1:8080/v1/threads/$THREAD_ID/runs \
  -H 'Content-Type: application/json' \
  -d '{"script_name":"m4_smoke"}'
```

Expected:

- HTTP `202 Accepted`
- response run status is `queued`
- event history contains `run_created` then `run_queued`
- the request returns before worker execution completes

## Worker processing smoke

Poll events or open SSE:

```bash
curl -N http://127.0.0.1:8080/v1/runs/$RUN_ID/events/stream
```

Expected happy path:

- `run_queued`
- `job_claimed`
- `lease_renewed`
- `pipeline_step_started`
- `pipeline_step_completed`
- `run_completed`

Expected recovery path after a lease expires:

- `job_recovering`
- `job_retry_scheduled`
- a fresh `job_claimed` with a higher ownership version
- `job_retry_exhausted` and `run_failed` after max attempts are exhausted

The final run should have one terminal outcome. Local simulated completion should still persist the assistant message through the existing runner path.

## Frontend smoke

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev
```

In the browser:

1. open Settings > Providers and confirm Provider Test Console shows configured providers or env guidance
2. use Test connection and confirm checking/success/failed states do not expose provider secrets
3. select or create a thread
4. send a message
5. confirm the Chat Canvas enters a waiting/queued state
6. confirm RunRail shows worker/job and pipeline events
7. open Background tasks and confirm it is a read-only observer, not a job control surface
8. confirm completion returns the canvas to a completed/history state

Do not mark UI validation complete unless the browser was actually exercised.

## Automated validation

```bash
go test ./...
zsh -o null_glob -c 'bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts'
bun run --cwd web build
bun run --cwd docs-site build
```

The raw quickstart glob may fail under zsh when a pattern has no matches. Use `zsh -o null_glob -c ...` or an explicit file list for local validation.

## Current limitations

Model-gateway jobs hydrate `message_id`, `provider_id`, and model override from durable queued metadata. Desktop runtime, tool execution, RAG/memory, plugins, and multi-agent orchestration remain outside M6.
