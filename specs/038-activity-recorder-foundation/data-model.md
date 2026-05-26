# Data Model: M30 Activity Recorder Foundation

## ActivityRecorderStatus

- `enabled`: whether Activity Recorder accepts new summaries
- `enabled_at`: timestamp for the most recent enable action
- `disabled_at`: timestamp for the most recent disable action
- `last_event_at`: timestamp for the newest stored activity event
- `last_cleared_at`: timestamp for the most recent cleanup action
- `event_count`: number of currently stored activity events
- `redaction_applied`: whether status metadata needed redaction

## ActivityEvent

- `id`: stable activity event id
- `kind`: supported activity kind, initially `app`, `browser`, `file`, `command`, or `note`
- `source`: bounded source label such as `manual`, `browser`, `workspace`, or `runtime`
- `summary`: bounded redacted summary
- `metadata`: optional bounded safe metadata map
- `created_at`: service timestamp
- `redaction_applied`: whether summary or metadata was redacted

## AppendActivityEventInput

- `kind`: required supported activity kind
- `source`: required bounded source label
- `summary`: required bounded summary
- `metadata`: optional bounded safe metadata

## ActivityRecorderAudit

- `event_type`: `activity_recorder_enabled`, `activity_recorder_disabled`, `activity_event_recorded`, or `activity_events_cleared`
- `summary`: safe audit summary
- `occurred_at`: event timestamp
- `redaction_applied`: whether audit metadata was redacted
