# Feature Specification: Memory Post-run Proposals

**Feature Branch**: `044-memory-post-run-proposals`
**Created**: 2026-05-26
**Status**: Candidate
**Input**: Continue Arkloop-parity memory slice by making "organize after each run" functional in Loomi.

## User Stories

### Story 1 - Generate a reviewable memory proposal after a run

As a user who enables memory and turns on per-run organization, I want Loomi to create a reviewable memory proposal after a successful run, so useful outcomes can be saved without direct automatic insertion.

**Acceptance Criteria**

1. Given memory is enabled, configured, and commit-after-run is true, when a run completes and has a persisted assistant message, then Loomi creates one pending memory write proposal linked to that run and thread.
2. Given the same completed run is processed again, when post-run proposal creation runs, then Loomi returns the existing proposal instead of creating a duplicate.
3. Given memory is disabled, unconfigured, or commit-after-run is false, when a run completes, then no post-run memory proposal is created.

### Story 2 - Preserve approval and audit boundaries

As a user managing memory, I want automatically generated run memories to remain pending until I approve them, so search and RunContext snapshots only include explicitly approved memory.

**Acceptance Criteria**

1. Pending post-run proposals do not appear in memory list/search.
2. Post-run proposals are visible in memory audit/history through `memory_write_proposed`.
3. The proposal uses existing redaction and safety classification.

## Requirements

- **FR-001**: The worker MUST attempt post-run proposal creation only after the runner has completed a run successfully.
- **FR-002**: The worker MUST require `MemoryProviderStatus.Enabled`, `Configured`, and `CommitAfterRun`.
- **FR-003**: The proposal MUST be `thread` scoped with `scope_id` equal to the source thread id.
- **FR-004**: The proposal MUST include `source_thread_id`, `source_run_id`, and an idempotency key derived from the run id.
- **FR-005**: Loomi MUST NOT auto-approve post-run proposals.
- **FR-006**: Failed, stopped, non-terminal, or assistant-message-less runs MUST NOT create post-run proposals.
- **FR-007**: Settings > Memory copy MUST explain that per-run organization creates approval-gated proposals.

## Non-goals

- No LLM summarization/distillation worker.
- No embedding/vector search implementation.
- No external semantic memory provider execution.
- No automatic approval.
- No new HTTP endpoint.
