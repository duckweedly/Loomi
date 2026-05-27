# Quickstart: M12.5 Real MCP Smoke Closeout

## Targeted Evidence Smoke

```bash
go test ./internal/httpapi -run TestM12RealLocalMCPApprovalSmoke
```

Expected evidence:

- local stdio fixture discovery uses MCP `Content-Length` frames
- discovery event includes `candidate_schema_hashes`
- provider-requested MCP tool blocks on approval
- scoped HTTP approve records `tool_call_approved`
- worker loads real MCP executor config from `LOOMI_MCP_SERVERS_JSON`
- one `tools/call` is executed
- result and continuation are redacted
- continuation completes with one final assistant message

## Required Closeout Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Browser Smoke

Run browser smoke only when the local API/database/provider fixture can be started without changing deployment shape. Expected UI evidence:

- Chat shows the approval card.
- Approve transitions Timeline/RunRail through executing and succeeded.
- Continuation/final assistant output appears after the redacted tool result.

If the browser smoke cannot run, record the exact reason in the M12 devlog and rely on backend/httpapi/runtime smoke for the same state sequence.

## Non-Goals

Do not add remote MCP, MCP HTTP/SSE/OAuth, marketplace/plugin install, DB-managed MCP server admin, shell/filesystem/browser automation, automatic execution, complex sandboxing, admin UI, or multi-step tool loops.
