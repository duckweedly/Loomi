# Contract: M12.5 Evidence Chain

The closeout smoke must prove this order:

```text
local stdio MCP discovery
  -> mcp_discovery_succeeded with candidate_schema_hashes
  -> persona allowed tool appears in enabled_tools
  -> provider requests mcp.<server>.<tool>
  -> tool_call_requested
  -> tool_call_approval_required
  -> scoped HTTP approve
  -> tool_call_approved
  -> worker StdioMCPToolExecutor executes one tools/call
  -> tool_call_executing
  -> tool_call_succeeded with redacted result
  -> provider continuation receives redacted tool result
  -> model_phase=continuation events
  -> run_completed and final assistant message
```

Required redaction assertions:

- raw fixture token is absent
- private path is absent
- raw env value is absent
- raw stderr is absent
- continuation request contains only redacted tool result content

Out-of-scope assertions:

- no remote MCP endpoint
- no MCP HTTP/SSE/OAuth
- no marketplace or plugin install
- no sandbox feature
- no shell/filesystem/browser automation tool support
- no multi-tool loop support
