---
title: Workspace Mutation Tools
description: Approval-gated workspace write and edit tool boundaries.
---

M23 starts the workspace mutation layer after M21 read tools and M22 bounded loops. It adds four approval-gated Work-mode tools: `workspace.write_file` for new files, `workspace.edit` for direct exact text replacement, and `workspace.patch_preview` / `workspace.patch_apply` for a review-first replacement flow.

## Boundary

Workspace mutation tools reuse the existing tool runtime path:

```text
provider tool request
-> tool_call_requested / approval_required
-> explicit approval
-> worker resume
-> ToolBroker
-> WorkspaceToolExecutor
-> tool_call_succeeded or tool_call_failed
-> provider continuation
```

The tools are Work-mode scoped through RunContext enabled tools and persona allowlists. Chat-mode runs must not gain workspace mutation access.

## Write File

`workspace.write_file` creates a new UTF-8 text file under the configured workspace root. It rejects existing targets, invalid UTF-8, NUL bytes, oversized content, traversal, absolute paths, sensitive paths, and symlink escapes.

The executor validates the parent directory through the existing workspace root policy, then creates the target with exclusive create semantics. Result metadata includes only relative path, operation, changed flag, byte count, line count, and redaction/truncation flags.

## Edit

`workspace.edit` applies one bounded replacement inside an existing UTF-8 text file under the configured workspace root. It rejects missing `old_text`, duplicate matches, invalid UTF-8, NUL bytes, oversized source/result content, traversal, absolute paths, sensitive paths, symlink escapes, edits before a same-run read, and edits after the target file changed since that read.

The executor requires a successful same-run `workspace.read` for the target file, checks the current file size and modification time against that read, normalizes CRLF/LF for matching, requires exactly one `old_text` match, preserves the original CRLF style when writing back, applies the matched indentation to multi-line replacement text, strips trailing spaces/tabs for non-Markdown files, writes the updated content with the existing file permissions, and returns metadata: relative path, operation, changed flag, before/after byte counts, before/after line counts, compact diff, changed-line snippet, match strategy, formatting flags, and redaction/truncation flags.

## Patch Preview And Apply

`workspace.patch_preview` uses the same bounded exact replacement logic as `workspace.edit`, but it does not write the file. It requires a same-run `workspace.read`, checks the read freshness, computes the normalized replacement, returns the compact diff/snippet, and records a run-local preview token from the path, normalized old/new text, current file content, size, and modification time.

`workspace.patch_apply` repeats the same preparation and then requires a matching fresh preview token before writing. If the file changed after preview, if the replacement text differs, or if the run never previewed the patch, apply fails without mutation. This gives the CLI and UI a review-first path without letting clients bypass the existing backend tool approval and worker execution boundary.

## Event Metadata

Mutation result summaries do not include host absolute root paths. `workspace.write_file` result summaries omit raw file content. `workspace.edit` result summaries include a bounded diff/snippet after approval. Tool-call event argument metadata uses a safe preview with relative path, byte counts, and presence flags, while execution still uses the approved tool-call payload.

## Current Limitations

- No overwrite mode.
- No append mode.
- No rich diff preview UI yet; result metadata carries a compact diff/snippet and approval rows identify high-risk mutation tools and the affected path.
- No shell command execution, browser automation, web fetch/search, or artifact runtime is included in M23.
