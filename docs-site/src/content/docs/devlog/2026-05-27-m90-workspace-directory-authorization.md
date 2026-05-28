---
title: M90 Workspace Directory Authorization
description: Require explicit folder authorization before real workspace tool execution.
---

M90 tightens the workspace tool boundary around the real selected folder.

Changes:

- `workspace.glob`, `workspace.grep`, `workspace.read`, and dependent workspace-scoped tools no longer fall back to the user's Home folder when no root is configured.
- The runtime now fails before reading with `workspace root is not authorized` unless the run captured a persisted desktop/API folder or explicit local-test `LOOMI_WORKSPACE_ROOT`.
- `StartRun` snapshots the selected workspace root into the background job; `PrepareRunContext` exposes that run-scoped root; approved-tool resume jobs preserve the original root; queued tool execution copies it into the tool invocation.
- `/v1/workspace/root` persists the selected folder but no longer calls `os.Setenv`, and startup no longer restores a saved folder into process-global env.
- Public run events redact `workspace_root_path`; raw job metadata keeps it only as the execution snapshot.
- Work prompt guidance now prefers `workspace_tree_summary` / `workspace_list_directory` for directory inventory. `workspace_glob` is only for filename matching or narrow follow-up discovery.
- `/v1/workspace/root` reports `configured=false` as `No folder selected` instead of implying Home access.
- Composer folder state now shows no selected folder until the user grants a directory.

Boundaries:

- No new workspace tool names, DB schema, shell access, or broad sandbox changes.
- Explicit local test roots through `LOOMI_WORKSPACE_ROOT` remain supported.
- Selected-root path traversal, sensitive-path, and symlink escape checks remain unchanged.
