---
title: M58 Memory Prompt Notebook Snapshot
description: Safe memory and notebook prompt blocks for model context.
---

## Completed

- Added `RunContext.NotebookSnapshot` and safe summary metadata.
- Built notebook snapshots during run-context preparation with `source_type=notebook`.
- Injected safe `<memory>` and `<notebook>` blocks into the gateway system prompt.
- Filtered notebook entries out of the semantic memory prompt block so structured notes stay separate.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestPrepareRunContextIncludesNotebookSnapshot|TestPrepareRunContextIncludesSafeMemorySnapshot|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots|TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestMemoryToolDefinitions|TestMemoryToolExecutorNotebookLifecycle|TestGatewayExposesCodeAgentToolsToProvider' -count=1`
