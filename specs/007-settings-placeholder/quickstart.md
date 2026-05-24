# Quickstart: M5.5 Settings Placeholder

## 1. Start mock desktop UI

```bash
bun run --cwd web desktop:dev
```

Expected:

- Loomi desktop window opens.
- Existing thread/workspace state is visible.

## 2. Open Settings

1. Click the sidebar Settings entry.
2. Select the Settings row if a menu appears.

Expected:

- Settings view opens in the main workspace area.
- Left settings navigation is visible.
- General category is selected by default.
- Existing selected thread is restored when returning from Settings.

## 3. Verify General working settings

Check these rows:

- interface language
- default workspace mode
- mock runtime scenario
- data source mode
- backend/model gateway status
- provider capability status when available
- provider draft Base URL, model ID, and masked API key presence

Expected:

- Chinese is the default interface language.
- Changing language to English updates current-session shell/settings copy.
- Changing default mode affects future local workspace/new conversation behavior.
- Changing mock runtime scenario affects future mock sends only.
- Data source/backend/provider capability rows are status/read-only and do not expose secrets.
- Provider draft Base URL and model ID can be filled for the current browser session only.
- Provider draft API key entry is masked, shows only whether a key was entered, and is not echoed as text.

## 4. Verify placeholder categories

Navigate every non-General category.

Expected:

- Placeholder categories show mock/preview/not connected copy.
- Placeholder controls do not trigger provider calls, tool execution, external writes, or persisted configuration changes.
- Placeholder categories outside Providers never ask for API keys.
- Providers draft fields do not trigger provider calls, backend writes, or persisted configuration changes.

## 5. Real API visibility smoke

If local API is running:

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev
```

Expected:

- Settings shows real API mode.
- Backend/provider capability status is displayed as user-safe read-only information.
- Provider draft entry remains current-session only, masks key entry, and does not send secrets to the backend.

## 6. Validation commands

```bash
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

## 7. Documentation check

Update docs-site pages for settings architecture/runbook/devlog/spec-kit status during implementation, then run docs build before completion.
