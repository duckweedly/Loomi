---
title: M10 Persona Skill Foundation
description: Built-in persona sync, run persona resolution, and safe persona observability for the M10 foundation slice.
---

M10 adds the smallest Persona/Skill foundation needed for real runs to execute with a durable persona choice. It does not add a marketplace, plugin install, MCP, Memory/RAG, Sandbox/Desktop Runtime, multi-agent orchestration, or broad skill packaging.

## Data Boundary

Product data owns persona records and versioned prompt/config snapshots:

- `personas`: stable persona identity, slug, name, description, source, default flag, and active version pointer.
- `persona_versions`: immutable version payload including system prompt, model route, allowed tool names, reasoning mode, and budget summary.
- `threads.persona_id`: optional thread-level persona preference.
- `runs.persona_id`: optional run-level override.
- `run_persona_snapshots`: immutable run snapshot of the resolved persona version.

The system prompt is stored for runtime use, but normal Timeline/debug summaries only expose safe fields.

## Built-In Sync

`cmd/loomi-api` syncs built-in personas at startup through `productdata.BuiltInPersonas()`. Sync is idempotent:

- an existing slug is reused
- a new version is inserted only when the version label changes
- default persona choice is maintained by built-in config
- unknown tool names are rejected before the persona becomes active

The current built-in default is `loomi-default`, with `runtime.get_current_time` as the only allowed MVP tool.

## Resolution Order

Run creation resolves persona in this order:

1. run override from `POST /v1/threads/{thread_id}/runs`
2. thread persona from `threads.persona_id`
3. default active persona
4. no persona only when the database has no synced default

The resolved snapshot is written to `run_persona_snapshots` when the run is created. Later built-in syncs do not change the persona version already attached to an older run.

## Runtime Use

`PrepareRunContext` attaches the resolved persona snapshot before worker runtime invocation. The persona can provide:

- system prompt for provider request construction
- model route defaults
- allowed tool names converted into runtime tool summaries
- reasoning mode and budget summary for future runtime policy

M10 only wires the route and tool allowlist into the current MVP runtime path. It does not execute arbitrary skills or install tool providers.

## Safe Observability

RunContext safe summaries include:

- persona id
- persona slug
- persona name
- persona description
- persona version
- reasoning mode
- budget summary
- allowed tool names

They must not include the raw system prompt. Pipeline events, Timeline, RunRail, and debug surfaces can show this safe summary for `prepare_context` evidence.
