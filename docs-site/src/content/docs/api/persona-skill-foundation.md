---
title: Persona Skill Foundation API
description: M10 persona list, thread selection, run override, and safe event metadata contracts.
---

M10 exposes enough API surface for the frontend to list personas and start a run with a selected persona. Persona prompt text is never returned through these read APIs.

## `GET /v1/personas`

Returns active persona summaries:

```json
{
  "personas": [
    {
      "id": "persona_...",
      "slug": "loomi-default",
      "name": "Default",
      "description": "General Loomi assistant persona.",
      "active_version": "2026-05-25.1",
      "is_default": true
    }
  ]
}
```

The response excludes system prompt text, model route internals, budget details, and raw skill configuration.

## Thread Persona

`POST /v1/threads` and `PATCH /v1/threads/{thread_id}` accept an optional `persona_id`:

```json
{
  "title": "Planning thread",
  "mode": "chat",
  "persona_id": "persona_..."
}
```

Thread persona is used only when a run does not provide its own override.

## Run Persona Override

`POST /v1/threads/{thread_id}/runs` accepts an optional `persona_id`:

```json
{
  "message_id": "msg_...",
  "source": "model_gateway",
  "provider_id": "custom",
  "model": "gpt-5.5",
  "persona_id": "persona_..."
}
```

Resolution order is run override, thread persona, then default persona. The resolved snapshot is durable and versioned at run creation.

## Safe Runtime Metadata

Pipeline `prepare_context` metadata may include safe persona fields:

```json
{
  "step": "prepare_context",
  "persona_id": "persona_...",
  "persona_slug": "loomi-default",
  "persona_name": "Default",
  "persona_description": "General Loomi assistant persona.",
  "persona_version": "2026-05-25.1",
  "persona_reasoning_mode": "balanced",
  "persona_budget_summary": "MVP default budget",
  "persona_allowed_tools": ["runtime.get_current_time"]
}
```

Prompt text, provider credentials, raw tool payloads, hidden local state, and future skill internals must not appear in ordinary event history or SSE.
