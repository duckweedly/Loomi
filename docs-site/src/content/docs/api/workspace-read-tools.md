---
title: Workspace Read Tools
description: M8 workspace read tool contracts and result payloads.
---

M8 does not add a new public file API. Workspace reads enter through the existing M7 tool-call request, approval, worker execution, run-event history, and SSE contracts.

## Tool Names

| Tool | Approval | Side effects |
| --- | --- | --- |
| `workspace.glob` | Required | Read-only |
| `workspace.grep` | Required | Read-only |
| `workspace.read_file` | Required | Read-only |

## Arguments

`workspace.glob`:

```json
{ "pattern": "**/*.go", "limit": 50 }
```

`workspace.grep`:

```json
{ "query": "ToolDefinition", "path": "internal/runtime", "limit": 50 }
```

`workspace.read_file`:

```json
{ "path": "internal/runtime/tools.go", "max_bytes": 4096 }
```

## Results

`workspace.glob`:

```json
{
  "matches": ["internal/runtime/tools.go"],
  "match_count": 1,
  "truncated": false
}
```

`workspace.grep`:

```json
{
  "matches": [
    { "path": "internal/runtime/tools.go", "line": 24, "preview": "type ToolDefinition struct {" }
  ],
  "match_count": 1,
  "truncated": false
}
```

`workspace.read_file`:

```json
{
  "path": "internal/runtime/tools.go",
  "size_bytes": 4096,
  "preview": "package runtime\n",
  "truncated": true
}
```

## Failure Contract

Unsafe inputs fail through `tool_call_failed` with redacted error metadata. Sensitive file contents are not read into event metadata.
