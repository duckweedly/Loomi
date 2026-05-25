# Contract: Work Artifact Evidence

## Local Seed Contract

M17 adds an explicit local-dev/test seed scenario:

```sh
LOOMI_SEED_SCENARIO=m17-work-artifact go run ./cmd/loomi-seed
```

Expected output includes:

- `thread_id`
- `message_id`
- `run_id`
- `event_id`

This is not an HTTP API and not a production write surface.

## Event Metadata Contract

```json
{
  "work_goal": "Close out M17 Work artifact evidence",
  "work_steps": [
    {
      "id": "m17-step-plan",
      "title": "Create repeatable evidence path",
      "status": "completed",
      "summary": "Seed writes existing thread/message/run/event data"
    }
  ],
  "work_artifacts": [
    {
      "id": "m17-artifact-evidence",
      "title": "M17 Work artifact evidence",
      "type": "markdown",
      "source_thread_id": "thr_m17_work_artifact",
      "source_run_id": "run_x",
      "summary": "Safe metadata-only evidence card",
      "created_at": "2026-05-25T00:00:00Z",
      "updated_at": "2026-05-25T00:00:00Z",
      "redaction_applied": true
    }
  ]
}
```

## Display Contract

Artifact cards may display only:

- id
- title
- type
- source thread/run
- summary
- created/updated
- redaction marker

Artifact cards must not display action buttons or executable controls.
