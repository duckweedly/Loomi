# Implementation Plan: M16 Work Mode Foundation

**Branch**: `[017-mcp-approval-gated-execution]` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/023-work-mode-foundation/spec.md`

## Summary

M16 makes Work mode minimally useful by adding a read-only Work Plan View for `Thread.mode = work`. The slice derives goal, steps, status, artifact references, and recent progress from existing messages, runs, and run events. It adds safe frontend projection and documentation only; no new backend API, worker queue, sandbox, file execution, browser automation, shell tools, or activity recorder.

## Technical Context

**Language/Version**: TypeScript/React/Bun frontend and docs; Go backend unchanged unless validation reveals an existing payload gap

**Primary Dependencies**: Existing React shell, Vite/Bun tests/build, current thread/message/run/event data model, Starlight docs site

**Storage**: Existing in-memory/mock and real API thread/run/event data; no new durable storage

**Testing**: Bun web tests for projection/rendering/isolation/redaction/event replay, Bun web build, docs-site build, git diff check, browser smoke

**Target Platform**: Local web shell

**Project Type**: web/desktop-feeling shell with existing backend API boundaries

**Performance Goals**: Work projection is synchronous over the selected thread's current messages and run events and remains small enough for immediate render

**Constraints**: Reuse `Thread.mode = work`; no new task system or execution environment; no file/shell/browser execution; no worker queue rewrite; redacted safe metadata only

**Scale/Scope**: One selected work thread, one current run, recent event preview, metadata-only artifact references

## Constitution Check

**I. Mechanism Parity, Original Expression**: Pass. M16 expresses Loomi's own Work mode vocabulary and avoids copying external product expression.

**II. Runnable Vertical Slices**: Pass. The feature is browser-visible and validated by focused tests plus smoke.

**III. Core Flow Before Platform Complexity**: Pass. It reuses existing thread/message/run/event foundations and explicitly excludes sandbox, activity recorder, multi-agent, marketplace, and queue rewrites.

**IV. Observable Agent Execution**: Pass. Work progress is projected from persisted/replayed run events.

**V. Safety, Permissions, and Data Boundaries**: Pass. Artifact references are metadata-only and redact secret-looking values; no execution controls are introduced.

## Project Structure

### Documentation (this feature)

```text
specs/023-work-mode-foundation/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
web/src/
├── domain.ts
├── mockData.ts
├── workModeProjection.ts
├── workModeProjection.test.ts
└── components/
    ├── ChatCanvas.tsx
    ├── WorkPlanView.tsx
    └── WorkPlanView.test.tsx

docs-site/src/content/docs/
├── architecture/work-mode-foundation.md
├── api/work-mode-foundation.md
├── runbooks/local-m16-work-mode.md
├── devlog/2026-05-25-m16-work-mode-foundation.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Add the Work Plan View as a frontend projection inside the existing ChatCanvas path for Work mode threads. Keep ChatCanvas, message history, run timeline, and right drawer behavior intact. Backend remains unchanged unless an existing API gap blocks the minimum projection.

## Complexity Tracking

No constitution violations or new complexity exceptions.
