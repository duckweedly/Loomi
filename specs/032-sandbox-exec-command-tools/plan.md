# Implementation Plan: M24 Sandbox Exec Command Tools

**Branch**: `[032-sandbox-exec-command-tools]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/032-sandbox-exec-command-tools/spec.md`

## Summary

M24 adds the first approval-gated command execution tool after workspace read/mutation tools. The slice introduces `sandbox.exec_command` as argv-form, Work-mode-only, approval-required, timeout/output-bounded execution under the configured workspace root. It reuses ToolCatalog, RunContext enabled snapshots, ToolBroker, tool-call approval, worker resume, run events, Settings, and RunRail. The feature explicitly avoids shell sessions, browser/web/artifact/plugin/multi-agent behavior.

## Technical Context

**Language/Version**: Go backend; TypeScript/React frontend; Astro/Starlight docs

**Primary Dependencies**: Existing Go stdlib process APIs, existing productdata/runtime/httpapi services, existing React Settings/RunRail components

**Storage**: Existing run events and tool_calls projections; no new tables for this slice

**Testing**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, browser smoke for catalog/timeline visibility

**Target Platform**: Local macOS/Darwin development first with localhost-compatible API/web commands

**Project Type**: Go API/worker plus web/desktop-feeling shell

**Performance Goals**: Commands must respect configured timeout and output bounds; unsafe validation must happen before spawning

**Constraints**: Work-mode only; approval required; argv only; relative cwd under workspace root; no model-supplied env; no destructive command patterns; bounded stdout/stderr; safe event metadata

**Scale/Scope**: First command execution slice for small local validation commands, not long-running terminal sessions or sandbox container orchestration

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. The feature uses Loomi-owned sandbox exec vocabulary and avoids copying product expression.
- **Runnable Vertical Slices**: PASS. US1 defines an approve -> execute -> continue smoke; US2 covers safety denial; US3 covers visible audit trail.
- **Core Flow Before Platform Complexity**: PASS. This builds on existing tool approval, worker, RunContext, catalog, and mutation foundations; browser/web/artifact/multi-agent runtime remain excluded.
- **Observable Agent Execution**: PASS. Commands go through persisted tool lifecycle events and visible Settings/RunRail surfaces.
- **Safety, Permissions, and Data Boundaries**: PASS. Execution is approval-gated, bounded, cwd-scoped, redacted, and denies destructive patterns.

## Project Structure

### Documentation (this feature)

```text
specs/032-sandbox-exec-command-tools/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── sandbox-exec-command.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/productdata/
├── models.go
├── builtin_personas.go
├── tool_catalog.go
├── tool_catalog_test.go
└── service_test.go

internal/runtime/
├── tools.go
├── tools_test.go
├── tool_broker.go
├── tool_broker_test.go
├── sandbox_tools.go
├── sandbox_tools_test.go
├── gateway.go
├── gateway_test.go
└── worker_test.go

internal/httpapi/
└── sandbox_exec_command_smoke_test.go

web/src/
├── components/SettingsView.tools.test.tsx
├── components/RunRail.runtime.test.ts
├── components/RunRail.tsx
├── mockApiClient.ts
└── mockData.ts

docs-site/src/content/docs/
├── api/sandbox-exec-command.md
├── architecture/sandbox-exec-command.md
├── runbooks/local-m24-sandbox-exec-command.md
├── devlog/2026-05-26-m24-sandbox-exec-command.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Extend the existing M18/M21/M23 tool runtime boundaries rather than introducing a separate command subsystem.

## Complexity Tracking

No constitution violations.
