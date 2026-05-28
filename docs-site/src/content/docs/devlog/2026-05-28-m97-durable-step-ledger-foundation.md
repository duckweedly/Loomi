---
title: 2026-05-28 M97 Durable Step Ledger Foundation
description: Loomi-native step projection for durable run, tool, approval, continuation, and terminal state.
---

Implemented:

- Added a Loomi-native run step projection over existing durable `run_events`.
- Added safe step metadata to known model, tool, approval, continuation, and terminal events.
- Added `RebuildRunStepState` so resume logic can separate completed tool results from pending tool calls and derive the next action.
- Updated worker retry continuation recovery to use the rebuilt step state instead of ad-hoc event suffix scanning.
- Preserved M80 behavior: already-succeeded tools are not executed again, pending tool calls are not serialized into provider continuation, later continuation/tool/terminal events suppress duplicate resume.

Arkloop comparison:

- Arkloop persists rollout JSONL items and reconstructs assistant/tool replay state plus pending tool calls from those items.
- Loomi keeps its own run/event/tool-call model. This slice adopts the mechanism, not Arkloop naming or private structure: persisted run events remain the source, `tool_calls` remains the current projection, and `RunStep` is a safe Loomi projection for orchestration decisions.

Boundaries:

- No Redis, external queue, worker pool rewrite, browser real tier, child agent execution, Docker/Firecracker, or run/event schema rename.
- No UI theme changes.
- No new provider protocol or raw tool-message persistence.

Focused validation:

```bash
go test ./internal/productdata ./internal/runtime -run 'TestRunStepLedger|TestQueuedRunRouterResumesContinuationAfterSucceededToolRestart|TestGatewayContinuationSkipsPendingToolCalls' -count=1
```
