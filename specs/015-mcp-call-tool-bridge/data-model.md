# Data Model: M13 MCP Call Tool Bridge

## MCP Call Arguments

- `server`: required string, currently only `local`
- `tool`: required string, currently only `echo`
- `arguments`: optional object

## local.echo Arguments

- `message`: required string, trimmed, 1-500 characters, must not look like a secret

## Tool Result

```json
{
  "server": "local",
  "tool": "echo",
  "message": "hello",
  "side_effect": "none"
}
```
