# Implementation Plan: Local Provider Autodetect Foundation

**Branch**: `026-local-provider-autodetect-foundation` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/026-local-provider-autodetect-foundation/spec.md`

## Summary

Add a conservative local provider autodetect slice that detects whether Claude Code and Codex appear locally available, exposes only redacted safe capability summaries through a read-only API, and shows those summaries in Settings > Providers without enabling or using them for model calls.

## Technical Context

**Language/Version**: Go backend; React + TypeScript frontend; Starlight docs

**Primary Dependencies**: Go standard library JSON/filepath/os APIs, existing `internal/httpapi`, existing `internal/runtime` provider capability patterns, existing `web/src` Settings components

**Storage**: None; detection is computed on request and not persisted

**Testing**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, `git diff --check`

**Target Platform**: Local Loomi API + web Settings shell

**Project Type**: Web app with Go local API

**Performance Goals**: Detection reads a bounded set of local config files and returns within ordinary local API latency; no network calls or CLI execution

**Constraints**: No secret output, no real keychain access, no OAuth refresh, no CLI install/execute, no automatic provider enablement, no default provider switch

**Scale/Scope**: Two local provider candidates: Claude Code and Codex

## Constitution Check

- **Mechanism Parity, Original Expression**: Pass. The slice studies a local provider availability mechanism but uses Loomi-owned names/copy and no third-party private interface expression.
- **Runnable Vertical Slices**: Pass. Detector tests, HTTP endpoint, Settings UI, and docs provide executable/visible outcomes.
- **Core Flow Before Platform Complexity**: Pass. This is a thin provider capability surface and does not pull in Tool Runtime, workspace tools, browser, sandbox, plugins, or marketplace.
- **Observable Agent Execution**: Pass. No new execution path is introduced; status is visible through Settings/API.
- **Safety, Permissions, and Data Boundaries**: Pass. Detection is read-only, redacted, opt-in before use, and no secrets are logged, persisted, or returned.

## Project Structure

### Documentation (this feature)

```text
specs/026-local-provider-autodetect-foundation/
├── plan.md
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── local-provider-detection-api.md
│   └── settings-provider-ui.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/runtime/
├── local_provider_detection.go
└── local_provider_detection_test.go

internal/httpapi/
├── runtime.go
├── runtime_test.go
└── server.go

web/src/
├── domain.ts
├── realApiClient.ts
├── apiClient.ts
├── mockApiClient.ts
├── state.ts
├── App.tsx
├── i18n.ts
└── components/
    ├── SettingsView.tsx
    └── SettingsView.runtime.test.tsx

docs-site/src/content/docs/
├── architecture/local-provider-autodetect.md
├── api/local-provider-autodetect.md
├── runbooks/local-m18-5-provider-autodetect.md
├── devlog/2026-05-25-m18-5-local-provider-autodetect.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Keep detector in `internal/runtime` because it describes provider capability, expose via `internal/httpapi`, and render in existing Settings > Providers without adding a new package or database migration.

## Complexity Tracking

No constitution violations.
