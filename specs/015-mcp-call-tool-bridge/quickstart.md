# Quickstart: M13 MCP Call Tool Bridge

## Focused Validation

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi
bun test --cwd web ./src/components/ToolCallCard.test.tsx
```

Expected:

- `mcp.call_tool` is allowlisted and approval-required.
- only `local.echo` is accepted.
- approved worker execution stores a bounded result summary.
- ToolCallCard renders MCP summaries without raw object output.

## Full Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```

## Browser Smoke

Start the web shell and open Settings > Tools. Confirm `mcp.call_tool` appears with group `mcp`, capability `call_tool`, risk `medium`, and side effect `mcp`.
