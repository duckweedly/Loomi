# Implementation Plan: M29 Multi-agent Runtime Foundation

**Branch**: `[037-multi-agent-runtime-foundation]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

## Summary

M29 adds a safe multi-agent coordination foundation through three builtin tools: `agent.spawn`, `agent.list`, and `agent.complete`. The slice reuses ToolCatalog, RunContext, approval, ToolBroker, worker continuation, Settings, and RunRail. It intentionally avoids autonomous sub-agent execution, external processes, parallel worker pools, remote orchestration, and cross-thread delegation.

## Technical Context

**Language/Version**: Go backend; TypeScript/React frontend; Astro/Starlight docs

**Primary Dependencies**: Existing productdata service interfaces, PostgresRepository, ToolBroker, worker approval resume path, React Settings/RunRail components

**Storage**: Productdata agent task records backed by both in-memory service and PostgreSQL `agent_tasks` table, with safe run-event summaries

**Testing**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, browser smoke for Settings Tools and RunRail agent lifecycle

**Constraints**: Work-mode only, approval required, bounded role/goal/result fields, no autonomous execution, no external processes, no network/filesystem side effects

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. Uses Loomi-owned child-task coordination vocabulary.
- **Runnable Vertical Slices**: PASS. The slice has spawn/list/complete backend proof and visible frontend state.
- **Core Flow Before Platform Complexity**: PASS. Adds explicit coordination records before autonomous multi-agent execution.
- **Observable Agent Execution**: PASS. Agent task requests/results go through run events and RunRail.
- **Safety, Permissions, and Data Boundaries**: PASS. Agent tools are approval-gated, bounded, Work-mode only, and non-autonomous.

## Project Structure

```text
specs/037-multi-agent-runtime-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── multi-agent-runtime.md
└── tasks.md
```

Source changes target:

```text
internal/productdata/
internal/runtime/
internal/httpapi/
web/src/components/
web/src/runtime/
docs-site/src/content/docs/
```

## Complexity Tracking

No constitution violations.
