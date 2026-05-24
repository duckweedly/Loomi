---
title: M5.5 Settings Placeholder Runbook
description: Local validation path for the temporary Settings surface and placeholder safety checks.
---

## Mock desktop UI

```bash
bun run --cwd web desktop:dev
```

Open the sidebar Settings entry and choose Settings. The main workspace should show the Settings surface with General selected by default and a back affordance that returns to the prior workspace.

## Settings navigation smoke

1. Open Settings from the sidebar footer.
2. Confirm the left Settings navigation shows Workspace, Agent Core, and Management groups.
3. Confirm General is selected by default.
4. Select placeholder categories such as Providers, MCP, Safety, Tools, Routes, and Advanced.
5. Use Back to workspace and confirm the previously selected thread is still visible.

## Working settings smoke

- Confirm the interface defaults to Chinese.
- Change Language to English and confirm Settings and primary shell labels switch languages.
- Change Default workspace mode and confirm future local thread creation uses that mode.
- Change Mock runtime scenario and confirm future mock sends use the selected script only.
- Confirm data source, backend capability, stream state, selected run status, and provider capability are read-only.
- Open Providers and fill a gateway Base URL and model ID; confirm they remain draft UI values for the current session.
- Type a throwaway value into the API key field and confirm the input is masked and only key presence is shown.

## Placeholder safety smoke

- Visit every non-General category and confirm the copy says mock, preview, disabled, or not connected.
- Confirm placeholder controls are disabled.
- Confirm no placeholder category outside Providers asks for keys, credentials, or provider setup values.
- Confirm Providers draft fields do not perform provider, tool, connector, filesystem, backend write, persistence, or external actions.

## Real API visibility smoke

If the local API is running, start the web app with:

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev
```

Open Settings and confirm Data source mode shows Real API. Backend and provider capability status should remain read-only and secret-free. Provider draft fields may be filled locally, but they must not trigger a backend write or provider request.

## Validation commands

```bash
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```
