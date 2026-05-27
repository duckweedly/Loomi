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
HTTP_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
WORKER_QUEUE_ENABLED=true
WORKER_QUEUE_PAUSED=false
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
go run ./cmd/loomi doctor --provider "$LOOMI_PROVIDER"
```

`doctor` now reports provider `check_stage`, `check_code`, HTTP status, and a direct fix for common upstream failures:

- `http=401` or `http=403`: refresh the provider API token.
- `http=429`: wait for quota reset or switch provider.
- `http=503`: retry later or switch provider.

## Real Smoke

Run the harness smoke with auto approval only for this smoke:

```bash
go run ./cmd/loomi smoke agent \
  --provider "$LOOMI_PROVIDER" \
  --model "$LOOMI_MODEL" \
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
events	... total, ... tool, ... approvals
last_events	...
```

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

For code changes in this area:

```bash
go test ./cmd/loomi ./internal/cli ./internal/runtime ./internal/httpapi -count=1
go test ./...
bun run --cwd docs-site build
git diff --check
```

