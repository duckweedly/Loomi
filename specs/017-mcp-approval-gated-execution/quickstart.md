# Quickstart: MCP Approval-Gated Execution

This quickstart is for the next implementation session. The current session only creates the Spec Kit design artifacts.

## Expected local setup

- M7 approval execution and tool-result continuation tests pass.
- M11 MCP stdio foundation can discover a local stdio MCP tool as a read-only namespaced candidate.
- A persona fixture includes the discovered namespaced MCP tool in allowed-tools.

## Validation flow

1. Run backend tests for approval projection and policy:

   ```bash
   go test ./internal/productdata ./internal/runtime
   ```

2. Run worker tests for ownership, cancellation, retry/recovery, and stdio lifecycle:

   ```bash
   go test ./internal/runtime
   ```

3. Run full backend validation:

   ```bash
   go test ./...
   ```

4. Run frontend replay tests after event mapping changes:

   ```bash
   bun test --cwd web
   ```

5. Build the frontend after replay/UI changes:

   ```bash
   bun run --cwd web build
   ```

6. Build the documentation site after docs-site updates:

   ```bash
   bun run --cwd docs-site build
   ```

## Manual smoke expectations

- A model-requested `mcp.<server_slug>.<tool_name>` call appears as approval-required and blocked.
- Deny records a denial and never starts a stdio process.
- Approve causes the worker to execute one local stdio MCP call under ownership.
- Timeline/debug shows execution started, succeeded or failed, and one continuation phase.
- Replaying history shows the same safe metadata as live SSE.
- Retrying after started/succeeded/failed/cancelled state does not execute the MCP tool again.
- Continuation asking for another tool records unsupported tool loop and runs no tool.

## Out-of-scope smoke

Do not validate or implement:

- remote MCP
- MCP HTTP/SSE/OAuth
- marketplace/plugin install
- DB-managed MCP server admin
- shell/filesystem/browser automation
- automatic execution
- complex sandbox
- admin UI
- multi-step tool loop
