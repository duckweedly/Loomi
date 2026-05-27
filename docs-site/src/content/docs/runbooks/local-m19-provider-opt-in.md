---
title: M19 Local Provider Opt-in Runbook
description: Validation path for explicit Local Codex session opt-in.
---

## Scope

Use this runbook to validate that detected Local Codex can be enabled as a session-local route candidate without claiming Chat readiness.

## Focused tests

```bash
go test ./internal/httpapi -run 'TestLocalProvider.*Enable|TestModelProvider'
bun test --cwd web realApiClient.test.ts state.test.ts components/SettingsView.runtime.test.tsx components/ChatCanvas.states.test.ts
```

Expected:

- detected-only Local Codex is absent from `GET /v1/model-providers`
- explicit enable adds Local Codex with `local_provider`, `session_local`, `credential_reference=redacted`, and `execution_state=unsupported`
- disable removes Local Codex from model providers
- unavailable/unsupported providers are rejected or remain not ready
- Settings exposes enable/disable state
- Chat remains blocked for enabled-but-unsupported Local Codex

## Full validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual smoke

1. Start local API and web.
2. Open Settings > Providers.
3. Click Detect local providers.
4. If Local Codex is detected, click Enable for this session.
5. Confirm Configured providers shows Local Codex as session-local, credential redacted, and execution unsupported.
6. Confirm Chat still blocks send.

## Safety checks

Do not run `codex`, `claude`, helper commands, or external login validation. Do not inspect keychain, `.env`, SSH/AWS credentials, or raw token values.
