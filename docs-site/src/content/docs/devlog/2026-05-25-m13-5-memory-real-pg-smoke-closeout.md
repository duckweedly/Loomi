---
title: 2026-05-25 M13.5 Memory Real PG Smoke Closeout
description: Real Postgres and HTTP API closeout evidence for M13 Memory Foundation.
---

## Completed

- Added `020-memory-real-pg-smoke-closeout` as a closeout/evidence Spec Kit slice.
- Added `internal/httpapi/memory_real_pg_smoke_test.go`.
- Proved the real Postgres repository plus HTTP handlers use migrated `memory_entries` and `memory_write_proposals`.
- Closed `specs/019-memory-foundation/spec.md` from Draft to Implemented.
- Updated M13 contracts and docs so current PG/API/RunContext behavior is labeled implemented, while distill/OpenViking/vector/RAG-style work remains deferred.
- Fixed the real PG API empty-list shape so memory list/search responses return `items: []` instead of `items: null`.

## Real PG/httpapi smoke coverage

`TestM13MemoryRealPGHTTPAPISmoke` covers:

- propose memory write through `POST /v1/memory/write-proposals`
- approve proposal through `POST /v1/memory/write-proposals/{proposal_id}/approve`
- approved memory visible through `GET /v1/memory` and `POST /v1/memory/search`
- `RunContext.MemorySnapshot` loads the approved safe memory from real PG
- delete through `DELETE /v1/memory/{entry_id}` tombstones the row
- list/search/RunContext immediately exclude the tombstoned row
- duplicate approve, deny, and delete do not duplicate entries or audit events
- out-of-scope delete returns not found without exposing the other entry id
- sensitive content is excluded from API responses, RunContext safe summary, and memory run-event metadata

## Settings > Memory browser smoke

Browser smoke was run with local API on `127.0.0.1:18080` and web on `127.0.0.1:5180`, because `5180` is in the local API CORS allowlist.

```bash
APP_ENV=local HTTP_ADDR=127.0.0.1:18080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080 bun run --cwd web dev --host 127.0.0.1 --port 5180
```

Result: Settings opened, Memory category was available, and the Memory panel rendered the real API empty state (`No memory entries`) with a search input. An initial attempt on `5182` hit the existing local CORS allowlist and exposed the PG `items:null` empty-list bug; after fixing the API response shape and rerunning on `5180`, the panel rendered successfully.

If the local API or web server cannot start because a port is occupied, the equivalent backend evidence is the real PG/httpapi smoke above plus `bun test --cwd web` and `bun run --cwd web build`.

## Boundaries

M13.5 does not add vector DB, embeddings, RAG orchestration, OpenViking, marketplace/plugin memory providers, browser/activity recorder ingestion, automatic memory distillation, multi-agent long-term memory automation, worker/job queue rewrite, sandbox, or MCP rewrite.
