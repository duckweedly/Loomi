---
title: 2026-05-27 M80 Durable Run Resume
description: Worker restart recovery for succeeded tool calls before provider continuation.
---

Implemented:

- Added failing-first runtime coverage for a worker retry after `tool_call_succeeded` was persisted but before provider continuation started.
- Added continuation reconstruction coverage proving pending `tool_call_requested` entries do not enter provider input.
- Preserved approval lifecycle order across resume: `approval_required -> approved -> executing -> succeeded`.
- Reused existing terminal run guards for late model/tool/approval events and kept the M76 multi-tool final assistant single-write smoke as the broad loop regression.
- Changed queued run resume so an already-succeeded tool call can resume missing continuation from durable run events, while already-started continuation, later tool requests, and terminal runs remain no-op boundaries.

Arkloop benchmark observation:

- Arkloop reconstructs replay state from durable rollout items, tracks pending tool calls separately, and syncs terminal run status from terminal events during lifecycle bootstrap.
- Loomi keeps its own simpler event/projection shape: `tool_calls` is the current-state projection, and run events are the ordered audit/replay source.

Boundaries:

- No new fields, migrations, providers, batch API, Redis, external queue, terminal shell runtime, or multi-agent behavior.
- Existing `terminal run cannot accept new events` and tool-call state guards remain the boundary for late provider/tool/approval writes.

Focused validation:

```bash
go test ./internal/runtime -run 'TestGatewayContinuationSkipsPendingToolCalls|TestQueuedRunRouterResumesContinuationAfterSucceededToolRestart' -count=1
```
