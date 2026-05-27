# Data Model: M16 Work Mode Foundation

## Work Thread

- Existing `Thread` with `mode = work`.
- Reuses existing thread id, title, run status, messages, and current run.
- No separate task or work item table.

## Work Plan Projection

- `goal`: Safe display string derived from work metadata or first user message.
- `steps`: Ordered `WorkStep[]`.
- `status`: Current run/work status derived from the current run and most recent relevant event.
- `artifacts`: Metadata-only `WorkArtifactReference[]`.
- `recentEvents`: Latest safe run events relevant to work progress.
- `emptyReason`: Optional display reason when no plan data exists.

## Work Step

- `id`: Stable projection id.
- `title`: Safe display string.
- `status`: `pending | running | completed | blocked | failed`.
- `summary`: Optional safe display string.

## Work Artifact Reference

- `id`: Stable projection id.
- `title`: Safe display string.
- `type`: Safe category string such as `note`, `plan`, `markdown`, `report`, or `artifact`.
- `sourceThreadId`: Optional thread id from safe metadata.
- `sourceRunId`: Optional run id from safe metadata.
- `summary`: Markdown-like safe preview text.
- `createdAt`: Optional display timestamp.
- `updatedAt`: Optional display timestamp.

## Recent Progress Event

- Existing run event id, type, detail, time, and status after frontend redaction.
- Events are sorted by existing event order and capped for display.
