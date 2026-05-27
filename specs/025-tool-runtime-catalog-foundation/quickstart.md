# Quickstart: M18 Tool Runtime + Tool Catalog Foundation

## Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Smoke Expectations

1. Builtin provider flow requests `runtime.get_current_time`.
2. API records approval-required tool call.
3. Approve through existing tool-call API.
4. Worker resumes and broker executes builtin tool.
5. MCP provider flow discovers `mcp.local-smoke.echo`.
6. Approve through existing tool-call API.
7. Worker resumes and broker executes local stdio MCP tool.
8. Replay API shows requested -> approval_required -> approved -> executing -> succeeded -> continuation -> completed.
9. Tools catalog API and Settings > Tools show safe catalog only.

## Explicitly Out of Scope

No workspace read tools, shell, sandbox, browser, web fetch/search, artifact runtime, plugin marketplace, remote MCP/OAuth, CLI installation, local provider autodetect, multi-agent, or worker queue rewrite.
