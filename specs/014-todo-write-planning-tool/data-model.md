# Data Model: M12 Todo Write Planning Tool

## Todo Item

- `title`: required string, trimmed, 1-160 characters
- `status`: optional string, defaults to `pending`

Allowed statuses:

- `pending`
- `in_progress`
- `completed`

## Tool Arguments

```json
{
  "items": [
    { "title": "Inspect current runtime tools", "status": "completed" },
    { "title": "Add todo_write tests", "status": "in_progress" }
  ]
}
```

## Tool Result

```json
{
  "total": 2,
  "pending_count": 0,
  "in_progress_count": 1,
  "completed_count": 1,
  "items": [
    { "title": "Inspect current runtime tools", "status": "completed" },
    { "title": "Add todo_write tests", "status": "in_progress" }
  ]
}
```
