# Plan: M66 Memory External Prompt Snapshot

## Scope

Add a runtime-only enrichment step in `Gateway.runWithContext` before `runSystemPrompt` is built. Reuse existing external provider read adapters instead of adding another provider client.

## Runtime

- Add `withExternalMemorySnapshot` on `Gateway`.
- Use the latest user provider message as query.
- Call `MemoryToolExecutor.externalMemorySearch`.
- Convert hits to safe `productdata.MemorySearchResult` values.
- Replace only the copied run context memory snapshot.

## Validation

```bash
go test ./internal/runtime -run 'TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
```
