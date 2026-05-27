# Quickstart: Local Provider Opt-in Bridge

## Backend focused tests

```sh
go test ./internal/httpapi -run 'TestLocalProvider.*Enable|TestModelProvider'
go test ./internal/runtime -run TestProvider
```

## Web focused tests

```sh
bun test --cwd web realApiClient.test.ts state.test.ts components/SettingsView.runtime.test.tsx components/ChatCanvas.states.test.ts
```

## Full validation

```sh
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual smoke

1. Start local API and web against real API mode.
2. Open Settings > Providers.
3. Click Detect local providers.
4. If Local Codex is detected, click Enable for this session.
5. Confirm Configured providers shows Local Codex as local/session-local/redacted/unsupported.
6. Confirm Chat still blocks sending because Local Codex execution bridge is not implemented.
