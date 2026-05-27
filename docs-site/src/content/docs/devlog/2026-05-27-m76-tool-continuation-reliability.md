---
title: 2026-05-27 M76 Tool Continuation Reliability
description: Stable six-tool continuation ordering, approval cleanup, and terminal-run guards.
---

Implemented:

- Added a failing-first HTTP smoke for a six-step code-agent run: `workspace.grep`, `workspace.read`, `workspace.patch_preview`, `workspace.patch_apply`, `sandbox.exec_command`, `workspace.read`, then final assistant message.
- Changed Gateway continuation projection to replay the ordered prefix of successful tool call/result pairs for the current run, not only the just-finished result.
- Covered one terminal tool state per `tool_call_id`, current-only pending approval state, ordered provider continuation messages, single final assistant message, and rejection of late terminal-run output/tool requests.

Arkloop benchmark observation:

- Arkloop rebuilds provider replay input from durable tool-call and tool-result state, and its resume/thread-run-state path avoids treating unfinished tool calls as valid continuation context.
- Loomi keeps its own event names, UI copy, and runtime shape. This slice only adopts the mechanism: durable ordered continuation reconstruction at the worker/Gateway boundary.

Focused validation:

```bash
go test ./internal/httpapi -run TestM76ToolContinuationReliabilitySmoke -count=1
go test ./internal/runtime ./internal/httpapi -run 'Test.*Continuation|Test.*Bounded|Test.*CodeAgent|Test.*Tool' -count=1
```
