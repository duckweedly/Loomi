---
title: Local M7 Tool Call Approval Runbook
description: Local validation commands and smoke expectations for the M7 tool-call approval and continuation slices.
---

M7 now has the minimal approval execution closure plus a tool-result continuation boundary: schema, product-data services, runtime tool definition, provider-to-approval-required conversion for `runtime.get_current_time`, scoped tool-call reads, idempotent approve/deny, worker resume, current-time execution, terminal result/error events, frontend approval controls, and provider-neutral continuation context for the second model phase.

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

## Approval execution smoke expectations

Validate the foundation, execution closure, and continuation boundary with automated tests and a local browser smoke:

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
- `POST /approve` records one approved event and queues exactly one resume.
- `POST /deny` records one denied event, stops the MVP run, and never writes executing.
- approved `runtime.get_current_time` writes `tool_call_executing` then `tool_call_succeeded`.
- tool failures write `tool_call_failed` with redacted error fields.
- ToolCallCard approve/deny buttons call the real API and show loading, disabled, and error states.
- tool-result continuation can build a provider request from `tool_call_requested` and `tool_call_succeeded`.
- OpenAI-compatible continuation requests include an assistant tool call followed by a matching tool result.
- continuation provider deltas are recorded with `model_phase = continuation`.
- a second tool request during continuation fails with `unsupported_tool_loop`.

## Validation commands

```bash
go test ./internal/productdata ./internal/runtime ./internal/db ./internal/httpapi ./cmd/...
bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/executionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

For the continuation slice specifically:

```bash
go test ./internal/runtime -run 'TestHTTPProviderSerializesOpenAIToolResultContinuation|TestGatewayBuildsContinuationContextFromToolResultEvents|TestGatewayContinuesAfterToolResultAndPersistsFinalAssistant|TestGatewayFailsWhenContinuationRequestsAnotherTool'
bun test ./web/src/runtime/realExecutionAdapter.test.ts
```

## Browser smoke

1. Open the local web shell against the local API.
2. Trigger a model or fake-provider run that emits `runtime.get_current_time`.
3. Confirm the ToolCallCard shows approval required.
4. Click Deny and confirm the card becomes denied and no executing event appears.
5. Trigger another run, click Approve, and confirm approved -> executing -> succeeded appears in ToolCallCard, RunRail, Timeline, and after refresh/SSE replay.
6. For continuation-capable provider/fake responses, confirm `tool_call_succeeded` is followed by continuation `model_output_delta`, one final assistant answer, and `run_completed`.

If a local Postgres integration database is available, also run the repository tests with the appropriate database environment for the Postgres path.

## Current non-goals

Do not smoke-test shell tools, filesystem tools, MCP, browser automation, arbitrary network tools, multi-agent execution, RAG/memory, multi-tool concurrency, or approval bypass in M7. Those capabilities are outside the approved M7 scope.
