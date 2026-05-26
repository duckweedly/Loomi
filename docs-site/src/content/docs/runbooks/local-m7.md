---
title: Local M7 Tool Call Approval Runbook
description: Local validation commands and smoke expectations for the M7 tool-call approval foundation.
---

M7 currently has the complete current-time approval slice: schema, product-data services, runtime tool definition, provider-to-approval-required conversion for `runtime.get_current_time`, scoped tool-call reads, idempotent approve/deny actions, worker resume after approval, stream replay tests, diagnostics counters, frontend approval/result/error/cancel controls, cancellation precedence, and mixed model/tool RunRail polish. Full multi-step model continuation remains outside M7.

## Start local services

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
APP_ENV=local HTTP_ADDR=127.0.0.1:8080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
```

`/readyz` should report ready after schema version `6` is applied.

## Configure a local model provider

In the desktop Settings provider panel, save a local OpenAI-compatible provider with Base URL, model ID, and API key. The API accepts the same shape directly:

```bash
curl -sS http://127.0.0.1:8080/v1/model-providers \
  -H 'Content-Type: application/json' \
  -d '{"base_url":"https://gateway.example.test/v1","model":"gpt-5.5","api_key":"sk-..."}'
```

The response and later provider list only expose redacted capability fields. The current implementation stores this provider in the running API process and updates the model gateway immediately; restart the API to clear it.

## Foundation smoke expectations

Validate the foundation, observable request slice, and approve/deny decision slice with automated tests:

- `tool_calls` migration creates a unique `(run_id, tool_call_id)` projection and rolls back cleanly.
- `RecordToolCallRequest` records `tool_call_requested` and `tool_call_approval_required` once for duplicate requests.
- terminal runs reject new tool calls.
- a second tool call in the same run is rejected for the M7 MVP.
- unknown tool names and unknown tool arguments are rejected.
- omitted or `UTC` timezone is accepted; non-UTC timezone is rejected.
- worker diagnostics count approval-blocked tool calls.
- frontend runtime replay maps `tool.call.*` events into one stable `ToolCall` view model.
- gateway provider tool-call events for `runtime.get_current_time` become `tool_call_requested` and `tool_call_approval_required` without assistant message persistence or tool execution.
- approve schedules one worker job, emits `tool_call_executing`, and completes approved `runtime.get_current_time` calls with `tool_call_succeeded` plus `run_completed`.
- tool execution failures emit `tool_call_failed` plus `run_failed` with stable redacted error fields.
- scoped `GET /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}` returns the redacted current projection.
- ToolCallCard renders disabled approval controls without handlers, enabled controls when approve/deny handlers are wired, and terminal denied/executing/succeeded/failed/cancelled states.

## Validation commands

```bash
go test ./internal/productdata ./internal/runtime ./internal/db ./internal/httpapi ./cmd/...
bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/executionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

If a local Postgres integration database is available, also run the repository tests with the appropriate database environment for the Postgres path.

## Current non-goals

Do not smoke-test shell tools, filesystem tools, MCP, browser automation, arbitrary network tools, multi-agent execution, RAG/memory, or approval bypass in M7. Those capabilities are outside the approved M7 scope.
