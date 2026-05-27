# Feature Spec: M58 Memory Prompt Notebook Snapshot

## Goal

Inject safe memory and notebook snapshots into model context, matching ArkLoop's runtime memory middleware shape more closely.

## Requirements

- Build a dedicated `RunContext.NotebookSnapshot` from approved visible notebook entries.
- Include safe memory summaries in a `<memory>` prompt block.
- Include safe notebook summaries in a separate `<notebook>` prompt block.
- Keep notebook entries out of the semantic `<memory>` prompt block.
- Redact prompt text and never include raw memory content, content hashes, secrets, local paths, provider traces, or tool output.

## Non-Goals

- Do not add automatic distillation or external provider snapshot refresh in this slice.
- Do not add separate notebook persistence tables.
- Do not change user-facing Settings pages.
