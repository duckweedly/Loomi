---
title: M21 Workspace Read Tools
description: Bounded read-only workspace tools through the existing catalog, approval, broker, worker, and timeline path.
---

M21 adds three read-only workspace tools: `workspace.glob`, `workspace.grep`, and `workspace.read`.

They are builtin workspace tools, but they still use the same runtime path as other approved tools:

```text
provider tool request
-> tool_call_requested
-> tool_call_approval_required
-> user approve
-> worker resume job
-> ToolBroker
-> workspace executor
-> tool_call_succeeded/failed
-> provider continuation
```

## Scope

The workspace root is resolved from `LOOMI_WORKSPACE_ROOT` when set. Otherwise local desktop/dev runtime defaults to the user's home directory so common requests like `Downloads`, `Desktop`, and `Documents` can run without a prior folder picker.

The desktop shell can update the runtime root after the user explicitly chooses a folder. Tool arguments are still normalized as relative workspace paths. Absolute paths, home expansion, `..` traversal, symlink escape, and paths outside the resolved root are rejected before content access.

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

`workspace.read` returns bounded UTF-8 text with `offset`, `limit`, `bytes_read`, `utf8_valid`, and `truncated` metadata. Binary content is summarized as unsupported and does not return raw bytes.

`workspace.glob` and `workspace.grep` walk with result limits and return relative paths only. Grep returns safe line excerpts and match counts.

## Catalog And Persona

The tool catalog exposes the three workspace tools under `group=workspace`, `source=builtin`, `risk_level=low`, and `approval_policy=always_required`. Safe metadata includes `scope=workspace` and `read_only=true`; it never includes the host absolute root.

The built-in persona may list workspace tools, but RunContext filters them out for Chat mode. Work mode can enable them through the existing persona/catolog resolution path.

## Non-Goals

M21 does not add shell execution, file writes, file edits, sandbox execution, browser automation, web search/fetch, artifact creation, multi-tool loops, or a broader sandbox architecture.
