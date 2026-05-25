---
title: 2026-05-25 M9 RunContext Pipeline Foundation
description: Implementation notes and validation for the M9 durable RunContext and linear pipeline trace slice.
---

## Completed

- Added a durable `RunContext` loader at the product-data boundary.
- Restored run, thread, ordered messages, job metadata, provider/model route, enabled MVP tool summary, and tool resume facts from durable state.
- Routed queued worker execution through linear stage records for `prepare_context`, `resolve_tools`, `invoke_runtime`, and `finalize`.
- Added `pipeline_step_failed` for redacted stage failures.
- Mapped stage failed/completed events into frontend runtime groups, Timeline, RunRail, and Background tasks/debug surfaces.
- Preserved the M7 `runtime.get_current_time` approval execution and tool-result continuation path.

## Validation

- `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...`
- `bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/RunTimeline.runtime.test.ts ./web/src/components/RunRail.runtime.test.ts`
- `bun run --cwd web build`
- `bun run --cwd docs-site build`

Browser smoke should confirm a real run shows context prepared, tools resolved, runtime invoked, and finalized in Timeline/debug trace, then still shows the same trace after refresh/history replay.

## Non-goals

This slice did not redo the M8 worker/job queue and did not add Redis, an external queue, hosted multi-worker platform behavior, Persona/Skill, MCP, Memory/RAG, Sandbox, Desktop Runtime, shell/filesystem/browser automation tools, or multi-agent orchestration.
