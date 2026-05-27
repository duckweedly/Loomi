# Implementation Plan: Memory Post-run Proposals

## Context

This slice follows M42 provider configuration and M43 memory tools. It activates the existing `commit_after_run` setting without changing the public memory API or approval model.

## Technical Approach

- Add a runtime helper that reads run status, memory provider status, and the assistant message for the completed run.
- Use the existing `ProposeMemoryWrite` service method with thread scope and a deterministic idempotency key.
- Invoke the helper from runtime success closeout before background job completion.
- Treat proposal failure as non-fatal to the already-completed run.
- Update Settings copy and docs to describe pending proposals.

## Data Model

No migration. Reuses:

- `memory_provider_configs.commit_after_run`
- `memory_write_proposals`
- `memory_audit_events`
- assistant `messages.metadata.run_id`

## Validation

```bash
go test ./internal/runtime -run 'TestWorkerProposesPostRunMemory|TestPostRunMemory'
bun test --cwd web src/components/SettingsView.runtime.test.tsx
bun run --cwd docs-site build
```

## Risks

- Proposal content is deterministic and bounded, not LLM-distilled. This is intentional for this slice.
- The source run is terminal when the proposal is created, so durable memory audit is the source of truth rather than a run timeline event.
