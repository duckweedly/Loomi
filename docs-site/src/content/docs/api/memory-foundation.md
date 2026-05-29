---
title: M13 Memory Foundation API and Events
description: Memory entry, search, write proposal, approval, delete, and durable audit contracts.
---

M13 exposes the first memory API under `/v1/memory`. All responses use safe summaries and omit raw memory content from list/search surfaces. M14 prepares the management/audit contract and unifies list/search/audit filters so the Settings > Memory UX can be implemented without fake history.

M14 blocker foundation adds durable `memory_audit_events` and scoped read/delete authorization for thread memories. Local smokes must apply migrations through version `10`; the API process still does not auto-apply migrations.

M42 adds `/v1/memory/provider` for backend-owned provider readiness. Local smokes must apply migrations through version `14` for persistent provider config.

M43 adds agent-facing `memory.*` tools through the runtime tool contract. It does not add new HTTP endpoints; the tools reuse the existing productdata memory service and approval/audit boundaries.

M44 wires the existing `commit_after_run` provider preference into runtime closeout. When enabled and configured, a completed run creates one idempotent pending memory write proposal from the run's assistant outcome. It does not approve the proposal or make it searchable.

M45 adds pending proposal review. `GET /v1/memory/write-proposals` returns safe pending proposal summaries for Settings > Memory; approval and denial continue through the existing decision endpoints.

M46 adds proposal editing before approval. `PATCH /v1/memory/write-proposals/{proposal_id}` updates only a pending proposal's title and safe summary; approval persists the edited text.

M47 expands `/v1/memory/provider` to carry the safe Nowledge and OpenViking configuration shape. It persists model/provider settings and key presence while continuing to omit raw secrets from all responses.

M48 expands the agent-facing memory tool set. New tools are still backed by local productdata and safe summaries; they do not execute external OpenViking/Nowledge adapters.

M49 adds snapshot and impression endpoints backed by local approved memories. These endpoints return bounded safe summaries and hit metadata only; they do not expose raw memory content or execute external provider adapters.

M50 adds a safe content view endpoint for `memory://` snapshot hits. It returns title and safe summary only, not raw stored memory content.

M51 adds a manual Notebook-style write path: the user can explicitly save one local memory from Settings > Memory.

M52 adds a safe recent-errors endpoint derived from memory provider diagnostics.

M53 adds a localhost Nowledge detector for Settings > Memory.

M54 adds provider-aware memory tool availability. This changes ToolCatalog and RunContext projections only; it does not add a new HTTP endpoint.

M57 adds runtime `notebook.*` tools. It does not add a new HTTP endpoint; notebook entries reuse the approved memory entry store with `source_type=notebook` projections.

M58 adds no HTTP endpoints. It changes run-context/provider prompt behavior by adding safe `<memory>` and `<notebook>` blocks from approved summaries.

M59 changes semantic snapshot projection only: `/v1/memory/snapshot` and `/v1/memory/impression` exclude notebook entries, while `source_type=notebook` remains available through memory list/search and run-context notebook injection.

M60 changes runtime tool execution only. Configured OpenViking and Nowledge providers can back `memory.search`, query-backed `memory.context`, and provider-URI `memory.read`. No raw provider config is exposed through HTTP.

M61 extends runtime tool execution for approved mutations. Configured OpenViking can back `memory.write`, provider-URI `memory.edit`, and provider-URI `memory.forget`; configured Nowledge can back `memory.write` and provider-URI `memory.forget`.

M62 extends Nowledge runtime reads: configured Nowledge can back `memory.connections`, `memory.timeline`, `memory.thread_search`, and `memory.thread_fetch`.

M63 changes `commit_after_run` behavior for external providers. Local memory still creates a pending write proposal; OpenViking and Nowledge use the provider write adapter and record a safe run event. The external closeout path checks the commit-completed marker by event type instead of replaying the full run event stream.

M65 extends OpenViking runtime reads: configured OpenViking can back `memory.connections` for `viking://...` IDs by reading direct child resources from `/api/v1/fs/ls`.

M66 changes runtime prompt assembly only. The Gateway can query the active external memory provider with the latest user message and inject safe hits into the initial `<memory>` block. No HTTP response shape changes.

M67 adds one run event type: `memory_external_snapshot_loaded`. It records successful external provider prompt recall with safe metadata only.

M68 adds Nowledge-specific regression coverage for the same external prompt recall path. It confirms Nowledge recall uses the latest user message, injects safe hits into `<memory>`, and records `memory_external_snapshot_loaded` without leaking query text, raw content, or credentials.

M69 extends `/v1/memory/errors` to include safe runtime provider failures. External prompt recall failures now write `memory_external_snapshot_failed` and remain non-fatal for the run.

M70 updates Settings > Memory rendering only. Recent errors now display optional safe `event_type` and `run_id` details when the API returns runtime provider failures.

## Entry APIs

| Method | Path | Meaning |
| --- | --- | --- |
| `GET` | `/v1/memory` | List approved, safe memory summaries. |
| `POST` | `/v1/memory` | Search approved, safe memory summaries by text. |
| `POST` | `/v1/memory/search` | Search approved, safe memory summaries by text. |
| `GET` | `/v1/memory/entries/{entry_id}` | Read one approved or tombstoned safe memory detail. |
| `POST` | `/v1/memory/entries` | Create one user-authored approved memory entry. |
| `DELETE` | `/v1/memory/entries/{entry_id}` | Tombstone a memory entry and remove it from future search/snapshots. |
| `GET` | `/v1/memory/audit` | List scoped safe memory audit/history items. |
| `GET` | `/v1/memory/errors` | List current safe memory provider diagnostic errors. |

## Snapshot And Impression APIs

| Method | Path | Meaning |
| --- | --- | --- |
| `GET` | `/v1/memory/snapshot` | Return the current safe overview snapshot built from approved memories. |
| `POST` | `/v1/memory/snapshot/rebuild` | Rebuild and return the safe overview snapshot. |
| `GET` | `/v1/memory/impression` | Return the current safe memory impression. |
| `POST` | `/v1/memory/impression/rebuild` | Rebuild and return the safe memory impression. |
| `GET` | `/v1/memory/content?uri=memory://{entry_id}&layer=overview` | Return safe content for a snapshot hit. |

Snapshot response:

```json
{
  "snapshot": {
    "memory_block": "- Preference: Prefers compact implementation slices",
    "hits": [{
      "uri": "memory://mem_abc",
      "entry_id": "mem_abc",
      "title": "Preference",
      "abstract": "Prefers compact implementation slices",
      "is_leaf": true,
      "updated_at": "2026-05-27T12:00:00Z"
    }],
    "updated_at": "2026-05-27T12:00:01Z",
    "rebuilt": true
  },
  "request_id": "req_..."
}
```

Impression response:

```json
{
  "impression": {
    "impression": "- Prefers compact implementation slices",
    "updated_at": "2026-05-27T12:00:02Z",
    "rebuilt": true
  },
  "request_id": "req_..."
}
```

Empty states return explicit safe text and `hits: []`. The `rebuilt` flag is true only for rebuild endpoints. Responses must not include raw `content`, proposal bodies, idempotency keys, provider traces, tool output, local paths, credentials, or secret-like values.

Content response:

```json
{
  "content": "Preference\n\nPrefers compact implementation slices",
  "layer": "read",
  "uri": "memory://mem_abc",
  "request_id": "req_..."
}
```

`layer=overview` returns the safe summary. `layer=read` prefixes the safe title when available. Unsupported URI schemes and missing or invalid layers return `invalid_request`. The endpoint uses the same memory entry authorization boundary as detail/delete, and still omits raw stored content.

## Provider API

| Method | Path | Meaning |
| --- | --- | --- |
| `GET` | `/v1/memory/provider` | Read the current safe memory provider status. |
| `PUT` | `/v1/memory/provider` | Update memory enablement, selected provider, and commit-after-run preference. |
| `GET` | `/v1/memory/provider/nowledge/detect` | Probe the default local Nowledge health endpoint and return a safe detection result. |
| `GET` | `/v1/memory/provider/openviking/detect` | Probe the default local OpenViking endpoint and return a safe detection result. |

Update request:

```json
{
  "enabled": true,
  "provider": "openviking",
  "commit_after_run": true,
  "openviking": {
    "base_url": "http://127.0.0.1:8282",
    "root_api_key": "stored-only",
    "embedding_selector": "default",
    "embedding_provider": "openai_compatible",
    "embedding_model": "text-embedding-3-large",
    "embedding_api_key": "stored-only",
    "embedding_api_base": "https://api.openai.com/v1",
    "embedding_dimension": 3072,
    "vlm_selector": "default",
    "vlm_provider": "openai_compatible",
    "vlm_model": "gpt-5.5",
    "vlm_api_key": "stored-only",
    "vlm_api_base": "https://api.openai.com/v1",
    "rerank_selector": "optional",
    "rerank_provider": "openai_compatible",
    "rerank_model": "",
    "rerank_api_key": "",
    "rerank_api_base": ""
  }
}
```

Nowledge uses the same endpoint:

```json
{
  "enabled": true,
  "provider": "nowledge",
  "commit_after_run": true,
  "nowledge": {
    "base_url": "http://127.0.0.1:7727",
    "api_key": "stored-only",
    "request_timeout_ms": 30000
  }
}
```

Status response:

```json
{
  "status": {
    "enabled": true,
    "provider": "openviking",
    "label": "OpenViking",
    "state": "healthy",
    "configured": true,
    "commit_after_run": true,
    "checked_at": "2026-05-26T12:00:00Z",
    "openviking": {
      "base_url": "http://127.0.0.1:8282",
      "root_api_key_set": true,
      "embedding_model": "text-embedding-3-large",
      "embedding_api_key_set": true,
      "embedding_dimension": 3072,
      "vlm_model": "gpt-5.5",
      "vlm_api_key_set": true
    },
    "diagnostic": {
      "code": "ok",
      "message": "Ready."
    }
  },
  "request_id": "req_..."
}
```

Known states are `disabled`, `available`, `unconfigured`, `healthy`, `unhealthy`, and `degraded`. Unknown stored providers return a safe local-provider fallback with `state=degraded`.

Provider responses must not include API keys, Authorization headers, endpoint credentials, raw provider traces, local file paths, tokens, or secret-like values. `semantic_endpoint` is accepted as legacy input for readiness projection but is not returned. OpenViking and Nowledge key fields are write-only; responses return only `*_key_set` booleans.

`commit_after_run=true` is a proposal toggle. It means runtime closeout may create a pending write proposal after a successful completed run when memory is enabled and configured. It never creates an approved `memory_entries` row by itself.

For external providers, `commit_after_run=true` means runtime closeout sends a bounded assistant outcome to the selected provider after run completion. The event metadata contains provider, operation, opaque entry id when available, status, and redaction flags only.

Nowledge detection response:

```json
{
  "detected": true,
  "base_url": "http://127.0.0.1:14242",
  "message": "Nowledge local instance detected.",
  "request_id": "req_..."
}
```

The detector only checks localhost with a short timeout. It does not discover API keys, scan remote hosts, or return secrets.

OpenViking detection response:

```json
{
  "detected": true,
  "base_url": "http://127.0.0.1:8282",
  "message": "OpenViking local instance detected.",
  "request_id": "req_..."
}
```

The detector only checks the default localhost OpenViking API with a short timeout. It does not send configured API keys, scan remote hosts, enable OpenViking, or return provider secrets.

## Runtime Memory Tools

| Tool | Required args | Effect |
| --- | --- | --- |
| `memory.search` | `query` | Search approved safe memory summaries. |
| `memory.list` | none | List approved safe memory summaries in the current scope. |
| `memory.read` | `entry_id` | Return one safe memory detail without raw content. |
| `memory.write` | `title`, `content` | Create a pending memory write proposal. |
| `memory.edit` | `title`, `content`, plus `proposal_id` or `entry_id` | Edit a pending proposal or create an approval-gated replacement proposal. |
| `memory.forget` | `entry_id` | Tombstone one visible memory entry and write audit. |
| `memory.context` | none | Return provider readiness plus bounded relevant memory summaries. |
| `memory.timeline` | none | Return safe memory audit timeline items. |
| `memory.connections` | `entry_id` or `query` | Return bounded related memory summaries. |
| `memory.thread_search` | `query` | Search local thread/message history with safe excerpts. |
| `memory.thread_fetch` | `thread_id` | Fetch safe excerpts from one local thread. |
| `memory.status` | none | Return safe memory provider readiness. |
| `notebook.read` | `entry_id` | Read one approved structured notebook entry without raw content. |
| `notebook.write` | `title`, `content` | Write one scoped structured notebook entry through the audited memory boundary. |
| `notebook.edit` | `entry_id`, `title`, `content` | Replace a notebook entry by tombstoning the old row and writing a new one. |
| `notebook.forget` | `entry_id` | Tombstone one structured notebook entry. |

All memory tools are Work-mode only, always approval-required, and exposed to providers with bounded schemas. Optional args follow the existing memory filter shape: `scope_type`, `scope_id`, `source_thread_id`, `source_run_id`, `source_type`, `source_event_id`, `idempotency_key`, `reason`, and `limit` where relevant.

Notebook tools are part of the same memory tool group and approval boundary. `source_type=notebook` filters notebook-backed entries; `source_type=manual` excludes them.

Tool result summaries are deliberately smaller than HTTP detail responses. They may include ids, title, summary, scope, status, safety state, counts, provider state, and redaction flags. They must not include raw memory `content`, `content_hash`, provider traces, local paths, Authorization headers, tokens, credentials, `.env` values, or secret-like fields.

External provider prompt hits follow the same safety rule. They are safe title/summary entries only and are not persisted as local approved memories.

External prompt recall event metadata:

```json
{
  "provider": "openviking",
  "status": "loaded_external",
  "entry_count": 1,
  "limit": 5,
  "redaction_applied": true
}
```

External provider tool results use provider URIs as `entry_id` values. OpenViking reads use `viking://...`; Nowledge reads use `nowledge://memory/{id}`. Callers should treat these ids as opaque and pass them back to `memory.read` unchanged.

External provider mutations return only accepted/deleted status, provider id, opaque URI when available, title/summary-safe metadata, and redaction flags. They do not return raw write content or provider response bodies.

OpenViking connections return direct child resource ids, safe titles, node type, relation, provider id, and redaction flags. They omit raw provider response bodies and do not recursively traverse provider directories.

Nowledge graph/timeline/thread results return safe ids, titles, excerpts, counts, relation labels, event labels, and timestamps. They omit raw provider response bodies and full message content.

Search request:

```json
{
  "query": "postgres memory",
  "scope_type": "thread",
  "scope_id": "thr_123",
  "source_thread_id": "thr_123",
  "source_run_id": "run_123",
  "source_type": "run",
  "include_tombstoned": false,
  "limit": 5
}
```

Search/list item:

```json
{
  "id": "mem_abc",
  "title": "Preference",
  "summary": "Prefers compact implementation slices",
  "scope_type": "user",
  "scope_id": "usr_local_dev",
  "status": "approved",
  "safety_state": "safe",
  "source_thread_id": "thr_123",
  "source_run_id": "run_123",
  "source_type": "run",
  "rank_reason": "text_match",
  "redaction_applied": true,
  "updated_at": "2026-05-25T00:00:00Z"
}
```

Empty list/search responses return `items: []`, not `null`, so the Settings > Memory surface can render empty state without special casing.

List/search requests with `scope_type=thread` require a non-empty `scope_id`. Missing `scope_id` returns `invalid_request` so callers do not confuse a malformed thread-scoped request with a genuine empty memory list.

Detail and delete for thread-scoped entries require a matching scope boundary. A request may use `scope_type=thread&scope_id={thread_id}`, `source_thread_id`, or `source_run_id`. User-scoped entries remain visible to the current user. Out-of-scope detail/delete returns generic `memory_not_found` and never echoes the target entry id.

Detail responses use the same safe projection and must not include raw `content`, `content_hash`, provider trace, tool output, local path, `.env`, `Authorization`, credential, token, or secret-like fields.

Manual create request:

```json
{
  "scope_type": "user",
  "title": "Preference",
  "content": "Prefers compact implementation slices"
}
```

Manual create is for user-authored Settings input and returns the same safe entry projection as detail/list. Agent/runtime-created writes still use `memory_write_proposals` and require approval before becoming searchable.

Errors response:

```json
{
  "errors": [{
    "code": "nowledge_unconfigured",
    "message": "Provider is not configured.",
    "provider": "nowledge",
    "state": "unconfigured",
    "checked_at": "2026-05-27T12:00:00Z",
    "run_id": "run_abc",
    "event_type": "memory_external_snapshot_failed"
  }],
  "request_id": "req_..."
}
```

The current implementation reports provider diagnostic state plus recent runtime provider failures. It does not include provider logs, request payloads, API keys, Authorization headers, tokens, prompt query text, memory bodies, local paths, or raw upstream traces.

## Write Proposal APIs

| Method | Path | Meaning |
| --- | --- | --- |
| `POST` | `/v1/memory/write-proposals` | Propose a memory write without making it searchable. |
| `GET` | `/v1/memory/write-proposals` | List safe memory write proposals, defaulting to pending proposals. |
| `PATCH` | `/v1/memory/write-proposals/{proposal_id}` | Edit a pending proposal's title and safe summary. |
| `POST` | `/v1/memory/write-proposals/{proposal_id}/approve` | Approve a pending proposal and create an entry. |
| `POST` | `/v1/memory/write-proposals/{proposal_id}/deny` | Deny a pending proposal without creating an entry. |

Proposal request:

```json
{
  "scope_type": "user",
  "title": "Preference",
  "content": "Prefers PostgreSQL-first memory",
  "source_thread_id": "thr_123",
  "source_run_id": "run_123",
  "idempotency_key": "agent-write-1"
}
```

Approve/deny request:

```json
{
  "reason": "user approved",
  "idempotency_key": "decision-1"
}
```

Pending and denied proposals do not appear in memory search. Approval responses include the proposal and the created safe entry summary.

Edit request:

```json
{
  "title": "Preference",
  "summary": "Prefers PostgreSQL-first memory"
}
```

Only pending proposals can be edited. The edit path accepts title and summary only, normalizes them through the memory redaction boundary, stores the edited summary as the proposal content used on approval, and returns the same safe proposal projection as the list API.

Proposal list responses include ids, title, summary, scope, status, safety state, source thread/run/event ids, decision metadata, and timestamps. They must omit raw `content`, idempotency keys, hidden user ids, and credentials.

The implemented public create path for agent memory is the write-proposal API. Direct approved-entry creation remains a service/repository boundary, not a public HTTP create endpoint.

Post-run proposals use the same write-proposal contract. They are thread-scoped, source-run-linked, and idempotent with a `post_run_memory:{run_id}` key. Re-running runtime closeout for the same completed run returns the same proposal rather than creating duplicates.

## Events

Memory audit is durably stored in `memory_audit_events`. When the source run exists, Loomi writes the audit row and related `run_events` timeline row in one transaction using the run-event sequence guard; failures are returned instead of silently dropping the timeline mirror. Run events are not the only audit store, so memory audit remains available after the source run has completed or failed.

| Event type | Meaning |
| --- | --- |
| `memory_snapshot_loaded` | Worker loaded a safe memory snapshot into `RunContext`. |
| `memory_write_proposed` | A memory write proposal was created. |
| `memory_write_approved` | A proposal was approved and an entry was created. |
| `memory_write_denied` | A proposal was denied. |
| `memory_entry_deleted` | A memory entry was tombstoned. |

Event metadata is restricted to ids, counts, scope, status, safety state, limits, and redaction flags. It must not include raw memory content or external tool/provider payloads.

For user-facing audit history, `memory_entry_deleted` is projected as `memory_deleted`.

Audit query filters accept the same thread/source shape used by list/search: `scope_type=thread&scope_id={thread_id}`, `source_thread_id={thread_id}`, `source_run_id={run_id}`, and `limit`. `thread_id` remains accepted as a direct audit filter. Thread-scoped audit requests must return only matching thread history and must not mix in same-user history from other threads.

Audit item:

```json
{
  "id": "evt_123",
  "event_type": "memory_write_approved",
  "summary": "Memory write approved",
  "thread_id": "thr_123",
  "run_id": "run_123",
  "memory_entry_id": "mem_123",
  "memory_proposal_id": "memprop_123",
  "status": "approved",
  "scope_type": "thread",
  "source_type": "run",
  "redaction_applied": true,
  "occurred_at": "2026-05-25T00:00:00Z"
}
```

Out-of-scope entry/proposal/run/thread ids return generic not found or empty scoped results. They must not reveal whether another user's memory exists.

## Current Filter Shape

Implemented now:

- `query` in JSON search requests and `q` in list query strings.
- `limit`.
- `scope_type` and `scope_id`.
- `source_thread_id`; for audit this is applied as the thread history boundary.
- `source_run_id`.
- `source_type` with `run`, `thread`, `manual`, or `any`.
- `include_tombstoned`.

Deferred:

- `workspace_id` and workspace-scoped memory. Docs and frontend must not send `workspace_id` until a workspace memory scope exists.
- Cursor pagination. Current list/search/audit responses are bounded by `limit`.
