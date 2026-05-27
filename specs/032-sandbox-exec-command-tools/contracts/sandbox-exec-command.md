# Contract: Sandbox Exec Command Tools

## Tool Catalog Entry

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
    "allowed_commands": ["pwd", "ls", "git status"],
    "argv_only": true,
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

Rules:

- `argv` is required and must be an array of non-empty strings.
- `cwd` is optional, relative, and defaults to `.`.
- `timeout_ms` is optional and bounded by the executor.
- `max_output_bytes` is optional and bounded by the executor.
- Shell strings, model-supplied env, file-reading commands, destructive commands, path-bearing `ls`, and out-of-root cwd values are rejected.

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

Failures use the existing `tool_call_failed` event path. Validation failures must happen before process spawn.
