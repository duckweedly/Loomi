# Quickstart: M27 Browser Automation Foundation

## Backend

```bash
go test ./...
```

Targeted tests while implementing:

```bash
go test ./internal/productdata -run 'Test.*Browser'
go test ./internal/runtime -run 'TestBrowser|TestToolBrokerExecutesBrowser'
go test ./internal/httpapi -run 'TestM27BrowserAutomation'
```

## Frontend

```bash
bun test --cwd web
bun run --cwd web build
```

Targeted tests while implementing:

```bash
bun test --cwd web SettingsView.tools RunRail.runtime runtimeScripts
```

## Docs

```bash
bun run --cwd docs-site build
```

## Manual Smoke

1. Start the API and web app.
2. Open Settings > Tools and confirm `browser.open`, `browser.snapshot`, and `browser.click_link` appear as builtin browser tools.
3. Trigger or replay a browser lifecycle and confirm RunRail labels browser rows as medium risk and public HTTP only.
4. Confirm no raw HTML, cookies, credentials, Set-Cookie values, or local paths appear in visible metadata.
