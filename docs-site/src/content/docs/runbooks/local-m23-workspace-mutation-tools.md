---
title: Local M23 Workspace Mutation Tools Validation
description: Local validation for the first approval-gated workspace mutation slice.
---

## Focused Checks

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesWorkspaceMutationTools|TestWorkspaceToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestWorkspaceWriteFile|TestWorkspaceEdit|TestToolBrokerExecutesWorkspaceMutation|TestWorkerExecutesApprovedWorkspace'
go test ./internal/httpapi -run TestM23WorkspaceMutationToolsSmoke
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

Expected evidence:

1. Tool catalog includes `workspace.write_file` and `workspace.edit` as high-risk, approval-required, write-capable workspace tools.
2. Work mode RunContext can enable mutation tools, while Chat mode keeps workspace tools out.
3. `workspace.write_file` creates a new UTF-8 text file only inside the workspace root.
4. Existing targets, traversal, sensitive paths, symlink escapes, invalid UTF-8, and oversized content fail without mutation.
5. `workspace.edit` applies exactly one replacement, rejects ambiguous/missing matches, and does not leak raw edit text into events.
6. Worker execution mutates only after explicit approval and then continues the provider.
7. HTTP smoke verifies approve -> execute -> final for both mutation tools.

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

M23 is not complete until full validation and browser smoke are finished.
