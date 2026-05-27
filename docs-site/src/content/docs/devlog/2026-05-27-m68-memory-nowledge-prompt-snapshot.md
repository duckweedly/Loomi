---
title: M68 Memory Nowledge Prompt Snapshot
description: Nowledge regression coverage for external prompt recall and safe run events.
---

## Completed

- Added a Nowledge-specific Gateway regression for external prompt memory recall.
- Verified the runtime calls Nowledge `/memories/search` with the latest user prompt and bounded limit.
- Verified safe Nowledge hits are injected into the initial `<memory>` prompt block.
- Verified `memory_external_snapshot_loaded` records `provider=nowledge` and safe count metadata.
- Verified prompt and event metadata do not leak Nowledge API keys, query text, raw content fields, provider traces, or local paths.

## Validation

- `go test ./internal/runtime -run 'TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestGatewayEnrichesPromptMemorySnapshotFromNowledgeProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1`
- `go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
- `git diff --check`

## Notes

- This slice adds no new HTTP response shape and no Settings UI changes.
- The implementation stays shared with OpenViking through `Gateway.withExternalMemorySnapshot` and `MemoryToolExecutor.externalMemorySearch`.
