---
title: Code-Agent Daily Loop
description: Work-mode code task loop over existing workspace, patch, sandbox, approval, worker, and run-event paths.
---

M75 makes the Work-mode code-agent loop dogfoodable without adding new tools. A provider can now drive the existing chain:

```text
workspace.grep -> workspace.read -> workspace.patch_preview -> workspace.patch_apply -> sandbox.exec_command -> assistant summary
```

The loop stays on the normal Gateway, ToolBroker, worker, approval, continuation, and run-event path. Read-only tools may auto-approve; review/write/command steps remain approval-gated.

## Observable Contract

The timeline proves the loop finished when one run contains, in order:

- `tool_call_requested` / `tool_call_succeeded` for `workspace.grep`.
- `tool_call_requested` / `tool_call_succeeded` for `workspace.read`.
- `tool_call_approval_required` then `tool_call_approved` / `tool_call_executing` / `tool_call_succeeded` for `workspace.patch_preview`.
- `tool_call_approval_required` then execution/success for `workspace.patch_apply`.
- `tool_call_approval_required` then execution/success for `sandbox.exec_command`.
- continuation `model_request_started` / `model_output_completed`.
- final `run_completed` and an assistant message.

`loop_index` / `loop_max` marks each accepted tool call. The provider continuation receives the prior tool result through the existing continuation projection; no side channel is introduced.

## Safety Boundaries

`workspace.patch_preview` requires a same-run read and returns `changed=false` plus compact diff metadata. `workspace.patch_apply` requires the matching fresh preview and writes only after approval. Pending mutation approvals on terminal runs are rejected and do not resume execution. `sandbox.exec_command` stays argv-only, workspace-scoped, bounded, and approval-required.

Events and UI summaries redact paths, raw file content, stdout/stderr bodies, and patch snippets where needed while still showing safe state such as operation, changed status, exit code, truncation, and loop count.

## M76 Continuation Reliability

M76 keeps the same tool surface and raises no new orchestration layer. The reliability change is at the Gateway/worker boundary: after each successful tool execution, the next provider request receives the ordered prefix of all successful tool call/result pairs in the current run, ending at the just-finished `tool_call_id`.

That means a six-step run such as:

```text
workspace.grep -> workspace.read -> workspace.patch_preview -> workspace.patch_apply -> sandbox.exec_command -> workspace.read -> final
```

projects continuation context as:

```text
after grep: grep result
after read: grep result, read result
after preview: grep result, read result, preview result
after apply: grep result, read result, preview result, apply result
after exec: grep result, read result, preview result, apply result, exec result
after final read: all six results
```

Each tool call still has exactly one terminal state (`succeeded`, `failed`, `denied`, or `cancelled`). A new tool call cannot be recorded while another call is blocked, not started, or executing, so old approvals cannot leak into `PendingApprovals`. Terminal runs reject late tool requests and late model output events.

## Arkloop Benchmark Observation

Arkloop separates provider input replay from the UI stream: worker pipeline code rebuilds assistant tool-call messages and tool-result messages from durable rollout/run state, while thread run state and resume helpers prevent unfinished tool calls from being replayed as valid continuation context. Loomi does not copy Arkloop naming, UI contract, or private pipeline shape. This slice takes the mechanism lesson only: continuation should be rebuilt from durable ordered events, not from a single transient tool result.
