---
title: M57 Memory Notebook Tools
description: Structured notebook memory tools on the existing safe memory boundary.
---

## Completed

- Added `notebook.read`, `notebook.write`, `notebook.edit`, and `notebook.forget` to the built-in memory tool surface.
- Wired notebook tools through validation, ToolCatalog, RunContext resolution, provider schemas, provider name mapping, default persona allowlists, and Settings > Tools copy.
- Implemented notebook lifecycle execution on top of approved memory entries with a `notebook` source marker, scoped reads, replacement-by-tombstone edits, and audited forget.
- Added `source_type=notebook` filtering while keeping manual memories separate from notebook entries.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestMemoryToolDefinitions|TestMemoryToolExecutorNotebookLifecycle|TestGatewayExposesCodeAgentToolsToProvider' -count=1`
