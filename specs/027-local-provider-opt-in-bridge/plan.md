# Implementation Plan: Local Provider Opt-in Bridge

**Branch**: `027-local-provider-opt-in-bridge` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/027-local-provider-opt-in-bridge/spec.md`

## Summary

Add a session-local opt-in bridge for M18.5 local provider detection. The bridge lets Local Codex be explicitly enabled after manual detection, returns it as a redacted local provider route candidate, and blocks chat execution because the real Local Codex execution bridge is not implemented.

## Technical Context

**Language/Version**: Go backend, React/TypeScript web, Starlight docs
**Primary Dependencies**: Go stdlib HTTP, existing runtime provider model, existing Settings provider UI
**Storage**: In-process server memory only for local provider enablement
**Testing**: Go tests, Bun web tests/build, docs-site build
**Target Platform**: Local web/API development
**Project Type**: Web-first agent platform
**Performance Goals**: Enable/disable endpoints complete synchronously using existing safe detection logic
**Constraints**: No auto detection on startup/page mount/chat send/provider list; no plaintext secrets; no CLI/keychain/OAuth refresh/external validation; no new tools
**Scale/Scope**: Local Codex first; Claude Code contract/unsupported only

## Constitution Check

- **Mechanism Parity, Original Expression**: Pass. Uses Loomi-owned provider capability language.
- **Runnable Vertical Slices**: Pass. Produces backend endpoints, UI state, tests, and docs.
- **Core Flow Before Platform Complexity**: Pass. Does not introduce sandbox/tools/CLI execution or durable credential storage.
- **Observable Agent Execution**: Pass. Explicitly does not add execution; unsupported route prevents misleading run events.
- **Safety, Permissions, and Data Boundaries**: Pass. All auth reads stay behind manual detection/enable and responses are redacted.

## Project Structure

```text
specs/027-local-provider-opt-in-bridge/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── local-provider-opt-in-api.md
│   └── settings-provider-ui.md
├── checklists/
│   └── requirements.md
└── tasks.md

internal/runtime/
├── local_provider_detection.go
└── providers.go

internal/httpapi/
├── runtime.go
├── runtime_test.go
└── server.go

web/src/
├── apiClient.ts
├── domain.ts
├── i18n.ts
├── mockApiClient.ts
├── realApiClient.ts
├── state.ts
└── components/
    ├── ChatCanvas.states.test.ts
    ├── SettingsView.runtime.test.tsx
    └── SettingsView.tsx
```

## Phase 0 Research

See [research.md](./research.md).

## Phase 1 Design

See [data-model.md](./data-model.md), [contracts/local-provider-opt-in-api.md](./contracts/local-provider-opt-in-api.md), and [contracts/settings-provider-ui.md](./contracts/settings-provider-ui.md).

## Post-Design Constitution Check

No violations introduced. The implementation remains a thin opt-in/status slice and leaves real Local Codex execution to a future bridge.
