---
title: Workspace Read Tools API
description: Catalog, event, and result contracts for M21 workspace.glob, workspace.grep, and workspace.read.
---

## Catalog

`GET /v1/tools/catalog` includes:

```json
{
  "name": "workspace.read",
  "display_name": "Workspace read",
  "description": "Read a bounded UTF-8 text slice from one workspace file.",
  "source": "builtin",
  "group": "workspace",
  "risk_level": "low",
  "approval_policy": "read_only",
  "enabled": true,
  "execution_state": "executable",
  "safe_metadata": {
    "arguments": ["path", "offset", "limit", "max_bytes"],
    "read_only": true,
    "scope": "workspace"
  }
}
```

The API never returns the local absolute workspace root.

## Workspace Root

`GET /v1/workspace/root` returns only safe state:

```json
{ "config": { "configured": false, "display_name": "Home" } }
```

`POST /v1/workspace/root` accepts an absolute folder path chosen by the local desktop shell, persists it for the local user, and updates the current process runtime root for subsequent workspace tool calls. `GET /v1/workspace/root` restores the persisted folder into the current process after restart when the folder is still available. Responses still return only `configured` and `display_name`; they do not echo the absolute path.

## Tool Arguments

`workspace.glob`:

```json
{ "pattern": "**/*.go", "path": "internal", "limit": 100 }
```

`workspace.grep`:

```json
{ "query": "ToolBroker", "path": "internal", "include": "*.go", "case_sensitive": true, "limit": 100 }
```

`workspace.read`:

```json
{ "path": "internal/runtime/tools.go", "offset": 0, "limit": 32768, "max_bytes": 32768 }
```

## Event Metadata

Workspace tool events use the existing tool lifecycle event names:

```text
tool_call_requested
tool_call_approved
tool_call_executing
tool_call_succeeded
tool_call_failed
```

`workspace.glob`, `workspace.grep`, and `workspace.read` are bounded read-only tools. After the user has selected a workspace root, they are recorded as auto-approved and executed by the worker without a per-call confirmation prompt. Mutating workspace tools still require explicit approval.

Metadata includes `tool_source=builtin`, `tool_group=workspace`, redacted arguments, approval status, execution status, and safe result/error metadata.

`workspace.glob` skips generated dependency/cache folders such as `node_modules`, `dist`, `build`, `.next`, `.vite`, `.venv`, `.claude`, and `.cache`, and returns `skipped_dir_count`. `workspace.grep` uses the same skip list, skips oversized files, stops after a bounded scanned-file budget, and returns `scanned_file_count` plus `skipped_file_count`.

If `workspace.read` is pointed at a directory, it succeeds with `kind=directory`, empty `content`, a bounded `entries` list, and a summary that tells the model to use `workspace.glob` for recursive listing or `workspace.read` on a file path. This prevents a mistaken directory read from failing the whole run.

## Failure Semantics

Traversal, absolute escape, sensitive files, symlink escape, unavailable paths, invalid grep patterns, and unsupported binary content do not expose file contents. Execution failures are persisted as `tool_call_failed` and terminal run failures through the existing approved-tool path.
