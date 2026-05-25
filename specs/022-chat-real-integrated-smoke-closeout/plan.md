# Implementation Plan: M15 Chat Real Integrated Smoke Closeout

**Branch**: `[017-mcp-approval-gated-execution]` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/022-chat-real-integrated-smoke-closeout/spec.md`

## Summary

M15 adds a gated, deterministic closeout smoke proving Chat's real backend path from HTTP run creation through worker execution, approved memory RunContext snapshot, MCP approval/execution, provider continuation, final assistant message, and persisted replay evidence. The implementation reuses existing Go API/service/worker/runtime boundaries and adds only fixture/test/runbook evidence needed to make the slice repeatable.

## Technical Context

**Language/Version**: Go 1.x backend, TypeScript/React/Bun frontend/docs

**Primary Dependencies**: existing Go stdlib/http stack, productdata repository/service, runtime provider/worker/MCP components, Bun/Vite/Starlight

**Storage**: existing test repository path used by API/runtime tests; no new durable storage

**Testing**: Go tests with explicit M15 smoke gate, existing Bun web tests/build, docs-site build

**Target Platform**: local development backend and web shell

**Project Type**: web-service plus web/desktop-feeling shell

**Performance Goals**: smoke completes in seconds locally and is deterministic across repeated runs

**Constraints**: no external paid model dependency; no new sandbox/filesystem/shell/browser automation tools; no worker queue rewrite; no sensitive fixture values in shareable surfaces

**Scale/Scope**: one integrated happy-path closeout/evidence scenario plus redaction and replay assertions

## Constitution Check

**I. Mechanism Parity, Original Expression**: Pass. The slice is Loomi-specific backend evidence and does not copy external product expression.

**II. Runnable Vertical Slices**: Pass. The feature is defined by one runnable smoke command and optional browser evidence.

**III. Core Flow Before Platform Complexity**: Pass. Reuses Chat/API/run/event/worker/context/MCP/memory foundations and explicitly excludes sandbox, activity recorder, RAG, marketplace, multi-agent, and queue rewrites.

**IV. Observable Agent Execution**: Pass. Required evidence is persisted events plus replay/timeline visibility.

**V. Safety, Permissions, and Data Boundaries**: Pass. Tool execution remains approval-gated; fixture canaries verify redaction across API, RunContext, events, tool summaries, docs, and optional UI.

## Project Structure

### Documentation (this feature)

```text
specs/022-chat-real-integrated-smoke-closeout/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
internal/httpapi/
├── chat_real_integrated_smoke_test.go
└── existing API handlers and runtime tests

internal/runtime/
├── existing deterministic provider, runner, worker, MCP, memory helpers
└── focused helper/test adjustments only if current boundaries cannot expose evidence

internal/productdata/
├── existing repository/service memory and event helpers
└── focused helper/test adjustments only if current boundaries cannot expose evidence

web/
└── existing real API timeline mapping/tests when replay behavior needs coverage

docs-site/src/content/docs/
├── devlog/
├── runbooks/
├── roadmap/current-status.md
├── spec-kit/workflow.md
└── api/ or architecture/ only for behavior differences
```

**Structure Decision**: Use the existing backend/API/runtime/productdata boundaries and add an explicit gated smoke test. Frontend code changes are only needed if replayed event types are not already visible. Docs-site updates are required for closeout evidence.

## Complexity Tracking

No constitution violations or new complexity exceptions.
