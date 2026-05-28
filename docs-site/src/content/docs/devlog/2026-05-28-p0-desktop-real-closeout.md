---
title: P0 Desktop Real Closeout
description: Real backend, renderer, Local Codex, workspace tools, approvals, patch apply, and bounded validation were verified end to end.
---

## Result

- Real backend and renderer were started against Postgres schema version 17.
- `/readyz`, database/schema readiness, provider readiness, tool catalog readiness, and workspace selection were all verified before dogfood.
- Provider: `local_codex`, status `available`, execution `supported`, model `gpt-5.5`.
- Workspace label: `Loomi-p0-dogfood-workspace`.
- Successful dogfood thread: `thr_1779933664527158000_24e9c76e31e3`.
- Successful dogfood run: `run_1779933664541429000_182b799e1ec9`.

## Dogfood Chain

The successful Work-mode run used this tool order:

```text
workspace.tree_summary -> workspace.read -> workspace.patch_preview -> workspace.patch_apply -> sandbox.exec_command
```

Evidence:

- Directory inventory started with `workspace.tree_summary`, not broad grep.
- `workspace.read` read `README.md`.
- `workspace.patch_preview` created preview `patch_1a5b2d2837382d8e`.
- `workspace.patch_apply` applied the approved README change.
- `sandbox.exec_command` ran `go test ./...` inside the selected workspace and exited `0`.
- `approvals list` for the final run returned no pending approvals.
- Final assistant content persisted as normal Markdown and was visible after refreshing the desktop renderer.

## Repeatable Closeout Path

The CLI smoke path now carries the repeatable evidence contract for this closeout:

```bash
go run ./cmd/loomi doctor --host http://127.0.0.1:18080 --provider local_codex --desktop
go run ./cmd/loomi smoke agent \
  --host http://127.0.0.1:18080 \
  --provider local_codex \
  --workspace /Users/xuean/Repos/personal-projects/Loomi \
  --auto-approve \
  --timeout 4m \
  --failure-log tmp/loomi-smoke/desktop-closeout-failure.json \
  --prompt "请读取 AGENTS.md，然后列目录确认当前 workspace，并用一个 Markdown 表格总结这个项目。"
```

Passing evidence must include:

- `thread_id` and `run_id`.
- `tool_order` with the persisted tool order from replayed run events.
- `pending_approvals	0`.
- `final_persisted	ok`.
- `replay	ok events=<n> terminal=run_completed`.
- A failure log path when the command exits non-zero.

This path reuses the existing API, SSE, approval, worker, provider, workspace-root, message-list, and run-event replay endpoints. It does not call the ToolBroker or database directly and does not introduce a mock success path.

## Fixes

- Local development CORS now covers `/readyz`, so the desktop readiness probe does not collapse into browser-level `Failed to fetch`.
- Built-in persona sync now refreshes same-version tool definitions on startup, preventing stale Postgres persona rows from hiding newer workspace tools.
- Sandbox argument validation now allows exact `./...` for bounded test commands while still rejecting traversal.

## Validation

- `go test ./internal/httpapi -run TestReadyzIncludesLocalDevCORSHeaders -count=1`
- `go test ./internal/productdata -run 'TestSyncBuiltInPersonasCreatesDefaultVersionIdempotently|TestSyncBuiltInPersonasUpdatesExistingVersionDefinition|TestSyncBuiltInPersonasPreservesOldVersionForRunSnapshots' -count=1`
- `LOOMI_TEST_DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable go test ./internal/productdata -run TestPostgresSyncBuiltInPersonasUpdatesExistingVersionDefinition -count=1`
- `go test ./internal/productdata -run TestValidateSandboxToolCallArgumentsUsesBoundedAllowlist -count=1`
- `go test ./cmd/loomi -run 'TestSmokeAgentCommandReportsProviderBoundaryBlocker|TestSmokeAgentCommandRunsWithAutoApprovalAndPrintsSummary' -count=1`

## Remaining Limits

- The closeout does not change release packaging or durable rollout design.
- The smoke workspace is intentionally temporary and only proves the current local desktop dogfood path.
- Packaged Electron automation is still manual: the CLI proves persisted replay/finalization, then the desktop/web renderer must be reloaded against the same API host to visually verify the thread timeline.
