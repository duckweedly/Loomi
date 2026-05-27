# Data Model: M21 Workspace Read Tools

## Workspace Tool Definition

- `name`: one of `workspace.glob`, `workspace.grep`, `workspace.read`
- `display_name`: safe human label
- `source`: `builtin`
- `group`: `workspace`
- `risk_level`: `low`
- `approval_policy`: `read_only`
- `execution_state`: `executable`
- `safe_metadata`: argument names, read-only flag, and scope label only

## Workspace Scope

- `root`: resolved process-local root from persisted local user config, `LOOMI_WORKSPACE_ROOT`, or local desktop/dev Home fallback
- `display_scope`: safe non-absolute label for UI/catalog metadata
- `relative_path`: normalized slash path under root
- `deny_reason`: stable code for traversal, outside root, symlink escape, sensitive path, directory-as-file, invalid text, or unsupported arguments

Rules:
- Tool arguments never choose a new root.
- All paths are normalized relative to the root.
- Symlinks are resolved before content access.
- Sensitive paths are rejected even when inside root.

## Workspace Tool Arguments

### `workspace.glob`

- `pattern`: required glob pattern relative to root
- `path`: optional relative directory filter
- `limit`: optional count cap

### `workspace.grep`

- `query`: required literal or regexp text
- `path`: optional relative directory/file filter
- `include`: optional glob filter
- `case_sensitive`: optional boolean
- `limit`: optional match cap

### `workspace.read`

- `path`: required relative file path
- `offset`: optional byte offset
- `limit`: optional byte limit
- `max_bytes`: optional maximum returned bytes

## Workspace Tool Result

Common fields:
- `tool`: tool name
- `scope`: safe label, never an absolute path
- `truncated`: boolean
- `limit`: effective caller/server limit
- `error_code`: only on failure

Specific fields:
- Glob: `matches` with relative paths and file type, `match_count`
- Grep: `matches` with relative path, line number, and safe line excerpt, `match_count`
- Read: `path`, `content`, `bytes_read`, `offset`, `truncated`, `utf8_valid`

## Tool Call Event

Existing events remain source of truth:

```text
tool_call_requested
tool_call_approved
tool_call_executing
tool_call_succeeded
tool_call_failed
model_output_delta/model_output_completed
run_completed/run_failed
```

Workspace events include `tool_source=builtin`, `tool_group=workspace`, tool name, redacted arguments, status fields, and safe result metadata.
