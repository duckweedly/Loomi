---
title: M95 Real Desktop Usability Smoke
description: Closeout smoke hardening for real API/provider/workspace/tool/final-message usability.
---

M95 keeps the existing `loomi smoke agent` path and makes it usable as the repeatable real desktop closeout smoke. The command now verifies API readiness, optional workspace-root selection, provider readiness, run/SSE completion, tool-chain shape, and persisted final assistant message.

The smoke intentionally fails closed when a completed run has no persisted assistant message, or when the final assistant message is exactly `[redacted]`. That catches the regressions that make a terminal run look successful while the user still has no useful answer.

Closeout prompts are documented in the M79 harness runbook:

- `你好` should complete without workspace/sandbox tools.
- `帮我分类当前目录` should start with `workspace.tree_summary` or `workspace.list_directory`, not grep/glob.
- A GitHub URL analysis should produce a natural-language final answer, not `[redacted]`.
- A workspace read flow should end with a natural-language summary after `workspace.read` and optional list/grep.

Validation added:

```bash
go test ./cmd/loomi ./internal/cli -count=1
```

M95 follow-up tightened the desktop real API path:

- Browser JSON requests now include a bearer token from `VITE_LOOMI_API_TOKEN` or `localStorage['loomi.api_token']`.
- Run event streaming moved from browser `EventSource` to fetch-based SSE so protected API sessions keep the same Authorization behavior for continuation events.
- Raw browser `Failed to fetch` is converted into a Loomi API connectivity diagnostic that names the active API base and points to backend/API-base setup.
- Local API CORS now allows the Authorization header for local dev origins and accepts alternate `127.0.0.1` / `localhost` renderer ports during real desktop smoke.

Additional validation:

```bash
bun test --cwd web ./src/realApiClient.test.ts ./src/runtime/adapterSelection.test.ts
go test ./internal/httpapi -run 'Test.*CORS|Test.*Health|Test.*Runtime' -count=1
```
