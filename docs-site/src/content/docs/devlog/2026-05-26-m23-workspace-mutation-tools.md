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

Validation run:

```bash
go test ./internal/productdata ./internal/runtime
go test ./internal/httpapi
bun test --cwd web ./src/components/SettingsView.tools.test.tsx
bun test --cwd web ./src/components/RunRail.runtime.test.ts
```

Still pending: full validation, docs build, and browser smoke.
