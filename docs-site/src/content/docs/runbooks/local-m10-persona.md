---
title: Local M10 Persona Validation
description: Local validation checklist for Persona/Skill foundation.
---

## Scope

This runbook validates the M10 minimal slice:

- built-in persona syncs into product data
- installed local `SKILL.md` manifests are discoverable through the read-only API
- thread/run can choose or inherit a persona
- RunContext stores a durable persona snapshot/version
- Timeline/debug show only safe persona summary
- prompt text is not exposed in ordinary frontend runtime views

## Commands

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...
bun test ./web/src/realApiClient.test.ts ./web/src/runtime/realExecutionAdapter.test.ts ./web/src/runtime/runtimeEventGroups.test.ts ./web/src/components/Composer.test.ts ./web/src/components/RunTimeline.runtime.test.ts ./web/src/components/RunRail.runtime.test.ts
bun run --cwd web build
bun run --cwd docs-site build
```

## API Smoke

Start local dependencies, apply migrations, and run `loomi-api`.

Check persona list:

```bash
curl -s http://127.0.0.1:8080/v1/personas
```

Expected:

- at least one default persona
- `slug` and `active_version` are present
- no system prompt text appears

Check installed skill discovery:

```bash
curl -s http://127.0.0.1:8080/v1/skills
```

Expected:

- project, Codex, Claude Code, or plugin `SKILL.md` manifests appear when installed locally
- each item includes `name`, `source`, `source_label`, `path`, and `installed`
- full skill instruction bodies do not appear

Create or update a thread with a persona id, then create a model-gateway run with either inherited persona or `persona_id` override.

Expected:

- run creation succeeds
- run event history includes a `prepare_context` pipeline event
- safe persona fields include name/version
- raw prompt text is absent
- unknown or inactive thread `persona_id` returns `400 invalid_request` without SQL or foreign-key details

## Browser Smoke

Open the web app in real API mode.

1. Select the default persona in the composer selector, or leave the default selected.
2. Send a message that creates a run.
3. Open Timeline/debug.
4. Confirm the run shows a safe persona summary with persona name/version.
5. Refresh and confirm history replay still shows the same persona version.

Do not treat this smoke as proof of marketplace, plugin install, MCP, Memory/RAG, Sandbox/Desktop Runtime, or multi-agent behavior.
