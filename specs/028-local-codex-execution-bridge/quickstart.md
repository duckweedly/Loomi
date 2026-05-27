# Quickstart: Local Codex Execution Bridge

## Automated Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Local Manual Smoke

1. Start the API with local worker enabled.
2. Start the desktop/web shell.
3. Open Settings > Providers.
4. Click Detect local providers.
5. If Local Codex is detected, click Enable for this session.
6. Confirm Configured providers shows `local_codex` as available and supported.
7. Return to Chat, type a short message, and send.
8. Confirm RunTimeline shows model request, output delta, and completion or a clear provider failure.
9. Confirm assistant message appears only on successful provider output.
10. Check page and logs for token/path canaries; none should appear.

## Fixture Smoke

Use a temporary `CODEX_HOME` with `auth.json` pointing at a local OpenAI-compatible test server. The fixture must include token/key canaries and assert no canary appears in API responses, run events, assistant metadata, or captured logs.
