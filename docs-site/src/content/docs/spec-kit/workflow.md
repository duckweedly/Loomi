---
title: Spec Kit Workflow Status
description: Loomi feature workflow status and implementation notes.
---

## M4 Run/Event/SSE

Spec Kit artifacts live in `specs/003-m4-run-event-sse/`:

- `spec.md` defines the local run/event/SSE user stories and deferred capabilities.
- `plan.md` selects deterministic local simulation, PostgreSQL persistence, and history-first SSE.
- `data-model.md` defines Run, Run Event, Event Stream Cursor, Stop Request, Deterministic Local Simulation, Stream State, and M4 Schema Revision.
- `contracts/` contains HTTP, SSE, migration, and frontend data-source contracts.
- `tasks.md` tracks implementation status.

Implementation status on 2026-05-23: core M4 run/event/SSE slice is implemented in the feature worktree. Remaining validation is recorded in the M4 devlog as commands are run.

## Workflow rule

For future non-trivial Loomi work, follow constitution order: specify, clarify, plan, tasks, optional analyze, then implement. Documentation updates are part of the same implementation session whenever code changes architecture, API, data model, runtime behavior, workflow, UI flow, or safety boundaries.
