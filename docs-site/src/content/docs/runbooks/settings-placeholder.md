---
title: M5.5 Settings Placeholder Runbook
description: Local validation path for the temporary Settings surface and placeholder safety checks.
---

## Desktop UI

```bash
bun run --cwd web desktop:dev
```

Open the sidebar Settings entry and choose Settings. The main workspace should show the Settings surface with General selected by default and a back affordance that returns to the prior workspace.

## Settings navigation smoke

1. Open Settings from the sidebar footer.
2. Confirm the left Settings navigation shows Workspace, Agent Core, and Management groups.
3. Confirm General is selected by default.
4. Select real categories such as Providers, Web Search, Skill, MCP, Memory, Tools, and About.
5. Confirm Plugins, Activity Recorder, Context, Safety, Routes, and Advanced are not present in the Settings navigation while they are placeholder-only.
6. Use Back to workspace and confirm the previously selected thread is still visible.

## Working settings smoke

- Confirm the interface defaults to Chinese.
- Change Language to English and confirm Settings and primary shell labels switch languages.
- Change Default workspace mode and confirm future local thread creation uses that mode.
- Change Theme and confirm the current workspace switches between light and dark appearance.
- Confirm General does not show backend capability, stream state, selected run status, or provider capability diagnostics.
- Open Providers and confirm Configured providers shows backend provider id, family, model, status, and any safe detail message.
- Click Test connection for a configured provider and confirm the button is disabled while checking, then shows success or failed without secrets.
- If no providers are configured, confirm the page shows local API environment configuration guidance.
- Confirm the Base URL, model ID, and API key inputs are under Local draft and clearly state they are not saved and do not affect real model calls.
- Fill a gateway Base URL and model ID; confirm they remain draft UI values for the current session.
- Type a throwaway value into the API key field and confirm the input is masked and only key presence is shown.

## Hidden placeholder smoke

- Confirm no generic placeholder panel is reachable from Settings navigation.
- Confirm no placeholder category asks for keys, credentials, or provider setup values.
- Confirm Providers draft fields do not perform provider, tool, filesystem, backend write, persistence, or external actions.

## Real API visibility smoke

If the local API is running, start the web app with:

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev
```

Open Settings and confirm General does not expose mock runtime, data source mode, backend capability, stream, selected run, or provider capability diagnostics. In Providers, Configured providers should list backend provider capability and Test connection should call the backend provider check endpoint. Provider draft fields may be filled locally, but they must not trigger a backend write or provider request.

When the local API has no available model provider, Chat should show “模型 Provider 未配置或不可用” / “Model provider is not configured or unavailable” before sending. Click Open Settings / 打开设置 and confirm Settings opens directly to Providers.

Local API CORS is available only when the API runs in `APP_ENV=local` or `APP_ENV=development`, and only for `http://127.0.0.1:5173` / `http://localhost:5173`. From those origins, browser requests to `/v1/threads` and `/v1/model-providers` should pass OPTIONS preflight and normal GET responses should include the matching `Access-Control-Allow-Origin` header.

## Validation commands

```bash
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```
