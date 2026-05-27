# Feature Specification: Memory Proposal Review

**Feature Branch**: `045-memory-proposal-review`
**Created**: 2026-05-26
**Status**: Candidate
**Input**: Continue the memory parity work by letting users review post-run memory proposals.

## User Stories

### Story 1 - See pending memory proposals

As a user with post-run memory organization enabled, I want pending proposals to appear in Settings > Memory, so I can decide what becomes saved memory.

**Acceptance Criteria**

1. Given pending proposals exist, when Settings > Memory loads, then it shows proposal title, safe summary, scope, redaction state, and created time.
2. Given proposals are listed, then raw proposal content and idempotency keys are not exposed to the frontend.
3. Given no proposals exist, then the page shows a neutral empty state.

### Story 2 - Approve or reject a proposal

As a user reviewing memory, I want to save or reject each proposal from Settings > Memory, so long-term memory remains user-controlled.

**Acceptance Criteria**

1. Approving a pending proposal creates an approved memory entry through the existing approval API and refreshes saved memory, proposals, and audit history.
2. Rejecting a pending proposal removes it from the pending list and refreshes audit history.
3. Approval and rejection do not use raw content in UI state.

## Requirements

- **FR-001**: Productdata MUST expose a safe pending proposal list with status/scope/source filters and bounded limit.
- **FR-002**: HTTP MUST expose `GET /v1/memory/write-proposals`.
- **FR-003**: Proposal list responses MUST omit `content`, idempotency keys, and hidden user fields.
- **FR-004**: Frontend API clients MUST map pending proposals and expose approve/deny methods.
- **FR-005**: Settings > Memory MUST render pending proposals above approved memories.
- **FR-006**: Settings > Memory MUST refresh proposals, audit, and entries after approval/denial.

## Non-goals

- No bulk approve or bulk deny.
- No editing proposal content.
- No direct approved memory creation from UI.
- No LLM distillation or semantic/vector provider execution.
