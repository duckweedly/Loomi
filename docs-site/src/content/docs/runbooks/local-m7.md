---
title: Local M7 Tool Call Approval Runbook
description: Local validation commands and smoke expectations for the M7 tool-call approval foundation.
---

M7 currently has the Phase 2 foundation plus the US1 observable request slice: schema, product-data services, runtime tool definition, provider-to-approval-required conversion for `runtime.get_current_time`, scoped tool-call reads, stream replay tests, diagnostics counters, and frontend approval-required placeholders. Full approve/deny UI and execution smoke paths are later M7 tasks.

## Start local services

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
APP_ENV=local HTTP_ADDR=127.0.0.1:8080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
```

`/readyz` should report ready after schema version `6` is applied.

## Foundation smoke expectations

Until approve/deny endpoints land, validate the foundation and observable request slice with automated tests:

- `tool_calls` migration creates a unique `(run_id, tool_call_id)` projection and rolls back cleanly.
- `RecordToolCallRequest` records `tool_call_requested` and `tool_call_approval_required` once for duplicate requests.
- terminal runs reject new tool calls.
- a second tool call in the same run is rejected for the M7 MVP.
- unknown tool names and unknown tool arguments are rejected.
- omitted or `UTC` timezone is accepted; non-UTC timezone is rejected.
- worker diagnostics count approval-blocked tool calls.
- frontend runtime replay maps `tool.call.*` events into one stable `ToolCall` view model.
- gateway provider tool-call events for `runtime.get_current_time` become `tool_call_requested` and `tool_call_approval_required` without assistant message persistence or tool execution.
- scoped `GET /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}` returns the redacted current projection.
- ToolCallCard renders approval-required placeholders with disabled controls until approve/deny handlers land.

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
