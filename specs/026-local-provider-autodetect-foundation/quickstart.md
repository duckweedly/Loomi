# Quickstart: Local Provider Autodetect Foundation

## Detector tests

```bash
go test ./internal/runtime -run TestLocalProvider
```

Expected:

- Claude primary API key fixture returns `available`/`api_key`.
- Claude settings env fixture returns safe `available` status and model/base-url summary without token values.
- Claude `apiKeyHelper` fixture returns `unsupported` and no helper execution.
- Codex API key auth fixture returns `available`/`api_key`.
- Codex OAuth token fixture returns `available`/`oauth`.
- `CODEX_API_KEY` env wins over auth file.
- Missing files return `unavailable`.
- Tests use temp HOME/CODEX_HOME/CLAUDE_CONFIG_DIR.

## HTTP API tests

```bash
go test ./internal/httpapi -run TestLocalProviderDetection
```

Expected:

- `GET /v1/local-provider-detections` returns safe providers.
- Response does not include `sk-`, `Bearer`, `access_token`, `refresh_token`, or private paths.
- Unsupported/disabled states are stable.

## Web tests

```bash
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/realApiClient.test.ts src/state.test.ts
```

Expected:

- Settings > Providers source contains Local Claude Code and Local Codex UI.
- Detected/not detected status and explicit opt-in/no secrets copy are present.
- Local detection does not populate configured model providers or switch the current provider.

## Full validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```
