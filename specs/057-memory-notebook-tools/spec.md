# Feature Spec: M57 Memory Notebook Tools

## Goal

Add ArkLoop-style structured notebook memory tools beside Loomi's semantic `memory.*` tools.

## Requirements

- Add `notebook.read`, `notebook.write`, `notebook.edit`, and `notebook.forget` as first-class runtime tools.
- Expose these tools through ToolCatalog, RunContext resolution, provider schemas, provider name mapping, and default persona allowlists.
- Store notebook entries through the existing memory durability, redaction, scope, tombstone, and audit boundary.
- Mark notebook-backed entries distinctly so memory search/list can filter `source_type=notebook`.
- Keep tool results safe-summary-only and omit raw content, hashes, secrets, local paths, and provider traces.

## Non-Goals

- Do not add a separate notebook table yet.
- Do not add external OpenViking/Nowledge notebook adapters in this slice.
- Do not add bulk delete/export/import controls.
