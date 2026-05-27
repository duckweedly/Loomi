# Contract: Memory Provider

Status: current implemented contract for `019-memory-foundation`.

## Scope

The MemoryProvider boundary lets RunContext, API, and runtime code depend on safe memory operations without depending on PG query details. V1 currently implements only the product data/Postgres provider. OpenViking and other providers remain deferred.

## Interface expectations

```text
SearchMemory(request) -> MemorySearchResult[]
BuildSnapshot(run_context_seed) -> MemorySnapshot
CreateWriteProposal(proposal) -> MemoryWriteProposal
ApproveWrite(decision) -> MemoryEntry
DenyWrite(decision) -> MemoryWriteProposal
DeleteMemory(entry_id, actor) -> MemoryTombstone
```

## Provider invariants

- Provider outputs are already scoped, authorized, and safe for the caller.
- Provider search excludes pending, denied, tombstoned, disabled, unsafe, and out-of-scope entries.
- Provider delete creates a tombstone or returns an existing tombstone idempotently.
- Provider write approval creates or links one Memory Entry exactly once.
- Provider errors return safe error codes without raw SQL, raw content, credentials, paths, or provider internals.

## PG provider v1

PG provider owns:

- `memory_entries` persistence.
- `memory_write_proposals` persistence.
- Scope/status/text/metadata filtering.
- Tombstone update.
- Safe audit event persistence hooks.

PG provider does not own:

- Embedding generation.
- Vector indexes.
- RAG orchestration.
- Distillation jobs.
- External provider sync.
- OpenViking compatibility.

## Future provider constraints

Any future provider must preserve the same safety contract:

- user deletion/control remains authoritative;
- deleted entries cannot reappear through sync;
- redaction happens before RunContext/API/UI/event exposure;
- audit events use safe metadata;
- provider text is untrusted data, not instructions.
