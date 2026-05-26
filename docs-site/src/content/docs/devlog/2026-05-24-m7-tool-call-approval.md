---
title: 2026-05-24 M7 Tool Call Approval Foundation Devlog
description: Phase 2 implementation notes, validation results, and limitations for M7 tool-call approval core.
---

## Completed scope

M7 Phase 2 foundation records approval-gated internal tool requests without executing broad tools. The US1 observable request slice now routes allowlisted provider `runtime.get_current_time` requests through that boundary and blocks the run on approval.

Implemented slice:

- migration `000006_m7_tool_call_approval` with `blocked_on_tool_approval` run state and `tool_calls`
- schema readiness target updated to version `6`
- backend tool-call approval/execution statuses and lifecycle event constants
- scoped `GetToolCall` and idempotent `RecordToolCallRequest` service/repository methods
- one-tool-per-run MVP guard
- terminal-run guard for tool request recording
- allowlisted `runtime.get_current_time` definition and executor
- strict timezone schema: omitted or `UTC` only
- unknown tool argument rejection
- key- and value-based redaction before metadata persistence
- worker diagnostics counters for blocked and resumable tool calls
- history-first stream tests for M7 tool event ordering
- frontend domain types for tool-call lifecycle, statuses, and metadata
- frontend API mapping for backend `tool_call_*` event names
- frontend runtime replay of M7 tool events into a stable `ToolCall` view model
- RunRail event grouping support for `tool-call`
- provider gateway conversion for allowlisted `runtime.get_current_time` tool requests
- non-executing approval-required run blocking for provider tool requests
- scoped tool-call read handler and real API client mapping
- ToolCallCard approval-required placeholder with disabled controls
- AgentStateMotion confirm state for `blocked_on_tool_approval`
- local desktop Settings save path for the OpenAI-compatible `custom` provider, with redacted capability responses and in-process gateway refresh
- provider Settings UI copy and card styling aligned with the desktop shell instead of draft-only placeholder language
- idempotent `ApproveToolCall` and `DenyToolCall` service/repository transitions
- scoped `POST /tool-calls/{tool_call_id}/approve` and `/deny` HTTP actions
- `tool_call_approved`, `tool_call_denied`, and denial `run_stopped` event recording
- frontend real API approve/deny methods and enabled ToolCallCard controls when handlers are provided
- frontend mapping for approved and denied tool-call states
- approval wake-up through the existing M6 worker/job pipeline with one resumable job per approved tool call
- approved `runtime.get_current_time` execution through the M7 internal tool executor
- `tool_call_executing`, `tool_call_succeeded`, `tool_call_failed`, `run_completed`, and `run_failed` lifecycle recording
- redacted result and error summaries on terminal tool-call projections
- frontend runtime replay for executing, succeeded, failed, and cancelled tool-call states
- ToolCallCard rendering for executing, result, redacted error, denied, and cancelled states
- cancellation precedence for pending, approved, and executing tool calls through `StopRun`
- `tool_call_cancelled` projection and event publishing before `run_stopped`
- worker guard that skips already-cancelled tool jobs without writing duplicate success/failure terminal events
- RunRail tool lifecycle summaries and default grouping for `tool.call.*` events without requiring explicit backend group metadata
- mixed model/tool/final stream ordering regression coverage

## Safety notes

This slice does not add shell tools, filesystem tools, MCP, browser automation, arbitrary network tools, multi-agent behavior, long-term memory/RAG, or approval bypass. The only executable tool definition is `runtime.get_current_time`, and it is still approval-required for M7 smoke coverage.

Tool summaries and event metadata are redacted before persistence. Sensitive metadata keys such as `api_key`, `authorization`, `password`, `secret`, `token`, and `credential` are always replaced with `[redacted]`.

## Validation log

Validated during Phase 2 implementation:

```bash
go test ./internal/productdata ./internal/runtime ./internal/db ./internal/httpapi ./cmd/...
bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/executionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts
bun run --cwd web build
```

Latest recorded results before docs build:

- backend and command tests passed for the listed packages
- frontend runtime/API/state/component tests passed with 79 tests and 267 expectations
- `bun run --cwd web build` passed

Latest US2 decision-slice validation:

```bash
go test ./...
bun test web/src
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Results:

- Go backend and command tests passed across all packages
- frontend tests passed with 202 tests and 624 expectations
- web production build passed
- docs-site build passed with 45 pages
- whitespace diff check passed

Latest US3 execution-slice validation:

```bash
go test ./...
bun test web/src
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Results:

- Go backend and command tests passed across all packages
- frontend tests passed with 206 tests and 645 expectations
- web production build passed
- docs-site build passed with 45 pages
- whitespace diff check passed

Latest cancellation and US4 polish validation:

```bash
go test ./internal/productdata ./internal/runtime -run 'TestStopRunCancelsPendingApprovedAndExecutingToolCalls|TestWorkerSkipsCancelledApprovedToolCall|TestMergeHistoryThenLiveKeepsMixedModelToolAndFinalOrder'
go test ./internal/httpapi -run 'TestStopRunHandlerCancelsPendingToolCall|TestStopRunHandlerPublishesStopEvents'
bun test web/src/runtime/runtimeEventGroups.test.ts web/src/components/RunRail.polish.test.ts
```

Results:

- pending, approved, and executing tool calls are cancelled exactly once on stop
- cancelled tool calls reject later success/failure overwrite attempts without appending duplicate terminal events
- HTTP stop publishes `tool_call_cancelled` before stop events and leaves the scoped projection cancelled
- mixed model/tool/final replay preserves event order and skips duplicate live replay
- RunRail groups ungrouped `tool.call.*` events under Tool call and labels tool lifecycle rows distinctly

Final M7 local validation:

```bash
go test ./...
bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Results:

- Go backend and command tests passed across all packages
- frontend tests passed with 207 tests and 652 expectations
- web production build passed
- docs-site build passed with 45 pages
- whitespace diff check passed

Browser smoke:

- Codex in-app browser was unavailable in this session, so local smoke used Playwright against `http://127.0.0.1:5173/`.
- Verified the actual `ToolCallCard` module rendered approval-required, executing, succeeded, failed, denied, and cancelled states in the Vite page.
- Verified approval and deny buttons invoked their handlers.
- Verified no browser console errors after the successful smoke pass.
- Final local Vite smoke used Playwright against `http://127.0.0.1:5174/`, captured `m7-runrail-smoke.png`, and current-page console errors were zero.

## Current limitations

M7 implementation is now complete at the local test/documentation level. A final full validation pass and browser smoke should be rerun after any follow-up changes before moving to the next milestone.

The US3 MVP stops at a terminal run after the tool result. It intentionally does not yet feed the tool result back into a multi-step model continuation loop.

The default test run does not exercise the Postgres repository path unless a local integration database environment is configured.
