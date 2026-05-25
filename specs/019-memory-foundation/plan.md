# Implementation Plan: M13 Memory Foundation

**Branch**: `main` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/019-memory-foundation/spec.md`

## Summary

M13 Memory Foundation adds the first PG-backed memory slice: approved `memory_entries`, scoped `memory_search`, approval-gated `memory_write`, and a bounded safe memory snapshot attached to RunContext. The design prioritizes privacy, user deletion/control, redaction, auditability, and thin implementation slices. It plans a MemoryProvider boundary but implements only PG provider in v1, and explicitly defers embeddings/vector search/RAG, OpenViking, automated distillation, marketplace/plugin, sandbox/browser/activity recorder, multi-agent long-term memory automation, worker/job queue rewrite, and MCP rewrite.

## Technical Context

**Language/Version**: Go backend/runtime/API/worker integration; TypeScript/React frontend for minimal memory management and event replay; Starlight docs-site with Bun validation.

**Primary Dependencies**: Existing `internal/productdata` repository/service patterns, migrations, local identity/thread/run ownership, M9 RunContext/Pipeline preparation, existing run events/SSE/history replay, M7-style approval/audit concepts where applicable, frontend real API adapter and shell settings/management surfaces, docs-site.

**Storage**: PostgreSQL only for v1. Add `memory_entries` for approved/tombstoned memory and a write-proposal boundary for `memory_write` approval state. Search uses PG text/metadata matching and explicit scope filters; no vector DB, embedding, external provider, or RAG subsystem.

**Testing**: Backend tests for migrations/repository/search/scope/delete/write approval/redaction/RunContext snapshot/events. API contract tests for list/search/read/delete/propose/approve/deny. Frontend tests for minimal memory list/search/delete and event mapping if implemented. Docs validation with `bun run --cwd docs-site build`. Go/web full test suites are implementation-session tasks, not required for this design-only session.

**Target Platform**: Local Loomi development stack with Go API/runtime, local PostgreSQL, web renderer, and docs-site.

**Project Type**: Local web application plus Go API/backend runtime and durable product data.

**Performance Goals**: RunContext memory snapshot stays bounded by count and byte size; search returns paged results; PG queries use scope/status filters before text ranking; memory failures degrade to safe empty/error events without blocking unrelated run execution unless the implementation explicitly marks memory required.

**Constraints**: Approval-gated agent writes, immediate deletion exclusion, safe audit metadata, user scope enforcement, no raw sensitive data in events/API/UI/docs, no new external dependency for memory v1, no worker/job queue or MCP redesign.

**Scale/Scope**: First implementation slice targets one PG provider, scoped memory entries and proposals, one bounded RunContext snapshot per run, minimal memory management API/UI, and design-only provider/distill notes.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Mechanism Parity, Original Expression**: PASS. The design uses Loomi's own Memory/RunContext/audit terms and does not copy another product's expression layer.
- **II. Runnable Vertical Slices**: PASS. US1 is independently demonstrable through seeded PG memory and RunContext snapshot; US2 adds approval-gated writes; US3 adds user control.
- **III. Core Flow Before Platform Complexity**: PASS. Memory follows completed RunContext/Pipeline foundations and defers vector/RAG, OpenViking, sandbox, recorder, marketplace, MCP, and worker rewrites.
- **IV. Observable Agent Execution**: PASS. Search, snapshot load, proposed write, approval/denial, and deletion have safe event/audit surfaces.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Writes are approval-gated, deletion is user-controlled, memory content is untrusted data, and sensitive data is redacted before persistence/UI/RunContext.
- **Technical Constraints**: PASS. The plan reuses Go/PostgreSQL/web/docs boundaries and justifies the small MemoryProvider interface as a future-proof boundary with PG as the only v1 implementation.
- **Development Workflow**: PASS. Specs, plan, contracts, data model, tasks, checklist, quickstart, and docs-site planned status are included before implementation.
- **Documentation Definition of Done**: PASS. docs-site planned updates and docs build validation are included.

## Project Structure

### Documentation (this feature)

```text
specs/019-memory-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── memory-api.md
│   ├── memory-events.md
│   └── memory-provider.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── productdata/
│   ├── migrations/              # memory_entries and memory write proposal schema
│   ├── models.go                # memory entry/proposal/search/snapshot/audit models
│   ├── repository.go            # scoped PG memory persistence and search
│   ├── service.go               # authorization, approval, deletion, redaction orchestration
│   └── *_test.go
├── runtime/
│   ├── context_loader.go        # RunContext memory snapshot loading
│   ├── memory.go                # MemoryProvider interface and PG adapter boundary
│   ├── memory_redaction.go      # shared safe summary/redaction helpers
│   └── *_test.go
└── httpapi/
    ├── memory.go                # minimal memory list/search/read/delete/write approval APIs
    └── memory_test.go

web/src/
├── realApiClient.ts             # memory API calls
├── runtime/
│   ├── realExecutionAdapter.ts  # memory event replay if surfaced in timeline/debug
│   └── runtimeEventGroups.ts
└── components/
    └── MemoryPanel.tsx          # minimal planned UI boundary for list/search/delete

docs-site/src/content/docs/
├── architecture/memory-foundation.md        # planned/design-only until implemented
├── api/memory-foundation.md                 # planned/design-only until implemented
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Memory persistence and authorization live in `internal/productdata`; RunContext snapshot assembly lives in `internal/runtime`; minimal API boundaries live in `internal/httpapi`; frontend work is a thin management surface and event replay only. The MemoryProvider abstraction is planned in `internal/runtime` but PG is the only v1 implementation.

## Phase 0: Research Summary

Research is recorded in [research.md](./research.md). Key decisions:

- Use PG-first scoped text/metadata memory, not embeddings/vector/RAG.
- Approval-gate all agent-proposed writes before search/RunContext eligibility.
- Use tombstone deletion for immediate exclusion plus safe audit evidence.
- Treat memory as untrusted data in every prompt/context boundary.
- Keep MemoryProvider small and PG-only in v1.
- Defer distillation, OpenViking, automation, recorder, sandbox, marketplace, MCP, and worker rewrites.

## Phase 1: Design Summary

- [data-model.md](./data-model.md) defines Memory Entry, Search Request/Result, Write Proposal, Approval Decision, Snapshot, Tombstone, Audit Event, and Provider Boundary.
- [contracts/memory-api.md](./contracts/memory-api.md) defines minimal list/search/read/delete/propose/approve/deny API contracts.
- [contracts/memory-events.md](./contracts/memory-events.md) defines safe run/audit events and forbidden fields.
- [contracts/memory-provider.md](./contracts/memory-provider.md) defines PG v1 provider behavior and future provider constraints.
- [quickstart.md](./quickstart.md) defines design validation and implementation smoke expectations.

## Post-Design Constitution Check

- **Runnable Vertical Slice**: PASS. US1 can ship with seeded PG memories and RunContext snapshot before writes/UI.
- **Core Flow Before Platform Complexity**: PASS. PG text search and approval-gated writes avoid vector/RAG/provider expansion.
- **Observable Agent Execution**: PASS. Snapshot/search/write/delete states have event/audit contracts.
- **Safety/Data Boundaries**: PASS. Deletion, redaction, scope, audit, and untrusted-data treatment are explicit.
- **Documentation**: PASS. docs-site status updates are planned-only and build validation is required.

## Complexity Tracking

No constitution violations. The only new abstraction is `MemoryProvider`, constrained to a small interface so PG v1 remains concrete while future providers can be designed later without changing RunContext callers.
