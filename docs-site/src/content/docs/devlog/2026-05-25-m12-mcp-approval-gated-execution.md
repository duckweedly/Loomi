---
title: 2026-05-25 M12 MCP Approval-Gated Execution
description: Implementation notes and validation for the M12 local MCP execution slice.
---

## Completed

- Allowed namespaced MCP tool-call projections through the existing M7 approval boundary.
- Required both prior discovery and persona allowed-tools resolution before approval.
- Added safe MCP metadata on tool-call events: `tool_source`, `server_slug`, and `candidate_schema_hash`.
- Added a bounded local stdio MCP `tools/call` executor facade using the same `Content-Length` framing as discovery.
- Wired the real API worker path to `StdioMCPToolExecutor` with local configs from `LOOMI_MCP_SERVERS_JSON`.
- Added worker routing for one approved MCP execution and one provider continuation.
- Prevented duplicate execution after `executing`, `succeeded`, `failed`, or `cancelled` states.
- Extended frontend replay grouping for `tool.call.*` events and MCP metadata.

## Validation

Targeted validation run during implementation:

```bash
go test ./internal/runtime -run 'TestGatewayRecordsApprovalRequiredMCP|TestGatewayRejectsMCP|TestGatewayRejectsPersonaAllowedMCP|TestWorkerExecutesApprovedMCP|TestWorkerDoesNotReexecuteMCP|TestGatewayBuildsMCPContinuation|TestStdioMCPToolExecutor'
go test ./internal/productdata -run 'TestRepositoryContractCoversMCPToolCallRequestProjection|TestRepositoryContractCoversM7ToolCallRequestProjection|TestRecordToolCallRequestValidatesM7SafetyBoundary|TestToolCallExecutionEventsRedactResultAndErrors'
bun test --cwd web src/runtime/realExecutionAdapter.test.ts src/runtime/runtimeEventGroups.test.ts
```

Full validation passed:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Boundaries

M12 executes only already-discovered local stdio MCP candidates after user approval. It does not add remote MCP, MCP HTTP/SSE/OAuth, marketplace/plugin install, DB-managed MCP server admin, shell/filesystem/browser automation, automatic execution, complex sandboxing, admin UI, or multi-step tool loops.
