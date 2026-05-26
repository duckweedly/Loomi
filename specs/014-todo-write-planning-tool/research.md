# Research: M12 Todo Write Planning Tool

## Decision: Runtime Tool, Not Workspace Tool

`runtime.todo_write` is internal planning metadata. It does not read or mutate workspace files and does not execute processes.

## Decision: Approval Required

Current Loomi tool lifecycle requires tool calls to start blocked on approval. M12 keeps that behavior so planning updates remain auditable and consistent with M7-M11.

## Decision: No Durable Todo Table

The MVP stores todo output in the existing tool call result summary. Durable cross-run todo management would need separate product semantics and is out of scope.

## Decision: Bounded Items

The tool accepts 1-20 items. Titles are trimmed and capped at 160 characters. Status is limited to `pending`, `in_progress`, and `completed`.
