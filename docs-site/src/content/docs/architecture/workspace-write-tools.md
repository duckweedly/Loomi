---
title: M9 Workspace Write Tools
description: Approval-gated text mutation tools for the next code-agent slice.
---

M9 adds the first workspace mutation tools behind the existing tool-call lifecycle:

- `workspace.write_file`
- `workspace.edit`

The tools reuse M7 approval, M8 workspace boundaries, worker resume, run events, history-first SSE replay, ToolCallCard, RunRail, and Timeline. They are not shell execution and do not grant broad filesystem permission.

## Boundary

Every M9 write tool is approval-required. The worker executes a mutation only after the user approves the pending tool call. Denial, cancellation, duplicate approval, and terminal guards continue to use existing tool-call behavior.

M9 rejects unsafe inputs before mutation:

- absolute paths
- `..` traversal
- symlink escapes
- paths outside the workspace root
- sensitive paths such as `.env*`, `.ssh`, `.aws`, `secrets`, `credentials`, private key basenames, and `.pem`
- missing parent directories
- directories as file targets
- oversized text payloads

## Tools

`workspace.write_file` writes bounded UTF-8 text to a relative file path whose parent directory already exists. It reports the relative path, bytes written, and whether the file was created.

`workspace.edit` replaces `old_text` with `new_text` only when `old_text` appears exactly once in the target UTF-8 text file. Missing or duplicate matches fail without changing the file.

## Non-Goals

M9 does not add shell execution, command execution, broad patch parsing, fuzzy edit matching, binary writes, directory creation, MCP, browser automation, multi-agent delegation, or external upload.
