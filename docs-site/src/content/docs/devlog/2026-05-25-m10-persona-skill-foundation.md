---
title: 2026-05-25 M10 Persona Skill Foundation
description: Implementation notes and validation for the M10 persona foundation slice.
---

## Completed

- Added persona and persona-version product data models.
- Added built-in persona sync with idempotent version creation.
- Added default built-in `loomi-default` persona with safe tool allowlist.
- Added thread persona selection and run-level persona override.
- Resolved and persisted a run persona snapshot/version when a run is created.
- Attached persona snapshot to `RunContext`.
- Applied persona model route and allowed tool names to the worker runtime context.
- Exposed safe persona summaries through pipeline metadata, Timeline, and debug views without prompt text.
- Added a minimal frontend persona selector wired into real API `persona_id` run creation.
- Added docs for architecture, API, runbook, and Spec Kit status.

## Validation

- `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...`
- `bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/Composer.test.ts ./web/src/components/RunTimeline.runtime.test.ts ./web/src/components/RunRail.runtime.test.ts`
- `bun run --cwd web build`
- `bun run --cwd docs-site build`

Browser smoke should confirm that selecting the default persona or relying on the default creates a run whose Timeline/debug surface shows safe persona name/version and whose RunContext preserves the selected version across history replay.

## Non-goals

M10 did not add a Skill marketplace, plugin install, MCP, Memory/RAG, Sandbox/Desktop Runtime, multi-agent orchestration, complex permissions, or raw persona prompt exposure in ordinary Timeline.
