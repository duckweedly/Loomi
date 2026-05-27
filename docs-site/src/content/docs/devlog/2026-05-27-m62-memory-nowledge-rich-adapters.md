---
title: M62 Memory Nowledge Rich Adapters
description: Nowledge graph, timeline, and thread adapters for memory tools.
---

## Completed

- Routed `memory.connections` for `nowledge://memory/...` entries to Nowledge graph expansion.
- Routed `memory.timeline` to Nowledge feed events.
- Routed `memory.thread_search` and `memory.thread_fetch` to Nowledge thread APIs.
- Kept local memory fallback behavior unchanged.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesNowledgeProvider|TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchReadStatusWriteAndForget|TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable' -count=1`
