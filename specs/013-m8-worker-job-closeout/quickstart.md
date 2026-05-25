# Quickstart: M8 Worker Job Closeout

## Targeted retry/backoff validation

```bash
go test ./internal/productdata ./internal/runtime
```

Expected:

- recovery tests show stale jobs are rescheduled with a future `scheduled_at`
- immediate stale-owner complete/fail writes remain rejected
- retry exhaustion still reaches terminal failed state

## Requested backend validation

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
```

Expected:

- productdata, runtime, httpapi, and command packages pass
- no frontend validation is required unless frontend state changes are introduced

## Docs validation

```bash
bun run --cwd docs-site build
```

Expected:

- Starlight docs build succeeds after roadmap/devlog updates

## Closeout reading path

1. Read `specs/013-m8-worker-job-closeout/contracts/m8-audit.md`.
2. Confirm `docs-site/src/content/docs/roadmap/current-status.md` says original M8 is covered and closeout passed.
3. Confirm `docs-site/src/content/docs/devlog/2026-05-25-m8-worker-job-closeout.md` records evidence, patch, validations, and non-goals.
