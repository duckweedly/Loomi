---
title: Local M22 Bounded Agent Loop Validation
description: Local validation for the first bounded workspace tool continuation loop.
---

## Focused Smoke

```bash
go test ./internal/httpapi -run TestM22BoundedAgentLoopWorkspaceSmoke
go test ./internal/httpapi -run TestM22CodeAgentReadEditExecReadLoopSmoke
go test ./internal/runtime -run 'TestGatewayContinuation|TestWorkerExecutesApprovedTodoWrite|TestExecuteTodoWrite|TestAppend.*Todo'
```

Expected backend evidence:

1. Work mode run starts through the normal model gateway queue.
2. Provider requests `workspace.glob`; Loomi records approval-required and does not execute before approval.
3. HTTP approval queues the worker resume job.
4. Worker executes the first tool through ToolBroker and records `tool_call_succeeded`.
5. Continuation provider requests `workspace.read`; Loomi records a second approval-required tool call instead of `unsupported_tool_loop`.
6. Second approval queues another worker resume job.
7. Worker executes `workspace.read`, continues the provider, and persists one final assistant message.
8. Run events include continuation `model_phase`, both tool call ids, `loop_index`, `loop_max`, and no fixture secrets or workspace root paths.
9. Code-agent smoke covers `workspace.read -> workspace.edit -> sandbox.exec_command -> workspace.read -> final` through the same HTTP approval and worker resume path.
10. Provider request tool schemas include enabled workspace mutation, bounded command, LSP, web, and todo tools using provider-safe names such as `workspace_edit`, `sandbox_exec_command`, and `todo_write`.
11. Tool request and tool success paths append durable runtime-derived `work.todo.updated` snapshots for Work threads, and approved `todo.write` appends a provider-maintained `work.todo.updated` snapshot.
12. WorkPlanView can replay real task state from those events without seed/mock metadata.

## Boundary Checks

Runtime tests cover:

- loop limit failure with `tool_loop_limit_reached`
- repeated continuation `tool_call_id` failure with `duplicate_tool_call_id`
- workspace tool outside the run enabled-tool snapshot failing before approval
- one non-terminal tool-call invariant
- enabled runtime tools can participate in continuation when present in the run snapshot
- unsupported continuation still failing with `unsupported_tool_loop`
- bounded command and LSP tools can participate in continuation when they are enabled for the run
- safe todo metadata normalization, runtime todo snapshot emission, provider `todo.write` snapshot emission, and Work-mode-only replay projection
- RunRail loop index, continuation, loop-limit, stopped state, and Stop action while blocked on approval
- stopped run semantics when a tool was approved and claimed but the run stops before execution

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```
