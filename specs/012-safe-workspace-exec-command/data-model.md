# Data Model: M10 Safe Workspace Exec Command

## Exec Command Tool Definition

- `name`: `workspace.exec_command`
- `approval_policy`: `always_required`
- `safety_class`: `workspace_exec`
- `argument_schema`: argv command, optional cwd, optional timeout
- `result_policy`: bounded stdout/stderr previews and exit status

## Arguments

- `command`: required argv list, for example `["go", "test", "./internal/runtime"]`
- `cwd`: optional relative workspace directory, default `.`
- `timeout_seconds`: optional bounded integer

## Result

- `cwd`: relative cwd
- `exit_code`: process exit code, or `-1` for timeout/start errors
- `stdout`: bounded stdout preview
- `stderr`: bounded stderr preview
- `timed_out`: whether timeout killed the command
- `stdout_truncated`: whether stdout was truncated
- `stderr_truncated`: whether stderr was truncated
