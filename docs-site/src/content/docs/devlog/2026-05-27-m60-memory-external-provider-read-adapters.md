---
title: M60 Memory External Provider Read Adapters
description: Safe read-side OpenViking and Nowledge memory tool execution.
---

## Completed

- Added an internal `GetMemoryProviderConfig` boundary so runtime can use write-only provider credentials without exposing them through HTTP.
- Routed configured OpenViking `memory.search` and `memory.read` through `/api/v1/search/find` and `/api/v1/content/read`.
- Routed configured Nowledge `memory.search` and `memory.read` through `/memories/search` and `/memories/{id}`.
- Kept tool results safe-summary-only and redacted provider errors.
- Added local `httptest` coverage for both adapters.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestMemoryProviderStatusDefaultsFallbackAndRedaction|TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider|TestMemoryToolExecutorSearchReadStatusWriteAndForget|TestMemoryToolExecutorNotebookLifecycle|TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestGatewayExposesCodeAgentToolsToProvider' -count=1`
