---
title: M13 Memory 本地 Smoke
description: M13/M13.5 memory migration、real PG/httpapi smoke、Settings > Memory 浏览器 smoke 和验证命令。
---

## 环境变量

```bash
APP_ENV=local
HTTP_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
LOOMI_TEST_DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080
```

## 启动 Postgres 并应用 migration

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

M14 blocker foundation 需要 version `10` 且 clean。API 不会在启动时自动执行 migration。

## Real PG/httpapi smoke

```bash
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/httpapi -run TestM13MemoryRealPGHTTPAPISmoke -count=1 -v
```

这个 smoke 使用真实 Postgres repository 和 HTTP handlers，覆盖：

- create/propose memory write
- approve proposal
- approved memory list/search
- RunContext safe memory snapshot
- delete 后 tombstone 并立即从 list/search/RunContext 排除
- duplicate approve/deny/delete 不重复 entry/audit
- out-of-scope 不泄露存在性
- terminal run 后 proposal/approve/deny/delete audit 仍可从 `/v1/memory/audit` 查询
- sensitive content 不进入 API response、RunContext safe summary、memory audit metadata

## M14 blocker foundation checks

```bash
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/productdata -run TestPostgresMemoryEntryScopeAndTerminalAudit -count=1 -v
```

这个 smoke 使用真实 Postgres repository，覆盖同一用户下 thread A 不能 read/delete thread B memory、thread A scope 可 read/delete、terminal run 后 approve/deny/delete audit 仍存在，以及重复 deny/delete 不重复 audit。

## Settings > Memory 浏览器 smoke

M14 full done gate uses a seeded memory entry and real API mode. It must cover list/search/filter, detail drawer or modal, delete confirmation, error/empty/loading/tombstoned states, and real audit history. Do not mock audit history in the UI.

启动 API：

```bash
APP_ENV=local HTTP_ADDR=127.0.0.1:8080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
```

启动真实 API 模式 web：

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev --host 127.0.0.1
```

浏览器检查：

- 打开 web shell。
- 进入 Settings > Memory。
- 验证 Memory category 可打开。
- 验证 list/search/delete 状态连接真实 API，删除后条目消失。
- M14 full smoke additionally verifies detail, filters, delete confirmation, and audit history from `/v1/memory/audit`.
- DevTools console 无红色错误。

如果本地端口被占用导致无法启动其中一个服务，记录端口/进程原因，并用 real PG/httpapi smoke、`bun test --cwd web`、`bun run --cwd web build` 作为等价后端与前端边界证据。

## Validation commands

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Deferred

M13/M13.5/M14 不包含 vector DB、embedding、RAG、OpenViking、自动 distill、activity recorder、sandbox、MCP rewrite、worker/job queue rewrite 或多 agent 自动记忆。
