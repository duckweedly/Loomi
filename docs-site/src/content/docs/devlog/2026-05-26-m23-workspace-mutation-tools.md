---
title: 2026-05-26 M23 Workspace Mutation Tools
description: Initial workspace mutation implementation notes.
---

M23 adds approval-gated workspace mutation tools on top of M21 read tools and M22 bounded continuation.

Implemented so far:

- Added `specs/031-workspace-mutation-tools/` with spec, plan, research, data model, contract, quickstart, and task list.
- Added `workspace.write_file` and `workspace.edit` tool identities to productdata catalog metadata.
- Marked mutation tools as `risk_level=high`, `approval_policy=always_required`, `read_only=false`, and `write_capable=true`.
- Added frontend Settings Tools mock/catalog visibility for write-capable workspace tools.
- Implemented `workspace.write_file` executor for new UTF-8 text files under the workspace root.
- Implemented `workspace.edit` executor for one exact UTF-8 text replacement inside an existing workspace file.
- Reused workspace root, traversal, sensitive path, and symlink escape boundaries.
- Added ToolBroker, worker, gateway, and HTTP smoke tests for approval-gated mutation execution and provider continuation.
- Added RunRail copy for high-risk, write-capable workspace mutation lifecycle rows.
- Changed mutation tool-call event metadata to expose path and byte-count previews instead of raw file/edit content.

Follow-up hardening:

- `workspace.edit` now requires a same-run `workspace.read` of the target file before mutation.
- The executor rejects stale edits when the target file size or modification time changed after that read.
- Successful edits now return a compact diff and changed-line snippet so the model can verify what changed after approval.
- `workspace.edit` now normalizes CRLF/LF for matching, preserves original CRLF files on write-back, applies matched indentation to multi-line replacements, and strips trailing spaces/tabs outside Markdown files.
- Tool catalog metadata marks `workspace.edit` with `requires_read_before_edit=true`, `returns_diff=true`, `normalizes_line_endings=true`, `preserves_indentation=true`, and `strips_trailing_whitespace_except_markdown=true`.

Patch preview/apply follow-up:

- Added `workspace.patch_preview` as an approval-gated review tool that requires same-run read freshness, computes the same compact diff/snippet as `workspace.edit`, returns `changed=false`, and records a run-local preview token without writing.
- Added `workspace.patch_apply` as the matching write tool that requires a fresh preview for the same run, file, normalized replacement, and file state before writing.
- Exposed provider-safe names `workspace_patch_preview` and `workspace_patch_apply` in builtin provider schemas.
- Updated Work-mode tool allowlists, Settings mock catalog, RunRail labels, and workspace todo titles for the patch tools.

Validation run:

```bash
go test ./internal/productdata ./internal/runtime
go test ./internal/httpapi
bun test --cwd web ./src/components/SettingsView.tools.test.tsx
bun test --cwd web ./src/components/RunRail.runtime.test.ts
```

Patch follow-up focused validation:

```bash
go test ./internal/productdata ./internal/runtime -run 'TestToolCatalogIncludesWorkspaceMutationTools|TestWorkModeScopedToolsOnlyEnabledForWorkModeRunContext|TestWorkspaceToolDefinitionsSeparateReadAndMutationRisk|TestWorkspacePatch|TestGatewayExposesCodeAgentToolsToProvider'
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

Still pending: full validation, docs build, and browser smoke.
