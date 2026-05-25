---
title: M19 Local Provider Opt-in Bridge
description: Session-local opt-in route candidate boundary for Local Codex.
---

M19 turns M18.5 detection into the first explicit opt-in bridge. Local provider detection still runs only from Settings > Providers user actions. A detected provider is not automatically configured.

## Boundary

The bridge stores only a process-local redacted enablement snapshot. `GET /v1/model-providers` reads that snapshot and never scans local auth files.

Enabled Local Codex is returned as a provider route candidate with:

- `local_provider: true`
- `session_local: true`
- `credential_reference: "redacted"`
- `execution_state: "unsupported"`
- `status: "unavailable"`

The unavailable status is intentional. M19 does not implement the Local Codex execution bridge, so Chat must remain blocked.

## Enable flow

1. User clicks Detect local providers.
2. UI receives safe detection results.
3. User clicks Enable for this session on an available Local Codex card.
4. Backend re-runs safe detection as part of the explicit enable action.
5. Backend stores only the redacted Local Codex capability snapshot in memory.
6. Configured providers list includes Local Codex as session-local and execution unsupported.

Disable removes the in-memory snapshot.

## Non-goals

M19 does not execute `codex` or `claude`, install CLIs, read keychain, refresh OAuth, call external endpoints to validate login, persist tokens, add sandbox/browser/filesystem/shell/workspace tools, or make Local Codex ready for Chat.
