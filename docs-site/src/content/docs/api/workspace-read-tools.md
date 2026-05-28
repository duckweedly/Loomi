---
title: Workspace Read Tools API
description: Catalog, event, and result contracts for workspace directory, grep, and read tools.
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
{ "config": { "configured": false, "display_name": "No folder selected" } }
```

`POST /v1/workspace/root` accepts an absolute folder path chosen by the local desktop shell and persists it for the local user. It does not update `LOOMI_WORKSPACE_ROOT` or any process-global execution boundary. Each new run snapshots the selected root into the background job metadata, `RunContext`, and tool invocation; approved-tool resume jobs preserve that same snapshot. Responses still return only `configured` and `display_name`; they do not echo the absolute path. When no folder has been selected and no explicit local-test `LOOMI_WORKSPACE_ROOT` is present at run creation, workspace tools fail before reading instead of defaulting to the user's home directory.

RunContext also carries a safe `workspace_label`, derived from the selected folder basename, for provider prompts and UI summaries. If a previous thread used `Arkloop` and the current thread starts after selecting `Downloads`, the new run uses the `Downloads` snapshot and label. User phrases such as `current directory`, `this directory`, `selected directory`, `just selected directory`, `当前目录`, `这个目录`, and `刚选目录` always mean the selected workspace root for the current run. `download directory` / `下载目录` only means Downloads when the selected workspace label is `Downloads`; otherwise the agent must ask the user to choose Downloads before using workspace tools.

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

`workspace.list_directory`:

```json
{ "path": ".", "max_entries": 200, "depth": 2, "include_hidden": false, "sort": "name" }
```

`workspace.tree_summary`:

```json
{ "path": ".", "max_entries": 200, "depth": 3, "include_hidden": false, "sort": "modified" }
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

`workspace.glob`, `workspace.grep`, `workspace.read`, `workspace.list_directory`, and `workspace.tree_summary` are bounded read-only tools. After the user has selected a workspace root, they are recorded as auto-approved and executed by the worker without a per-call confirmation prompt. Mutating workspace tools still require explicit approval.

Metadata includes `tool_source=builtin`, `tool_group=workspace`, redacted arguments, approval status, execution status, and safe result/error metadata.

Successful workspace tool results include `workspace_label` so RunRail and tool cards can say which selected workspace is being read without exposing the host absolute path:

```json
{
  "tool": "workspace.read",
  "scope": "workspace",
  "workspace_label": "Downloads",
  "path": "receipt.txt",
  "bytes_read": 128,
  "truncated": false
}
```

`workspace.glob` skips generated dependency/cache folders such as `node_modules`, `dist`, `build`, `.next`, `.vite`, `.venv`, `.claude`, `.cache`, `.git`, `target`, and `vendor`, and returns `skipped_dir_count`. `workspace.grep`, `workspace.list_directory`, and `workspace.tree_summary` use the same skip list. Grep skips oversized files and stops after a bounded scanned-file budget; directory tools stop after `max_entries` and cap depth at `3`.

`workspace.list_directory` returns safe relative entries:

```json
{
  "operation": "list_directory",
  "path": ".",
  "total_entries_seen": 12,
  "returned_entries": 12,
  "truncated": false,
  "directories_count": 3,
  "files_count": 9,
  "entries": [{ "path": "src/main.go", "kind": "file", "depth": 2, "size": 128 }]
}
```

`workspace.tree_summary` returns the same bounded scan plus classification:

```json
{
  "operation": "tree_summary",
  "total_entries_seen": 12,
  "returned_entries": 12,
  "truncated": false,
  "directories_count": 3,
  "files_count": 9,
  "by_extension": { ".go": 2, ".md": 1 },
  "by_kind": { "code": 2, "document": 1, "image": 0, "video": 0, "audio": 0, "archive": 0, "app": 0, "other": 0 },
  "largest_files": [{ "path": "src/main.go", "kind": "code", "size": 128 }],
  "recent_files": [{ "path": "docs/spec.md", "kind": "document", "size": 64 }]
}
```

Directory results never include the host absolute root. Hidden entries are omitted unless `include_hidden=true`. Secret-looking filename segments such as `token`, `secret`, `password`, `credential`, and API-key variants are replaced with `[redacted]`.

If `workspace.read` is pointed at a directory, it succeeds with `kind=directory`, empty `content`, a bounded `entries` list, and a summary that tells the model to use `workspace.tree_summary` or `workspace.list_directory` for directory inventory, or `workspace.read` on a file path. This prevents a mistaken directory read from failing the whole run.

## Failure Semantics

Missing workspace authorization, traversal, absolute escape, sensitive files, symlink escape, unavailable paths, invalid grep patterns, and unsupported binary content do not expose file contents. Execution failures are persisted as `tool_call_failed` and terminal run failures through the existing approved-tool path.
