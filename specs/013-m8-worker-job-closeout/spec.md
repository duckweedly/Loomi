# Feature Specification: M8 Worker Job Closeout

**Feature Branch**: `[013-m8-worker-job-closeout]`

**Created**: 2026-05-25

**Status**: Draft

**Input**: User description: "Create a Spec Kit feature to audit whether the original M8 Worker + Job Queue roadmap is already covered by specs/008-worker-job-pipeline and current code, implement only the smallest uncovered M8 gap, and document the closeout."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Audit M8 Coverage (Priority: P1)

As a Loomi maintainer, I want the original M8 worker/job queue scope checked against the already implemented M6 worker job pipeline, so that the roadmap can move forward without duplicate implementation.

**Why this priority**: The highest value is deciding whether M8 is already covered and avoiding a second worker queue implementation.

**Independent Test**: Review the M8 route-map items and verify each item has an existing spec, code, test, or documentation evidence entry.

**Acceptance Scenarios**:

1. **Given** the original M8 list, **When** the audit compares it with M6 worker/job evidence, **Then** each item is marked covered, partially covered, or missing with a cited source.
2. **Given** an item is already covered, **When** tasks are generated, **Then** no duplicate implementation task is created for that item.
3. **Given** a missing item is found, **When** implementation begins, **Then** only that smallest M8 gap is changed.

---

### User Story 2 - Close the Smallest Gap (Priority: P2)

As a developer validating worker reliability, I want retry scheduling to include a real backoff delay when recovering stale work, so that retry/backoff in the M8 roadmap is represented by behavior instead of only by an event name.

**Why this priority**: The audit found durable jobs, enqueue, claim, lease renewal, recovery, terminal failure, and stale-owner guards are present; retry/backoff was the only behavior-level gap.

**Independent Test**: Simulate an expired lease, recover the job, verify the job is not immediately claimable, then advance to the scheduled retry time and verify the next worker can claim it.

**Acceptance Scenarios**:

1. **Given** a worker-owned job has an expired lease, **When** recovery reschedules it, **Then** the job receives a future retry time.
2. **Given** the retry time has not arrived, **When** a worker attempts to claim the job, **Then** the claim does not succeed.
3. **Given** the retry time has arrived, **When** a worker claims again, **Then** the job is claimable and ownership version protection still applies.

---

### User Story 3 - Record M8 Closeout (Priority: P3)

As a future Loomi maintainer, I want roadmap and devlog documentation to state that original M8 is closed by the M6 pipeline plus the closeout patch, so that M9 work does not re-open old worker-queue questions.

**Why this priority**: The project constitution treats docs as part of done, and this feature is primarily a roadmap closeout decision.

**Independent Test**: Read the docs-site roadmap and devlog and confirm they state the M8 closeout decision, evidence, validation commands, and remaining non-goals.

**Acceptance Scenarios**:

1. **Given** the docs-site roadmap is opened, **When** the M8 status is read, **Then** it states that original M8 is covered and closeout passed.
2. **Given** the devlog is read, **When** validation evidence is needed, **Then** it lists the audit result, the minimal gap patched, and the exact validation commands.
3. **Given** later roadmap items are reviewed, **When** M9 is considered, **Then** the docs keep RunContext/Pipeline work out of this closeout.

### Edge Cases

- The audit discovers a capability described in docs but not backed by code or tests.
- A retry backoff patch makes worker recovery too slow for local validation.
- A stale worker attempts to complete, fail, or retry after recovery has reassigned ownership.
- Documentation could imply M9 RunContext/Pipeline work was completed by this closeout.
- Existing untracked local agent/config files must remain untouched.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The audit MUST cover jobs table, transactionally enqueued run execution work, pending-job claim, lease heartbeat or renewal, retry/backoff, failed terminal state, lost-lock ownership guard, immediate run creation acknowledgement, worker crash recovery, and stale-worker terminal write prevention.
- **FR-002**: The audit MUST identify whether each original M8 item is covered by current M6 worker job pipeline behavior before creating implementation work.
- **FR-003**: The feature MUST NOT implement a duplicate worker queue when existing behavior already covers an M8 item.
- **FR-004**: The feature MUST patch retry/backoff only if recovery scheduling lacks an actual delay before the next claim.
- **FR-005**: The retry/backoff patch MUST preserve worker crash recovery, max-attempt terminal failure, and stale-owner completion/failure guards.
- **FR-006**: The feature MUST NOT add M9 RunContext/Pipeline middleware, Redis, external queues, multi-worker platformization, MCP, Memory, Desktop Runtime, or new tool execution behavior.
- **FR-007**: Documentation MUST update roadmap/current-status and add a devlog entry with evidence and validation results.
- **FR-008**: Validation MUST include targeted worker/job tests, the requested Go package test command, and docs-site build when docs change.

### Key Entities *(include if feature involves data)*

- **M8 Audit Item**: One original roadmap requirement and its current coverage status.
- **Background Job**: Existing durable job created for run execution.
- **Worker Lease**: Existing ownership record used for renewal, recovery, and stale-owner protection.
- **Retry Schedule**: The next eligible claim time after a stale lease is recovered.
- **Closeout Record**: Roadmap and devlog evidence that original M8 has passed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 10 of 10 original M8 audit items have a documented covered or patched result.
- **SC-002**: A recovered stale job is scheduled in the future in 100% of targeted retry/backoff tests.
- **SC-003**: A stale worker cannot complete or fail a recovered job in 100% of targeted ownership guard tests.
- **SC-004**: The requested Go validation command and docs-site build complete successfully, or exact blockers are recorded.
- **SC-005**: Roadmap and devlog state that M8 closeout passed without claiming M9 RunContext/Pipeline completion.

## Assumptions

- `specs/008-worker-job-pipeline` is the authoritative implemented worker/job pipeline feature for this audit.
- Current M7/M7.5 code may extend run statuses, but original M8 closeout is judged against worker/job queue reliability only.
- A small fixed/exponential local retry backoff is sufficient for the first closeout; hosted production retry policy remains later work.
- Browser smoke is not required unless frontend state changes are made.
