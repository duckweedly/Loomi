# Contract: Persona Resolution and RunContext Snapshot

## Purpose

Resolve the persona used by a run and attach the exact persona snapshot/version to RunContext before provider/runtime invocation.

## Resolution Input

```json
{
  "run_id": "run_...",
  "thread_id": "thread_...",
  "job_id": "job_...",
  "run_persona_id": "persona_optional",
  "thread_persona_id": "persona_optional"
}
```

## Resolution Order

```text
run_persona_id
-> thread_persona_id
-> active default built-in persona
-> prepare_context failure
```

Rules:

- Run override wins over thread selection.
- Thread selection wins over default built-in.
- Resolved persona must be active and belong to the correct scope/source.
- Resolved version is captured before runtime invocation.
- Existing run/provider route overrides may remain more specific than persona model route only when already supported by current run creation semantics.

## RunContext Output

```json
{
  "persona": {
    "id": "persona_...",
    "slug": "loomi-default",
    "version": "2026-05-25.1",
    "name": "Loomi Default",
    "description": "General Loomi assistant behavior.",
    "model_route": {
      "provider_id": "custom",
      "model": "selected-model"
    },
    "allowed_tool_names": ["runtime.get_current_time"],
    "reasoning_mode": "balanced",
    "budget_summary": "Default local development budget.",
    "resolved_from": "default"
  }
}
```

Runtime-only fields:

- `systemPrompt` can be used to construct provider/system context.
- `systemPrompt` must not be included in normal persisted pipeline event metadata or Timeline/debug summaries.

## Failure Semantics

Resolution fails before runtime invocation when:

- selected persona does not exist
- selected persona is inactive
- selected persona has no active version
- persona scope/source does not match the current run boundary
- model route cannot be resolved
- allowed tools include unsupported names and no safe intersection can be formed
- no default persona is available

Failure output:

```json
{
  "stage": "prepare_context",
  "error_code": "persona_resolution_failed",
  "message": "Persona could not be resolved for this run."
}
```

The failure must be redacted and must use existing worker/job terminal failure semantics.
