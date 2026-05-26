# Implementation Plan: M30 Activity Recorder Foundation

**Branch**: `[038-activity-recorder-foundation]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

## Summary

M30 turns Activity Recorder from a placeholder into a safe foundation slice: disabled by default, explicit enable/disable, bounded redacted activity summaries, user-visible list/status, and idempotent cleanup. The slice avoids automatic OS capture and stores only local safe summaries.

## Technical Context

**Language/Version**: Go backend; TypeScript/React frontend; Astro/Starlight docs

**Primary Dependencies**: Existing productdata memory service patterns, HTTP server routing, redaction helpers, Settings panel, real/mock API clients

**Storage**: In-memory productdata activity recorder state and activity summaries for this foundation slice

**Testing**: TDD with `go test ./internal/productdata`, `go test ./internal/httpapi`, targeted `bun test --cwd web`, then full closeout commands

**Target Platform**: Local web/API development environment

**Project Type**: Web app with Go API and React frontend

**Performance Goals**: Bounded list operations with default limit 20 and hard cap 100

**Constraints**: disabled by default, explicit opt-in, redaction before persistence, no screenshots/keystrokes/clipboard/raw browser/raw shell/file contents/full paths, idempotent cleanup

**Scale/Scope**: One foundation slice for safe local activity summary state, not full desktop activity capture

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. Uses Loomi-owned activity summary vocabulary and does not copy external expression.
- **Runnable Vertical Slices**: PASS. API and Settings UI can demonstrate enable -> append -> list -> disable/clear.
- **Core Flow Before Platform Complexity**: PASS. Comes after tool runtime, sandbox, browser, artifact, and multi-agent foundations.
- **Observable Agent Execution**: PASS. Recorder state and summaries are inspectable in Settings and can be tested through API.
- **Safety, Permissions, and Data Boundaries**: PASS. Explicit opt-in, redaction, no raw desktop capture, and cleanup path are core requirements.

## Project Structure

```text
specs/038-activity-recorder-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── activity-recorder.md
└── tasks.md
```

Source changes target:

```text
internal/productdata/
internal/httpapi/
web/src/
docs-site/src/content/docs/
```

## Complexity Tracking

No constitution violations.
