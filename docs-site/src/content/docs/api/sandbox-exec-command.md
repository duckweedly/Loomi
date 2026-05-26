---
title: Bounded Read-only Command API
description: Catalog, request, and result contracts for M24 sandbox.exec_command.
---

M24 uses the existing tool-call approval and run-event APIs. No new command endpoint is added. Despite the historical tool name, the current implementation is a bounded read-only command slice, not an isolated sandbox.

## Catalog

```json
{
  "name": "sandbox.exec_command",
  "source": "builtin",
  "group": "sandbox",
  "risk_level": "high",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "bounded_read_only_command",
    "exec_capable": true,
    "read_only": true,
    "isolated_sandbox": false,
    "argv_only": true,
    "allowed_commands": ["pwd", "ls", "git status"],
    "arguments": ["argv", "cwd", "timeout_ms", "max_output_bytes"]
  }
}
```

## Request Arguments

```json
{
  "argv": ["pwd"],
  "cwd": ".",
  "timeout_ms": 5000,
  "max_output_bytes": 8192
}
```

`argv` is required and must be an array of non-empty strings. The first argv element must match the intentionally tiny allowlist: `pwd`, `ls`, or `git status`. `ls` accepts no path argument except `.`. `cwd` defaults to `.` and must stay under the configured workspace root. `timeout_ms` and `max_output_bytes` are optional and executor-bounded.

Unsupported request fields, shell-form commands, model-supplied env, file-reading commands such as `cat`/`head`/`tail`/`sed`/`wc`/`rg`, write/network/script/package-manager commands, unsafe git subcommands, path-bearing `ls`, absolute/traversal path arguments, and out-of-scope cwd values are rejected before spawn.

## Result Summary

```json
{
  "tool": "sandbox.exec_command",
  "scope": "bounded_read_only_command",
  "operation": "exec_command",
  "argv": ["pwd"],
  "cwd": ".",
  "exit_code": 0,
  "timed_out": false,
  "stdout": ".\n",
  "stderr": "",
  "stdout_bytes": 2,
  "stderr_bytes": 0,
  "stdout_truncated": false,
  "stderr_truncated": false,
  "redaction_applied": false
}
```

Non-zero exit codes are reported through `exit_code` and still use `tool_call_succeeded` because the command ran to completion. Validation and spawn failures use `tool_call_failed`.

## Event Safety

Tool-call request events expose only safe command previews. Normal API/UI events must not include host absolute workspace roots, credentials, raw environment values, provider raw payloads, or unbounded output.
