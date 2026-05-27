# Contract: Tools Catalog API

## `GET /v1/tools/catalog`

Returns a read-only safe catalog summary.

### Response

```json
{
  "tools": [
    {
      "name": "runtime.get_current_time",
      "display_name": "Current time",
      "description": "Returns current UTC time.",
      "source": "builtin",
      "group": "runtime",
      "input_schema_hash": "sha256:...",
      "risk_level": "low",
      "approval_policy": "always_required",
      "enabled": true,
      "execution_state": "executable",
      "safe_metadata": {
        "arguments": ["timezone"]
      }
    }
  ],
  "request_id": "req_..."
}
```

### Safety

For MCP candidates, the API uses the latest successful discovery metadata for a namespaced tool. API catalog entries do not claim worker executor availability; when executor availability is unknown, `execution_state` is `non_executable`.

The response MUST NOT include raw args, raw results, MCP command, env, stderr, access tokens, API keys, local credential paths, or provider traces.
