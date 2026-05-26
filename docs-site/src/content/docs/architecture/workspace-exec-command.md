---
title: M10 Workspace Exec Command
description: Approval-gated bounded command execution for the code-agent loop.
---

M10 adds `workspace.exec_command`, the first local command execution tool. It runs through the existing tool-call lifecycle: request, approval, worker execution, terminal event, SSE replay, and frontend rendering.

## Boundary

The command tool is approval-required and argv-only. Loomi does not invoke a shell, allocate a PTY, create persistent terminal sessions, or run background daemons.

Execution is bounded by:

- cwd resolved inside the workspace root
- timeout capped by validation
- bounded stdout/stderr previews
- redacted metadata
- rejection of shell wrappers and obvious destructive commands

## Execution Flow

1. Provider/runtime emits `workspace.exec_command` with argv, cwd, and timeout.
2. Product data validates the argument shape and blocks unsafe command names.
3. Loomi records approval-required tool events.
4. The user approves or denies the command.
5. Approved calls execute through the worker with the configured workspace root.
6. The runtime uses `exec.CommandContext` without a shell.
7. Loomi records exit code, stdout/stderr previews, timeout state, and truncation flags.

## Non-Goals

M10 does not add shell strings, `sh -c`, PTY interaction, persistent terminal sessions, process management, MCP, browser automation, multi-agent delegation, or external upload.
