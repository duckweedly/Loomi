# Feature Specification: Memory Proposal Edit

**Feature Branch**: `046-memory-proposal-edit`
**Created**: 2026-05-26
**Status**: Candidate
**Input**: Continue memory parity by letting users adjust pending memory proposals before saving them.

## User Stories

### Story 1 - Edit a pending proposal before saving

As a user reviewing pending memory, I want to refine the proposed title and summary before approval, so saved memory reflects what I actually want Loomi to remember.

**Acceptance Criteria**

1. Given a pending proposal is visible in Settings > Memory, when I open edit mode, then I can change its title and summary.
2. Given I save edits, then the pending card updates without approving the proposal.
3. Given I approve after editing, then the approved memory entry uses the edited title and summary.

### Story 2 - Keep editing safe and bounded

As a user, I want proposal editing to preserve the memory safety boundary, so raw hidden fields and cross-user proposals are not exposed or mutable.

**Acceptance Criteria**

1. Only pending proposals can be edited.
2. The edit API accepts only title and summary.
3. List responses still omit raw content, idempotency keys, and hidden user fields.

## Requirements

- **FR-001**: Productdata MUST expose an update operation for pending memory proposals.
- **FR-002**: HTTP MUST expose `PATCH /v1/memory/write-proposals/{proposal_id}`.
- **FR-003**: Updating a proposal MUST normalize and redact title/summary through the existing memory content boundary.
- **FR-004**: Approving an edited proposal MUST create the memory entry from the edited summary.
- **FR-005**: Settings > Memory MUST allow editing title and summary before save/decline.
- **FR-006**: Frontend save/cancel edit state MUST not expose raw content or idempotency keys.

## Non-goals

- No bulk edit.
- No editing already approved or denied proposals.
- No direct raw-content reveal.
- No semantic/vector retrieval changes.
