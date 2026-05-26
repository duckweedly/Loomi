---
title: Workspace Mutation Tools API
description: Catalog and event contracts for M23 workspace mutation tools.
---

M23 introduces write-capable workspace tool catalog entries. They use the existing tool-call approval and run-event APIs; no separate mutation endpoint is added.

## Catalog

`workspace.write_file`:

```json
{
  "name": "workspace.write_file",
  "source": "builtin",
  "group": "workspace",
  "risk_level": "high",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "workspace",
    "read_only": false,
    "write_capable": true,
    "arguments": ["path", "content", "max_bytes"]
  }
}
```

`workspace.edit`:

```json
{
  "name": "workspace.edit",
  "source": "builtin",
  "group": "workspace",
  "risk_level": "high",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "workspace",
    "read_only": false,
    "write_capable": true,
    "arguments": ["path", "old_text", "new_text", "max_bytes"]
  }
}
```

## Result Summary

Successful `workspace.write_file` execution records safe result metadata:

```json
{
  "tool": "workspace.write_file",
  "scope": "workspace",
  "operation": "write_file",
  "path": "src/generated.txt",
  "changed": true,
  "bytes_written": 12,
  "line_count_after": 2,
  "truncated": false,
  "redaction_applied": false
}
```

Successful `workspace.edit` execution records safe result metadata:

```json
{
  "tool": "workspace.edit",
  "scope": "workspace",
  "operation": "edit",
  "path": "src/notes.txt",
  "changed": true,
  "bytes_before": 19,
  "bytes_after": 20,
  "line_count_before": 2,
  "line_count_after": 2,
  "truncated": false,
  "redaction_applied": false
}
```

Tool-call events expose mutation argument previews instead of raw content:

```json
{
  "arguments_summary": {
    "path": "src/generated.txt",
    "content_bytes": 8,
    "content_provided": true
  }
}
```

Normal API/UI event metadata must not include raw file content, host absolute workspace roots, credentials, shell output, provider raw payloads, or hidden local state.

## Failure Behavior

Validation failures leave files unchanged and surface through the existing `tool_call_failed` and terminal run failure path. Typical causes include existing target, path escape, sensitive target, invalid UTF-8, oversized content, missing edit match, and duplicate edit match.
