# Implementation Plan: M12 Real MCP Smoke Closeout

**Branch**: `018-m12-real-mcp-smoke-closeout` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/018-m12-real-mcp-smoke-closeout/spec.md`

## Summary

Add M12.5 closeout evidence by exercising the existing M12 local stdio MCP approval path end to end: real `Content-Length` discovery fixture, persona-allowed provider tool request, approval-required block, scoped HTTP approve, worker resume with `StdioMCPToolExecutor` loaded from `LOOMI_MCP_SERVERS_JSON`, redacted result, continuation, and final message. Update docs-site to record the evidence and keep all broader MCP platform capabilities out of scope.

## Technical Context

**Language/Version**: Go 1.24 backend tests; TypeScript/React/Bun frontend and docs validation

**Primary Dependencies**: Existing `internal/runtime`, `internal/productdata`, `internal/httpapi`, Vite/React web app, Starlight docs site

**Storage**: In-memory product service for smoke evidence; no new persistence

**Testing**: Go tests, Bun frontend tests, web build, docs build, diff whitespace check

**Target Platform**: Local development runtime

**Project Type**: Web app with Go API/worker backend

**Performance Goals**: Fixture smoke completes within bounded MCP timeout and normal test runtime

**Constraints**: No remote MCP, marketplace/plugin install, sandbox, shell/filesystem/browser automation, multi-tool loop, new MCP admin UI, or new platform capability

**Scale/Scope**: One closeout feature directory, one narrow local smoke path, docs evidence updates

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Mechanism Parity, Original Expression**: Pass. This is Loomi's own local validation evidence and docs.
- **Runnable Vertical Slices**: Pass. The feature exists only to add a runnable end-to-end smoke.
- **Core Flow Before Platform Complexity**: Pass. It validates the existing M12 slice and explicitly defers broader MCP platform work.
- **Observable Agent Execution**: Pass. The smoke checks persisted events and provider continuation.
- **Safety, Permissions, and Data Boundaries**: Pass. Approval, redaction, local-only config, and non-goals are the feature boundaries.

## Project Structure

### Documentation (this feature)

```text
specs/018-m12-real-mcp-smoke-closeout/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── evidence-chain.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/httpapi/
├── runtime_test.go
└── mcp_real_smoke_test.go

docs-site/src/content/docs/
├── runbooks/local-m12-mcp-approval-execution.md
├── devlog/2026-05-25-m12-mcp-approval-gated-execution.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Keep implementation in existing runtime/httpapi boundaries and documentation pages; the closeout smoke lives in `internal/httpapi` so it can exercise the scoped HTTP approve endpoint while still using the real runtime worker and stdio executor.

## Complexity Tracking

No constitution violations.
