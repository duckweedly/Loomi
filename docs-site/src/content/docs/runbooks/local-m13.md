---
title: Local M13 MCP Call Tool Bridge
description: Validate mcp.call_tool locally.
---

## Focused Validation

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi
bun test --cwd web ./src/components/ToolCallCard.test.tsx
```

Expected:

- `mcp.call_tool` is allowlisted and approval-required.
- only `local.echo` is accepted.
- secret-looking messages are rejected before approval.
- approved worker execution stores a bounded echo result.
- ToolCallCard renders MCP arguments/results without raw object output.
- `GET /v1/tools/catalog` includes `mcp.call_tool`.

## Full Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```

## Browser Smoke

Start the web shell:

```bash
bun run --cwd web dev -- --host 127.0.0.1 --port 5173
```

Open Settings > Tools and confirm `mcp.call_tool` appears with capability `call_tool`, risk `medium`, side effect `mcp`, and safety class `mcp_bridge`.
