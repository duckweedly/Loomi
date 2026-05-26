# Contract: M9 Workspace Write Tools

## Tool Names

- `workspace.write_file`
- `workspace.edit`

## Lifecycle

All M9 write tools use the existing lifecycle:

1. `tool_call_requested`
2. `tool_call_approval_required`
3. `tool_call_approved` or `tool_call_denied`
4. `tool_call_executing`
5. `tool_call_succeeded`, `tool_call_failed`, or `tool_call_cancelled`

## Safety Rules

- Approval is always required.
- Inputs are untrusted data.
- Paths are relative to workspace root.
- Symlink escapes and sensitive paths are denied before mutation.
- Content is UTF-8 text and bounded.
- Failed edits leave the file unchanged.
- Shell, command execution, network, MCP, browser automation, and external upload are not allowed.

## Result Shapes

```json
{
  "tool_name": "workspace.write_file",
  "result": {
    "path": "internal/example.txt",
    "bytes_written": 24,
    "created": true,
    "truncated": false
  }
}
```

```json
{
  "tool_name": "workspace.edit",
  "result": {
    "path": "internal/example.txt",
    "replacements": 1,
    "bytes_before": 24,
    "bytes_after": 25
  }
}
```
