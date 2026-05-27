---
title: Bounded Command API
description: Catalog, request, and result contracts for approval-gated sandbox command and process tools.
---

`sandbox.exec_command`, `sandbox.start_process`, `sandbox.continue_process`, and `sandbox.terminate_process` use the existing tool-call approval and run-event APIs. No new command endpoint is added. Despite the historical sandbox group name, the current implementation is a bounded command/process slice, not an isolated sandbox.

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
    "scope": "bounded_command",
    "exec_capable": true,
    "read_only": false,
    "validation_capable": true,
    "isolated_sandbox": false,
    "argv_only": true,
    "allowed_commands": ["pwd", "ls", "cat", "head", "tail", "sed -n", "wc", "rg", "git status", "git diff", "git log", "git show", "go test", "bun test", "bun run build", "npm test", "npm run build", "pnpm test", "pnpm run build"],
    "arguments": ["argv", "cwd", "timeout_ms", "max_output_bytes"]
  }
}
```

Process tools are advertised with the same builtin group and approval policy:

```json
[
  {
    "name": "sandbox.start_process",
    "source": "builtin",
    "group": "sandbox",
    "risk_level": "high",
    "approval_policy": "always_required",
    "execution_state": "executable",
    "safe_metadata": {
      "scope": "bounded_process",
      "exec_capable": true,
      "read_only": false,
      "validation_capable": true,
      "isolated_sandbox": false,
      "argv_only": true,
      "arguments": ["argv", "cwd", "timeout_ms", "max_output_bytes", "stdin"],
      "stdin_capable": true
    }
  },
  {
    "name": "sandbox.continue_process",
    "source": "builtin",
    "group": "sandbox",
    "risk_level": "high",
    "approval_policy": "always_required",
    "execution_state": "executable",
    "safe_metadata": {
      "scope": "bounded_process",
      "exec_capable": true,
      "read_only": false,
      "validation_capable": true,
      "cursor_capable": true,
      "stdin_capable": true,
      "isolated_sandbox": false,
      "argv_only": true,
      "arguments": ["process_id", "cursor", "stdin_text", "input_seq", "close_stdin"]
    }
  },
  {
    "name": "sandbox.terminate_process",
    "source": "builtin",
    "group": "sandbox",
    "risk_level": "high",
    "approval_policy": "always_required",
    "execution_state": "executable",
    "safe_metadata": {
      "scope": "bounded_process",
      "exec_capable": true,
      "read_only": false,
      "validation_capable": true,
      "isolated_sandbox": false,
      "argv_only": true,
      "arguments": ["process_id"]
    }
  }
]
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

`argv` is required and must be an array of non-empty strings. Product-data validation and the executor both enforce the bounded workspace inspection/validation allowlist. `cwd` defaults to `.` and must stay under the configured workspace root. `timeout_ms` and `max_output_bytes` are optional and executor-bounded.

Unsupported request fields, shell-form commands, model-supplied env, write/network/package-install commands, unsafe git subcommands, absolute/traversal/sensitive path arguments, hidden-search flags, output-writing validation flags, and out-of-scope cwd values are rejected before spawn.

`sandbox.start_process` accepts the same request shape and allowlist as `sandbox.exec_command`, but its timeout is capped at 120 seconds and it returns a run-scoped process handle. It also accepts `stdin: true` for the narrow stdin-enabled process slice. That slice currently allows only argv-form `["cat"]`, without shell, PTY, env injection, or arbitrary command expansion.

`sandbox.continue_process` accepts:

```json
{
  "process_id": "sp_0123456789abcdef",
  "cursor": 48,
  "stdin_text": "next input\n",
  "input_seq": 2,
  "close_stdin": false
}
```

`cursor` is optional and returns stdout bytes after that bounded buffer offset. `stdin_text` is optional, capped by product-data validation, and requires a monotonically increasing `input_seq` so duplicate writes can be rejected. `close_stdin` closes the stdin pipe after any supplied text is written.

`sandbox.terminate_process` accepts only `process_id`.

The process handle is available only to the originating run. Unknown handles and cross-run handles fail before any process operation.

## Result Summary

```json
{
  "tool": "sandbox.exec_command",
  "scope": "bounded_command",
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

Process tool results use `bounded_process` scope:

```json
{
  "tool": "sandbox.continue_process",
  "scope": "bounded_process",
  "operation": "continue_process",
  "process_id": "sp_0123456789abcdef",
  "argv": ["go", "test", "./..."],
  "cwd": ".",
  "status": "running",
  "exit_code": -1,
  "timed_out": false,
  "stdout": "ok github.com/sheridiany/loomi/internal/runtime\n",
  "stderr": "",
  "stdout_bytes": 48,
  "stderr_bytes": 0,
  "stdout_truncated": false,
  "stderr_truncated": false,
  "next_cursor": 48,
  "stdin_open": false,
  "input_seq": 2,
  "redaction_applied": false
}
```

`status` is one of `running`, `exited`, or `terminated`. `next_cursor` is the next stdout offset to pass back into `sandbox.continue_process` for incremental reads. Timed-out processes are reported with `timed_out: true` after the process exits due to the bounded context.

## Event Safety

Tool-call request events expose only safe command previews. Normal API/UI events must not include host absolute workspace roots, credentials, raw environment values, provider raw payloads, or unbounded output.
