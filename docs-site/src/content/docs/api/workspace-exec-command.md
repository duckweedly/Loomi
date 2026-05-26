---
title: Workspace Exec Command
description: M10 approval-gated command execution tool contract.
---

M10 does not add a standalone terminal API. Commands enter through the existing tool-call request, approval, worker execution, run-event history, and SSE contracts.

## Tool Name

| Tool | Approval | Execution |
| --- | --- | --- |
| `workspace.exec_command` | Required | argv-only, no shell |

## Arguments

```json
{
  "command": ["go", "test", "./internal/runtime"],
  "cwd": ".",
  "timeout_seconds": 30
}
```

## Result

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

## Failure Contract

Unsafe validation fails before execution. Timeout uses `exit_code: -1` and `timed_out: true`. Non-zero process exits are still terminal tool results with the process exit code.
