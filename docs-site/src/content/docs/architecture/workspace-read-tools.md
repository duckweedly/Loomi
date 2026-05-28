---
title: M21 Workspace Read Tools
description: Bounded read-only workspace tools through the existing catalog, broker, worker, and timeline path.
---

M21 adds three read-only workspace tools: `workspace.glob`, `workspace.grep`, and `workspace.read`.

They are builtin workspace tools and use the same broker/worker continuation path as other tools:

```text
provider tool request
-> tool_call_requested
-> tool_call_approved
-> ToolBroker
-> workspace executor
-> tool_call_succeeded/failed
-> provider continuation
```

## Scope

The workspace root is resolved from the persisted local user workspace root when a run starts. That value is snapshotted into the background job metadata, surfaced through `RunContext`, and copied into each workspace tool invocation by the queued runner. Approved-tool resume jobs preserve the original run snapshot, so changing the desktop-selected folder during an active run does not move that run's execution boundary.

`/v1/workspace/root` no longer mutates `LOOMI_WORKSPACE_ROOT`; the environment variable is only a local/test fallback read at run creation when no persisted folder exists. If no folder has been saved and no explicit local-test `LOOMI_WORKSPACE_ROOT` is present, workspace tools fail before reading. Loomi no longer treats the user's home directory as an implicit workspace root; the desktop folder picker or explicit local test environment is the authorization boundary.

The desktop shell can update the persisted runtime root after the user explicitly chooses a folder. Tool arguments are still normalized as relative workspace paths. Absolute paths, home expansion, `..` traversal, symlink escape, and paths outside the resolved root are rejected before content access.

Sensitive paths are denied before opening files:

- `.env*`
- `secrets/**`
- `credentials/**`
- `id_rsa*`
- `id_ed25519*`
- `*.pem`
- `.git/**`
- `.ssh`, `.aws`, `.gnupg`

## Results

`workspace.read` returns bounded UTF-8 text with `offset`, `limit`, `bytes_read`, `utf8_valid`, and `truncated` metadata. Binary content is summarized as unsupported and does not return raw bytes. Directory reads are treated as a safe successful summary with `kind=directory`, empty `content`, bounded relative `entries`, and guidance to use `workspace.glob` or read a concrete file path; they do not terminate the run.

`workspace.glob` and `workspace.grep` walk with result limits and return relative paths only. Both skip generated dependency/cache folders before walking into them. Grep also has a scanned-file budget for no-match searches, skips oversized files, and returns safe line excerpts and scan counters.

## Catalog And Persona

The tool catalog exposes the three workspace tools under `group=workspace`, `source=builtin`, `risk_level=low`, and `approval_policy=read_only`. Safe metadata includes `scope=workspace` and `read_only=true`; it never includes the host absolute root.

Once the desktop shell has selected a workspace root, `workspace.glob`, `workspace.grep`, and `workspace.read` are auto-approved bounded reads inside that root. Workspace mutation tools, sandbox commands, browser actions, artifacts, agents, MCP, memory writes, and notebook writes remain approval-gated.

The built-in persona may list workspace tools, but RunContext filters them out for Chat mode. Work mode enables tools through the existing persona/catalog resolution path and then narrows the callable surface to the latest user intent. A casual greeting should not expose workspace, sandbox, agent, artifact, browser, or web tools. A folder listing/classification request exposes the bounded workspace read tools and hides sandbox/process tools unless the user explicitly asks to run a command.

For folder listing or classification, the model is guided to use `workspace.tree_summary` or `workspace.list_directory` first with bounded depth and entry limits, then read a few representative files if needed. `workspace.glob` is reserved for filename pattern matching or narrow follow-up discovery, and `workspace.grep` is reserved for content search.

## Non-Goals

M21 does not add shell execution, file writes, file edits, sandbox execution, browser automation, web search/fetch, artifact creation, multi-tool loops, or a broader sandbox architecture.
