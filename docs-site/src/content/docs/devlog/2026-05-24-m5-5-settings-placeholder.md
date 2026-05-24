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
- Providers configured-provider console showing backend provider id, family, model, status, safe detail message, and per-provider Test connection action
- Providers draft fields for current-session Base URL, model ID, and masked API key presence without persistence or provider calls, explicitly labeled Local draft
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
- earlier real API visibility smoke partially passed but was blocked by missing backend CORS headers for `http://127.0.0.1:5173` preflight requests to `/v1/threads` and `/v1/model-providers`.
- follow-up local CORS slice allows only `http://127.0.0.1:5173` and `http://localhost:5173` in local/development API mode, covering OPTIONS preflight and normal GET response headers for Settings/Chat browser smoke.
- provider test console slice adds Configured providers, a per-provider Test connection button backed by the existing provider check endpoint, checking/success/failed UI state, safe error display, and clearer Local draft copy for non-persistent notes.
- Chat provider guidance slice warns in real API mode when no available provider is configured, keeps mock mode unaffected, and opens Settings directly to Providers from the warning CTA.
- runtime/capability i18n cleanup moves runtime event group labels and backend capability title/detail copy into `web/src/i18n.ts`, threads locale through RunTimeline/RunRail, and reserves M6 worker/job status copy in English and Chinese without adding dependencies.

## Known limitations

Settings are session-local in M5.5. Provider draft entry is not real provider management: Base URL and model ID are browser-session notes, API key entry is masked, and only key presence is retained. Persistent settings, backend secret storage, account/team settings, tool permissions, memory/RAG controls, activity recorder controls, and route management remain deferred to later specs.

Real API visibility smoke now depends on running the API in `APP_ENV=local` or `APP_ENV=development` and using the supported Vite dev origins: `http://127.0.0.1:5173` or `http://localhost:5173`. Other local ports remain outside the CORS allowlist.
