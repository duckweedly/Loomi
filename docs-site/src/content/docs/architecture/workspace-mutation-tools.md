---
title: Workspace Mutation Tools
description: Approval-gated workspace write and edit tool boundaries.
---

M23 starts the workspace mutation layer after M21 read tools and M22 bounded loops. It adds two approval-gated Work-mode tools: `workspace.write_file` for new files and `workspace.edit` for exact text replacement.

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

`workspace.edit` applies one exact replacement inside an existing UTF-8 text file under the configured workspace root. It rejects missing `old_text`, duplicate matches, invalid UTF-8, NUL bytes, oversized source/result content, traversal, absolute paths, sensitive paths, and symlink escapes.

The executor reads the existing file after path resolution, requires exactly one `old_text` match, writes the updated content with the existing file permissions, and returns safe metadata: relative path, operation, changed flag, before/after byte counts, before/after line counts, and redaction/truncation flags.

## Event Metadata

Mutation result summaries do not include raw file content or host absolute root paths. Tool-call event argument metadata uses a safe preview with relative path, byte counts, and presence flags, while execution still uses the approved tool-call payload.

## Current Limitations

- No overwrite mode.
- No append mode.
- No diff preview UI yet; approval rows identify high-risk/write-capable mutation tools and the affected path.
- No shell command execution, browser automation, web fetch/search, or artifact runtime is included in M23.
