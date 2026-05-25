---
title: Local M20 Codex Execution Runbook
description: Validate explicit Local Codex execution through Chat.
---

## Automated Checks

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual Smoke

1. Start the API and desktop/web shell.
2. Open Settings > Providers.
3. Click Detect local providers.
4. If Local Codex is detected, click Enable for this session.
5. Confirm Configured providers shows `local_codex` as available/supported, session-local, and redacted.
6. Return to Chat and send a short message.
7. Confirm RunTimeline/RunRail show model request, output delta, and completed events, or a clear provider failure.
8. Confirm an assistant message appears only after successful provider output.
9. Check page and logs for token/path canaries.

## Fixture Proof

The automated smoke uses a temporary `CODEX_HOME/auth.json` and a local OpenAI-compatible test server. It proves the Loomi bridge, provider registration, worker, Gateway, events, assistant persistence, and redaction path without touching the user's real auth file.

## Real Local Auth Note

M20 does not refresh OAuth and does not read keychain. If the local Codex token cannot call the configured compatible endpoint, execution fails through the provider failure path instead of fabricating a reply.
