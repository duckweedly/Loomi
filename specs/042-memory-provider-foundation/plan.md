# Implementation Plan: Memory Provider Foundation

**Branch**: `[042-memory-provider-foundation]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/042-memory-provider-foundation/spec.md`

## Summary

Add the first Arkloop-class memory mechanism slice in Loomi terms: a backend-owned memory provider configuration/status contract, safe runtime readiness metadata, and Settings visibility. This slice keeps the current local approved-memory store as the default provider and deliberately stops before agent-facing memory tools, external semantic write/read integration, and automatic post-run distillation.

## Technical Context

**Language/Version**: Go backend, TypeScript/React frontend, Astro/Starlight docs

**Primary Dependencies**: Existing `internal/productdata`, `internal/httpapi`, `internal/runtime`, Vite React app, Bun docs toolchain

**Storage**: Existing product data repository plus in-memory service fallback; no new external memory store in this slice

**Testing**: `go test` focused packages, Bun/Vitest focused web tests, browser smoke for Settings UI, `bun run build` for docs

**Target Platform**: Local Loomi API + desktop-feeling web shell

**Project Type**: Web app with Go API/runtime backend

**Performance Goals**: Provider status resolution is bounded and non-blocking for run preparation; status refresh must not make existing local memory list/search slower in normal Settings usage

**Constraints**: No secret leakage in API/UI/run metadata; unknown provider values must degrade safely; no Arkloop brand/copy/visual/private names; no memory tools or automatic distillation in this slice

**Scale/Scope**: One Settings > Memory provider foundation, one backend status API, one runtime readiness projection, existing memory management preserved

## Constitution Check

- Mechanism Parity, Original Expression: PASS. The feature implements provider/status mechanics using Loomi provider names and copy; no external brand or UI text is copied.
- Runnable Vertical Slices: PASS. US1 can be validated independently through Settings and backend status; US2 through deterministic provider health projections; US3 through run context/readiness tests; US4 through existing memory tests.
- Core Flow Before Platform Complexity: PASS. This slice reuses existing memory and runtime boundaries and does not introduce external semantic storage, background workers, or multi-provider tool execution.
- Observable Agent Execution: PASS. Run readiness is a safe summary that future run events/tools can reuse.
- Safety, Permissions, and Data Boundaries: PASS. Provider diagnostics are redacted and existing memory deletion/audit controls remain the durable boundary.

## Project Structure

### Documentation (this feature)

```text
specs/042-memory-provider-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── memory-provider-status.openapi.yaml
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── productdata/      # memory provider config/status models and persistence/service tests
├── httpapi/          # provider status/config endpoints and response tests
└── runtime/          # safe memory readiness projection for run preparation

web/src/
├── components/       # Settings > Memory provider UI
├── realApiClient.ts  # backend mapping
├── apiTypes.ts       # frontend-facing provider types
└── *.test.ts(x)      # focused Settings/API tests

docs-site/src/content/docs/
├── architecture/
├── api/
├── runbooks/
└── devlog/
```

**Structure Decision**: Reuse existing productdata/httpapi/runtime/web/docs-site boundaries. Do not add a new provider framework package until tools/distillation need provider polymorphism beyond config/status.

## Phase 0 Research

See [research.md](./research.md).

## Phase 1 Design

See [data-model.md](./data-model.md), [contracts/memory-provider-status.openapi.yaml](./contracts/memory-provider-status.openapi.yaml), and [quickstart.md](./quickstart.md).

## Complexity Tracking

No constitution violations.
