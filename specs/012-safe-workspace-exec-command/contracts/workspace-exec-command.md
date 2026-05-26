# Contract: M10 Workspace Exec Command

## Tool Name

- `workspace.exec_command`

## Arguments

```json
{
  "command": ["go", "test", "./internal/runtime"],
  "cwd": ".",
  "timeout_seconds": 30
}
```

## Safety Rules

- Approval is always required.
- Commands are argv arrays only.
- No shell wrapper execution.
- cwd must stay inside workspace root.
- Timeout is bounded.
- stdout/stderr are bounded and redacted.
- Obvious destructive commands are rejected before execution.

## Result Shape

```json
{
  "cwd": ".",
  "exit_code": 0,
  "stdout": "ok github.com/sheridiany/loomi/internal/runtime",
  "stderr": "",
  "timed_out": false,
  "stdout_truncated": false,
  "stderr_truncated": false
}
```
