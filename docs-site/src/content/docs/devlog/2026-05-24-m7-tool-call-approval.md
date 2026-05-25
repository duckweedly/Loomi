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

## Current limitations

Remaining M7 work includes approve/deny endpoints, enabled approval UI actions, worker execution of approved `runtime.get_current_time`, result/error/cancel terminal states, and final documentation updates after the full milestone is complete.

The default test run does not exercise the Postgres repository path unless a local integration database environment is configured.
