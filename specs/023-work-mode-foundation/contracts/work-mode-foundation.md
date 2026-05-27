# Contract: M16 Work Mode Foundation

M16 does not add a backend endpoint. The UI consumes existing thread, message, run, and event contracts and recognizes optional safe metadata fields when present.

## Existing Inputs

- `Thread.mode`: `chat | work`
- `Message.content`: text used as fallback goal/step source
- `Run.status`: current status
- `Run.events[]`: replayed progress source
- `RunEvent.metadata`: optional safe metadata map

## Optional Safe Metadata Shape

```json
{
  "work_goal": "Ship M16 work mode foundation",
  "work_steps": [
    { "id": "step-plan", "title": "Plan projection", "status": "completed", "summary": "Projection contract defined" }
  ],
  "work_artifacts": [
    {
      "id": "artifact-plan",
      "title": "Work mode plan",
      "type": "markdown",
      "source_thread_id": "thread-brief",
      "source_run_id": "run-1",
      "summary": "Safe metadata preview only",
      "created_at": "2026-05-25T10:00:00Z",
      "updated_at": "2026-05-25T10:05:00Z"
    }
  ]
}
```

## Safety Rules

- Unknown artifact fields are ignored.
- Secret-looking values are redacted before render.
- Paths, shell commands, browser automation hints, and executable controls are not rendered as actions.
