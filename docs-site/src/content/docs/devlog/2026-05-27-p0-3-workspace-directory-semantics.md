---
title: 2026-05-27 P0-3 Workspace Directory Semantics
description: Fixed current/download directory references to the selected workspace snapshot.
---

P0-3 fixes the user-facing meaning of workspace references in Work mode.

- Current runs now carry a safe `workspace_label` alongside the existing run-scoped workspace root snapshot.
- Provider instructions explicitly bind `当前目录` / `这个目录` / `刚选目录` / current directory / this directory / selected directory to the selected workspace root for the current run.
- `下载目录` / download directory is only treated as Downloads when the selected workspace label is `Downloads`; otherwise the agent must ask the user to choose Downloads instead of falling back to Loomi, Arkloop, a historical thread, or process cwd.
- Workspace tool results include `workspace_label`, so RunRail and ToolCallCard can show `正在读取：Downloads` / `workspace: Downloads` without exposing host absolute paths.
- Composer exposes the selected workspace name through a safe title/summary only.

Regression coverage:

```bash
go test ./internal/productdata -run 'TestPrepareRunContextExposesSafeWorkspaceLabel|TestNewThreadUsesCurrentWorkspaceInsteadOfPreviousThreadRoot' -count=1
go test ./internal/runtime -run 'TestRunSystemPromptUsesSelectedWorkspaceLabelForDirectoryReferences|TestWorkspaceReadResultIncludesSafeWorkspaceLabel' -count=1
bun test --cwd web src/components/ToolCallCard.test.tsx src/components/Composer.test.ts src/components/RunRail.runtime.test.ts src/workModeProjection.test.ts src/components/WorkPlanView.test.tsx
```
