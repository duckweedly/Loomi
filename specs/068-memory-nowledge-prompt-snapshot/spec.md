# Feature Spec: M68 Memory Nowledge Prompt Snapshot Regression

## Goal

Nowledge must have the same automatic prompt-memory recall evidence as OpenViking.

## User Story

As a user running with Nowledge selected, I can trust that Loomi recalls safe Nowledge memories before the model request and surfaces the recall in the run timeline.

## Functional Requirements

- Reuse the existing external prompt recall path for Nowledge.
- Search Nowledge with the latest user message and bounded limit.
- Inject safe Nowledge hits into the initial `<memory>` prompt block.
- Record the same `memory_external_snapshot_loaded` progress event with provider `nowledge`.
- Do not include query text, memory body fields, API keys, provider traces, or local paths in event metadata.

## Non-Goals

- No new Nowledge endpoint shape.
- No Settings UI redesign.
- No background cache or automatic provider install.

## Success Criteria

- Runtime regression proves Nowledge prompt recall, event metadata, and secret redaction.
