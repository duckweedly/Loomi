---
title: Workspace Write Tools
description: M9 workspace write and exact edit tool contracts.
---

M9 does not add a new public file API. Workspace writes enter through the existing tool-call request, approval, worker execution, run-event history, and SSE contracts.

## Tool Names

| Tool | Approval | Effect |
| --- | --- | --- |
| `workspace.write_file` | Required | Write bounded UTF-8 text |
| `workspace.edit` | Required | Replace one exact text match |

## Arguments

`workspace.write_file`:

```json
{
  "path": "internal/generated.txt",
  "content": "hello Loomi\n"
}
```

`workspace.edit`:

```json
{
  "path": "internal/generated.txt",
  "old_text": "hello",
  "new_text": "hello, safe"
}
```

## Results

`workspace.write_file`:

```json
{
  "path": "internal/generated.txt",
  "bytes_written": 12,
  "created": true,
  "truncated": false
}
```

`workspace.edit`:

```json
{
  "path": "internal/generated.txt",
  "replacements": 1,
  "bytes_before": 12,
  "bytes_after": 18
}
```

## Failure Contract

Unsafe validation and exact-edit failures are reported through `tool_call_failed` with redacted metadata. Missing-match and duplicate-match edit failures leave the file unchanged.
