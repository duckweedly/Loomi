# Implementation Plan: M48 Memory Tool Parity

## Slice

This slice extends the existing M43 runtime bridge. New tools reuse productdata memory entries, audit events, proposals, threads, and messages. External semantic providers remain a later adapter slice.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestMemoryTool|TestGatewayExposesCodeAgentToolsToProvider|TestMemoryToolDefinitions|TestWorkerExecutesApprovedMemory' -count=1`
- `bun test --cwd web src/components/SettingsView.tools.test.tsx src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts`
- `bun run --cwd web build`
- `bun run --cwd docs-site build`
