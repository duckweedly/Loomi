---
title: 2026-05-25 M22 Bounded Agent Loop
description: First bounded workspace continuation loop slice.
---

M22 implementation has started with the US1 backend loop path.

Implemented:

- Continuation can request later enabled Work-mode workspace, bounded command, LSP, web, browser, artifact, coordination, or todo tools after the previous tool call reaches terminal execution.
- Every later tool call is persisted as approval-required and remains blocked until explicit approval.
- The run keeps a hard limit of 24 accepted tool calls, enough for project survey/source-reading runs while still preventing unbounded autonomy.
- Over-limit continuation fails with `tool_loop_limit_reached` without recording the extra tool call.
- Repeated continuation `tool_call_id` fails with `duplicate_tool_call_id` without duplicating approval events.
- Unsupported or disabled continuation remains guarded by `unsupported_tool_loop`.
- Provider requests now receive schema definitions for enabled builtin code-agent tools, with provider-safe function names such as `workspace_read`, `workspace_edit`, `sandbox_exec_command`, `lsp_symbols`, and `todo_write`.
- In-memory and PostgreSQL tool-call request paths now permit a later tool call only when existing calls are terminal.
- Tool lifecycle events include `loop_index` and `loop_max` metadata for replay.
- HTTP smoke covers `workspace.glob -> approve -> result -> workspace.read -> approve -> result -> final`.
- Code-agent smoke covers `workspace.read -> workspace.edit -> sandbox.exec_command -> workspace.read -> final`.
- Work todo metadata now normalizes to safe `todo_items`, bounded item fields, stable statuses, `updated_by`, and `redaction_applied`.
- Runtime now appends durable `work.todo.updated` snapshots after tool requests and successful tool executions in Work threads.
- Approved `todo.write` calls now let the provider replace the Work todo snapshot with bounded normalized todo items, persisted as `work.todo.updated` with `updated_by=provider`.
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
