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

1. Fixture root can be set through `LOOMI_WORKSPACE_ROOT` at run creation; in desktop/API runs, `/v1/workspace/root` persists the chosen folder for the local user without mutating the process environment. Each run snapshots that root into the job, `RunContext`, and tool invocation. When unset, workspace tools fail before reading and the UI shows that no folder is selected.
2. Work mode run requests `workspace.glob`, `workspace.grep`, or `workspace.read`.
3. Run emits `tool_call_requested`, `tool_call_approved`, `tool_call_executing`, then success or failure for bounded read-only workspace tools.
4. Worker executes through ToolBroker without per-call confirmation after the workspace root has been selected.
5. Success triggers provider continuation and final assistant message, including auto-approved read-only calls created during continuation.
6. Sensitive paths and symlink escape fail without leaking fixture secret content.
7. Desktop Work composer can choose a folder; the local API persists it for subsequent new runs while active runs keep their original root snapshot.
8. A casual Work greeting does not expose workspace, sandbox, agent, artifact, browser, or web tools.
9. A folder listing/classification run prefers `workspace.tree_summary` or `workspace.list_directory` without approval. `workspace.glob` is used only for filename pattern matching or narrow follow-up discovery.

## UI Check

Open Settings > Tools and confirm workspace tools show:

- `builtin`
- `workspace`
- `read-only`
- `workspace scope`
- `low`
- `read_only`
- `executable`

No local absolute workspace path should be visible.
