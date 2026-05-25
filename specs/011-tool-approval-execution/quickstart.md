# Quickstart: M7 Tool Approval Execution Closure

## Backend Validation

```sh
go test ./internal/productdata ./internal/runtime ./internal/db ./internal/httpapi ./cmd/...
```

Expected: approve/deny idempotency, scoped rejection, worker execution, redaction, and SSE replay tests pass.

## Frontend Validation

```sh
bun test --cwd web
bun run --cwd web build
```

Expected: ToolCallCard actions and real execution adapter mapping tests pass; web build succeeds.

## Docs Validation

```sh
bun run --cwd docs-site build
```

Expected: Starlight docs build succeeds after M7 docs updates.

## Local Smoke

1. Start the local API, worker, and web shell using the current M7 runbook.
2. Start a model/fake-provider run that requests `runtime.get_current_time` with omitted timezone or `UTC`.
3. Verify the run enters `blocked_on_tool_approval` and ToolCallCard shows approval-required state.
4. Click Deny and verify the card shows denied, Timeline shows `tool.call.denied`, and the run stops without `tool.call.executing`.
5. Start a second run and click Approve.
6. Verify ToolCallCard moves through approved, executing, and succeeded.
7. Reconnect SSE or refresh the page and verify history-first replay reconstructs the same state.
8. Try a non-UTC timezone or unsupported tool in test fixtures and verify safe failure with no execution.
