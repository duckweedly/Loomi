---
title: 2026-05-27 M72 Provider 503 Diagnosis
description: Completion-smoke provider diagnosis for Settings and CLI doctor.
---

## Completed scope

M72 separates provider configuration from live completion health.

- `GET /v1/model-providers` now reports complete local config as `configured`, not completion-ready.
- `POST /v1/model-providers/check` runs a minimal non-streaming completion smoke and returns `completion-ok` or `completion-failed` with `check_code` such as `completion-failed-503`.
- Settings > Providers displays the provider check category on the provider card.
- `loomi doctor` calls the provider check endpoint and includes the normalized check code and HTTP status.

## Safety

The completion check discards provider bodies and exposes only the safe status category, HTTP status, and sanitized message. It does not return API keys, bearer tokens, raw provider bodies, prompt text, provider trace payloads, or local auth paths.

## Validation

Planned validation:

```bash
go test ./internal/runtime ./internal/httpapi ./internal/cli -count=1
bun test --cwd web ./src/realApiClient.test.ts ./src/components/SettingsView.runtime.test.tsx
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
```

Real provider smoke is optional and depends on a reachable configured provider.
