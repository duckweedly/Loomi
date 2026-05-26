---
title: 2026-05-25 M22 Bounded Agent Loop
description: First bounded workspace continuation loop slice.
---

M22 implementation has started with the US1 backend loop path.

Implemented:

- Continuation can request a later enabled Work-mode `workspace.*` read tool after the previous tool call reaches terminal execution.
- Every later workspace tool call is persisted as approval-required and remains blocked until explicit approval.
- The run keeps a hard limit of three accepted tool calls.
- Over-limit continuation fails with `tool_loop_limit_reached` without recording the extra tool call.
- Repeated continuation `tool_call_id` fails with `duplicate_tool_call_id` without duplicating approval events.
- Existing non-workspace continuation remains guarded by `unsupported_tool_loop`.
- In-memory and PostgreSQL tool-call request paths now permit a later tool call only when existing calls are terminal.
- Tool lifecycle events include `loop_index` and `loop_max` metadata for replay.
- HTTP smoke covers `workspace.glob -> approve -> result -> workspace.read -> approve -> result -> final`.
- Work todo metadata now normalizes to safe `todo_items`, bounded item fields, stable statuses, `updated_by`, and `redaction_applied`.
- Work mode projection and WorkPlanView replay/render todo snapshots while keeping Chat mode isolated.
- RunRail shows bounded loop position, continuation phase, loop-limit failure, stopped state, and exposes Stop while blocked on tool approval.
- Queued worker resume exits cleanly when a run is stopped between tool approval and execution; it does not execute the approved tool, continue the provider, or write failure events over the stopped terminal state.

Validation run:

```bash
go test ./internal/httpapi -run TestM22BoundedAgentLoopWorkspaceSmoke
go test ./internal/runtime ./internal/productdata ./internal/httpapi
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke passed on `http://127.0.0.1:5173/`: Work Plan showed todo replay, RunRail showed continuation plus two workspace tool timeline rows, and console errors were zero.
