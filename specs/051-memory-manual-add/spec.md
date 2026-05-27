# Feature Spec: M51 Manual Memory Add

## Goal

Add the local Notebook-style manual memory write path to Settings > Memory.

## Requirements

- Let the user explicitly add one approved user-scoped memory from Settings > Memory.
- Add a backend create endpoint under `/v1/memory/entries`.
- Return safe memory projections only; no raw content or content hash in the response.
- Refresh saved memory, history, and snapshot state after adding.
- Keep automatic memory writes approval-gated; this path is user-authored manual memory.

## Non-goals

- No bulk clear-all.
- No editing approved memories.
- No external provider write adapter.
