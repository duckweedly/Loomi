# Implementation Plan: Local Codex Execution Bridge

**Branch**: `028-local-codex-execution-bridge` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/028-local-codex-execution-bridge/spec.md`

## Summary

Turn explicit Local Codex opt-in into a real executable provider by adding a Local Codex runtime provider that reads auth only at explicit enable/execution boundaries, registers through the existing Gateway provider list, and emits normal model gateway events. Chat remains blocked for unavailable or unsupported local provider states.

## Technical Context

**Language/Version**: Go backend, React/TypeScript web, Starlight docs

**Primary Dependencies**: Go stdlib HTTP, existing runtime `Provider` interface, existing Gateway/Worker, existing Settings/Chat provider state

**Storage**: In-process local provider enablement only; no DB secret storage

**Testing**: Go tests and race-safe concurrency tests, Bun web tests/build, docs-site build

**Target Platform**: Local API + web/desktop-feeling shell

**Project Type**: Web-first agent platform

**Performance Goals**: Enable/list operations remain synchronous; provider streaming uses existing gateway flow

**Constraints**: No automatic local auth scanning; no plaintext secrets in persistent or frontend surfaces; no CLI/keychain/OAuth refresh; no new sandbox/browser/filesystem/shell/workspace tools

**Scale/Scope**: Local Codex first; other local providers stay unsupported unless later specs add execution bridges

## Constitution Check

- **Mechanism Parity, Original Expression**: Pass. Uses Loomi provider language and existing gateway concepts.
- **Runnable Vertical Slices**: Pass. Produces a Chat-sendable provider with API, worker, event, UI, and docs validation.
- **Core Flow Before Platform Complexity**: Pass. Reuses run/event/SSE/worker/gateway and avoids new platform tools.
- **Observable Agent Execution**: Pass. All execution goes through persisted gateway run events and timeline views.
- **Safety, Permissions, and Data Boundaries**: Pass. Explicit opt-in gates local auth reads and redaction forbids secret/path exposure.

## Project Structure

```text
specs/028-local-codex-execution-bridge/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── local-codex-execution-api.md
│   └── settings-chat-ui.md
├── checklists/
│   └── requirements.md
└── tasks.md

internal/runtime/
├── local_codex_provider.go
├── local_provider_detection.go
├── providers.go
└── gateway.go

internal/httpapi/
├── runtime.go
├── runtime_test.go
└── server.go

web/src/
├── runtime/backendCapabilityStatus.ts
├── components/ChatCanvas.states.test.ts
├── components/SettingsView.runtime.test.tsx
├── realApiClient.test.ts
└── realApiClient.ts

docs-site/src/content/docs/
├── architecture/
├── api/
├── runbooks/
├── devlog/
├── roadmap/
└── spec-kit/
```

## Phase 0 Research

See [research.md](./research.md).

## Phase 1 Design

See [data-model.md](./data-model.md), [contracts/local-codex-execution-api.md](./contracts/local-codex-execution-api.md), and [contracts/settings-chat-ui.md](./contracts/settings-chat-ui.md).

## Post-Design Constitution Check

No violations introduced. The implementation keeps M20 as a thin runnable provider slice over the existing Gateway instead of adding a parallel chat execution path.
