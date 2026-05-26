# Contract: M8 Workspace Read Tools

## Tool Names

- `workspace.glob`
- `workspace.grep`
- `workspace.read_file`

## Lifecycle

All M8 read tools use the M7 lifecycle:

1. `tool_call_requested`
2. `tool_call_approval_required`
3. `tool_call_approved` or `tool_call_denied`
4. `tool_call_executing`
5. `tool_call_succeeded`, `tool_call_failed`, or `tool_call_cancelled`

## Safety Rules

- Approval is always required.
- Inputs are untrusted data.
- Results use relative paths only.
- Sensitive paths are denied before reading.
- Output is bounded and redacted.
- Shell, write operations, network, MCP, browser automation, and external upload are not allowed.

## Result Shapes

```json
{
  "tool_name": "workspace.glob",
  "result": {
    "matches": ["internal/runtime/tools.go"],
    "match_count": 1,
    "truncated": false
  }
}
```

```json
{
  "tool_name": "workspace.grep",
  "result": {
    "matches": [{ "path": "internal/runtime/tools.go", "line": 42, "preview": "func Execute..." }],
    "match_count": 1,
    "truncated": false
  }
}
```

```json
{
  "tool_name": "workspace.read_file",
  "result": {
    "path": "internal/runtime/tools.go",
    "size_bytes": 2048,
    "preview": "package runtime\n",
    "truncated": true
  }
}
```
