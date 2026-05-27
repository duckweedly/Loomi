---
title: 2026-05-27 M91 Directory Classification Tools
description: Workspace directory listing and classification tool closeout.
---

M91 adds first-class directory browsing tools for selected workspace folders:

- Added `workspace.list_directory` for bounded relative directory listings.
- Added `workspace.tree_summary` for bounded counts, extension stats, kind classification, largest files, and recent files.
- Directory tools cap `max_entries` at 500 and `depth` at 3.
- Hidden files are excluded by default; generated/cache folders such as `node_modules`, `dist`, `build`, `.next`, `.vite`, `.venv`, `.cache`, `.git`, `target`, and `vendor` are skipped.
- Results never include the host absolute workspace root and redact secret-looking filename segments.
- Provider schemas and Work prompt now steer directory inventory/classification tasks toward list/tree before grep.
- RunRail and ToolCallCard show directory-safe labels and compact summaries.

Validation added:

- Runtime tests for listing, depth bounds, generated-dir skips, truncation, root escape, hidden defaults, and classification stats.
- Provider schema coverage for the new tools.
- HTTP bounded loop smoke for `list_directory/tree_summary/read -> final` without grep-only behavior.
- Frontend tests for safe directory summaries.

Known limits:

- Classification is extension/name based, not content-aware semantic classification.
- The tools summarize the selected workspace root only; paths outside the selected root still require selecting a different folder.
