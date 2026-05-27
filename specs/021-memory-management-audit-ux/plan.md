# Implementation Plan: M14 Memory Management Audit UX

**Branch**: `[021-memory-management-audit-ux]` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/021-memory-management-audit-ux/spec.md`

## Summary

Enhance M13's user-control surface by making Settings > Memory a usable management view and adding safe memory audit/history backed by real productdata memory events. The implementation extends the existing Go memory API and React Settings surface, reuses M13 redaction/scope/tombstone behavior, and keeps distillation, embeddings/RAG, OpenViking, activity recorder ingestion, MCP, workers, sandboxing, and multi-agent rewrites out of scope.

## Technical Context

**Language/Version**: Go 1.23 backend; TypeScript/React/Vite web shell; Starlight docs site.

**Primary Dependencies**: Existing `internal/httpapi`, `internal/productdata`, M13 memory service/repository, `web/src/realApiClient.ts`, `web/src/components/SettingsView.tsx`, Bun test/build tooling.

**Storage**: Existing PostgreSQL memory tables plus M14 `memory_audit_events` for durable user-readable memory audit.

**Testing**: Go unit/integration tests, Bun web tests, Vite build, docs-site build, browser smoke in local web.

**Target Platform**: Local Loomi web/API development environment.

**Project Type**: Existing Go API plus React web shell and Starlight docs site.

**Performance Goals**: Settings > Memory list/history should load bounded current-user pages quickly for local development; no new production throughput target.

**Constraints**: Do not touch `.env*`, secrets, `.serena`, `.claude`, `.superpowers`, or `docs-site/.dev`. Do not implement distillation, embeddings/RAG, OpenViking, activity recorder ingestion, MCP, worker queue, sandbox, or multi-agent rewrites.

**Scale/Scope**: One M14 vertical slice. This session completed the blocker foundation for scoped list/search/detail/delete and durable safe audit; the remaining slice is the Settings > Memory UX.

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. This extends Loomi's own memory management language and UI.
- **Runnable Vertical Slices**: PASS. The slice is demonstrable in Settings > Memory and via API/tests.
- **Core Flow Before Platform Complexity**: PASS. The work stays on current memory/API/UI boundaries and avoids future memory platform complexity.
- **Observable Agent Execution**: PASS. Memory history makes write/snapshot/delete behavior observable to users.
- **Safety, Permissions, and Data Boundaries**: PASS. Scope authorization, no existence leak, redaction, tombstone, and confirmation are explicit gates.
- **Documentation Definition of Done**: PASS. Architecture/API/runbook/devlog/status/spec-kit docs update in the same work session.

## Project Structure

### Documentation (this feature)

```text
specs/021-memory-management-audit-ux/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── memory-management-api.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/httpapi/
├── memory.go
├── memory_test.go
└── memory_real_pg_smoke_test.go

internal/productdata/
├── memory.go
└── memory_service_test.go

web/src/
├── realApiClient.ts
├── realApiClient.test.ts
├── memory.test.ts
└── components/
    ├── SettingsView.tsx
    └── SettingsView.*.test.tsx

docs-site/src/content/docs/
├── architecture/memory-foundation.md
├── api/memory-foundation.md
├── runbooks/local-m13-memory.md
├── devlog/2026-05-25-m14-memory-management-audit-ux.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Extend existing M13 files instead of adding a new memory subsystem. API work stays in `internal/httpapi`/`internal/productdata`; UI work stays in the existing Settings view/client tests; docs amend current memory foundation pages.

## Complexity Tracking

No constitution violations.
