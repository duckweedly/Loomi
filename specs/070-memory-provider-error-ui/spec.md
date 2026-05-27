# Feature Spec: M70 Memory Provider Error UI

## Goal

Settings > Memory recent errors should make runtime provider failures visibly distinguishable from configuration diagnostics.

## User Story

As a user looking at Memory settings, I can see whether a recent provider issue came from a run-time recall event and which safe run/event id it belongs to.

## Functional Requirements

- Preserve the existing recent errors panel.
- Include safe runtime `event_type` and `run_id` details when returned by `/v1/memory/errors`.
- Keep long ids readable without breaking the Settings layout.
- Do not expose prompt text, raw memory content, provider traces, API keys, tokens, or local paths.

## Non-Goals

- No navigation to run detail.
- No modal redesign.
- No raw log viewer.

## Success Criteria

- Component regression confirms runtime error fields are rendered.
- Frontend build passes.
