---
title: M61 Memory External Provider Write Adapters
description: Approval-gated external OpenViking and Nowledge memory mutations.
---

## Completed

- Routed configured OpenViking `memory.write` through session creation, message append, and commit.
- Routed configured OpenViking `memory.edit` for `viking://...` URIs through content replacement.
- Routed configured OpenViking `memory.forget` for `viking://...` URIs through provider delete.
- Routed configured Nowledge `memory.write` through `/memories` and `memory.forget` through provider delete.
- Preserved local memory proposal behavior for local memory.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider|TestMemoryToolExecutorSearchReadStatusWriteAndForget|TestMemoryToolExecutorNotebookLifecycle|TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestGatewayExposesCodeAgentToolsToProvider' -count=1`
