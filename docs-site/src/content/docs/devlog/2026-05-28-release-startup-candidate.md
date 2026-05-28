---
title: 2026-05-28 Release Startup Candidate
description: Local startup, doctor diagnostics, safe CLI output, and real smoke path closeout.
---

Implemented:

- Aligned the Go API default listen address with the CLI/web local default: `127.0.0.1:18080`.
- Added `scripts/start-local.sh` as the user-facing local release/startup candidate entry. It delegates to the existing `scripts/dev-local.sh` helper so the release path and development path share one implementation.
- Added `scripts/dev-local.sh` as a small local helper that applies migrations, starts API and web, runs `loomi doctor --desktop`, prints the real smoke command, and optionally runs it with `RUN_SMOKE=1`.
- Tightened the helper into a fail-fast startup path: missing `migrate`, API process exit before readiness, doctor readiness timeout, web process exit after startup, desktop doctor failure, and `RUN_SMOKE=1` smoke failure now exit non-zero instead of leaving the desktop shell open against an unready API/DB/provider.
- Expanded `loomi doctor` to read structured `/readyz` checks and report database/schema blockers separately before probing provider/tool endpoints.
- Added a web/API mismatch warning when `VITE_LOOMI_API_BASE_URL` differs from the doctor host.
- Kept Local Codex as explicit opt-in: detected-but-disabled is still a provider warning with an enable remedy, not automatic enablement.
- Tightened CLI output so doctor/config output does not print raw API tokens, provider error bodies, config paths, or Local Codex auth paths.

Recommended startup:

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
APP_ENV=local HTTP_ADDR=127.0.0.1:18080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080 bun run --cwd web dev --host 127.0.0.1 --port 5180
go run ./cmd/loomi doctor --desktop --provider local_codex
go run ./cmd/loomi smoke agent --host http://127.0.0.1:18080 --provider local_codex --workspace "$PWD" --auto-approve --timeout 4m --prompt "Read AGENTS.md with workspace.read, then reply with local startup smoke complete."
```

Helper path:

```bash
scripts/start-local.sh
RUN_SMOKE=1 scripts/start-local.sh
```

Helper failure diagnostics:

- `tmp/start-local/doctor.log` captures the latest readiness/doctor failure for the user-facing path.
- `tmp/start-local/api.log` captures API startup and DB/schema failures for the user-facing path.
- `tmp/start-local/web.log` exists only after doctor readiness succeeds and web starts.
- `scripts/dev-local.sh` keeps `tmp/dev-local/` as the development-helper default unless `LOOMI_DEV_LOG_DIR` is set.

Validation targets:

```bash
bash scripts/dev-local.test.sh
go test ./cmd/loomi ./cmd/loomi-api ./internal/cli ./internal/httpapi -count=1
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Boundaries:

- No UI redesign, runtime strategy change, Docker/Redis/vector DB addition, or destructive cleanup.
- `tmp/`, local screenshots, and local smoke logs remain local noise unless intentionally promoted.
