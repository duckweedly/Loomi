---
title: Local M23 Workspace Mutation Tools Validation
description: Local validation for the first approval-gated workspace mutation slice.
---

## Focused Checks

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesWorkspaceMutationTools|TestWorkspaceToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestWorkspaceWriteFile|TestWorkspaceEdit|TestWorkspacePatch|TestToolBrokerExecutesWorkspaceMutation|TestWorkerExecutesApprovedWorkspace'
go test ./internal/httpapi -run TestM23WorkspaceMutationToolsSmoke
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

Expected evidence:

1. Tool catalog includes `workspace.write_file`, `workspace.edit`, `workspace.patch_preview`, and `workspace.patch_apply` as high-risk, approval-required workspace tools.
2. Work mode RunContext can enable mutation tools, while Chat mode keeps workspace tools out.
3. `workspace.write_file` creates a new UTF-8 text file only inside the workspace root.
4. Existing targets, traversal, sensitive paths, symlink escapes, invalid UTF-8, and oversized content fail without mutation.
5. `workspace.edit` requires a same-run read, rejects stale/ambiguous/missing matches, applies exactly one replacement, preserves CRLF files, keeps multi-line replacement indentation aligned with the matched block, strips non-Markdown trailing spaces/tabs, and records a compact diff/snippet after approval.
6. `workspace.patch_preview` requires a same-run read, returns compact diff/snippet without writing, and records a run-local preview token.
7. `workspace.patch_apply` requires a fresh matching preview and rejects stale previews without mutation.
8. Worker execution mutates only after explicit approval and then continues the provider.
9. HTTP smoke verifies approve -> execute -> final for mutation tools.

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

M23 is not complete until full validation and browser smoke are finished.
