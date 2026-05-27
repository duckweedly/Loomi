---
title: 2026-05-25 M21 Workspace Read Tools
description: Bounded read-only workspace tools complete candidate.
---

M21 adds the first useful workspace read capability to the tool runtime: `workspace.glob`, `workspace.grep`, and `workspace.read`.

Implemented:

- Builtin workspace catalog entries with read-only safe metadata.
- Work-mode-only RunContext enablement for workspace tools.
- Approval-gated provider request path through existing tool-call events.
- ToolBroker dispatch to a Go stdlib workspace executor.
- Root resolution from `LOOMI_WORKSPACE_ROOT` or repo root.
- Relative path normalization, traversal rejection, symlink escape rejection, and sensitive path denylist.
- Bounded read with UTF-8 safe content and truncation metadata.
- Bounded glob/grep with relative paths only.
- Settings > Tools read-only workspace labels.
- RunRail timeline copy for workspace tool events.
- Regression coverage for Chat mode provider requests being rejected before approval.
- Render coverage proving Settings ignores absolute paths in tool safe metadata.

Validation target:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Known limits: no shell, write/edit file tools, sandbox, browser automation, web search/fetch, artifact creation, multi-tool loops, or multi-root workspace policy.
