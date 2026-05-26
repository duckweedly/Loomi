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
    "requires_read_before_edit": true,
    "returns_diff": true,
    "normalizes_line_endings": true,
    "preserves_indentation": true,
    "strips_trailing_whitespace_except_markdown": true,
    "arguments": ["path", "old_text", "new_text", "max_bytes"]
  }
}
```

`workspace.patch_preview`:

```json
{
  "name": "workspace.patch_preview",
  "source": "builtin",
  "group": "workspace",
  "risk_level": "high",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "workspace",
    "read_only": true,
    "write_capable": false,
    "requires_read_before_preview": true,
    "returns_diff": true,
    "preview_only": true,
    "normalizes_line_endings": true,
    "preserves_indentation": true,
    "strips_trailing_whitespace_except_markdown": true,
    "arguments": ["path", "old_text", "new_text", "max_bytes"]
  }
}
```

`workspace.patch_apply`:

```json
{
  "name": "workspace.patch_apply",
  "source": "builtin",
  "group": "workspace",
  "risk_level": "high",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "workspace",
    "read_only": false,
    "write_capable": true,
    "requires_patch_preview": true,
    "returns_diff": true,
    "normalizes_line_endings": true,
    "preserves_indentation": true,
    "strips_trailing_whitespace_except_markdown": true,
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
  "diff": "--- src/notes.txt\n+++ src/notes.txt\n-second\n+changed\n",
  "snippet": "changed",
  "match_strategy": "exact",
  "line_endings_preserved": false,
  "indentation_preserved": false,
  "trailing_whitespace_stripped": false,
  "truncated": false,
  "redaction_applied": false
}
```

Successful `workspace.patch_preview` records the same diff metadata without changing the file:

```json
{
  "tool": "workspace.patch_preview",
  "scope": "workspace",
  "operation": "patch_preview",
  "path": "src/notes.txt",
  "changed": false,
  "bytes_before": 19,
  "bytes_after": 20,
  "diff": "--- src/notes.txt\n+++ src/notes.txt\n-second\n+changed\n",
  "snippet": "changed",
  "preview_id": "patch_abcdef0123456789",
  "truncated": false,
  "redaction_applied": false
}
```

Successful `workspace.patch_apply` requires a fresh matching preview and then records `operation=patch_apply`, `changed=true`, the same compact diff/snippet, and the same `preview_id`.

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

Normal API/UI event metadata must not include host absolute workspace roots, credentials, shell output, provider raw payloads, or hidden local state. `workspace.write_file` omits raw content; `workspace.edit` and patch preview/apply include a compact diff/snippet after approval so the model and timeline can verify the proposed or applied change.

## Failure Behavior

Validation failures leave files unchanged and surface through the existing `tool_call_failed` and terminal run failure path. Typical causes include existing target, path escape, sensitive target, invalid UTF-8, oversized content, missing edit match, duplicate edit match, edit before same-run read, stale file state after read, patch apply before preview, and stale patch preview. `workspace.edit` and the patch tools normalize CRLF/LF while matching, write the original CRLF style back when present, apply the matched indentation to multi-line replacements, and strip trailing spaces/tabs for non-Markdown files.
