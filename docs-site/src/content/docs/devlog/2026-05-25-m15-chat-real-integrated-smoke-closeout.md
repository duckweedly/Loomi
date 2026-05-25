---
title: M15 Chat Real Integrated Smoke Closeout
description: Deterministic real Chat smoke evidence across memory, MCP approval execution, continuation, and replay.
---

M15 is a closeout/evidence slice for Chat's real integrated path. It does not add new platform features; it proves the existing M7/M9/M11/M12/M13/M14 pieces can run together through real API/service/worker boundaries.

## Completed In This Candidate

- Created `specs/022-chat-real-integrated-smoke-closeout/` with spec, plan, research, data model, contract, quickstart, tasks, and requirements checklist.
- Added gated backend smoke `TestM15ChatRealIntegratedSmoke`.
- The smoke creates the thread, user message, and model-gateway run through HTTP handlers.
- It prepares an approved thread-scoped memory entry, then verifies `RunContext.MemorySnapshot` through `memory_snapshot_loaded` and pipeline safe summary metadata.
- It discovers a deterministic local stdio MCP fixture and records `mcp_discovery_succeeded` with a stable `candidate_schema_hashes` entry.
- It uses a deterministic provider fixture that first requests the discovered persona-allowed MCP tool and later emits one final assistant message during continuation.
- The first worker pass blocks on `tool_call_approval_required`; no MCP `tools/call` runs before approval.
- The smoke approves through `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/approve`.
- The second worker pass executes exactly one local stdio MCP `tools/call`, records redacted `tool_call_succeeded`, performs one provider continuation, writes the assistant message, and completes the run.
- Replay evidence is fetched through `GET /v1/runs/{run_id}/events`.

## Evidence Chain

Expected event sequence includes:

- `run_queued`
- `job_claimed`
- `pipeline_step_started` / `pipeline_step_completed` for `prepare_context`
- `memory_snapshot_loaded`
- `mcp_discovery_succeeded`
- `tool_call_requested`
- `tool_call_approval_required`
- `tool_call_approved`
- `tool_call_executing`
- `tool_call_succeeded`
- continuation `model_request_started`
- continuation `model_output_completed`
- `run_completed`

The gated smoke logs the run id, tool-call id, final assistant message, and replay event types for local closeout evidence.

## Safety

The fixture injects sensitive-looking provider and MCP values, including token-shaped values, authorization-shaped values, stderr-like text, and local-path-shaped values. The smoke fails if those canaries appear in:

- run creation or fetch API responses
- event replay API responses
- messages API responses
- RunContext safe summary metadata
- tool result summary
- provider continuation request
- M15 docs examples

Docs examples use synthetic safe placeholders only.

## Validation

Required validation for this candidate:

```bash
LOOMI_M15_REAL_CHAT_SMOKE=1 go test ./internal/httpapi -run TestM15ChatRealIntegratedSmoke -count=1 -v
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Browser Smoke Status

The authoritative M15 evidence is the gated backend smoke because it exercises the real in-process HTTP handlers, productdata service, worker, MCP stdio subprocess fixture, provider fixture, continuation, and replay API without external model spend.

Browser smoke should be run only when the local API and web shell can be started together in real API mode. If it is not run, record the blocker and use the backend smoke output as equivalent evidence for this closeout slice.
