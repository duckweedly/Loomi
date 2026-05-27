# Feature Spec: M69 Memory Provider Runtime Errors

## Goal

Settings > Memory recent errors must include safe runtime provider failures, not only static provider configuration diagnostics.

## User Story

As a user checking memory health, I can see when external memory recall failed during a run without exposing prompts, secrets, or raw provider traces.

## Functional Requirements

- Record a safe runtime event when external prompt-memory recall fails.
- Keep provider recall failures non-fatal for the run.
- Include recent runtime memory provider failures in `/v1/memory/errors`.
- Preserve existing provider configuration diagnostic errors.
- Error response items may include safe run/event identifiers but must not include query text, raw memory content, upstream trace bodies, API keys, Authorization headers, tokens, or local paths.

## Non-Goals

- No retry scheduler.
- No provider process restart.
- No raw provider log viewer.
- No Settings UI redesign.

## Success Criteria

- Runtime regression proves failed external recall writes a safe provider error event and recent error item.
- HTTP regression proves `/v1/memory/errors` returns runtime provider failures without secrets.
