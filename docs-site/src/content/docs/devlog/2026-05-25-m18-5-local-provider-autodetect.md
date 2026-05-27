---
title: 2026-05-25 M18.5 Local Provider Autodetect
description: Detection-only local Claude Code and Codex provider capability slice.
---

## Completed

- Added Spec Kit artifacts under `specs/026-local-provider-autodetect-foundation/`.
- Added backend local provider detector for Claude Code and Codex safe capability summaries.
- Added read-only `GET /v1/local-provider-detections`.
- Added Settings > Providers local autodetect cards.
- Kept local autodetect separate from configured model providers.
- Documented architecture, API contract, and validation runbook.

## Safety boundary

The detector only reports presence-derived capability status. It does not output secrets, paths, Authorization headers, raw token fields, provider auth payloads, or account metadata.

The slice does not execute `apiKeyHelper`, call Claude/Codex/OpenAI endpoints, refresh OAuth tokens, read keychain data, install CLIs, persist tokens, enable providers, or switch model-gateway defaults.

## Validation

Targeted validation:

```bash
go test internal/runtime/local_provider_detection.go internal/runtime/local_provider_detection_test.go -run TestLocalProvider
go test ./internal/httpapi -run TestLocalProviderDetection
bun test --cwd web src/realApiClient.test.ts src/state.test.ts src/components/SettingsView.runtime.test.tsx
```

Full validation commands for closeout:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Known limitations

- Detected local providers are not usable for model-gateway runs yet.
- No persistent opt-in setting is implemented.
- Keychain-backed credentials are not inspected.
- OAuth refresh is intentionally unsupported.
- The current UI is read-only status evidence, not a provider connection flow.
