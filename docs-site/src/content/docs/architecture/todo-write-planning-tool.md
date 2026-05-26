---
title: Todo Write Planning Tool
description: M12 approval-gated structured planning tool.
---

M12 adds `runtime.todo_write`, a low-risk runtime tool that lets a model publish a bounded structured plan through the existing tool-call lifecycle.

## Boundary

`runtime.todo_write` is not a workspace file tool and does not execute commands. It writes no separate durable todo state. The plan is recorded as normal tool-call arguments and result summaries.

The tool accepts 1-20 items. Each item has a trimmed title and a status of `pending`, `in_progress`, or `completed`.

## Flow

1. Provider emits `runtime.todo_write` with `items`.
2. Product data validates the allowlist, approval requirement, bounded item count, titles, and statuses.
3. Loomi records `tool.call.requested` and `tool.call.approval_required`.
4. User approves or denies the request through the existing approval UI.
5. Worker executes the approved tool.
6. Runtime returns item summaries and status counts.
7. Tool lifecycle events and ToolCallCard render the result in the timeline.

## Safety

The tool remains approval-required for consistency with M7-M11. It has no file, process, provider, or network side effects, and it rejects unknown fields.

## Non-Goals

M12 does not add editable task management, a todo database table, auto-approval, MCP, LSP, spawn-agent, or multi-agent delegation.
