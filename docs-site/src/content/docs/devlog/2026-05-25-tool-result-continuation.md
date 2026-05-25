---
title: 2026-05-25 Tool Result Continuation
description: Implementation notes for the minimal M7 tool-result-to-model continuation boundary.
---

## Completed

- Added provider-neutral continuation context roles: `assistant_tool_call` and `tool_result`.
- OpenAI-compatible provider requests now serialize tool-result continuation as native assistant `tool_calls` plus matching `tool` result messages.
- Runtime can build continuation context from thread messages and persisted `tool_call_requested` / `tool_call_succeeded` events.
- Approved `runtime.get_current_time` worker execution now resumes the provider continuation automatically after `tool_call_succeeded`.
- Runtime can make one continuation provider request, persist second-phase model deltas with `model_phase = continuation`, persist one final assistant message, and complete the run.
- Denied and `tool_call_failed` paths are terminal and do not call continuation.
- Continuation provider failures are redacted, terminal, and do not create a duplicate final assistant message.
- Continuation provider requests for another tool fail safely with `unsupported_tool_loop`.
- Frontend runtime replay pauses pre-tool drafts after `tool.call.succeeded`, uses continuation deltas as the final assistant draft, keeps denied/failed paths non-final, and renders the tool result in ToolCallCard/Timeline/RunRail.

## Validation

Targeted validation:

```bash
go test ./internal/productdata ./internal/runtime ./internal/db ./internal/httpapi ./cmd/...
bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/executionAdapter.test.ts ./web/src/components/ToolCallCard.test.tsx ./web/src/components/RunRail.runtime.test.ts ./web/src/components/RunTimeline.runtime.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

Local browser smoke used a local OpenAI-compatible fake provider:

- success: approval-required -> approve -> executing -> succeeded -> continuation deltas -> one `run_completed` final and one assistant message
- denied: approval-required -> deny -> `run_stopped`, zero continuation events
- provider failure: approved tool succeeded, continuation provider failure recorded one redacted `run_failed`, zero `model_output_completed` final assistant events

Final closeout smoke on 2026-05-25 used the local API on `127.0.0.1:18080`, the Vite web shell on `127.0.0.1:5173`, Postgres schema version 6, and a local OpenAI-compatible fake provider on `127.0.0.1:19091`.

- success: browser sent a model-gateway message, showed `runtime.get_current_time` as `approval_required`, Approve advanced it through `tool_call_approved`, `tool_call_executing`, `tool_call_succeeded`, continuation `model_output_delta`, `model_output_completed`, and `run_completed`.
- persisted events: the successful run recorded 16 events including initial `model_request_started`, `tool_call_requested`, `tool_call_approval_required`, continuation `model_request_started` with `model_phase = continuation`, one final assistant message, and one `run_completed`.
- browser console: no warnings; only expected 404s for `runs/current` on a newly created empty thread before a run exists.
- observed UI limitation: the top capability chip can remain `Provider unavailable` after a successful model-gateway run because generic capability-signal derivation treats `model_request_started` metadata text containing `provider` as a provider-unavailable signal. The run itself completed successfully.

## Current limit

This slice still intentionally supports one approved tool call and one continuation provider call per run. It does not add shell/filesystem/MCP/browser automation tools, multi-agent loops, long-term memory, or RAG.
