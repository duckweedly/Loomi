---
title: Local Release Startup Candidate
description: One-path local startup, doctor, smoke, and cleanup flow for Loomi.
---

This runbook is the release/startup candidate path for a local user. It keeps the API, web renderer, CLI doctor, workspace selection, Local Codex opt-in, and real agent smoke on the same local base URL.

## Defaults

- API: `http://127.0.0.1:18080`
- API listen address: `HTTP_ADDR=127.0.0.1:18080`
- Web API base: `VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080`
- API bearer token sources: `LOOMI_API_TOKEN`, `loomi config set api_token <token>`, `VITE_LOOMI_API_TOKEN`, or browser `localStorage['loomi.api_token']`
- Provider: `local_codex` by default, but it must be detected and explicitly enabled
- Workspace root: selected through `/v1/workspace/root`, the desktop picker, or smoke `--workspace`

## Few-step Startup

Start Postgres and apply migrations:

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
```

Start API:

```bash
APP_ENV=local \
HTTP_ADDR=127.0.0.1:18080 \
DATABASE_URL="$DATABASE_URL" \
LOOMI_WORKER_QUEUE_ENABLED=true \
LOOMI_WORKER_QUEUE_PAUSED=false \
go run ./cmd/loomi-api
```

Start web in another terminal:

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080 bun run --cwd web dev --host 127.0.0.1 --port 5180
```

Diagnose:

```bash
go run ./cmd/loomi doctor --desktop --provider local_codex
```

If doctor reports Local Codex detected but disabled, enable it explicitly from Settings > Providers or:

```bash
curl -s -X POST http://127.0.0.1:18080/v1/local-provider-detections/local_codex/enable
```

Set or verify workspace:

```bash
go run ./cmd/loomi smoke agent \
  --host http://127.0.0.1:18080 \
  --provider local_codex \
  --workspace "$PWD" \
  --auto-approve \
  --timeout 4m \
  --prompt "Read AGENTS.md with workspace.read, then reply with local startup smoke complete."
```

Stop/cleanup: press `Ctrl-C` in the API and web terminals. Local smoke logs and temporary files under `tmp/` are local noise unless a specific file is intentionally promoted.

## One-command Helper

For the local release/startup candidate, use `scripts/start-local.sh`. It is a thin user-facing wrapper over the existing `scripts/dev-local.sh` helper, so the release path and development helper share one fail-fast implementation instead of drifting apart.

```bash
scripts/start-local.sh
```

Set `RUN_SMOKE=1` to run the real smoke after doctor:

```bash
RUN_SMOKE=1 scripts/start-local.sh
```

`scripts/start-local.sh` writes logs under `tmp/start-local/` by default. Override with `LOOMI_START_LOG_DIR` or `LOOMI_LOG_DIR`. `scripts/dev-local.sh` remains available for development and keeps its default `tmp/dev-local/` log directory unless `LOOMI_DEV_LOG_DIR` is set.

The helper requires the `migrate` CLI. If `migrate` is missing, install it or apply migrations manually before using the helper; the script exits before starting API or web.

After API starts, the helper waits up to 60 seconds for `loomi doctor --host "$LOOMI_HOST" --provider "$LOOMI_PROVIDER"` to succeed. If doctor never reports ready, the script exits non-zero and points to:

- `tmp/start-local/doctor.log` for the latest readiness/doctor failure.
- `tmp/start-local/api.log` for API startup and DB/schema errors.
- `tmp/start-local/web.log` only after doctor readiness succeeds and web is actually started.

The helper also checks that API and web child processes are still running at the readiness gates. If API exits before doctor readiness, web is not started. If web exits immediately after startup, desktop doctor and smoke are not run. The final `loomi doctor --desktop` is blocking: workspace, tool catalog, or web/API mismatch failures stop the helper and clean up the started API/web child processes instead of leaving the desktop shell open against an unready backend. With `RUN_SMOKE=1`, the real smoke is part of the startup contract and any smoke failure exits non-zero.

Common fixes:

- `migrate CLI not found`: install the local migrate CLI or run `migrate -path migrations -database "$DATABASE_URL" up` in a prepared environment, then retry.
- `api process exited before readiness`: inspect `tmp/start-local/api.log`; most failures are port conflicts, Postgres not running, unapplied migrations, or wrong `DATABASE_URL`.
- `web process exited after startup`: inspect `tmp/start-local/web.log`; most failures are missing Bun/web dependencies or a web port conflict.
- `doctor did not report ready`: inspect `tmp/start-local/doctor.log` first, then `tmp/start-local/api.log`; most failures are Postgres not running, unapplied migrations, wrong `DATABASE_URL`, or provider not enabled.
- `desktop doctor` provider/workspace failure: enable Local Codex or another provider explicitly, then select a workspace folder before running smoke.

The helper starts API and web only after readiness is confirmed, prints the exact smoke command, and stops child processes when it receives `Ctrl-C`.

## Doctor Blockers

`loomi doctor --desktop --provider local_codex` distinguishes these local blockers:

- API not connected: start `loomi-api` or fix `LOOMI_HOST`.
- DB/schema not ready: start Postgres and apply migrations.
- Provider not configured: set `LOOMI_PROVIDER`, save a provider, or enable Local Codex.
- Local Codex detected but disabled: explicitly enable it before using it as a provider.
- Provider `401`/`403`: refresh the provider token.
- Provider `429`: wait for quota reset or switch provider.
- Provider `503`: retry later or switch provider.
- Workspace not selected: choose a folder in the desktop UI or pass smoke `--workspace`.
- Tool catalog unavailable or empty: check `/v1/tools/catalog` and runtime tool registration.
- Web/API mismatch: align `VITE_LOOMI_API_BASE_URL` with the doctor host before starting web.

CLI output must not print raw API tokens, provider response bodies, Local Codex auth paths, config file paths, or absolute sensitive paths.
