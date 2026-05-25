# Implementation Plan: M17 Work Artifact Evidence Closeout

**Branch**: `[017-mcp-approval-gated-execution]` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/024-work-artifact-evidence-closeout/spec.md`

## Summary

M17 closes out Work artifact evidence by moving the M16 Work Plan View from mock-only validation to repeatable local evidence. The slice adds a local-dev/test-only seed scenario that uses existing product service methods to create/reuse a Work thread and message, start or reuse a current run, and append safe work metadata events. The frontend keeps projecting from the existing thread/message/run/event path and tightens artifact cards with redaction markers and no executable controls.

## Technical Context

**Language/Version**: Go backend/seed command, TypeScript/React/Bun frontend and docs

**Primary Dependencies**: Existing productdata service/repository, loomi-seed command, real API client, ChatCanvas, WorkPlanView, Starlight docs site

**Storage**: Existing PostgreSQL or in-memory test service; no new tables

**Testing**: Go tests for local seed evidence, Bun web tests for projection/rendering/mode isolation, web build, docs-site build, browser smoke

**Target Platform**: Local API plus local web shell

**Project Type**: Web-first agent platform with Go API/runtime and React frontend

**Performance Goals**: Seed and projection operate on one local thread/current run and a small recent event list

**Constraints**: Reuse `Thread.mode = work`; no production event-write endpoint; no task system, sandbox, artifact runtime, shell/filesystem/browser automation, activity recorder, multi-agent, plugin marketplace, or worker queue rewrite

**Scale/Scope**: One repeatable M17 Work evidence thread, one current run, one safe work metadata event payload, browser smoke evidence

## Constitution Check

**I. Mechanism Parity, Original Expression**: Pass. M17 expresses Loomi Work evidence in project vocabulary and does not copy external expression.

**II. Runnable Vertical Slices**: Pass. The slice is validated through local seed, real API mode, browser smoke, and focused tests.

**III. Core Flow Before Platform Complexity**: Pass. It reuses thread/message/run/event and avoids artifact runtime, sandbox, activity recorder, multi-agent, and queue rewrites.

**IV. Observable Agent Execution**: Pass. Work progress is observable through persisted/replayed run events.

**V. Safety, Permissions, and Data Boundaries**: Pass. Seeded artifacts are metadata-only; unsafe fields are redacted/omitted and no executable controls are added.

## Project Structure

### Documentation (this feature)

```text
specs/024-work-artifact-evidence-closeout/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
cmd/loomi-seed/
├── main.go
└── main_test.go

web/src/
├── domain.ts
├── workModeProjection.ts
├── workModeProjection.test.ts
└── components/
    ├── WorkPlanView.tsx
    └── WorkPlanView.test.tsx

docs-site/src/content/docs/
├── architecture/work-mode-foundation.md
├── api/work-mode-foundation.md
├── runbooks/local-m17-work-artifact-smoke.md
├── devlog/2026-05-25-m17-work-artifact-evidence-closeout.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Keep Work projection in the existing frontend path and add evidence creation only to `cmd/loomi-seed` behind an explicit local seed scenario. Do not add an HTTP event-write endpoint or new storage model.

## Complexity Tracking

No constitution violations or new complexity exceptions.
