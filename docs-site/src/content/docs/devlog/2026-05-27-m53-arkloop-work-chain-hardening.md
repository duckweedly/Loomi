---
title: 2026-05-27 M53 ArkLoop Work Chain Hardening
---

Status: candidate hardening slice.

Changes:

- Compared the local ArkLoop source under `tmp/Arkloop` and aligned Loomi's Work-mode prompt with ArkLoop's tool-first file workflow: folder/file tasks should use workspace tools before asking the user to run commands or paste listings.
- Persisted the desktop-selected workspace root for the local user and restored it into the API process after restart. API responses still expose only `configured` and `display_name`.
- Changed `workspace.glob`, `workspace.grep`, and `workspace.read` to bounded read-only auto-approved tools once the workspace root is selected. Mutating workspace tools remain approval-gated.
- Fixed worker continuation so an auto-approved read-only tool emitted after a prior tool result is executed in the same run chain instead of getting stuck as `approved/not_started`.
- Scoped Work-mode enabled tools from the latest user intent so casual chat does not receive workspace/sandbox/agent/artifact/browser tools, while file/folder tasks receive the workspace read tools without requiring shell commands.
- Added a continuation guard for folder listing: after a successful `workspace.glob`, later model continuations omit `workspace.glob` and keep targeted `workspace.grep` / `workspace.read` available.
- Hardened `tool.load_tools` so model requests that provide a single `queries` or `names` string are normalized instead of rejected, and provider-safe names such as `workspace_read` resolve to the internal dotted catalog name.
- Updated Work-mode prompt and workspace read tool descriptions to treat the selected folder as the root: folder-listing calls should use path `.` and should not repeat display names such as `Downloads`.
- Added a workspace executor guard for repeated selected-folder names: if the root is `Downloads` and a missing tool path starts with `Downloads/`, it resolves from the selected root instead of failing as `Downloads/Downloads`.

Smoke evidence:

- Temporary latest API on `127.0.0.1:18081` loaded the persisted workspace root as `Downloads`.
- A real Work run for “列一下当前工作目录里的文件，简单分类。” prepared only 6 intent-scoped tools, auto-executed one `workspace.glob`, required no user approval, completed, and returned a categorized answer from the selected folder.

Validation target:

- `go test ./internal/productdata ./internal/runtime ./internal/httpapi`
- full Go/Web/docs validation before closeout
