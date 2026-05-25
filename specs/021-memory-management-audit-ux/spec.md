# Feature Specification: M14 Memory Management Audit UX

**Feature Branch**: `[021-memory-management-audit-ux]`

**Created**: 2026-05-25

**Status**: Ready for implementation prep

**Input**: User description: "Create and complete M14 / 021-memory-management-audit-ux. Build on M13 Memory Foundation by improving Settings > Memory as a usable management surface and adding user-readable audit/history. Do not implement automatic memory distillation, embeddings/RAG, OpenViking, activity recorder ingestion, MCP/worker/sandbox/multi-agent rewrites."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manage approved memories (Priority: P1)

As a Loomi user, I want Settings > Memory to show my approved memories with search, filters, details, and safe deletion, so that I can understand and control what Loomi remembers.

**Why this priority**: User control is the central M14 value. The management UI must be useful before audit history becomes meaningful.

**Independent Test**: Seed approved, tombstoned, empty, and unauthorized memory states, open Settings > Memory, search/filter the list, inspect one detail panel, and delete one entry only after confirmation.

**Acceptance Scenarios**:

1. **Given** approved memories exist for the current user, **When** the user opens Settings > Memory, **Then** Loomi shows safe summaries, scope/source metadata, status, redaction markers, and clear loading/error/empty/tombstoned states.
2. **Given** the user enters a search query, **When** results are loaded, **Then** Loomi shows only scoped approved memories whose safe summaries or metadata match the query.
3. **Given** scope/source filters are changed, **When** the list reloads, **Then** Loomi applies at least the smallest grounded combination of thread/workspace/source_run/source_type filters supported by current data.
4. **Given** the user opens a memory detail panel, **When** details are displayed, **Then** the panel shows safe summary, scope, source run/thread, created/updated/deleted status, and redaction state without raw memory or unsafe traces.
5. **Given** the user deletes a memory, **When** they confirm deletion, **Then** Loomi tombstones the memory, removes it from active search/list results, and keeps a safe deleted state/audit record.

---

### User Story 2 - Review safe memory history (Priority: P2)

As a Loomi user, I want to see a readable history of memory write proposals, approvals, denials, deletions, and snapshot loads, so that I can audit how memory changed without seeing unsafe raw data.

**Why this priority**: Memory trust depends on knowing what happened and why. Audit history must use real backend data rather than UI-only fake events.

**Independent Test**: Create real memory proposal, approval, denial, deletion, and snapshot-load events, then call the scoped audit API and open Settings > Memory history to verify safe metadata only.

**Acceptance Scenarios**:

1. **Given** memory audit events exist, **When** the user views Memory history, **Then** Loomi lists `memory_write_proposed`, `memory_write_approved`, `memory_write_denied`, `memory_deleted`, and `memory_snapshot_loaded` events using safe metadata.
2. **Given** no independent audit-list API exists, **When** M14 is implemented, **Then** Loomi adds the smallest scoped backend read endpoint backed by existing productdata memory events rather than fabricating UI-only history.
3. **Given** audit metadata contains source references, **When** the event is returned, **Then** Loomi shows only safe scope/source/status/redaction metadata and excludes raw memory, secrets, provider traces, tool output, and local paths.
4. **Given** a memory operation references a run that is already terminal, **When** audit history is loaded, **Then** the memory audit item remains visible instead of being dropped by terminal-run event restrictions.

### Edge Cases

- The memory list is empty, still loading, fails to load, or contains only tombstoned entries.
- Search query is empty, long, or contains prompt-like text.
- Filter values reference a thread/run/source outside the current user's scope.
- A detail request or delete request targets an out-of-scope or already tombstoned memory.
- Delete is clicked accidentally or retried after a timeout.
- Audit events contain raw memory-like fields, provider traces, tool output, local paths, or secret-looking values.
- Existing productdata audit/event rows are missing for older entries.
- Snapshot history exists without matching visible memory entries.
- Existing terminal runs receive memory proposals, approvals, denials, or deletes after completion.
- Redaction input contains `/home/...`, Windows paths, `stdout`, `stderr`, or provider trace strings.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST create `specs/021-memory-management-audit-ux/` with Spec Kit artifacts for M14.
- **FR-002**: Settings > Memory MUST list approved memory entries visible to the current user with safe summary, scope, source metadata, status, timestamps, and redaction markers.
- **FR-003**: Settings > Memory MUST support search over safe summary/metadata through the real API boundary.
- **FR-004**: Settings > Memory MUST support scope/source filtering with at least the grounded subset currently backed by memory data among `thread`, `workspace`, `source_run`, and `source_type`.
- **FR-005**: Settings > Memory MUST provide a memory detail drawer or modal with safe summary, scope, source run/thread, created/updated/deleted state, and redaction markers.
- **FR-006**: Memory deletion MUST require explicit confirmation and MUST NOT happen on a single accidental click.
- **FR-007**: Memory UI MUST render clear loading, empty, error, and tombstoned/deleted states.
- **FR-008**: Loomi MUST expose user-readable memory history for write proposed, approved, denied, deleted, and snapshot loaded events using real backend data.
- **FR-009**: If no existing audit list endpoint can provide the history, Loomi MUST add the smallest scoped read endpoint backed by existing productdata memory audit/event data.
- **FR-010**: Memory list/detail/history responses MUST NOT expose raw memory, secrets, provider traces, tool output, local paths, credentials, or unsafe unredacted text.
- **FR-011**: Backend memory management and audit reads MUST preserve no-existence-leak behavior and scope authorization.
- **FR-012**: Duplicate delete, deny, and approve requests MUST remain idempotent.
- **FR-013**: M14 MUST fix or explicitly gate known blockers before full UX implementation: thread-scoped memory read/delete authorization, terminal-run audit preservation, expanded redaction for `/home`, Windows paths, stdout/stderr, and provider traces, and unified list/search scope API shape.
- **FR-014**: Settings > Memory component tests MUST cover list/search/filter/detail/delete confirmation/history and major empty/error/loading states.
- **FR-015**: Documentation MUST update memory architecture, API, local runbook, devlog, roadmap/current-status, and Spec Kit workflow for M14.
- **FR-016**: M14 MUST NOT implement automatic distillation, OpenViking provider, vectors/embedding/RAG, browser/activity recorder ingestion, MCP, worker queue, sandbox, or multi-agent rewrites.

### Key Entities

- **Memory Management Item**: Safe UI/API projection of an approved or tombstoned memory entry, including summary, scope, source metadata, status, timestamps, and redaction markers.
- **Memory Detail**: Safe expanded projection for one memory entry with deletion state and provenance.
- **Memory Filter**: Scoped search/filter request over safe memory metadata.
- **Memory Audit Item**: Safe event projection for memory write proposal, approval, denial, deletion, or snapshot load.
- **Delete Confirmation**: UI state that separates delete intent from confirmed tombstone action.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A browser smoke can open Settings > Memory and exercise list, search, filter, detail, delete confirmation, and audit history.
- **SC-002**: Backend tests prove memory list/detail/history remain scoped, thread-scoped entries are readable/deletable by their owner, terminal-run audit is retained, and out-of-scope requests do not leak existence.
- **SC-003**: Redaction tests prove list/detail/history omit raw memory, secrets, provider traces, tool output, local paths, `/home`, Windows paths, stdout, stderr, and Authorization/env-like content.
- **SC-004**: UI tests cover loading, empty, error, tombstoned/deleted, detail, delete confirmation, search/filter, and audit states.
- **SC-005**: Required validation commands pass: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.
- **SC-006**: The resulting diff contains no implementation of distillation, OpenViking, vector/embedding/RAG, activity recorder ingestion, MCP, worker queue, sandbox, or multi-agent rewrites.

## Assumptions

- M13 memory tables, services, redaction helpers, and Settings > Memory placeholder/API boundaries remain available.
- Existing user/thread/workspace/run authorization patterns are reused.
- Audit history can be sourced from current productdata memory events or minimally extended service reads.
- Filtering prioritizes grounded current fields over speculative UI controls.
- Memory content is always treated as untrusted data, even after approval.
