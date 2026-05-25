# Implementation Plan: M13.5 Memory Real PG Smoke Closeout

**Branch**: `[020-memory-real-pg-smoke-closeout]` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/020-memory-real-pg-smoke-closeout/spec.md`

## Summary

Close M13 by adding evidence rather than new capability: real Postgres migration/API smoke coverage for memory lifecycle, status cleanup in `019-memory-foundation`, and docs-site evidence/runbook updates. The code change is limited to one integration smoke test that exercises the already-implemented repository and HTTP handlers.

## Technical Context

**Language/Version**: Go 1.23 backend; TypeScript/React web shell; Starlight docs site.

**Primary Dependencies**: Existing `pgxpool`, `productdata.PostgresRepository`, `httpapi.Server`, explicit `migrate` SQL workflow, existing Bun/Vite/docs tooling.

**Storage**: Existing local PostgreSQL database migrated through `migrations/000009_m13_memory_foundation.up.sql`.

**Testing**: Go real PG smoke gated by `LOOMI_TEST_DATABASE_URL`, full `go test ./...`, web Bun tests/build, docs-site build, `git diff --check`.

**Target Platform**: Local development on the existing Loomi web/API workspace.

**Project Type**: Existing Go API plus React web shell and Starlight docs site.

**Performance Goals**: Smoke completes within normal integration-test latency on local Postgres; no production throughput target is introduced.

**Constraints**: Do not touch `.env*`, secrets, `.serena`, `.claude`, `.superpowers`, or `docs-site/.dev`. Do not introduce new memory platform features.

**Scale/Scope**: One closeout/evidence slice for M13. No vector DB, embedding, RAG, OpenViking, automatic distill, activity recorder, sandbox, MCP rewrite, worker rewrite, or multi-agent memory automation.

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. This is internal closeout evidence for Loomi behavior.
- **Runnable Vertical Slices**: PASS. The closeout adds an executable real PG/httpapi smoke and docs evidence.
- **Core Flow Before Platform Complexity**: PASS. The work proves the existing PG/RunContext/API slice and explicitly avoids platform expansion.
- **Observable Agent Execution**: PASS. RunContext snapshot and memory audit events are included in smoke evidence.
- **Safety, Permissions, and Data Boundaries**: PASS. Redaction, tombstone, idempotency, and out-of-scope non-leakage are tested.
- **Documentation Definition of Done**: PASS. docs-site runbook/devlog/status updates and docs build validation are required.

## Project Structure

### Documentation (this feature)

```text
specs/020-memory-real-pg-smoke-closeout/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── real-pg-httpapi-smoke.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/httpapi/
└── memory_real_pg_smoke_test.go

specs/019-memory-foundation/
├── spec.md
└── contracts/

docs-site/src/content/docs/
├── api/memory-foundation.md
├── architecture/memory-foundation.md
├── devlog/
├── roadmap/current-status.md
├── runbooks/
└── spec-kit/workflow.md
```

**Structure Decision**: The smoke belongs in `internal/httpapi` because it proves the HTTP handler path while sharing the real `productdata.PostgresRepository` used by the API process. Spec and docs changes are status/evidence updates only.

## Complexity Tracking

No constitution violations.
