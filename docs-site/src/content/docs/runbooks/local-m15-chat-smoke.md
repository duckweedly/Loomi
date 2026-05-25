---
title: Local M15 Chat Smoke
description: Local validation for the deterministic real Chat integrated closeout smoke.
---

## Scope

This runbook validates the M15 closeout path:

- real HTTP/API Chat run creation
- worker-owned RunContext preparation
- approved memory snapshot loaded into the run
- deterministic provider fixture requesting one persona-allowed discovered local stdio MCP tool
- approval required before execution
- HTTP approve
- one MCP `tools/call`
- redacted tool result
- one provider continuation
- final assistant message
- replayable run history

M15 does not add sandboxing, filesystem/shell/browser automation tools, activity recording, OpenViking/vector/RAG/distill, marketplace/plugin install, multi-agent behavior, or a worker queue rewrite.

## Gated Smoke Command

```bash
LOOMI_M15_REAL_CHAT_SMOKE=1 go test ./internal/httpapi -run TestM15ChatRealIntegratedSmoke -count=1 -v
```

Expected output:

- test passes
- log includes `run_id`, `tool_call_id`, final assistant text, and event types
- event types include memory, MCP discovery, approval, execution, continuation, and completion milestones

Without `LOOMI_M15_REAL_CHAT_SMOKE=1`, the test skips intentionally so normal package tests stay cheap.

## Smoke Mechanics

The test uses the Go test binary itself as the local stdio MCP fixture. The fixture speaks MCP `Content-Length` frames for discovery and `tools/call`, increments a temporary count file on execution, and returns a result containing sensitive-looking canaries. The count file proves no MCP execution occurs before HTTP approval and exactly one execution occurs afterward.

The deterministic provider fixture has two phases:

1. Initial model phase: requests the discovered namespaced MCP tool.
2. Continuation phase: after redacted tool result, emits `M15 chat smoke complete.`

## Replay Checks

The smoke fetches `GET /v1/runs/{run_id}/events` and verifies:

- `run_queued`
- `job_claimed`
- `pipeline_step_started` and `pipeline_step_completed`
- `memory_snapshot_loaded`
- `mcp_discovery_succeeded` with `candidate_schema_hashes`
- `tool_call_approval_required`
- `tool_call_approved`
- `tool_call_executing`
- `tool_call_succeeded`
- continuation `model_request_started`
- continuation `model_output_completed`
- `run_completed`

## Redaction Checks

The smoke fails if sensitive canaries appear in:

- API run responses
- API event replay responses
- API message responses
- RunContext safe summary metadata
- tool result summary
- provider continuation request
- M15 docs examples

Use only safe synthetic placeholders in docs and logs.

## Full M15 Validation

```bash
LOOMI_M15_REAL_CHAT_SMOKE=1 go test ./internal/httpapi -run TestM15ChatRealIntegratedSmoke -count=1 -v
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Browser Smoke

If a local API and web dev server are available, start them in real API mode and open Chat:

```bash
APP_ENV=local HTTP_ADDR=127.0.0.1:18080 go run ./cmd/loomi-api
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080 bun run --cwd web dev --host 127.0.0.1 --port 5173
```

Verify the real timeline can show the same replay states. If browser smoke is blocked, cite the blocker and use the gated backend smoke as the closeout evidence.
