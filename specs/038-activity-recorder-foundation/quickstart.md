# Quickstart: M30 Activity Recorder Foundation

## Focused Validation

```bash
go test ./internal/productdata -run 'TestActivityRecorder'
go test ./internal/httpapi -run 'TestActivityRecorder'
bun test --cwd web ActivityRecorderPanel.test.tsx SettingsView.activity.test.tsx realApiClient.test.ts mockApiClient.test.ts state.test.ts
```

## Full Closeout

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual Smoke

1. Start API and web app.
2. Open Settings > Activity Recorder.
3. Confirm initial state is disabled and no event rows are shown.
4. Enable recording and confirm status changes.
5. In mock mode, confirm safe seeded activity rows render with redaction markers.
6. Confirm no raw screenshots, keystrokes, clipboard data, credentials, full local paths, raw shell output, raw browser HTML, or file contents appear.

M30 does not add automatic OS-level capture, screenshots, keystroke logging, clipboard capture, authenticated browser profile recording, upload/sync, or cross-device activity history.
