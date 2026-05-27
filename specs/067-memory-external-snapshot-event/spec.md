# Feature Spec: M67 Memory External Snapshot Event

## Goal

External provider prompt recall must be observable in the run timeline, not hidden inside prompt assembly.

## User Story

As a user or developer inspecting a run, I can see that Loomi loaded an external memory snapshot before the model request, including provider and safe count metadata.

## Functional Requirements

- Record a progress run event after external provider prompt recall succeeds.
- Use event type `memory_external_snapshot_loaded`.
- Metadata must include provider, status, entry count, limit, and redaction flag.
- Metadata must not include query text, raw hit content, credentials, provider traces, or local paths.
- Provider failures or empty results must keep the previous no-event fallback.

## Non-Goals

- No UI redesign.
- No memory audit history entry.
- No background snapshot cache.

## Success Criteria

- Runtime test proves the event is recorded with safe metadata.
