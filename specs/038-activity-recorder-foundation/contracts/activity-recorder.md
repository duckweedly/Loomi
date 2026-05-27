# Contract: M30 Activity Recorder Foundation

## Status

`GET /v1/activity-recorder/status`

```json
{
  "status": {
    "enabled": false,
    "event_count": 0,
    "redaction_applied": false
  }
}
```

## Enable

`POST /v1/activity-recorder/enable`

```json
{
  "status": {
    "enabled": true,
    "enabled_at": "2026-05-26T00:00:00Z",
    "event_count": 0,
    "redaction_applied": false
  }
}
```

## Disable

`POST /v1/activity-recorder/disable`

Returns the same status shape with `enabled = false`.

## Append Event

`POST /v1/activity-recorder/events`

```json
{
  "kind": "browser",
  "source": "manual",
  "summary": "Reviewed docs page",
  "metadata": {
    "title": "Docs",
    "url_host": "example.com"
  }
}
```

Response:

```json
{
  "event": {
    "id": "act_...",
    "kind": "browser",
    "source": "manual",
    "summary": "Reviewed docs page",
    "metadata": {
      "title": "Docs",
      "url_host": "example.com"
    },
    "created_at": "2026-05-26T00:00:00Z",
    "redaction_applied": false
  },
  "status": {
    "enabled": true,
    "event_count": 1,
    "redaction_applied": false
  }
}
```

## List Events

`GET /v1/activity-recorder/events?limit=20`

```json
{
  "items": [],
  "status": {
    "enabled": false,
    "event_count": 0,
    "redaction_applied": false
  }
}
```

## Clear Events

`DELETE /v1/activity-recorder/events`

```json
{
  "status": {
    "enabled": true,
    "event_count": 0,
    "last_cleared_at": "2026-05-26T00:00:00Z",
    "redaction_applied": false
  }
}
```

## Rejections

- append while disabled
- unsupported kind
- missing source or summary
- oversized summary or metadata
- raw credentials, Authorization headers, private paths, screenshots, keystrokes, clipboard data, raw shell output, raw browser HTML, or file contents
