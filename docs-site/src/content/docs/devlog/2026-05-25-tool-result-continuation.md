---
title: 2026-05-25 Tool Result Continuation
description: Implementation notes for the minimal M7 tool-result-to-model continuation boundary.
---

## Completed

- Added provider-neutral continuation context roles: `assistant_tool_call` and `tool_result`.
- OpenAI-compatible provider requests now serialize tool-result continuation as native assistant `tool_calls` plus matching `tool` result messages.
- Runtime can build continuation context from thread messages and persisted `tool_call_requested` / `tool_call_succeeded` events.
- Runtime can make one continuation provider request, persist second-phase model deltas with `model_phase = continuation`, persist one final assistant message, and complete the run.
- Continuation provider requests for another tool fail safely with `unsupported_tool_loop`.
- Frontend runtime replay pauses pre-tool drafts after `tool.call.succeeded` and uses continuation deltas as the final assistant draft.

## Validation

Targeted validation:

```bash
go test ./internal/runtime
go test ./internal/productdata
bun test src/runtime/realExecutionAdapter.test.ts
```

## Current limit

This slice does not add approve/deny endpoints or approved tool execution. It is ready to be called after the M7 execution slice records a redacted `tool_call_succeeded` result.
