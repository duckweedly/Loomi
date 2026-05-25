# Quickstart: Persona Skill Foundation

## Purpose

Validate the M10 foundation slice after implementation: built-in personas sync to DB, thread/run resolution chooses a persona, RunContext records the persona snapshot/version, and Timeline/debug shows a safe summary without raw prompt text.

## Backend Validation

Run:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
```

Expected coverage:

- Built-in persona sync creates and updates durable persona/version records idempotently.
- Duplicate sync does not create duplicate active versions.
- Thread/run/default persona resolution follows the required precedence.
- RunContext contains resolved persona snapshot/version before runtime invocation.
- Missing/inactive/cross-scope persona references fail before provider/runtime invocation.
- Persona allowed tools intersect with the existing runtime allowlist.
- Persona model route is applied through existing provider/model route behavior.
- Timeline/debug event metadata contains safe persona summary and excludes raw system prompt.

## Web Validation

Run related runtime/UI tests, including any new or updated tests under:

```text
web/src/realApiClient.test.ts
web/src/runtime/realExecutionAdapter.test.ts
web/src/runtime/runtimeEventGroups.test.ts
web/src/components/Composer.test.ts
web/src/components/RunTimeline.runtime.test.ts
web/src/components/RunRail.runtime.test.ts
```

Then run:

```bash
bun run --cwd web build
```

Expected coverage:

- Persona selection or read-only default display is visible in the chosen MVP surface.
- Live and replayed Timeline/debug show the same safe persona summary.
- Raw persona prompt text is absent from normal UI state and metadata rendering.

## Docs Validation

Run:

```bash
bun run --cwd docs-site build
```

Expected docs updates:

- `docs-site/src/content/docs/architecture/persona-skill-foundation.md`
- `docs-site/src/content/docs/api/persona-skill-foundation.md`
- `docs-site/src/content/docs/runbooks/local-m10-persona.md`
- `docs-site/src/content/docs/roadmap/current-status.md`
- `docs-site/src/content/docs/spec-kit/workflow.md`
- `docs-site/src/content/docs/devlog/2026-05-25-m10-persona-skill-foundation.md`

## Browser Smoke

1. Start the local API/worker and web app in real API mode.
2. Ensure built-in persona sync has run and at least one default persona exists.
3. Create or open a thread.
4. Select a persona if the MVP selector is implemented, or confirm the default persona display if the read-only path is chosen.
5. Create a run.
6. Open Timeline/debug details for the run.
7. Confirm the run's safe persona summary includes:
   - persona name
   - persona version
   - model route label
   - allowed tool names or count
   - reasoning mode
   - budget summary
8. Confirm RunContext/debug metadata records the persona snapshot/version.
9. Confirm raw persona system prompt text is not visible in normal Timeline/debug.
10. Refresh/reconnect and confirm history replay shows the same safe persona summary.

## Non-Goals to Verify

- No full Skill marketplace.
- No plugin installation.
- No MCP.
- No Memory/RAG.
- No Sandbox/Desktop Runtime.
- No multi-agent behavior.
- No new worker queue.
- No broad permission framework.
- No raw persona prompt in normal Timeline/debug.

## Validation Results

Completed on 2026-05-25:

- `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...` passed.
- `bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/Composer.test.ts ./web/src/components/RunTimeline.runtime.test.ts ./web/src/components/RunRail.runtime.test.ts` passed with 57 tests.
- `bun run --cwd web build` passed.
- `bun run --cwd docs-site build` passed.

Browser/API smoke:

- API ran locally on `127.0.0.1:18080` because `127.0.0.1:8080` was already occupied.
- Web dev smoke ran on `127.0.0.1:5173`.
- `GET /v1/personas` returned the built-in default persona `loomi-default` at version `2026-05-25.1`.
- A thread using that persona created a local simulated run whose history replay showed `persona_name: Loomi Default`, `persona_version: 2026-05-25.1`, model route `custom/gpt-5.5`, reasoning mode `balanced`, budget summary, and `runtime.get_current_time` allowlist summary.
- Browser RunRail showed the persona selector as `Loomi Default v2026-05-25.1` and Timeline/debug displayed the safe persona summary after refresh/history replay.
- Browser snapshot did not contain `system_prompt` or raw prompt text.
