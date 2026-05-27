---
title: M18.5 Local Provider Autodetect Runbook
description: Validation path for safe Claude Code and Codex local provider detection.
---

## Scope

Use this runbook to validate the M18.5 detection-only slice. It should prove local provider status is visible without using local credentials for model calls.

## Detector tests

```bash
go test ./internal/runtime -run TestLocalProvider
```

Expected coverage:

- Claude `.claude.json` `primaryApiKey` fixture returns `available` and `api_key`
- Claude `settings.json` env fixture returns safe status and model candidate
- Claude `apiKeyHelper` returns `unsupported` without execution
- Codex auth-file API key fixture returns `available` and `api_key`
- Codex OAuth token fixture returns `available` and `oauth`
- `CODEX_API_KEY` env wins over auth file
- missing files return `unavailable`
- tests use temp HOME/CODEX_HOME/CLAUDE_CONFIG_DIR

## HTTP API tests

```bash
go test ./internal/httpapi -run TestLocalProviderDetection
```

Expected:

- `GET /v1/local-provider-detections` returns Local Claude Code and Local Codex
- response contains only safe capability fields
- response excludes API keys, bearer tokens, token field names, and private paths
- unsupported and disabled statuses are stable

## Web tests

```bash
bun test --cwd web src/realApiClient.test.ts src/state.test.ts src/components/SettingsView.runtime.test.tsx
```

Expected:

- real API client maps local provider detection
- state stores local detections separately from configured providers
- Settings > Providers shows local autodetect cards
- UI copy says explicit opt-in is required and no secrets are shown
- current configured provider list is not replaced by local detections

## Full validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Safety checks

Do not run `claude`, `codex`, `rtk`, or helper commands. Do not inspect `.env`, SSH/AWS credentials, keychain entries, or real auth token values. Do not paste or upload local auth content. Detection is shape and presence only.
