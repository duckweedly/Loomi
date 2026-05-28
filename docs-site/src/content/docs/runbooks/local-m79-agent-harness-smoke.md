---
title: Local M79 Agent Harness Smoke
description: Real CLI/API to worker/provider/tool/approval/final-message smoke path.
---

M79 validates the real harness path, not another simulated provider fixture:

```text
loomi CLI -> API -> thread/message/run -> worker queue -> Gateway provider -> tool call -> approval -> tool execution -> provider continuation -> final assistant message
```

## Required Environment

The API must run with a real product store and worker queue enabled:

```bash
APP_ENV=local
HTTP_ADDR=127.0.0.1:18080
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
LOOMI_WORKER_QUEUE_ENABLED=true
LOOMI_WORKER_QUEUE_PAUSED=false
LOOMI_WORKSPACE_ROOT=/Users/xuean/Repos/personal-projects/Loomi
```

At least one model provider must be configured:

```bash
LOOMI_PROVIDER=custom
LOOMI_MODEL=<model>
```

For the saved `custom` provider route, configure the API base URL and token through Settings or `/v1/model-providers`. For the local Codex bridge, enable `local_codex` in the current API session after local provider detection succeeds. Do not print provider tokens in logs or smoke output.

Optional web-search tools need one of:

```bash
LOOMI_TAVILY_API_KEY=<token>
LOOMI_BRAVE_SEARCH_API_KEY=<token>
```

## Preflight

```bash
go run ./cmd/loomi-api
```

In another terminal:

```bash
go run ./cmd/loomi doctor --desktop --provider "$LOOMI_PROVIDER"
```

`doctor --desktop` reports the same live desktop readiness boundaries the renderer needs: API `/readyz` and DB/schema, provider availability or Local Codex enablement, tool catalog availability, and workspace root selection. `doctor` also reports provider `check_stage`, `check_code`, HTTP status, and a direct fix for common upstream failures:

- `http=401` or `http=403`: refresh the provider API token.
- `http=429`: wait for quota reset or switch provider.
- `http=503`: retry later or switch provider.
- `workspace`: choose a folder in the desktop UI or set `LOOMI_WORKSPACE_ROOT`.

## Real Smoke

Run the harness smoke with auto approval only for this smoke:

```bash
go run ./cmd/loomi smoke agent \
  --provider "$LOOMI_PROVIDER" \
  --model "$LOOMI_MODEL" \
  --workspace "$LOOMI_WORKSPACE" \
  --auto-approve \
  --timeout 2m \
  --prompt "Read AGENTS.md with workspace.read, then reply with M79 smoke complete."
```

Expected output includes:

```text
smoke ok
stage	run_completed
thread_id	...
run_id	...
final_stage	run_completed
provider	... check_stage=completion ...
workspace	...
events	... total, ... tool, ... approvals
tool_chain	...
final_message	...
last_events	...
```

The smoke fails closed when the run completes without a persisted assistant message or when the final assistant message is exactly `[redacted]`. This keeps finalization regressions visible instead of treating a terminal run event as enough.

It also fails closed when a terminal run still has unresolved approval events, or when the final assistant message is the generated failure placeholder `未生成成功回复`.

## One-command Real Desktop Acceptance

Use this command as the P0-2 regression check for a local desktop-feeling run. It exercises the same API, provider, workspace root, worker, tool approval, tool execution, provider continuation, persisted final message, and reloadable history path that the desktop renderer depends on:

```bash
loomi doctor --host http://127.0.0.1:18080 --provider local_codex && \
loomi smoke agent \
  --host http://127.0.0.1:18080 \
  --provider local_codex \
  --workspace /Users/xuean/Repos/personal-projects/Loomi \
  --auto-approve \
  --timeout 4m \
  --failure-log tmp/loomi-smoke/desktop-closeout-failure.json \
  --prompt "请读取 AGENTS.md，然后列目录确认当前 workspace，并用一个 Markdown 表格总结这个项目。"
```

Expected CLI evidence:

- `doctor ok` and `local_codex` is available/supported.
- `smoke ok`.
- `thread_id`, `run_id`, and `final_stage run_completed` are printed.
- `workspace` matches the selected workspace.
- `tool_order` contains the actual persisted tool order, with at least one workspace tool such as `workspace.read`, `workspace.list_directory`, or `workspace.tree_summary`.
- `pending_approvals	0`.
- `final_persisted	ok`.
- `replay	ok events=<n> terminal=run_completed`, proving the run can be reconstructed from persisted events after refresh/replay.
- `final_message` is non-empty, not `[redacted]`, and not `未生成成功回复`.
- On failure, the command prints `failure_log	tmp/loomi-smoke/desktop-closeout-failure.json`; attach that JSON evidence to the closeout notes instead of relying on terminal scrollback.

After the command, open or reload the desktop/web renderer against the same API host and verify the created thread still shows the final assistant message and the tool timeline. The current automated coverage is CLI + frontend render regression; a packaged Electron/browser automation smoke is the next step.

## Real Desktop Usability Closeout

Use the same CLI path for the desktop/web closeout so the browser and CLI exercise the same API, provider, workspace root, worker, tool, SSE, and final-message boundaries. Set `VITE_LOOMI_API_BASE_URL` for the renderer, point it at the same API host, and select the same workspace folder in the desktop UI or pass it to the CLI:

```bash
export LOOMI_WORKSPACE=/absolute/path/to/workspace
export LOOMI_PROVIDER=local_codex
export LOOMI_MODEL=gpt-5
export VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080
# Optional when the local API is behind a bearer-protected proxy:
# export VITE_LOOMI_API_TOKEN=<token>

go run ./cmd/loomi smoke agent --provider "$LOOMI_PROVIDER" --model "$LOOMI_MODEL" --workspace "$LOOMI_WORKSPACE" --timeout 2m --prompt "你好"
go run ./cmd/loomi smoke agent --provider "$LOOMI_PROVIDER" --model "$LOOMI_MODEL" --workspace "$LOOMI_WORKSPACE" --auto-approve --timeout 4m --failure-log tmp/loomi-smoke/classify-failure.json --prompt "帮我分类当前目录"
go run ./cmd/loomi smoke agent --provider "$LOOMI_PROVIDER" --model "$LOOMI_MODEL" --workspace "$LOOMI_WORKSPACE" --auto-approve --timeout 4m --failure-log tmp/loomi-smoke/github-url-failure.json --prompt "分析这个 GitHub URL: https://github.com/openai/openai-go"
go run ./cmd/loomi smoke agent --provider "$LOOMI_PROVIDER" --model "$LOOMI_MODEL" --workspace "$LOOMI_WORKSPACE" --auto-approve --timeout 4m --failure-log tmp/loomi-smoke/readme-failure.json --prompt "读取 README，必要时列目录或搜索关键词，然后用自然语言总结这个项目。"
```

Expected closeout evidence:

- `你好`: `smoke ok`, `tool_chain` absent, and `final_message` is natural language.
- `帮我分类当前目录`: first tool in `tool_chain` is `workspace.tree_summary` or `workspace.list_directory`; it must not start with `workspace.grep` or `workspace.glob`.
- `刚选目录` / `当前目录` means the selected `--workspace` snapshot for that run. If a previous smoke used another folder, the new run must still show the current workspace label and tool results must stay under that root.
- `下载目录` is only accepted as Downloads when the selected workspace basename is `Downloads`; otherwise the expected behavior is a clear prompt to choose Downloads rather than silently reading the repo root.
- GitHub URL prompt: `final_message` is natural language and is not `[redacted]`.
- README prompt: `tool_chain` includes `workspace.read` and may include `workspace.list_directory` or `workspace.grep`; `final_message` summarizes the project instead of dumping raw tool JSON.
- Browser/desktop requests and run event streams both use the same bearer token source when configured, so protected local API sessions do not fail only after SSE connects.

If the provider is unavailable, the command stops at the provider boundary and exits non-zero:

```text
smoke blocked
stage	provider_check
provider	custom status=completion-failed check_stage=completion check=completion-failed-auth http=401 ...
blocked_reason	provider_auth
fix	Refresh the provider API token, then run loomi doctor again.
```

That blocked result is still a valid provider-boundary smoke: it proves CLI/API configuration and the live provider completion check path, but it does not prove worker/tool/approval/final-message execution.

## Validation

For the current release/startup candidate path, use [Local Release Startup Candidate](./local-release-startup/). It keeps `loomi-api`, web, doctor, workspace selection, Local Codex enablement, and real agent smoke on `127.0.0.1:18080`.

For code changes in this area:

```bash
go test ./cmd/loomi ./internal/cli ./internal/runtime ./internal/httpapi -count=1
go test ./...
bun run --cwd docs-site build
git diff --check
```
