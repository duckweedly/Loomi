# Contract: Workspace Mutation Tools

## Tool Catalog Entries

`workspace.write_file`:

- `source`: `builtin`
- `group`: `workspace`
- `risk_level`: `high`
- `approval_policy`: `always_required`
- `execution_state`: `executable`
- `safe_metadata.scope`: `workspace`
- `safe_metadata.read_only`: `false`
- `safe_metadata.write_capable`: `true`
- Required arguments: `path`, `content`
- Optional arguments: `max_bytes`

`workspace.edit`:

- `source`: `builtin`
- `group`: `workspace`
- `risk_level`: `high`
- `approval_policy`: `always_required`
- `execution_state`: `executable`
- `safe_metadata.scope`: `workspace`
- `safe_metadata.read_only`: `false`
- `safe_metadata.write_capable`: `true`
- Required arguments: `path`, `old_text`, `new_text`
- Optional arguments: `max_bytes`

## Request Argument Summary

Accepted summaries are provider-neutral maps. Persisted event summaries must be redacted.

`workspace.write_file` request:

```json
{
  "path": "src/new-file.txt",
  "content": "text content",
  "max_bytes": 65536
}
```

Safe metadata may include:

```json
{
  "path": "src/new-file.txt",
  "operation": "write_file",
  "bytes_requested": 12,
  "line_count": 1,
  "preview_available": true,
  "redaction_applied": false
}
```

`workspace.edit` request:

```json
{
  "path": "src/existing.txt",
  "old_text": "before",
  "new_text": "after",
  "max_bytes": 65536
}
```

Safe metadata may include:

```json
{
  "path": "src/existing.txt",
  "operation": "edit",
  "old_text_bytes": 6,
  "new_text_bytes": 5,
  "preview_available": true,
  "redaction_applied": false
}
```

## Result Summary

Successful `workspace.write_file` result:

```json
{
  "tool": "workspace.write_file",
  "scope": "workspace",
  "operation": "write_file",
  "path": "src/new-file.txt",
  "changed": true,
  "bytes_written": 12,
  "line_count_after": 1,
  "truncated": false,
  "redaction_applied": false
}
```

Successful `workspace.edit` result:

```json
{
  "tool": "workspace.edit",
  "scope": "workspace",
  "operation": "edit",
  "path": "src/existing.txt",
  "changed": true,
  "bytes_before": 20,
  "bytes_after": 19,
  "line_count_before": 2,
  "line_count_after": 2,
  "truncated": false,
  "redaction_applied": false
}
```

## Failure Contract

Failures must leave file contents unchanged when validation fails before mutation. Failure messages are redacted and may use stable codes such as:

- `workspace_path_out_of_scope`
- `workspace_sensitive_path`
- `workspace_target_exists`
- `workspace_target_missing`
- `workspace_binary_or_invalid_utf8`
- `workspace_edit_ambiguous`
- `workspace_write_too_large`
- `tool_execution_failed`

## UI Contract

Settings Tools and RunRail should identify mutation entries as:

- workspace scoped
- write capable
- approval required
- high risk

UI-visible event detail must not include host absolute root paths or raw file contents.
