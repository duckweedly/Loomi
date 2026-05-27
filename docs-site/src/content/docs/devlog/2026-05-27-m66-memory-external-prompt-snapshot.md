---
title: M66 Memory External Prompt Snapshot
description: External provider recall before the initial model request.
---

## Completed

- Added runtime prompt enrichment for configured external memory providers.
- Reused existing OpenViking/Nowledge read adapters through `MemoryToolExecutor.externalMemorySearch`.
- Injected safe external hits into the existing `<memory>` prompt block.

## Safety

- The latest user message is used as the bounded recall query.
- Provider failures do not fail the model run.
- Prompt entries contain safe title/summary projections only.
- Raw provider payloads, credentials, and provider traces are not injected.

## Validation

- `go test ./internal/runtime -run 'TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1`
- `go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1`
- `bun run --cwd web build`
- `bun run build` from `docs-site/`
