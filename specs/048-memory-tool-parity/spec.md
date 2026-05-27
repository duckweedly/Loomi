# Feature Spec: M48 Memory Tool Parity

## Goal

Expand Loomi's agent-facing memory tools toward the target mechanism while preserving Loomi's safety model: every tool is approval-gated, safe-summary-only, and backed by local productdata until external provider adapters exist.

## Requirements

- Keep existing `memory.search`, `memory.read`, `memory.write`, `memory.forget`, and `memory.status`.
- Add `memory.list`, `memory.edit`, `memory.context`, `memory.timeline`, `memory.connections`, `memory.thread_search`, and `memory.thread_fetch`.
- Expose all memory tools in ToolCatalog, provider schema mapping, worker execution, and Settings > Tools.
- Treat `memory.edit` as mutation: it may edit a pending proposal or create a replacement proposal, but it must not directly overwrite approved memory.
- Return safe summaries/excerpts only. No raw content, credentials, provider traces, file paths, or secret-like fields.

## Non-goals

- No external OpenViking/Nowledge execution.
- No notebook tools.
- No snapshot/impression builder.
- No activity recorder ingestion.
