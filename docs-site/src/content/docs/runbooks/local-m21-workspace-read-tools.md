---
title: Local M21 Workspace Read Tools Validation
description: Local validation and smoke expectations for workspace.glob, workspace.grep, and workspace.read.
---

## Commands

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Focused Smoke

```bash
go test ./internal/httpapi -run TestM21WorkspaceReadToolsSmoke
go test ./internal/runtime -run 'TestWorkspaceReadTools|TestToolBrokerExecutesWorkspaceTool'
```

Expected backend evidence:

1. Fixture root is set through `LOOMI_WORKSPACE_ROOT`; when unset in local desktop/dev, the runtime defaults to the user's home directory.
2. Work mode run requests `workspace.glob`, `workspace.grep`, or `workspace.read`.
3. Run emits `tool_call_requested` and `tool_call_approval_required`.
4. No filesystem content is returned before approval.
5. HTTP approval queues the worker resume job.
6. Worker executes through ToolBroker and emits `tool_call_executing` then success or failure.
7. Success triggers provider continuation and final assistant message.
8. Sensitive paths and symlink escape fail without leaking fixture secret content.
9. Desktop Work composer can choose a folder; the local API updates the runtime workspace root for subsequent tool calls.

## UI Check

Open Settings > Tools and confirm workspace tools show:

- `builtin`
- `workspace`
- `read-only`
- `workspace scope`
- `low`
- `always_required`
- `executable`

No local absolute workspace path should be visible.
