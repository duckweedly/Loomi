# Data Model: M17 Work Artifact Evidence Closeout

## Work Evidence Seed

- Reuses existing Work thread with `mode = work`.
- Reuses existing user message as the seed anchor.
- Starts or reuses the current run for that thread.
- Appends a progress event containing Work metadata when not already present.
- Scope: local-dev/test seed command only.

## Work Event Metadata

- `work_goal`: string.
- `work_steps`: ordered array of step references.
- `work_artifacts`: ordered array of safe artifact references.
- Recent progress: existing run event sequence and summaries.

## Work Step

- `id`: stable local seed identifier.
- `title`: display label.
- `status`: pending, running, completed, blocked, or failed.
- `summary`: optional safe text.

## Artifact Evidence Reference

- `id`: stable artifact identifier.
- `title`: safe display title.
- `type`: safe category.
- `sourceThreadId`: existing source thread identifier.
- `sourceRunId`: existing source run identifier.
- `summary`: safe short summary.
- `createdAt`: optional timestamp.
- `updatedAt`: optional timestamp.
- `redactionApplied`: true when unsafe fields or values were redacted/omitted.

## State Rules

- Chat threads never produce a Work Plan projection.
- Work projection reads from current run events first, then message/thread fallback.
- Artifact evidence is never executable.
- Redaction is one-way for display; raw unsafe values are not rendered.
