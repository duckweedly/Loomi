---
title: 2026-05-25 M7 approval execution
description: Minimal safe approve/deny and current-time tool execution closure.
---

## Completed

- Created Spec Kit feature `011-tool-approval-execution`.
- Added idempotent approve and deny product-data transitions.
- Added scoped approve and deny HTTP endpoints under thread/run/tool-call paths.
- Deny writes `tool_call_denied`, stops the MVP run, and cancels pending run jobs.
- Approve writes `tool_call_approved`, moves the run back to queued, and schedules one resume job.
- Worker resume jobs with `tool_call_id` execute only `runtime.get_current_time`.
- Execution writes `tool_call_executing` before invocation and then `tool_call_succeeded` or `tool_call_failed`.
- Result and error summaries are redacted before projection and event persistence.
- Frontend real API client maps approved, denied, executing, succeeded, and failed states.
- ToolCallCard approve/deny controls call real actions and expose disabled/loading/error states.
- Active run tool calls render in ChatCanvas and replay through existing SSE state updates.

## Safety boundaries

M7 still allows only `runtime.get_current_time` with omitted timezone or `UTC`. It does not add shell, filesystem, arbitrary network, MCP, browser automation, multi-tool concurrency, multi-agent loops, memory/RAG, or approval bypass.

## Validation

Completed in this work session:

- `go test ./internal/productdata ./internal/runtime ./internal/db ./internal/httpapi ./cmd/...` passed.
- `bun test --cwd web` passed.
- `bun run --cwd web build` passed.
- `bun run --cwd docs-site build` passed.
- Browser smoke opened the local web shell and verified ToolCallCard rendering in the app. A real-API browser smoke fixture exposed the history replay gap fixed in this slice; the in-app browser environment did not provide `window.fetch`, so approve/deny click-through could not be completed there. The approve/deny execution paths are covered by backend, adapter, and component tests; repeat full click-through in a normal browser or once the in-app browser fetch surface is available.
