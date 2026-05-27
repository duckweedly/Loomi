# Quickstart: M13.5 Real PG Memory Smoke Closeout

## 1. Start Postgres and apply migrations

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

Expected: version `9` and clean.

## 2. Run the real PG/httpapi smoke

```bash
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/httpapi -run TestM13MemoryRealPGHTTPAPISmoke -count=1 -v
```

Expected: the smoke covers proposal, approval, list/search, RunContext snapshot, delete/tombstone exclusion, duplicate approve/deny/delete idempotency, out-of-scope non-leakage, and sensitive redaction.

## 3. Settings > Memory browser smoke

Start API and web in real API mode:

```bash
APP_ENV=local HTTP_ADDR=127.0.0.1:8080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev --host 127.0.0.1
```

Open the web shell, go to Settings > Memory, and verify the Memory surface loads without console errors and can render list/search/delete state against the API.

## 4. Required closeout validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Deferred

This closeout does not add vector DB, embedding, RAG, OpenViking, automatic distill, activity recorder, sandbox, MCP rewrite, worker rewrite, or multi-agent automatic memory.
