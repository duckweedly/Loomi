# Contract: runtime.todo_write

## Tool

| Tool | Approval | Capability | Risk | Side Effect |
| --- | --- | --- | --- | --- |
| `runtime.todo_write` | Required | `plan` | `low` | `none` |

## Arguments

```json
{
  "items": [
    { "title": "Read failing tests", "status": "completed" },
    { "title": "Implement minimal tool execution", "status": "in_progress" }
  ]
}
```

Rules:

- `items` is required.
- `items` length is 1-20.
- every item title is trimmed, non-empty, and at most 160 characters.
- status defaults to `pending`.
- status must be `pending`, `in_progress`, or `completed`.

## Result

The worker returns counts and sanitized items through `result_summary`. No separate endpoint or database table is added.
