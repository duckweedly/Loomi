---
title: P0 Desktop Readiness Diagnostics
description: Real desktop startup now exposes API, DB, provider, Local Codex, tool catalog, and workspace blockers.
---

## What changed

- Added a desktop readiness model for the renderer so real API failures do not collapse into raw `Failed to fetch`.
- Chat Canvas now shows the highest-priority blocker with next-step actions: retry, open settings, detect local provider, enable Local Codex, or choose folder.
- The web client reads `/readyz` directly, including `not_ready` schema responses, so DB/schema failures can be separated from API-unreachable failures.
- Added `loomi doctor --desktop` to include workspace root readiness alongside API, provider, and tool catalog checks.

## Validation

- Targeted web tests covered `realApiClient`, readiness derivation, and Chat Canvas rendering.
- Targeted Go tests covered desktop doctor workspace diagnostics.

## Limits

- The readiness panel uses the existing Settings and workspace-selection flows; it does not auto-start Postgres or the API.
- Tool catalog readiness is based on whether the frontend can load executable tools from `/v1/tools/catalog`; deeper tool execution is still validated by the M79/M92 smoke paths.
