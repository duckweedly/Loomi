# Feature Spec: M62 Memory Nowledge Rich Adapters

## Goal

Expose Nowledge-specific graph, timeline, and thread capabilities through Loomi memory tools.

## Requirements

- Route `memory.connections` for `nowledge://memory/...` entries to Nowledge graph expansion.
- Route `memory.timeline` to Nowledge feed events when Nowledge is selected and configured.
- Route `memory.thread_search` to Nowledge thread search.
- Route `memory.thread_fetch` to Nowledge thread fetch.
- Return only safe excerpts and summary fields.

## Non-Goals

- Do not add equivalent OpenViking graph adapters without a matching provider API.
- Do not expose raw provider traces or full message content.
- Do not change Settings UI.
