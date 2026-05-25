---
title: M13 Memory Foundation Architecture
description: PG-backed memory entries, safe RunContext snapshots, approval-gated writes, and user-controlled deletion.
---

M13 adds the first local memory slice for Loomi. The implemented boundary is intentionally small: PostgreSQL-backed memory rows, text search, approval-gated memory writes, a safe memory snapshot in `RunContext`, and a minimal Settings UI for view/search/delete.

## Boundary

`productdata` owns memory durability and safety filtering. Runtime code sees memory through a `MemoryProvider` interface and receives only safe summaries in `RunContext.MemorySnapshot`.

The first provider is the existing product data service and PostgreSQL repository. OpenViking, embeddings, vector search, automatic distillation, marketplace/plugin memory providers, browser/activity recorder ingestion, and multi-agent long-term automation are outside this slice.

## Data Model

M13 adds:

- `memory_entries`: approved, tombstoned, or disabled user/thread-scoped memories.
- `memory_write_proposals`: pending/approved/denied write intents created by an agent or API caller.

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

When a proposal references a source run, M13 appends safe audit run events:

- `memory_write_proposed`
- `memory_write_approved`
- `memory_write_denied`
- `memory_entry_deleted`

Audit metadata includes ids, scope, status, and safety state only. Raw memory content, credentials, provider payloads, file contents, shell output, and browser/desktop captured state must not enter audit events.

## User Control

The first UI boundary is Settings > Memory:

- list approved memories
- text-search safe summaries
- delete an entry by tombstoning it

The UI does not expose automatic distillation, provider selection, OpenViking, vector search, or bulk import controls.

## Redaction

Memory creation and proposal creation run content through the same product data redaction helper used by events. Secret-looking input becomes redacted or blocked before it can be returned by HTTP, placed in a RunContext snapshot, or written to audit event metadata.
