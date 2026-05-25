# Implementation Plan: Persona Skill Foundation

**Branch**: `main` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/015-persona-skill-foundation/spec.md`

## Summary

M10 adds the smallest Persona/Skill foundation on top of the M9 RunContext/Pipeline baseline. Built-in persona definitions sync into durable product data with versions, thread/run creation resolves a persona through run override -> thread selection -> default built-in fallback, the worker records the resolved persona snapshot/version in RunContext before runtime invocation, and Timeline/debug shows only a safe persona summary. The slice deliberately stops before Skill marketplace, plugin install, MCP, Memory/RAG, Sandbox/Desktop Runtime, multi-agent, broad permissions, or exposing raw persona prompt text in normal Timeline metadata.

## Technical Context

**Language/Version**: Go backend/runtime/worker; TypeScript/React frontend; Starlight docs-site with Bun validation.

**Primary Dependencies**: Existing `internal/productdata` run/thread/message/job/provider data boundary and M9 `PrepareRunContext`; existing `internal/runtime` worker, queued runner, pipeline recorder, gateway, and tool allowlist; existing `internal/httpapi` thread/run endpoints and SSE/history replay; existing `web/src/realApiClient.ts`, runtime adapters, `RunTimeline`, `RunRail`, and thread/run composer surfaces.

**Storage**: Existing PostgreSQL migrations plus new durable persona tables or columns as narrowly required for `personas`, persona versions/snapshots, and thread/run persona references. Built-in persona config is repository-local data synced into DB; no external marketplace or plugin registry storage.

**Testing**: `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...`; related web runtime/UI tests for persona selection/display and safe Timeline/debug summary; `bun run --cwd web build`; `bun run --cwd docs-site build`; browser smoke selecting a persona or using default persona to create a run, then checking Timeline/debug persona summary and RunContext persona version.

**Target Platform**: Local Loomi development environment with Go API/worker, local PostgreSQL, web renderer, and docs-site.

**Project Type**: Local web application plus Go API/backend runtime and durable product data.

**Performance Goals**: Persona sync is idempotent and suitable for local startup or explicit seed path. Persona resolution adds no API request blocking beyond existing run creation reads/writes. Timeline/debug summary appears through normal run-event ordering in live SSE and history replay.

**Constraints**: Reuse M6/M9 worker/job, RunContext, pipeline, provider route, SSE/history, and MVP tool allowlist. No full Skill marketplace, plugin install, MCP, Memory/RAG, Sandbox/Desktop Runtime, multi-agent, new queue, large framework, broad permission system, or raw persona prompt exposure in normal Timeline/debug metadata.

**Scale/Scope**: One local built-in persona sync path, at least one default built-in persona, one resolved persona snapshot per run, thread/run selection fallback, safe summary in current Timeline/debug surfaces.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. Uses Loomi's own Persona, Persona Version, Persona Snapshot, RunContext, Timeline, and debug terminology.
- **II. Runnable Vertical Slices**: PASS. The feature produces a verifiable run path from built-in sync through run creation, worker RunContext, and visible safe persona summary.
- **III. Core Flow Before Platform Complexity**: PASS. It follows worker/job and RunContext foundations and explicitly defers marketplace, plugins, MCP, memory, sandbox, desktop runtime, and multi-agent behavior.
- **IV. Observable Agent Execution**: PASS. Persona selection/version is recorded as safe run context metadata and rendered in Timeline/debug.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Raw system prompt is runtime-only and excluded from normal Timeline/debug summaries; persona allowed tools can only narrow existing allowlisted runtime tools.
- **Technical Constraints**: PASS. The plan reuses productdata/runtime/httpapi/web/docs boundaries and avoids new dependencies.
- **Development Workflow**: PASS. Spec, plan, tasks, analyze, then implementation only after user confirmation.
- **Documentation Definition of Done**: PASS. Implementation tasks include docs-site architecture/API/runbook/roadmap/devlog/spec-kit updates and docs build.

## Project Structure

### Documentation (this feature)

```text
specs/015-persona-skill-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── persona-sync.md
│   ├── persona-resolution.md
│   └── frontend-persona-summary.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── productdata/
│   ├── models.go              # Persona, PersonaVersion/Snapshot, thread/run references
│   ├── repository.go          # durable persona sync, resolution, snapshot persistence
│   ├── repository_test.go
│   ├── service.go             # built-in sync and persona resolution use cases
│   └── service_test.go
├── runtime/
│   ├── pipeline.go            # include safe persona summary in prepare_context metadata
│   ├── pipeline_test.go
│   ├── queued_runner.go       # consume persona snapshot from RunContext
│   ├── tools.go               # intersect persona tools with existing runtime allowlist
│   └── worker_test.go
└── httpapi/
    ├── product.go             # thread persona selection if implemented through thread API
    ├── runtime.go             # run persona override and event/history exposure
    └── runtime_test.go

cmd/
├── loomi-api/
│   └── main.go                # startup or seed hook for built-in persona sync if chosen
└── loomi-seed/
    └── main.go                # explicit seed path if startup hook is not chosen

web/src/
├── realApiClient.ts                         # persona fields and safe event metadata mapping
├── runtime/
│   ├── realExecutionAdapter.ts              # replay persona summary into runtime state
│   └── runtimeEventGroups.ts                # persona summary grouping for debug/timeline
└── components/
    ├── Composer.tsx                         # minimal run persona selector if selected path
    ├── RunTimeline.tsx                      # safe persona summary display
    ├── RunRail.tsx                          # debug detail surface
    └── SettingsView.tsx                     # optional read-only persona catalog if selected path

docs-site/src/content/docs/
├── architecture/persona-skill-foundation.md
├── api/persona-skill-foundation.md
├── runbooks/local-m10-persona.md
├── roadmap/current-status.md
├── spec-kit/workflow.md
└── devlog/2026-05-25-m10-persona-skill-foundation.md
```

**Structure Decision**: Persona persistence and resolution belong in `internal/productdata` because thread/run ownership and durable snapshots are product data. Runtime code consumes the resolved snapshot through the existing M9 RunContext path and only applies model/tool behavior through current gateway/tool boundaries. HTTP API changes stay narrow: expose selection inputs and safe summaries. Frontend work stays in current composer/timeline/debug surfaces so the slice is browser-smokeable without a new admin console.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Store built-in personas as durable records plus versioned snapshot bodies.
- Resolve persona at run creation/worker context time using run override, thread selection, then default built-in fallback.
- Keep raw system prompt out of normal Timeline/debug and event metadata.
- Intersect persona allowed tools with existing runtime tool allowlist.
- Let persona model route select within existing provider/model routing; do not add provider administration.
- Implement minimal frontend selector or read-only resolved persona display based on the smallest existing UI surface that supports a browser smoke.

## Phase 1: Design Summary

- [data-model.md](./data-model.md) defines Persona, Persona Version, Persona Selection, Persona Snapshot, RunContext Persona Summary, and Built-in Persona Config.
- [contracts/persona-sync.md](./contracts/persona-sync.md) defines built-in config sync behavior and idempotency.
- [contracts/persona-resolution.md](./contracts/persona-resolution.md) defines thread/run/default fallback and RunContext snapshot semantics.
- [contracts/frontend-persona-summary.md](./contracts/frontend-persona-summary.md) defines safe Timeline/debug and minimal UI expectations.
- [quickstart.md](./quickstart.md) defines validation commands and browser smoke path requested by the user.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart validates built-in sync, run resolution, RunContext snapshot/version, and Timeline/debug display.
- **Core Flow Before Platform Complexity**: PASS. Contracts explicitly defer marketplace, plugins, MCP, memory, sandbox, desktop runtime, and multi-agent capabilities.
- **Observable Agent Execution**: PASS. Safe persona summary has event/API/frontend contracts.
- **Safety/Data Boundaries**: PASS. Prompt text is not part of normal Timeline/debug metadata; allowed tools are constrained by existing allowlist.
- **Documentation**: PASS. docs-site update targets and validation are included in tasks.

## Complexity Tracking

No constitution violations. New durable persona storage is justified because M10 requires versioned Persona DB and old-run attribution; no new runtime platform, marketplace, plugin layer, MCP, memory system, sandbox, or permission framework is justified for this slice.
