---
title: Background Task UX
description: Read-only M6.5 observer for worker job state, diagnostics, and timeline events.
---

M6.5 upgrades the Background tasks panel from a placeholder into a read-only observer. Its purpose is visibility, not worker control.

## Snapshot selection

The panel chooses one observable snapshot:

1. job for the selected Chat run
2. empty state when no selected run job exists

The panel must not expose retry, recover, cancel, claim, or lease controls.

## Empty state

When there is no selected run job, the panel explains that no background task is running and that real model messages will surface queued jobs and worker events.

## Job and observation states

The productized labels cover:

- queued
- leased
- retrying
- recovering
- completed
- failed
- cancelled
- dead

`recovering` can be a backend job status or a read-only observation derived from `job_recovering`. M6.5 does not require rewriting the M6 job data model.

## Worker diagnostics

Worker diagnostics are summarized as display-only metadata. Useful fields include worker id, lease state, attempt count, heartbeat time, and sanitized diagnostic message when available.

## Timeline relationship

RunRail/Timeline remains the event-detail surface. Background tasks should align with timeline labels for:

- `job_claimed`
- `lease_renewed`
- `job_recovering`
- `job_retry_scheduled`
- `job_attempt_failed`
- `job_retry_exhausted`
- cancellation
- worker diagnostics
- unknown worker/job events

M6.5 preserves incoming event order and keeps raw event type visible or accessible for debugging.
