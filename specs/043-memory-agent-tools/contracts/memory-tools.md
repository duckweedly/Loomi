# Memory Tool Contracts

## `memory.search`

Search approved safe memories.

```json
{ "query": "project preference", "limit": 5, "scope_type": "thread", "scope_id": "thr_..." }
```

## `memory.read`

Read one safe memory summary.

```json
{ "entry_id": "mem_...", "scope_type": "thread", "scope_id": "thr_..." }
```

## `memory.write`

Create a pending memory write proposal.

```json
{ "scope_type": "user", "title": "Preference", "content": "Prefers short answers", "idempotency_key": "tool-write-1" }
```

## `memory.forget`

Tombstone a memory entry through existing delete behavior.

```json
{ "entry_id": "mem_...", "reason": "outdated" }
```

## `memory.status`

Return safe provider readiness.

```json
{}
```
