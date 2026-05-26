---
title: Bounded Read-only Command
description: Approval-gated bounded read-only command execution through the Loomi tool runtime.
---

M24 adds `sandbox.exec_command` after workspace read and mutation tools. The tool name is historical: this slice is not an isolated sandbox. It is intentionally narrow: argv-form command, Work-mode only, explicit approval, relative workspace cwd, tiny read-only command allowlist, bounded timeout, bounded stdout/stderr previews, and no shell session.

## Runtime Boundary

```text
provider tool request
-> tool_call_requested / approval_required
-> explicit approval
-> worker resume
-> ToolBroker
-> SandboxToolExecutor
-> tool_call_succeeded or tool_call_failed
-> provider continuation
```

The tool is enabled by the same RunContext snapshot used by workspace tools. Chat-mode runs remove sandbox tools from enabled tools, and Gateway rejects sandbox exec requests that are not present in the enabled-tool snapshot.

## Execution Model

The executor accepts only argv arrays. It resolves `cwd` through the existing workspace root boundary, rejects absolute/traversal/sensitive cwd values, rejects unlisted commands, and executes the command with a bounded context timeout. Output is captured after completion, clipped to configured byte limits, converted to safe UTF-8, and scrubbed of the absolute workspace root before event persistence.

Non-zero process exits are successful tool executions with `exit_code` metadata; validation or spawn failures become `tool_call_failed`.

## Safety Rules

- Shell-form commands are rejected.
- Model-supplied env maps are rejected.
- Only `pwd`, `ls`/`ls .`, and `git status` are allowed.
- File-reading commands such as `cat`, `head`, `tail`, `sed`, `wc`, and `rg` are not allowed in this slice because they need workspace resolver and sensitive-path handling before they can be safely exposed.
- Python, Node, Bun, Go, npm/pnpm, brew, curl/wget, ssh, shells, chmod, kill, rm, write-capable git commands, package managers, network clients, and script execution are rejected before spawn.
- Absolute and traversal path arguments are rejected before spawn.
- Denied, stopped, terminal, duplicate, or out-of-scope tool calls do not execute.
- Normal run events do not include raw env values, host absolute roots, provider raw payloads, or hidden local state.

## Current Limitations

- No streaming stdout/stderr.
- No persistent shell session.
- No container sandbox, syscall isolation, or per-command filesystem snapshot yet.
- No browser, web fetch/search, artifact runtime, or multi-agent orchestration in M24.
