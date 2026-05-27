# Data Model: Memory Provider Foundation

## MemoryProviderConfig

- `enabled`: whether memory participates in future run preparation.
- `provider`: selected provider id. Known values for this slice: `local`, `semantic`.
- `commit_after_run`: preference reserved for later distillation; exposed now so Settings and backend agree on the state.
- `semantic_endpoint`: optional non-secret endpoint/base URL used only for readiness projection in this slice.
- `semantic_configured`: derived boolean; true only when required non-secret configuration exists.
- `updated_at`: server timestamp.

Validation:

- Missing config defaults to enabled local memory.
- Unknown provider normalizes to local and emits a degraded diagnostic.
- Secret fields are not part of the response model.

## MemoryProviderStatus

- `enabled`
- `provider`
- `label`
- `state`: `disabled`, `available`, `unconfigured`, `healthy`, `unhealthy`, or `degraded`
- `configured`
- `checked_at`
- `diagnostic`: safe code/message pair, no raw headers, keys, tokens, or provider traces

Rules:

- Disabled memory always returns `state=disabled`.
- Local provider returns `state=available`.
- Semantic provider with missing required config returns `state=unconfigured`.
- Semantic provider with failed health check returns `state=unhealthy`.
- Unknown provider returns local provider with `state=degraded`.

## MemoryReadinessSnapshot

- `enabled`
- `provider`
- `state`
- `configured`
- `checked_at`
- `diagnostic`

Rules:

- Snapshot is safe to include in run metadata/events.
- Snapshot must not cause run preparation to fail.
- Snapshot must not include endpoint credentials or raw provider errors.
