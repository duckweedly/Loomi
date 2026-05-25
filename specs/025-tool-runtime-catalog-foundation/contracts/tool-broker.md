# Contract: Tool Broker

## ToolExecutor

```text
ExecuteTool(context, ToolInvocation) -> ToolResult
```

The same broker entrypoint must route builtin and MCP tool execution.

The worker execution catalog must be derived from the prepared RunContext for the current run, not from a global historical catalog, so MCP schema hash validation uses the same discovery projection that enabled the tool.

## Required Broker Checks

- tool call belongs to the same thread/run/tool-call scope
- approval status is approved
- execution status is executing
- tool exists in catalog
- tool is enabled and executable
- tool is present in persona allowed tools for the run
- MCP tool was discovered and candidate schema hash matches
- arguments summary is redacted before result/event persistence

## Failure

Failures return safe `error_code` and `error_message`; broker must not expose raw executor output.
