# Contract: Built-in Persona Sync

## Purpose

Load repository-local built-in persona definitions into durable product data so run resolution can use versioned persona records.

## Built-in Config Shape

```json
{
  "slug": "loomi-default",
  "name": "Loomi Default",
  "description": "General Loomi assistant behavior.",
  "system_prompt": "Runtime-only prompt text.",
  "model_route": {
    "provider_id": "custom",
    "model": "selected-model"
  },
  "allowed_tool_names": ["runtime.get_current_time"],
  "reasoning_mode": "balanced",
  "budget_summary": "Default local development budget.",
  "version": "2026-05-25.1",
  "is_default": true
}
```

## Sync Behavior

Rules:

- Sync validates required fields before writing.
- Sync upserts the persona identity by stable built-in slug.
- Sync creates the version when `(persona, version)` does not exist.
- Sync makes the configured version active for new runs.
- Sync is idempotent when the same config is applied repeatedly.
- Sync must not delete historical versions used by old runs.
- Sync must leave enough safe summary data for list/read surfaces without exposing prompt text.

## Validation Failures

Sync fails before activation when:

- name, system prompt, model route, or version is missing
- allowed tool name is unknown to the existing runtime allowlist
- more than one built-in config is marked default
- no active default persona exists after sync

Failure output should include a stable error code and safe message. It must not include raw prompt text.

## Expected Result

```json
{
  "synced": 1,
  "created_personas": 1,
  "created_versions": 1,
  "activated_versions": 1,
  "default_persona_slug": "loomi-default"
}
```
