# Implementation Plan: M8 Safe Workspace Read Tools

**Branch**: `010-safe-workspace-read-tools` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/010-safe-workspace-read-tools/spec.md`

## Summary

M8 extends the M7 approval-gated tool boundary with the first local code-agent read tools: `workspace.glob`, `workspace.grep`, and `workspace.read_file`. The implementation reuses the existing tool-call lifecycle, approval APIs, worker resume path, run events, SSE replay, ToolCallCard, RunRail, and Timeline. It adds only read-only workspace execution with strict root containment, sensitive path rejection, bounded results, and redaction.

## Technical Context

**Language/Version**: Go 1.23 for API/runtime/worker execution; TypeScript/React/Vite in `web/`; Bun for frontend/docs commands.

**Primary Dependencies**: Existing M7 productdata/runtime/httpapi tool-call flow; Go standard library filesystem/path matching; existing frontend runtime mapping and tool UI. No new runtime dependency is required.

**Storage**: Reuse existing `tool_calls` projection and run events. No migration is expected unless the current tool schema requires extending allowlisted tool names in persisted validation.

**Testing**: TDD required. Add backend tests for validation, sensitive path denial, bounded glob/grep/read results, approval execution, idempotency, cancellation, and history replay. Add frontend tests for workspace tool event mapping and UI labels. Validate with `go test ./...`, `bun test --cwd web`, web build, docs build, `git diff --check`, and browser smoke.

**Target Platform**: Local macOS/Darwin development with the local repo as workspace root.

**Constraints**: Read-only; approval required; no writes, shell, network, MCP, browser automation, hidden capture, or external upload. All paths are relative to workspace root in result payloads. Sensitive files are denied even when inside the root.

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. The feature uses Loomi's own tool names and execution contracts.
- **Runnable Vertical Slices**: PASS. Each tool can be approved, executed, replayed, and inspected in UI.
- **Core Flow Before Platform Complexity**: PASS. This follows M7 approval and stays inside safe read tools before write/exec/MCP/browser/multi-agent features.
- **Observable Agent Execution**: PASS. All states use persisted tool-call run events and existing UI replay.
- **Safety, Permissions, and Data Boundaries**: PASS. Approval, root containment, sensitive path denial, bounded output, and redaction are mandatory.
- **Documentation Definition of Done**: PASS. Architecture, API, runbook, devlog, roadmap, and spec-kit docs must be updated during implementation.

## Project Structure

```text
specs/010-safe-workspace-read-tools/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── workspace-read-tools.md
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
├── service.go
└── service_test.go

internal/httpapi/
└── runtime_test.go

web/src/
├── domain.ts
├── runtime/executionAdapter.ts
├── runtime/executionAdapter.test.ts
├── components/ToolCallCard.tsx
├── components/ToolCallCard.test.tsx
├── components/RunRail.tsx
└── components/RunRail.polish.test.ts
```

## Complexity Tracking

No constitution violations. The only new execution logic is a small internal workspace reader in `internal/runtime`; existing approval and event machinery remains the safety boundary.
