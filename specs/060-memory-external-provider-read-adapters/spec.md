# Feature Spec: M60 Memory External Provider Read Adapters

## Goal

Make configured OpenViking and Nowledge providers executable for safe read-side memory tools.

## Requirements

- Add an internal raw provider config read boundary for runtime only.
- Route `memory.search` to OpenViking `/api/v1/search/find` when OpenViking is selected and configured.
- Route `memory.read` for `viking://...` URIs to OpenViking `/api/v1/content/read`.
- Route `memory.search` to Nowledge `/memories/search` when Nowledge is selected and configured.
- Route `memory.read` for `nowledge://memory/...` URIs to Nowledge `/memories/{id}`.
- Return safe summaries only and keep raw provider payloads, credentials, local paths, and content hashes out of tool results.

## Non-Goals

- Do not execute external writes during validation.
- Do not add automatic distillation or session commit adapters in this slice.
- Do not expose raw provider config through HTTP.
