# Contract: Runtime Event Groups

## Purpose

Define how run events are grouped in Timeline and Debug surfaces as backend events become richer.

## Primary groups

| Group | Purpose | Example event types |
|-------|---------|---------------------|
| Run lifecycle | High-level run state and terminal outcomes | `run.created`, `run.started`, `run.completed`, `run.stopped`, `run.cancelled`, `run.retrying`, `run.recovering` |
| Model stream | Model generation progress and usage | `model.started`, `model.delta`, `model.final`, `model.usage`, `assistant.drafting`, `assistant.message.completed` |
| Worker/job | Queue and worker ownership activity | `job.queued`, `job.claimed`, `worker.claimed`, `job.retrying`, `worker.released` |
| Error | Provider, stream, backend, and execution failures | `provider.error`, `model.error`, `stream.error`, `backend.unavailable`, `run.failed` |

## Grouping rules

- Every displayed event maps to one primary group.
- Error group takes precedence when an event has both lifecycle and error semantics.
- Unknown non-error events default to Run lifecycle only if they describe run state; otherwise they appear in the closest existing group by event prefix.
- Unknown error-like events map to Error when their type, status, or severity indicates failure.
- Event groups should remain visible and stable even when a run has no events in a group.

## Event display fields

- `label`: Short human-readable label.
- `detail`: Concise detail suitable for debug review.
- `time`: User-visible event time.
- `status`: Event status or severity when available.
- `usage`: Optional usage summary for model token events.

## Ordering rules

- Preserve chronological ordering within each group.
- Preserve the run's overall terminal status outside the group list.
- Replayed events must not create duplicates.

## Acceptance checks

- Lifecycle and model stream events appear in separate groups.
- Queue, worker claim, and retry events appear in Worker/job.
- Provider and stream errors appear in Error and are visually distinct.
- Token usage appears in Model stream or debug detail, not in the main assistant message text.
