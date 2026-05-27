---
title: M65 Memory OpenViking Connections
description: OpenViking fs/ls routing for memory.connections.
---

## Completed

- Routed `memory.connections` for `viking://...` IDs to the active OpenViking provider.
- Added a safe `/api/v1/fs/ls` adapter that projects child resources into bounded connection items.
- Kept Nowledge and local connection behavior unchanged.

## Safety

- Tool results contain only opaque child URI, safe title, node type, relation, count, provider, and redaction flag.
- Raw OpenViking payloads, provider traces, and credentials are not returned.

## Validation

- `go test ./internal/runtime -run TestMemoryToolExecutorSearchesOpenVikingProvider -count=1`
- `go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider|TestMemoryToolExecutorSearchReadStatusWriteAndForget|TestMemoryToolExecutorNotebookLifecycle' -count=1`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
