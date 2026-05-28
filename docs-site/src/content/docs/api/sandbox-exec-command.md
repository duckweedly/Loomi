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

`cursor` is optional and returns stdout bytes after that absolute stdout byte cursor. `next_cursor` advances with total stdout bytes captured, not with the current preview length, so callers can poll long output without replaying already-read bytes. The stdout buffer is bounded and keeps the latest retained bytes; if a caller passes a cursor older than the retained window, Loomi returns the oldest retained safe tail and marks `stdout_truncated: true` instead of growing memory without limit. `stdin_text` is optional, capped by product-data validation, and requires a monotonically increasing `input_seq` so duplicate writes can be rejected. `close_stdin` closes the stdin pipe after any supplied text is written.

`sandbox.terminate_process` accepts only `process_id`.

The process handle is available only to the originating run. Unknown handles and cross-run handles fail before any process operation.

Read-only continue is allowed after a registry rebuild when the stored process is terminal. Continue requests that include `stdin_text` or `close_stdin` are mutations and are rejected for terminal process records or terminal runs. `sandbox.start_process` is also rejected for terminal runs.

Durable process-store writes are not best-effort for user-visible process tools. If `sandbox.start_process`, `sandbox.continue_process`, `sandbox.terminate_process`, process completion, or restored-record reconciliation cannot save the safe process record, the tool returns a safe `sandbox process durable state could not be saved` failure through the existing tool-call failure path. This prevents a successful tool result from claiming a process can be recovered after restart when the durable record was not written.

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
  "argv_summary": ["go", "test", "./..."],
  "cwd": ".",
  "cwd_alias": ".",
  "status": "running",
  "exit_code": null,
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
  "terminal_summary": "",
  "started_at": "2026-05-27T09:00:00Z",
  "updated_at": "2026-05-27T09:00:01Z",
  "ended_at": "",
  "redaction_applied": false
}
```

`status` is one of `running`, `exited`, `terminated`, `failed`, `expired`, or `lost`. `next_cursor` is the next absolute stdout cursor to pass back into `sandbox.continue_process` for incremental reads. Timed-out, TTL-expired, or lost restored processes report terminal status and a `terminal_summary`.

`sandbox.terminate_process` also returns `terminal_summary`, for example `terminated exit_code=-1`, so RunRail and ToolCallCard can show a compact lifecycle outcome without exposing raw process details. A later `sandbox.continue_process` on an exited or terminated process is read-only: it skips stdin writes/close actions and returns the stored terminal status, cursor, byte counts, and summary.

## Process Record

The runtime keeps the minimum durable handle record needed for `SandboxProcessStore` rebuild and audit. In local tests this can use the memory repository; in the API/worker path with Postgres available it uses `productdata.PostgresRepository` and the `sandbox_process_records` table:

```json
{
  "run_id": "run_123",
  "process_id": "sp_0123456789abcdef",
  "argv_summary": ["cat"],
  "cwd_alias": ".",
  "status": "exited",
  "cursor": 2,
  "started_at": "2026-05-27T09:00:00Z",
  "updated_at": "2026-05-27T09:00:01Z",
  "ended_at": "2026-05-27T09:00:01Z",
  "exit_code": 0,
  "stdout_tail": "a\n",
  "stderr_tail": "",
  "stdout_bytes": 2,
  "stderr_bytes": 0,
  "terminal_summary": "exited exit_code=0"
}
```

The durable record stores only `run_id`, `process_id`, safe argv summary, relative cwd alias, lifecycle status, cursor, byte counters, bounded output tails, timestamps, stdin state, and terminal summary. It does not store raw env, raw secret-looking output, host absolute workspace paths, shell strings, or a restartable command session.

After API process restart, terminal records can be restored as safe summaries. `sandbox.continue_process` on a restored terminal record returns stored status, cursor, byte counts, redacted output tail, and `terminal_summary`; it does not write stdin and does not execute a replacement process. A restored `running` record has no live `exec.Cmd`, so reconciliation marks it `lost` with `terminal_summary: "lost process missing after registry restore"` instead of pretending it is still running. Records older than the configured max lifetime or idle timeout are marked `expired`. Cleanup deletes only Loomi-owned `sandbox_process_records`; Loomi never scans or kills unrelated host processes.

## Event Safety

Tool-call request events expose only safe command previews. Normal API/UI events must not include host absolute workspace roots, credentials, raw environment values, provider raw payloads, or unbounded output.

M93 keeps the same output scrubber before returning one-shot and process results: invalid UTF-8 is dropped, the configured workspace root is replaced with `.`, host absolute paths and secret-looking content are redacted, and `redaction_applied` is set when the preview changes. The bounded byte counts and cursor values still reflect captured stdout/stderr bytes, not the redacted preview length.

ToolCallCard allows safe process metadata such as `argv_summary`, `cwd_alias`, `status`, `next_cursor`, timestamps, and `terminal_summary`, but treats raw `stdout`, `stderr`, paths, cwd, and secret-looking values as redacted preview fields.
