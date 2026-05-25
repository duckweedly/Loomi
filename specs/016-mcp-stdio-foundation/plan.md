# Implementation Plan: MCP Stdio Foundation

**Branch**: `main` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/016-mcp-stdio-foundation/spec.md`

## Summary

M11 adds the smallest MCP stdio foundation after M10 Persona/Skill and M9 RunContext. Loomi can read explicit local stdio MCP server configs, run bounded discovery/list-tools, validate discovered schemas, map tools into namespaced read-only ToolSpec candidates, let persona allowed-tools reference those candidates as disabled-by-default, and surface safe MCP availability in RunContext plus Timeline/debug. The slice stops before MCP execution, HTTP/SSE/OAuth, remote MCP, marketplace install, shell/filesystem/browser automation, sandboxing, or any approval bypass.

## Technical Context

**Language/Version**: Go backend/runtime/worker; TypeScript/React frontend if Timeline/debug labels need mapping; Starlight docs-site with Bun validation.

**Primary Dependencies**: Existing `internal/productdata` config/run/thread/persona/tool-call data boundary; existing `internal/runtime` ToolSpec/ToolRegistry, M7 approval concepts, M9 RunContext/pipeline, M10 persona allowed-tool resolution, worker/job runner, run events, SSE/history replay, and frontend runtime adapters.

**Storage**: Explicit local MCP server configuration is repository/local-dev configuration for this slice. Implementation may persist or project discovery safe status if needed for history/debug replay. Persist only safe summaries: server slug/id, enabled state, discovery status, safe candidate names, counts, timestamps, and redacted error codes. Do not persist admin-managed server config, env values, raw args, raw stderr, tokens, credentials, or secret-looking paths.

**Testing**: Go tests for config validation, discovery parser, schema redaction, ToolSpec mapping, persona allowed-tool references, and RunContext MCP availability summaries. Web tests for Timeline/debug labels only if UI mapping is touched. `bun run --cwd docs-site build` after planned docs-site updates.

**Target Platform**: Local Loomi development environment with Go API/worker, local PostgreSQL where needed, web renderer, and docs-site.

**Project Type**: Local web application plus Go API/backend runtime and durable product data.

**Performance Goals**: Discovery is bounded by per-server timeout and candidate limits. RunContext preparation uses cached or recent safe discovery summaries when available and does not block normal runs on optional MCP discovery failure unless the run/persona explicitly requires MCP availability in a later slice.

**Constraints**: Only local stdio MCP servers. No HTTP/SSE/OAuth, remote network MCP, marketplace/plugin install, shell/filesystem/browser automation, automatic tool execution, complex sandbox, new queue, broad permission framework, or reimplementation of Persona/Skill, RunContext, Worker queue, or M7 approval. Discovery output is untrusted data.

**Scale/Scope**: One local-development MCP config model, one bounded discovery/list-tools path, one ToolSpec mapping path, one safe RunContext availability summary, required Timeline/debug labels for discovery state, and a documented future execution boundary.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. Uses Loomi's MCP Server Config, MCP Tool Candidate, ToolSpec, RunContext, Timeline, and approval terminology without copying another product's expression.
- **II. Runnable Vertical Slices**: PASS. The slice is runnable through local config validation, discovery/list-tools, mapping, and visible safe availability summary without execution.
- **III. Core Flow Before Platform Complexity**: PASS. M11 follows M7/M9/M10 foundations and explicitly defers HTTP/SSE/OAuth, marketplace, sandbox, desktop runtime, and automation tools.
- **IV. Observable Agent Execution**: PASS. Discovery success/failure and availability state are recorded as safe metadata visible through Timeline/debug.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Env/args/stderr/tokens/paths are sensitive; discovery output is external data; execution remains approval-gated future work.
- **Technical Constraints**: PASS. The plan reuses productdata/runtime/httpapi/web/docs boundaries and avoids new platform layers.
- **Development Workflow**: PASS. Specify, plan, tasks, analyze completed first; implementation proceeded only after user confirmation.
- **Documentation Definition of Done**: PASS. Implementation tasks include docs-site architecture/API or approval extension/runbook/roadmap/devlog/spec-kit updates and docs build.

## Project Structure

### Documentation (this feature)

```text
specs/016-mcp-stdio-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── mcp-server-config.md
│   ├── mcp-discovery-tool-mapping.md
│   ├── mcp-run-context-observability.md
│   └── mcp-execution-boundary.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── productdata/
│   ├── models.go              # discovery safe projection types if persisted here
│   ├── repository.go          # optional discovery safe status reads/writes
│   ├── service.go             # config validation and persona allowed-tool candidate resolution
│   └── *_test.go
├── runtime/
│   ├── mcp_config.go          # local stdio config validation and redacted summaries
│   ├── mcp_discovery.go       # bounded list-tools discovery/session lifecycle
│   ├── mcp_tools.go           # MCP schema to read-only ToolSpec candidate mapping
│   ├── tools.go               # merge read-only candidates into ToolRegistry summaries
│   ├── pipeline.go            # include MCP availability in safe RunContext/pipeline metadata
│   └── *_test.go
└── httpapi/
    ├── runtime.go             # existing history/SSE exposure of safe discovery events if needed
    └── runtime_test.go

web/src/
├── realApiClient.ts                         # map safe MCP discovery metadata if exposed
├── runtime/
│   ├── realExecutionAdapter.ts              # replay discovery/availability events if exposed
│   └── runtimeEventGroups.ts                # Timeline/debug labels for MCP discovery
└── components/
    ├── RunTimeline.tsx                      # visible discovery labels if needed
    └── RunRail.tsx                          # debug detail surface if needed

docs-site/src/content/docs/
├── architecture/mcp-stdio-foundation.md
├── api/mcp-stdio-foundation.md              # or extend api/tool-call-approval.md
├── runbooks/local-m11-mcp.md
├── roadmap/current-status.md
├── spec-kit/workflow.md
└── devlog/2026-05-25-m11-mcp-stdio-foundation.md
```

**Structure Decision**: MCP config validation, discovery, and ToolSpec mapping belong near runtime/tool registry boundaries. Server configuration stays local-dev explicit config in this slice; only discovery safe status may cross into product data if replay needs it. Persona references remain product data. RunContext only receives safe availability summaries. Frontend work stays limited to existing Timeline/debug replay surfaces and is required for the visible M11 debug labels.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Support only explicit local stdio MCP configs.
- Treat env, args, command paths, stderr, tokens, and secret-looking paths as sensitive.
- Run bounded discovery/list-tools only; do not execute MCP tools.
- Map MCP tools to namespaced read-only ToolSpec candidates.
- Let Persona allowed tools reference MCP candidate names, but keep execution disabled by default.
- Attach safe MCP availability to RunContext and Timeline/debug.
- Document future execution as M7 approval + audit + redacted arguments/results.

## Phase 1: Design Summary

- [data-model.md](./data-model.md) defines MCP Server Config, Discovery Session, Tool Candidate, ToolSpec Mapping, Availability Summary, Safety Error, and Execution Boundary.
- [contracts/mcp-server-config.md](./contracts/mcp-server-config.md) defines local stdio config validation and safe summaries.
- [contracts/mcp-discovery-tool-mapping.md](./contracts/mcp-discovery-tool-mapping.md) defines discovery/list-tools parsing, namespacing, and ToolSpec mapping.
- [contracts/mcp-run-context-observability.md](./contracts/mcp-run-context-observability.md) defines RunContext and Timeline/debug safe metadata.
- [contracts/mcp-execution-boundary.md](./contracts/mcp-execution-boundary.md) documents what future MCP execution must satisfy before any invocation is allowed.
- [quickstart.md](./quickstart.md) defines validation commands and browser/debug smoke expectations.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. Quickstart validates config, discovery, ToolSpec mapping, RunContext availability, and Timeline/debug labels without execution.
- **Core Flow Before Platform Complexity**: PASS. No HTTP/SSE/OAuth, marketplace, sandbox, automation tools, new queue, or tool execution.
- **Observable Agent Execution**: PASS. Discovery success/failure and safety errors have event/debug contracts.
- **Safety/Data Boundaries**: PASS. Sensitive config/process fields are redacted before persistence/UI; discovery output is untrusted.
- **Documentation**: PASS. docs-site update targets and docs build are included in tasks.

## Complexity Tracking

No constitution violations. A small MCP discovery boundary is justified because M11 explicitly requires MCP stdio config, session lifecycle, discovery, schema mapping, and registry visibility. Execution is intentionally deferred to avoid permission and sandbox complexity in this slice.
