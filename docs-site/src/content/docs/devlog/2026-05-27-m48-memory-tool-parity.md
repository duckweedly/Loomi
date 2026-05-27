---
title: M48 Memory Tool Parity
description: Expanded agent-facing memory tools backed by safe local summaries.
---

M48 expands Loomi's agent memory tools toward the target mechanism while preserving the local approval and redaction boundary.

## Completed

- Added `memory.list`, `memory.edit`, `memory.context`, `memory.timeline`, `memory.connections`, `memory.thread_search`, and `memory.thread_fetch`.
- Added validation, ToolCatalog entries, provider schema names, gateway mappings, and built-in persona allowlist entries.
- Implemented safe local execution in `MemoryToolExecutor`.
- Updated Settings > Tools labels and mock catalog coverage.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestMemoryTool|TestGatewayExposesCodeAgentToolsToProvider|TestMemoryToolDefinitions|TestWorkerExecutesApprovedMemory' -count=1`
- `bun test --cwd web src/components/SettingsView.tools.test.tsx src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts`
- `bun run --cwd web build`

## Still Deferred

- External OpenViking/Nowledge adapter execution.
- Notebook tools.
- Snapshot and impression builder.
