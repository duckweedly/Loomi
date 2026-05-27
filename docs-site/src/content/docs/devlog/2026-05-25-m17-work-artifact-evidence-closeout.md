---
title: 2026-05-25 M17 Work Artifact Evidence Closeout
description: Work mode 从 mock-only projection 推进到可重复 local evidence seed 和真实浏览器 smoke。
---

## Completed

- Added `specs/024-work-artifact-evidence-closeout/` with spec, plan, research, data model, contract, quickstart, checklist, and tasks.
- Added explicit `LOOMI_SEED_SCENARIO=m17-work-artifact` local-dev/test seed path in `cmd/loomi-seed`.
- Seed path creates or reuses a Work thread/message, starts or reuses current run, and appends repeatable `work.plan.updated` metadata event.
- Added seed evidence tests for Work thread mode, run/event metadata, artifact metadata, and idempotency.
- Added `redactionApplied` artifact projection and UI marker.
- Added frontend tests for M17 seeded projection, safe artifact metadata, no executable artifact controls, event replay, and Chat/Work isolation.
- Updated Work mode architecture/API docs, local runbook, current status, and Spec Kit workflow.

## Scope Kept Out

- No artifact execution/runtime.
- No sandbox.
- No shell/filesystem/browser automation tools.
- No activity recorder.
- No multi-agent runtime.
- No plugin marketplace.
- No new task system.
- No worker queue rewrite.
- No production event-write HTTP API.

## Validation

Final command results and browser smoke evidence are recorded in the session final report for M17.
