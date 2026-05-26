---
title: M8 Workspace Read Tools
description: Approval-gated read-only workspace tools for the first code-agent inspection slice.
---

M8 adds the first local code-agent read tools behind the M7 tool-call lifecycle:

- `workspace.glob`
- `workspace.grep`
- `workspace.read_file`

These tools reuse existing approval, worker resume, run event, SSE replay, ToolCallCard, RunRail, and Timeline paths. They do not add a separate file browser or hidden local scanner.

## Boundary

Every M8 workspace tool is read-only and approval-required. The worker executes the tool only after the user approves the pending tool call. Denial, cancellation, duplicate approval, and terminal guards continue to use M7 behavior.

The implementation rejects unsafe inputs before reading content:

- absolute paths
- `..` traversal
- paths outside the workspace root
- sensitive path segments such as `.env*`, `.ssh`, `.aws`, `secrets`, `credentials`, private key basenames, and `.pem`
- binary files for `workspace.read_file`

Results use relative paths and bounded summaries. M8 does not write files, run shell commands, call the network, invoke MCP, automate browsers, upload content, or grant persistent broad filesystem permissions.

## Execution Flow

1. Provider/runtime emits an allowlisted workspace tool request.
2. Product data validates the tool name and safe argument shape.
3. Loomi records `tool_call_requested` and `tool_call_approval_required`.
4. The user approves or denies the call.
5. Approved calls enqueue the existing worker resume job.
6. The worker records `tool_call_executing`.
7. The runtime tool executor performs bounded read-only inspection.
8. Loomi records `tool_call_succeeded`, `tool_call_failed`, or `tool_call_cancelled`.
9. History-first SSE and frontend replay render the same lifecycle.

## Result Policy

`workspace.glob` returns bounded relative path matches and a truncation flag.

`workspace.grep` returns bounded `{ path, line, preview }` rows with relative paths.

`workspace.read_file` returns a bounded UTF-8 preview, size metadata, relative path, and truncation flag.
