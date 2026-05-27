# Plan: M65 Memory OpenViking Connections

## Scope

Reuse the existing `MemoryToolExecutor.connections` boundary. Add an OpenViking-specific preflight before Nowledge/local fallback.

## Runtime

- Add `externalOpenVikingConnections`.
- Add `listOpenVikingMemoryDir` for `/api/v1/fs/ls?uri=...`.
- Project directory entries into safe connection items.
- Keep the existing `boundedLimit` cap.

## Validation

```bash
go test ./internal/runtime -run TestMemoryToolExecutorSearchesOpenVikingProvider -count=1
go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider|TestMemoryToolExecutorSearchReadStatusWriteAndForget|TestMemoryToolExecutorNotebookLifecycle' -count=1
```
