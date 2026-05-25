---
title: Local M9 RunContext Pipeline Runbook
description: Local validation path for durable RunContext loading, linear pipeline stage trace, and browser smoke.
---

M9 adds the smallest RunContext + Pipeline foundation on top of the existing worker/job queue. It does not add a new queue, Redis, external workers, Persona/Skill, MCP, Memory/RAG, Sandbox, Desktop Runtime, shell/filesystem/browser automation, or multi-agent behavior.

## Start local services

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
APP_ENV=local HTTP_ADDR=127.0.0.1:8080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
```

Worker queue should remain enabled:

```bash
export LOOMI_WORKER_QUEUE_ENABLED=true
export LOOMI_WORKER_QUEUE_PAUSED=false
```

## Expected stage trace

Create a run from the web UI or API. Event history should include worker/job events plus these pipeline stages:

```text
pipeline_step_started      step=prepare_context
pipeline_step_completed    step=prepare_context
pipeline_step_started      step=resolve_tools
pipeline_step_completed    step=resolve_tools
pipeline_step_started      step=invoke_runtime
pipeline_step_completed    step=invoke_runtime
pipeline_step_started      step=finalize
pipeline_step_completed    step=finalize
```

For model-gateway runs, `prepare_context` restores the durable run, thread, message history, job metadata, provider/model route, and enabled MVP tool summary before runtime invocation. For approved tool resume jobs, RunContext carries the tool-call id and lets the existing M7 continuation path run after `runtime.get_current_time` succeeds.

## Failure smoke

A malformed model-gateway job without a provider route should fail before provider invocation:

- `pipeline_step_started` with `step=prepare_context`
- `pipeline_step_failed` with `step=prepare_context`
- a redacted terminal `run_failed`
- no `model_request_started`

## Browser smoke

1. Start the local API/worker and web app in real API mode.
2. Create or open a thread with a user message.
3. Start a run with the current provider/model route.
4. Open Timeline or Background tasks/debug details.
5. Confirm the trace shows context prepared, tools resolved, runtime invoked, and finalized.
6. Refresh the page and confirm history replay shows the same stage trace.
7. Confirm stage details are safe summaries and do not expose provider credentials, raw tool results, file contents, shell output, or hidden local state.

## Validation commands

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/RunTimeline.runtime.test.ts ./web/src/components/RunRail.runtime.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```
