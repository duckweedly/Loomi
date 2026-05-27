# Data Model: M24 Sandbox Exec Command Tools

## Sandbox Exec Tool

- `name`: `sandbox.exec_command`
- `source`: builtin
- `group`: sandbox
- `risk_level`: high
- `approval_policy`: always_required
- `execution_state`: executable when enabled by Work-mode RunContext
- `safe_metadata`: bounded read-only command scope, argv-only marker, exec-capable marker, non-isolated marker, current allowed command list, timeout/output limits, argument names

## Exec Request Summary

Safe preview derived from model arguments:

- `argv`: ordered command tokens, excluding env values
- `cwd`: relative workspace directory, default `.`
- `timeout_ms`: bounded timeout
- `max_output_bytes`: bounded combined stdout/stderr preview limit
- `redaction_applied`: whether preview redaction occurred

Validation:

- argv must be non-empty.
- argv[0] must be one of the tiny allowlist commands: `pwd`, `ls`, or `git`.
- `ls` may only run without path args or with `.`; `git` may only run `git status`.
- cwd must resolve under the workspace root.
- model-supplied env is not accepted.
- file-reading commands and destructive command patterns are rejected before spawn.

## Exec Result Summary

Safe result metadata:

- `tool`: `sandbox.exec_command`
- `scope`: `bounded_read_only_command`
- `operation`: `exec_command`
- `argv`: safe argv preview
- `cwd`: relative cwd
- `exit_code`: integer when process exits
- `timed_out`: boolean
- `stdout`: bounded preview
- `stderr`: bounded preview
- `stdout_bytes`: captured stdout bytes before truncation
- `stderr_bytes`: captured stderr bytes before truncation
- `stdout_truncated`: boolean
- `stderr_truncated`: boolean
- `redaction_applied`: boolean

No raw env map, host absolute root, provider raw payload, or hidden local state is stored in normal events.
