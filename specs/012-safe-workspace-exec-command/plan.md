# Implementation Plan: M10 Safe Workspace Exec Command

**Branch**: `012-safe-workspace-exec-command` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/012-safe-workspace-exec-command/spec.md`

## Summary

M10 adds `workspace.exec_command`, the first approval-gated command execution tool. It reuses tool-call approval, worker resume, run events, SSE replay, and frontend tool UI. The implementation runs argv commands without a shell, inside a workspace-contained cwd, with timeout and bounded output.

## Technical Context

**Language/Version**: Go 1.23 backend/runtime; TypeScript/React/Vite frontend; Bun for frontend/docs.

**Primary Dependencies**: Go standard library `os/exec`, `context`, workspace path helpers, existing productdata tool-call lifecycle and runtime worker. No new dependency is required.

**Storage**: Reuse existing tool_calls, run events, and background jobs. No migration expected.

**Testing**: TDD required. Add runtime tests for validation, safe command, dangerous command rejection, cwd escape rejection, timeout, and output truncation. Add productdata tests for tool-name/argument validation. Add worker tests for approved command terminal events. Add frontend tests for readable exec summaries.

**Constraints**: Approval required; no shell; no PTY; bounded timeout; bounded output; cwd inside workspace; reject destructive command names; no persistent sessions or background daemons.

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS.
- **Runnable Vertical Slices**: PASS. Safe command request -> approval -> execution -> terminal event -> UI replay.
- **Core Flow Before Platform Complexity**: PASS. Follows read/write tools before sandbox, MCP, browser, and multi-agent.
- **Observable Agent Execution**: PASS. Command lifecycle remains persisted and replayable.
- **Safety, Permissions, and Data Boundaries**: PASS. Approval, bounded execution, command rejection, timeout, and redaction are mandatory.
- **Documentation Definition of Done**: PASS.
