---
title: 2026-05-24 M5.5 Settings Placeholder Devlog
description: Implementation notes, validation results, limitations, and next steps for the temporary Settings surface.
---

## Completed scope

M5.5 introduces a temporary Settings surface for the Loomi shell. The slice focuses on visible information architecture, current-session local controls, and safe mock placeholders for future platform settings.

Implemented slice:

- Settings catalog with Workspace, Agent Core, and Management category groups
- in-app two-column Settings view opened from the existing sidebar Settings entry
- General category selected by default with a Back to workspace affordance
- session-local default workspace mode control for future local thread creation
- session-local language control with Chinese as default and English as an option
- mock runtime scenario control wired to future mock sends only
- read-only data source, backend capability, stream state, selected thread/run, and redacted provider capability rows
- Providers draft fields for current-session Base URL, model ID, and masked API key presence without persistence or provider calls
- placeholder panels for future platform categories with disabled preview controls
- placeholder safety tests preventing provider/tool/connector/file/backend write paths from SettingsView

## Validation log

Validated during implementation:

```bash
bun test ./web/src/useWorkspaceShellState.test.ts ./web/src/state.runtime.test.ts ./web/src/components/settingsCatalog.test.ts ./web/src/components/SettingsView.layout.test.tsx ./web/src/components/SettingsView.navigation.test.tsx ./web/src/components/SettingsView.runtime.test.tsx ./web/src/components/SettingsView.placeholders.test.tsx ./web/src/App.settings.test.tsx ./web/src/App.threadModes.test.ts ./web/src/components/ThreadSidebar.actions.test.ts
bun run --cwd web build
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

Latest recorded results:

- settings-focused frontend tests passed with 36 tests and 84 expectations.
- full frontend Bun suite passed with 103 tests and 236 expectations after provider draft coverage.
- frontend build passed.
- docs build passed.
- mock web browser smoke passed: Settings opens from the sidebar, Providers is marked mixed, Base URL/model draft fields can be filled, API key entry is a masked password input that only shows key presence, MCP placeholder shows disabled preview controls, and Back to workspace restores the chat canvas.
- real API visibility smoke partially passed: Settings shows Real API mode and read-only status rows; thread and provider loading are blocked by missing backend CORS headers for `http://127.0.0.1:5173` preflight requests to `/v1/threads` and `/v1/model-providers`.

## Known limitations

Settings are session-local in M5.5. Provider draft entry is not real provider management: Base URL and model ID are browser-session notes, API key entry is masked, and only key presence is retained. Persistent settings, backend secret storage, account/team settings, tool permissions, memory/RAG controls, activity recorder controls, and route management remain deferred to later specs.

Real API visibility smoke depends on backend CORS allowing the Vite dev origin. In this run, `http://127.0.0.1:8080/readyz` responded, but browser preflight requests from `http://127.0.0.1:5173` to `/v1/threads` and `/v1/model-providers` failed because the response did not include `Access-Control-Allow-Origin`.
