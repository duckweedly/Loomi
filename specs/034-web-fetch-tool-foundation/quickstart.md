# Quickstart: M26 Web Fetch Tool Foundation

## Backend

```bash
go test ./...
```

Targeted tests while implementing:

```bash
go test ./internal/runtime -run 'TestWeb'
go test ./internal/productdata -run 'Test.*Web'
go test ./internal/httpapi -run 'TestM26WebFetch'
```

## Frontend

```bash
bun test --cwd web
bun run --cwd web build
```

Targeted tests while implementing:

```bash
bun test --cwd web SettingsView.tools RunRail.runtime
```

## Docs

```bash
bun run --cwd docs-site build
```

## Manual Smoke

1. Start the API and web dev server.
2. Open Settings > Tools and confirm `web.fetch` appears as builtin, web-scoped, read-only, approval-required, medium risk.
3. Trigger or replay a Work mode `web.fetch` lifecycle and confirm RunRail shows approval, execution, status/final URL/truncation metadata, and no raw body/cookie/credential content.
