# Data Model: M23 Workspace Mutation Tools

## Workspace Mutation Tool

Builtin tool catalog entry for a write-capable workspace operation.

Fields:

- `name`: `workspace.write_file` or `workspace.edit`
- `group`: `workspace`
- `source`: `builtin`
- `risk_level`: high for mutation tools
- `approval_policy`: always required
- `execution_state`: executable when enabled by RunContext
- `safe_metadata.scope`: `workspace`
- `safe_metadata.read_only`: `false`
- `safe_metadata.write_capable`: `true`
- `safe_metadata.operations`: operation labels shown in Settings/RunRail

Validation:

- Only enabled for Work mode run context.
- Must be present in persona allowed tool names before execution.
- Must not be enabled in Chat mode.

## Mutation Request Summary

Redacted provider argument summary used for approval projection and execution.

Fields:

- `tool_call_id`
- `tool_name`
- `path`: relative workspace path only
- `operation`: `write_file` or `edit`
- `bytes_requested`
- `line_count`
- `old_text_bytes` for edits
- `new_text_bytes`
- `preview_available`
- `redaction_applied`

Validation:

- Path must remain inside workspace root after cleaning and symlink checks.
- Sensitive path patterns are denied.
- Content must be UTF-8 text and within configured size limits.
- `workspace.write_file` target must not exist.
- `workspace.edit` target must exist, be text, and contain exactly one match for `old_text`.

## Mutation Result Summary

Redacted execution result persisted after approved mutation execution.

Fields:

- `tool`
- `scope`
- `operation`
- `path`
- `changed`
- `bytes_written`
- `bytes_before`
- `bytes_after`
- `line_count_before`
- `line_count_after`
- `truncated`
- `redaction_applied`

Validation:

- Result summary must not contain host absolute root paths.
- Result summary must not contain raw file content.
- Failed results must not include sensitive path or content values.

## State Transitions

Mutation tools reuse the existing tool-call lifecycle:

```text
requested -> approval_required -> approved -> executing -> succeeded
requested -> approval_required -> denied -> cancelled/stopped
approved/executing -> failed
```

Terminal tool calls must not be re-executed or re-applied on worker retry.
