# Plan: M68 Memory Nowledge Prompt Snapshot Regression

## Scope

Add a Nowledge-specific regression over the existing external prompt recall implementation.

## Runtime

- Configure Nowledge against a local `httptest` provider.
- Assert the Gateway calls `/memories/search` with the latest user prompt and safe limit.
- Assert safe prompt memory text is injected.
- Assert `memory_external_snapshot_loaded` metadata uses `provider=nowledge` and omits unsafe fields.

## Documentation

- Update memory API/architecture/runbook/current-status/spec-kit pages.
- Add a devlog entry with validation evidence.

## Validation

```bash
go test ./internal/runtime -run 'TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestGatewayEnrichesPromptMemorySnapshotFromNowledgeProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```
