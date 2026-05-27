---
title: M59 Memory Notebook Snapshot Separation
description: Keep semantic memory snapshots separate from notebook entries.
---

## Completed

- Filtered `source_type=notebook` entries out of semantic overview snapshots and memory impressions.
- Kept notebook context available through `RunContext.NotebookSnapshot` and `<notebook>` prompt injection.
- Extended snapshot tests so notebook entries cannot leak into the semantic snapshot cards.

## Validation

- `go test ./internal/productdata ./internal/runtime -run 'TestMemoryOverviewAndImpressionSnapshotsAreSafe|TestPrepareRunContextIncludesNotebookSnapshot|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1`
