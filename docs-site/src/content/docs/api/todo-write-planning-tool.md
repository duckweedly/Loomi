---
title: Todo Write Planning Tool
description: M12 runtime.todo_write tool contract.
---

## Tool Name

| Tool | Approval | Capability | Risk | Side Effect |
| --- | --- | --- | --- | --- |
| `runtime.todo_write` | Required | `plan` | `low` | `none` |

## Arguments

```json
{
  "items": [
    { "title": "Inspect runtime registry", "status": "completed" },
    { "title": "Add todo_write tests", "status": "in_progress" },
    { "title": "Run validation" }
  ]
}
```

Rules:

- `items` is required.
- `items` must contain 1-20 objects.
- `title` is required, trimmed, and capped at 160 characters.
- `status` defaults to `pending`.
- `status` must be `pending`, `in_progress`, or `completed`.

## Result

```json
{
  "total": 3,
  "pending_count": 1,
  "in_progress_count": 1,
  "completed_count": 1,
  "items": [
    { "title": "Inspect runtime registry", "status": "completed" },
    { "title": "Add todo_write tests", "status": "in_progress" },
    { "title": "Run validation", "status": "pending" }
  ]
}
```

The result is stored in the existing tool-call `result_summary` projection.
