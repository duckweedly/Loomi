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

## Failure Semantics

Traversal, absolute escape, sensitive files, symlink escape, directories as files, unavailable paths, invalid grep patterns, and unsupported binary content do not expose file contents. Execution failures are persisted as `tool_call_failed` and terminal run failures through the existing approved-tool path.
