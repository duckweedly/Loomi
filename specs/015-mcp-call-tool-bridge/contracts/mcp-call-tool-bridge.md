# Contract: mcp.call_tool

## Tool

| Tool | Approval | Capability | Risk | Side Effect |
| --- | --- | --- | --- | --- |
| `mcp.call_tool` | Required | `call_tool` | `medium` | `mcp` |

## Arguments

```json
{
  "server": "local",
  "tool": "echo",
  "arguments": {
    "message": "hello"
  }
}
```

Rules:

- only `server: "local"` is accepted in M13.
- only `tool: "echo"` is accepted in M13.
- `arguments.message` must be 1-500 characters after trimming.
- secret-looking messages are rejected.

## Result

```json
{
  "server": "local",
  "tool": "echo",
  "message": "hello",
  "side_effect": "none"
}
```
