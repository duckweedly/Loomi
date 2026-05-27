# Feature Spec: M59 Memory Notebook Snapshot Separation

## Goal

Keep semantic memory overview/impression snapshots separate from structured notebook entries.

## Requirements

- Exclude `source_type=notebook` entries from overview snapshot hits and memory block text.
- Exclude `source_type=notebook` entries from memory impression text.
- Preserve notebook prompt injection through `RunContext.NotebookSnapshot`.
- Keep all snapshot outputs safe-summary-only.

## Non-Goals

- Do not add a dedicated notebook management page.
- Do not add new HTTP endpoints.
- Do not change approved memory entry persistence.
