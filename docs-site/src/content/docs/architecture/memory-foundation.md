---
title: M13 Memory Foundation Architecture
description: PG-backed memory entries, safe RunContext snapshots, approval-gated writes, and user-controlled deletion.
---

M13 adds the first local memory slice for Loomi. The implemented boundary is intentionally small: PostgreSQL-backed memory rows, text search, approval-gated memory writes, a safe memory snapshot in `RunContext`, and a minimal Settings UI for view/search/delete.

M13.5 closes the implementation with a real Postgres/httpapi smoke that applies migrations through the M13 tables and exercises the HTTP proposal/approval/list/search/delete path against the same repository used by `RunContext`.

M14 blocker foundation adds the minimum hardening needed before the full UX: thread-scoped read/delete authorization, durable memory audit rows, broader redaction, and one grounded search/list filter shape across docs, backend, and the frontend client.

M14 is the next UX/API contract slice. Its goal is not distillation or RAG; it makes Settings > Memory usable as a management surface and adds scoped, user-readable audit history backed by real memory events.

M42 adds the provider foundation slice. It makes the selected memory mode backend-owned, exposes safe provider readiness, and renders that readiness in Settings > Memory.

M43 adds the agent-facing memory tool slice. It does not add automatic distillation or semantic/vector retrieval; it routes bounded `memory.*` tools through the existing ToolCatalog, provider schema, ToolBroker, approval, worker continuation, and safe event/result boundaries.

M44 adds the post-run proposal slice. It connects the existing `commit_after_run` provider setting to runtime closeout and creates one pending, approval-gated write proposal after a successful completed run.

M45 adds the proposal review slice. It exposes a safe pending proposal list and renders approval/denial controls in Settings > Memory.

M46 adds proposal editing. It keeps the same `memory_write_proposals` table and lets users edit pending proposal title/summary before approval.

M47 expands provider configuration to match the Nowledge/OpenViking mechanism shape. It adds safe persistence and Settings controls for service URLs, key presence, embedding model, VLM model, and optional rerank model. This is still a configuration/status slice; external provider read/write adapters are not executed yet.

M48 expands the agent memory tool surface. It adds list, edit, context, timeline, connections, thread search, and thread fetch tools on top of the M43 tools. These tools are backed by local productdata and return only safe summaries/excerpts.

M49 adds Memory Snapshot and Memory Impression surfaces for Settings > Memory. The implementation deliberately reuses approved local memory search instead of adding a new graph/vector store.

M50 lets snapshot hits open a safe content view. It reuses memory entry authorization and returns only title plus safe summary.

M51 adds a user-authored manual memory path. It creates one approved local memory from Settings input and does not change the approval-gated agent write flow.

M52 adds a recent-errors read model derived from provider diagnostics. It is intentionally small until external provider adapters produce durable failure events.

M53 adds a local Nowledge detector. It is a UI convenience for localhost setup and does not change provider execution.

M54 makes memory tool exposure provider-aware. Nowledge keeps the semantic memory subset and disables `memory.edit`; disabled or unconfigured memory disables memory tools in the catalog and removes them from prepared run contexts.

M55 moves provider-specific Memory configuration into a modal so the main Settings > Memory surface stays focused on enablement, post-run organization, provider choice, status, and the Configure action.

M56 changes provider choice from a segmented control to selectable provider cards, matching the reference Memory service layout more closely while keeping Loomi's own copy and boundaries.

M57 adds ArkLoop-style Notebook tools beside semantic memory tools. Notebook entries use the same approved memory store and audit boundary, but are marked with `source_event_id=notebook` so they can be filtered as `source_type=notebook`.

M58 injects safe memory context into provider prompts. Semantic memory summaries are wrapped in `<memory>`, and structured notebook summaries are wrapped separately in `<notebook>`.

M59 keeps semantic overview/impression snapshots separate from notebook entries. Notebook rows remain available to `RunContext.NotebookSnapshot` and `source_type=notebook` searches, but do not appear in the Memory Snapshot or Memory Impression cards.

M60 adds read-side external adapters. When OpenViking or Nowledge is selected and configured, runtime memory search/read tools call the configured provider and return only redacted summaries.

M61 adds approval-gated external provider mutations. The tool still has to pass Loomi's tool approval boundary; after approval, OpenViking and Nowledge mutations execute against the selected configured provider.

M62 maps richer Nowledge read APIs into Loomi's expanded memory tools: graph connections, feed timeline, thread search, and thread fetch.

M63 aligns external `commit_after_run` with provider commit behavior. Local memory still creates a pending proposal, while OpenViking/Nowledge commit the post-run assistant outcome through the selected external provider and record a safe terminal-run commit event.

M64 adds an OpenViking localhost detector so Settings > Memory can probe both external memory service families from the same explicit user action pattern. The detector is status-only and does not read or return provider secrets.

M65 maps OpenViking directory listing into `memory.connections`. When OpenViking is active, `viking://...` IDs use `/api/v1/fs/ls` and return bounded child resource summaries instead of falling back to local text search.

M66 lets external providers contribute to the initial model prompt. Before the first provider request, Gateway uses the latest user message to query the active external memory provider and injects safe hits into the existing `<memory>` block.

M67 makes that external prompt recall observable by recording `memory_external_snapshot_loaded` with safe provider/count metadata after a successful external snapshot load.

M68 locks the Nowledge side of that path with provider-specific regression coverage. The implementation remains shared with OpenViking, but tests now prove Nowledge search routing, prompt injection, safe event metadata, and secret redaction.

M69 extends recent memory errors from static diagnostics to runtime provider failures. External prompt recall failures emit `memory_external_snapshot_failed`, keep the run alive, and appear in `/v1/memory/errors` as redacted provider error items.

M70 keeps the same Settings > Memory panel but renders optional runtime `event_type` and `run_id` details for recent provider errors. Long ids wrap inside the existing panel instead of resizing the layout.

## Boundary

The current implemented boundary is `productdata`: the service and repository own memory durability, search, safety filtering, audit persistence, and the safe summaries loaded into `RunContext.MemorySnapshot`.

`memory_provider_configs` stores the current provider preference for the local identity. The default is enabled local memory, which preserves the existing approved-entry store. M47 adds Nowledge and OpenViking config fields while keeping the legacy `semantic` placeholder readable for older local data. Provider readiness only reports configured/unconfigured/healthy/degraded state; external provider read/write execution remains a later slice.

M60 keeps provider secrets write-only at the HTTP boundary but adds an internal runtime-only config read method. This lets tool execution use provider credentials while API responses continue to expose only key-presence booleans.

`MemoryProvider` remains the future extraction point described by the M13 design contract. External semantic storage, embeddings, vector search, automatic distillation, marketplace/plugin memory providers, browser/activity recorder ingestion, and multi-agent long-term automation are outside this slice.

`MemoryToolExecutor` is the current runtime bridge for agent calls. It delegates to the productdata memory service and returns safe summaries only. `memory.write` creates a pending write proposal and `memory.forget` calls the tombstone API path; neither bypasses user approval or audit. `memory.edit` can edit a pending proposal or create a replacement proposal for an approved entry, but it never overwrites an approved memory directly.

Notebook tools are agent-facing structured note tools on the same safe boundary. `notebook.write` creates an approved scoped notebook entry with the `notebook` source marker. `notebook.edit` tombstones the old notebook entry and writes a replacement entry. `notebook.forget` tombstones a notebook entry. `notebook.read` returns the same safe summary projection as memory reads and rejects tombstoned or non-notebook entries.

Tool availability is computed from the selected memory provider. Local, semantic, and OpenViking keep the full Loomi memory tool set. Nowledge omits `memory.edit` because that provider shape exposes semantic recall and write surfaces without direct edit semantics. If memory is disabled or the selected provider is unconfigured, memory tools are disabled in Settings > Tools and omitted from prepared run contexts.

For external providers, `memory.search` and query-backed `memory.context` use the selected provider when configured. `memory.read` supports provider URIs returned by search: `viking://...` for OpenViking and `nowledge://memory/{id}` for Nowledge. Results are transformed into the same safe tool result shape as local memory and never include raw provider payloads or credentials.

Initial prompt enrichment reuses the same external read adapters. It copies the prepared run context, replaces only the safe memory snapshot for the provider request, and leaves the durable local prepared context unchanged. Provider failures or empty results fall back to the existing local snapshot without failing the run.

External prompt recall emits a progress event only after successful recall. The event stores provider, status, entry count, limit, and redaction flag. It does not store the recall query, raw hit text, credentials, provider traces, or local paths.

Nowledge and OpenViking use the same prompt-enrichment boundary. Provider-specific differences stay inside the external adapter: OpenViking calls `/api/v1/search/find`; Nowledge calls `/memories/search` with the latest user message and bounded limit.

If external prompt recall fails, Gateway records `memory_external_snapshot_failed` as a safe error event and returns the original prepared context. This makes Settings > Memory recent errors reflect real runtime failures without blocking the model request or leaking the user query/provider response.

External mutation routing is provider-specific. OpenViking supports `memory.write`, `memory.edit` for `viking://...` URIs, and `memory.forget` for `viking://...` URIs. Nowledge supports `memory.write` and `memory.forget` for `nowledge://memory/{id}`; `memory.edit` remains disabled for Nowledge. Local memory keeps the existing proposal-first write/edit flow and tombstone delete flow.

Nowledge-rich routing is also provider-specific. When Nowledge is active, `memory.connections` expands graph neighbors for `nowledge://memory/{id}`, `memory.timeline` reads feed events, and thread search/fetch use Nowledge's thread APIs. These results are projected into the same safe Loomi tool result shape with excerpts instead of raw provider payloads.

OpenViking connection routing is URI-specific. When OpenViking is active, `memory.connections` for `viking://...` calls `/api/v1/fs/ls?uri=...` and returns direct children as safe connection items. It does not recursively walk provider trees and does not expose raw OpenViking response bodies.

Post-run proposal creation lives in runtime closeout. The runtime helper checks productdata provider status after the runner or gateway writes the final event: memory must be enabled, configured, and `commit_after_run` must be true. Only completed runs with a persisted assistant message for that run are eligible.

For external providers, the same closeout hook calls the external write adapter instead of creating a local proposal. Success/failure is recorded with `memory_provider_commit_completed` or `memory_provider_commit_failed`, which are allowed after terminal run completion and carry only safe metadata.

Provider detectors are UI conveniences only. Nowledge detection probes the local health path; OpenViking detection probes the local filesystem listing path. Both use short localhost-only requests, return only detected/base URL/message/request id, and do not enable providers, discover keys, or validate remote credentials.

## Data Model

M13 adds:

- `memory_entries`: approved, tombstoned, or disabled user/thread-scoped memories.
- `memory_write_proposals`: pending/approved/denied write intents created by an agent or API caller.
- `memory_audit_events`: durable user-readable memory audit records for proposal, approval, denial, delete, and snapshot events.
- `memory_provider_configs`: user memory enablement, selected provider, commit-after-run preference, legacy semantic endpoint placeholder, OpenViking base/root/model settings, Nowledge base/timeout settings, write-only provider keys, safe diagnostic, and update timestamp.

Memory rows store a title, redacted summary, redacted content, scope, source thread/run/event ids, safety state, content hash, timestamps, and tombstone metadata. Delete is a tombstone operation: content is cleared, summary becomes `[deleted]`, and deleted metadata is retained for audit.

M57 reuses `memory_entries` for notebook rows rather than adding a new table. The durable distinction is `source_event_id='notebook'`; search/list projections expose this as `source_type=notebook`. Manual user-added memories exclude that marker and continue to show as `manual`.

## Safe Snapshot

When the worker prepares a `RunContext`, it searches up to five approved safe memories visible to the run's thread:

- user-scoped memories for the same local identity
- thread-scoped memories for the active thread
- approved entries only
- blocked entries and tombstones excluded

The snapshot records `loaded`, `empty`, or `unavailable`, and appends a `memory_snapshot_loaded` run event with only counts/status/limits/redaction flags.

Run preparation also resolves `RunContext.MemoryReadiness`. The readiness summary contains only enabled/provider/state/configured/diagnostic-code metadata, so later memory tools can check the same boundary without re-reading unsafe config. Memory disabled or provider-unavailable states do not block the run.

Thread search/fetch tools read only Loomi's local thread/message store and return bounded excerpts. They are not desktop activity capture and do not import browser or shell history.

RunContext also builds `NotebookSnapshot` from approved visible entries with `source_type=notebook`. The gateway appends safe summaries to the system prompt as:

```xml
<memory>
- Preference: Keep answers short.
</memory>

<notebook>
- Project notebook: Use structured notes for durable facts.
</notebook>
```

These blocks use title and safe summary only. Raw content, content hashes, provider traces, local paths, credentials, shell output, and tool output are never included.

## Overview Snapshot And Impression

`MemorySnapshotService` exposes four operations: get/rebuild overview snapshot and get/rebuild impression. The in-memory service and Postgres repository both build these outputs from bounded `SearchMemory` results:

- overview snapshot: `memory_block`, safe hit metadata, updated time, and rebuild flag
- impression: concise safe summary text, updated time, and rebuild flag

The current M49 builders are deterministic projections, not external distillation. Rebuild means "refresh the projection from current approved local memories"; it does not write approved memories, call an embedding model, or connect to Nowledge/OpenViking.

Notebook-backed entries are excluded from these semantic projections. The Notebook layer has its own run-context snapshot and prompt block, so structured notes do not get flattened into the ordinary memory impression.

Snapshot and impression outputs follow the same safety rules as RunContext memory snapshots: approved entries only, tombstones and unsafe entries excluded, no raw content, no proposal bodies, no tool output, no provider traces, no local paths, and no credentials.

Snapshot hit content uses `memory://{entry_id}` URIs. The content endpoint resolves the entry through `GetMemoryEntry`, so scope checks stay centralized. The returned text is a presentation projection of title and summary; raw stored memory content remains internal to productdata and is not sent to the frontend.

## Provider Status

Provider status states:

- `disabled`: memory is turned off for run readiness.
- `available`: local approved-memory store is available.
- `unconfigured`: selected non-local provider is missing required setup. OpenViking requires base URL, root key, embedding model, and VLM model; Nowledge requires base URL; legacy semantic requires its endpoint.
- `healthy`: selected provider setup is present for this configuration slice.
- `unhealthy`: reserved for deterministic provider health failures in the full semantic adapter slice.
- `degraded`: unknown stored provider normalized back to local memory.

Diagnostics are redacted in productdata before HTTP, frontend state, or run summaries can see them. API keys, Authorization headers, token-like values, provider traces, local paths, and secret-like markers must not appear in provider status responses. Provider key fields are write-only at the HTTP boundary; status responses expose only key-presence booleans such as `root_api_key_set` and `api_key_set`.

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

Post-run proposals are created after the run has reached `completed`, so their durable memory audit row is the source of truth. They are idempotent per run with `post_run_memory:{run_id}` and remain `pending` until the user approves or denies them.

Pending proposal review reads from `memory_write_proposals` through a safe projection: title, summary, scope, status, safety state, source ids, and timestamps only. The review surface must not expose proposal content, idempotency keys, or hidden user fields.

Editing is allowed only while a proposal is pending. The update operation accepts title and summary, runs them through the same memory content normalization/redaction boundary, and writes the edited summary back as the proposal content. Approval therefore creates the memory entry from the user-confirmed text instead of the original generated proposal body.

## User Control

The first UI boundary is Settings > Memory:

- list approved memories
- text-search safe summaries
- delete an entry by tombstoning it

The UI does not expose automatic distillation, vector search, or bulk import controls.

M14 expands this surface to list/search/filter, detail drawer or modal, explicit delete confirmation, loading/empty/error/tombstoned states, and a real audit/history panel. The audit surface is backed by productdata memory events and must not fabricate UI-only history. The default UI keeps engineering filters collapsed and folds routine `memory_snapshot_loaded` events into a small system snapshot note, while write/delete/proposal events remain visible as human-readable history.

M42 adds the Settings > Memory service panel above the management list. It reads `/v1/memory/provider`, can refresh status, and updates enabled/provider/commit-after-run through the backend. The panel uses Loomi copy and provider labels; it must not copy external product names or private provider wording.

M43 makes Settings > Tools list the memory tools as `memory` scope, medium risk, approval-gated, and safe-summary-only. The Memory page itself remains the management surface for provider status, entries, search, delete, and history.

M44 makes the Settings > Memory "Organize after each run" toggle active for local runtime closeout. The UI text describes the result as an approval-gated proposal so users do not confuse it with direct long-term memory insertion.

M45 shows pending proposals above saved memories. Saving a proposal calls the existing approve endpoint and refreshes proposals, saved entries, and audit history. Rejecting a proposal calls the existing deny endpoint and refreshes proposals and audit history.

M46 adds inline edit controls to each pending proposal card. Users can adjust the title and summary, save those edits without approving, then save or reject the proposal. The UI keeps raw proposal content and idempotency keys out of state.

M49 adds snapshot and impression cards above provider configuration. Each card shows a safe preview, last update time, rebuild action, and basic hit metadata. The cards are read-only projections; memory mutation still flows through proposal approval or explicit delete.

M50 turns snapshot hit chips into view actions. The modal is read-only and closes locally; it does not edit approved entries or bypass proposal review.

M51 adds a compact manual-add form to Settings > Memory. This is treated as direct user control, so it may create an approved memory entry immediately. Agent-generated and post-run writes still go through pending proposals.

M52 shows current provider diagnostic errors in the provider panel. M69 also includes recent runtime provider failures from safe run events. The list is redacted and does not include upstream logs, provider request payloads, prompt query text, memory bodies, or secrets.

The Settings renderer treats runtime ids as opaque diagnostic strings. It does not create executable controls or links from provider error items.

M53 adds a detect action in the Nowledge config section. A successful localhost health check fills the base URL through the existing provider update path; API keys remain user-entered and write-only.

M54 reflects provider support in Settings > Tools. Unsupported memory tools remain visible as disabled catalog entries with a safe reason code, while runnable contexts only receive supported memory tools.

M55 keeps Nowledge/OpenViking details behind Configure. The modal reuses the existing update endpoint and write-only key fields; it does not add install, restart, process-control, or bulk-delete actions.

M56 keeps provider card selection as a pure UI affordance over the same update endpoint. It does not change provider persistence or add external adapter execution.

M57 makes notebook tools visible in Settings > Tools as approval-gated memory-scope tools. It does not add a separate Notebook page, export/import controls, or destructive bulk actions.

## M14 Management And Audit Flow

Settings > Memory reads safe projections from the memory API. List/search share one filter shape: `q` or JSON `query`, `scope_type`, `scope_id`, `source_thread_id`, `source_run_id`, `source_type`, `include_tombstoned`, and `limit`. `workspace_id` is deferred until workspace-scoped memory exists. Detail reads return only safe metadata: summary/title, scope, source thread/run/event ids, source type, status, timestamps, and redaction state.

Detail/delete authorization follows the memory scope. User-scoped memory is visible to the same user. Thread-scoped memory requires a matching `scope_type=thread&scope_id`, `source_thread_id`, or `source_run_id`; wrong or missing context returns generic not found.

Audit history reads `memory_write_proposed`, `memory_write_approved`, `memory_write_denied`, `memory_deleted`, and `memory_snapshot_loaded` from scoped `memory_audit_events`. Memory audit survives terminal source runs because users still need history after the run completes.

## Redaction

Memory creation and proposal creation run content through the same product data redaction helper used by events. Secret-looking input becomes redacted or blocked before it can be returned by HTTP, placed in a RunContext snapshot, or written to audit event metadata.

M14 expands the redaction requirement to common local and provider-output forms: `/Users/...`, `/home/...`, Windows paths, stdout/stderr dumps, tool output, provider traces, `.env` values, Authorization headers, tokens, credentials, key/env markers, and secret-like values.

The post-run proposal body is bounded before entering productdata and then passes through the same proposal redaction and safety classification as manual or agent-created proposals.

## M13.5 Evidence

`TestM13MemoryRealPGHTTPAPISmoke` is the current end-to-end backend evidence. It verifies migrated `memory_entries` and `memory_write_proposals`, approval-gated creation, search/list visibility, RunContext safe snapshot load, tombstone exclusion, duplicate approve/deny/delete idempotency, out-of-scope non-leakage, and sensitive redaction across API responses and run-event metadata.
