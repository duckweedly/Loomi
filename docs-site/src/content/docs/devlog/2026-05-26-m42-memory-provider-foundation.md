---
title: M42 Memory Provider Foundation
description: Backend-owned memory provider status, safe run readiness, and Settings > Memory provider controls.
---

M42 is the first memory provider parity slice. It keeps Loomi's existing local approved-memory store as the default provider and adds a backend-owned provider status contract for later semantic memory tools and distillation.

## Completed

- Added `memory_provider_configs` persistence for enabled state, selected provider, commit-after-run preference, semantic endpoint placeholder, safe diagnostic, and update timestamp.
- Added `GET /v1/memory/provider` and `PUT /v1/memory/provider`.
- Added provider status states: `disabled`, `available`, `unconfigured`, `healthy`, `unhealthy`, and `degraded`.
- Unknown provider values normalize to local memory with a degraded diagnostic.
- Run preparation now exposes safe memory readiness in `RunContext.SafeSummary`.
- Settings > Memory now shows a backend-derived Memory Service panel with enablement, provider, state, configured status, diagnostic, refresh, and update controls.
- Existing approved memory list/search/detail/audit/delete behavior remains on the same M13/M14 APIs.

## Safety

Provider diagnostics are redacted in productdata before they can reach HTTP, frontend state, or run summaries. API keys, Authorization headers, endpoint credentials, raw provider traces, local paths, tokens, and secret-like values must not be returned.

## Non-goals

This slice does not implement agent-facing memory tools, automatic conversation distillation, embeddings, vector search, external semantic storage, activity-recorder ingestion, or multi-agent long-term automation.
