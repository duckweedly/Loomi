# Feature Spec: M49 Memory Snapshot And Impression

## Goal

Expose Arkloop-style Memory Snapshot and Memory Impression surfaces in Loomi while keeping the current local productdata safety boundary.

## Requirements

- Add HTTP endpoints for current and rebuilt memory overview snapshots.
- Add HTTP endpoints for current and rebuilt memory impressions.
- Return only approved memory summaries, hit metadata, timestamps, and rebuild state.
- Render snapshot and impression cards in Settings > Memory above provider configuration.
- Allow users to rebuild snapshot and impression from Settings > Memory.
- Keep raw memory content, proposal content, tool output, provider traces, local paths, credentials, and secret-like values out of API responses and UI state.

## Non-goals

- No external Nowledge/OpenViking adapter execution.
- No vector graph, notebook, desktop activity ingestion, or browser history ingestion.
- No automatic approval of generated memories.
