# Quickstart: Memory Provider Foundation

## Focused Backend Validation

```bash
go test ./internal/productdata ./internal/httpapi ./internal/runtime
```

## Focused Web Validation

```bash
bun test web/src/components/SettingsView.runtime.test.tsx web/src/memory.test.ts
```

## Browser Smoke

1. Start the API and web app with the repo's normal local commands.
2. Open Settings > Memory.
3. Verify provider enablement, selected provider, readiness, diagnostic, and refresh state come from backend data.
4. Toggle memory disabled/enabled and confirm the UI does not claim semantic recall when disabled or unconfigured.

## Documentation Validation

```bash
cd docs-site
bun run build
```
