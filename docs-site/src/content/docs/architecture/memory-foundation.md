---
title: M13 Memory Foundation Architecture
description: PG-backed memory entries, safe RunContext snapshots, approval-gated writes, and user-controlled deletion.
---

M13 adds the first local memory slice for Loomi. The implemented boundary is intentionally small: PostgreSQL-backed memory rows, text search, approval-gated memory writes, a safe memory snapshot in `RunContext`, and a minimal Settings UI for view/search/delete.

M13.5 closes the implementation with a real Postgres/httpapi smoke that applies migrations through the M13 tables and exercises the HTTP proposal/approval/list/search/delete path against the same repository used by `RunContext`.

M14 blocker foundation adds the minimum hardening needed before the full UX: thread-scoped read/delete authorization, durable memory audit rows, broader redaction, and one grounded search/list filter shape across docs, backend, and the frontend client.

M14 is the next UX/API contract slice. Its goal is not distillation or RAG; it makes Settings > Memory usable as a management surface and adds scoped, user-readable audit history backed by real memory events.

## Boundary

The current implemented boundary is `productdata`: the service and repository own memory durability, search, safety filtering, audit persistence, and the safe summaries loaded into `RunContext.MemorySnapshot`.

`MemoryProvider` remains the future extraction point described by the M13 design contract, but runtime code has not been moved behind an independent provider abstraction yet. OpenViking, embeddings, vector search, automatic distillation, marketplace/plugin memory providers, browser/activity recorder ingestion, and multi-agent long-term automation are outside this slice.

## Data Model

M13 adds:

- `memory_entries`: approved, tombstoned, or disabled user/thread-scoped memories.
- `memory_write_proposals`: pending/approved/denied write intents created by an agent or API caller.
- `memory_audit_events`: durable user-readable memory audit records for proposal, approval, denial, delete, and snapshot events.

Memory rows store a title, redacted summary, redacted content, scope, source thread/run/event ids, safety state, content hash, timestamps, and tombstone metadata. Delete is a tombstone operation: content is cleared, summary becomes `[deleted]`, and deleted metadata is retained for audit.

## Safe Snapshot

When the worker prepares a `RunContext`, it searches up to five approved safe memories visible to the run's thread:

- user-scoped memories for the same local identity
- thread-scoped memories for the active thread
- approved entries only
- blocked entries and tombstones excluded

The snapshot records `loaded`, `empty`, or `unavailable`, and appends a `memory_snapshot_loaded` run event with only counts/status/limits/redaction flags.

## Write Gating

Agent memory writes must enter `memory_write_proposals` first. A proposal is searchable by neither `GET /v1/memory` nor `POST /v1/memory/search` until approved.

Approval creates a new approved `memory_entries` row. Denial leaves no entry. If proposal content matches unsafe patterns, the proposal is marked denied/blocked and cannot be approved.

When a proposal references a source run, Loomi stores durable safe audit rows:

- `memory_write_proposed`
- `memory_write_approved`
- `memory_write_denied`
- `memory_entry_deleted`

Audit metadata includes ids, scope, status, and safety state only. Raw memory content, credentials, provider payloads, file contents, shell output, and browser/desktop captured state must not enter audit events.

If the source run is still non-terminal, Loomi also writes a related run timeline event. If the run is terminal, audit still succeeds and remains queryable from `/v1/memory/audit`; the run timeline is treated as an associated view, not the audit source of truth.

## User Control

The first UI boundary is Settings > Memory:

- list approved memories
- text-search safe summaries
- delete an entry by tombstoning it

The UI does not expose automatic distillation, provider selection, OpenViking, vector search, or bulk import controls.

M14 expands this surface to list/search/filter, detail drawer or modal, explicit delete confirmation, loading/empty/error/tombstoned states, and a real audit/history panel. The audit surface is backed by productdata memory events and must not fabricate UI-only history. The default UI keeps engineering filters collapsed and folds routine `memory_snapshot_loaded` events into a small system snapshot note, while write/delete/proposal events remain visible as human-readable history.

## M14 Management And Audit Flow

Settings > Memory reads safe projections from the memory API. List/search share one filter shape: `q` or JSON `query`, `scope_type`, `scope_id`, `source_thread_id`, `source_run_id`, `source_type`, `include_tombstoned`, and `limit`. `workspace_id` is deferred until workspace-scoped memory exists. Detail reads return only safe metadata: summary/title, scope, source thread/run/event ids, source type, status, timestamps, and redaction state.

Detail/delete authorization follows the memory scope. User-scoped memory is visible to the same user. Thread-scoped memory requires a matching `scope_type=thread&scope_id`, `source_thread_id`, or `source_run_id`; wrong or missing context returns generic not found.

Audit history reads `memory_write_proposed`, `memory_write_approved`, `memory_write_denied`, `memory_deleted`, and `memory_snapshot_loaded` from scoped `memory_audit_events`. Memory audit survives terminal source runs because users still need history after the run completes.

## Redaction

Memory creation and proposal creation run content through the same product data redaction helper used by events. Secret-looking input becomes redacted or blocked before it can be returned by HTTP, placed in a RunContext snapshot, or written to audit event metadata.

M14 expands the redaction requirement to common local and provider-output forms: `/Users/...`, `/home/...`, Windows paths, stdout/stderr dumps, tool output, provider traces, `.env` values, Authorization headers, tokens, credentials, key/env markers, and secret-like values.

## M13.5 Evidence

`TestM13MemoryRealPGHTTPAPISmoke` is the current end-to-end backend evidence. It verifies migrated `memory_entries` and `memory_write_proposals`, approval-gated creation, search/list visibility, RunContext safe snapshot load, tombstone exclusion, duplicate approve/deny/delete idempotency, out-of-scope non-leakage, and sensitive redaction across API responses and run-event metadata.
