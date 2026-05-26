# Implementation Plan: M9 Safe Workspace Write Tools

**Branch**: `011-safe-workspace-write-tools` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/011-safe-workspace-write-tools/spec.md`

## Summary

M9 extends the approval-gated workspace tool boundary with the first text mutation tools: `workspace.write_file` and `workspace.edit`. It reuses M7 approval, M8 workspace root validation shape, worker resume, run events, SSE replay, and existing frontend tool UI. The slice stays deliberately smaller than shell execution or patch orchestration.

## Technical Context

**Language/Version**: Go 1.23 for backend/runtime/worker execution; TypeScript/React/Vite in `web/`; Bun for frontend/docs commands.

**Primary Dependencies**: Go standard library filesystem/path/text utilities; existing productdata tool-call lifecycle; existing runtime worker; existing ToolCallCard/RunRail/Timeline mapping. No new dependency is required.

**Storage**: Reuse existing `tool_calls`, run events, and background jobs. No migration expected.

**Testing**: TDD required. Add runtime tests for write/edit validation, root containment, sensitive path denial, symlink escape denial, exact-match edit behavior, and no-mutation failure cases. Add productdata tests for tool name/argument validation and approval-required behavior. Add worker test for approved write/edit terminal events. Add frontend tests for readable write/edit summaries.

**Target Platform**: Local macOS/Darwin development with the API process working directory as the development workspace root.

**Constraints**: Approval required; text-only; bounded content; no shell, network, MCP, browser automation, external upload, or broad directory creation. Writes must be auditable through existing tool-call events.

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. Tool names and UX copy use Loomi's own domain language.
- **Runnable Vertical Slices**: PASS. Each write/edit tool can be requested, approved, executed, replayed, and displayed.
- **Core Flow Before Platform Complexity**: PASS. M9 follows M7 approval and M8 read tools; shell/exec/MCP/browser/multi-agent remain deferred.
- **Observable Agent Execution**: PASS. All states remain persisted run events and visible frontend replay.
- **Safety, Permissions, and Data Boundaries**: PASS. Approval, root containment, sensitive path denial, bounded content, and no-mutation failures are mandatory.
- **Documentation Definition of Done**: PASS. Architecture, API, runbook, devlog, roadmap, and spec-kit docs must be updated during implementation.

## Project Structure

```text
specs/011-safe-workspace-write-tools/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── workspace-write-tools.md
│   └── docs-update-plan.md
└── tasks.md
```

```text
internal/runtime/
├── tools.go
├── tools_test.go
├── worker.go
└── worker_test.go

internal/productdata/
├── models.go
└── service_test.go

web/src/
├── runtime/executionAdapter.test.ts
├── components/ToolCallCard.tsx
└── components/ToolCallCard.test.tsx
```

## Complexity Tracking

No constitution violations. The mutation primitive is exact text replacement and full text write only; generalized patch parsing and command execution are intentionally deferred.
