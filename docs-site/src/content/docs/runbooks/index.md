---
title: Runbooks 入口
description: Loomi 本地开发、验证和排错手册入口。
---

这里记录可执行操作：如何启动服务、如何跑测试、如何执行 migration、如何排查事件流、如何清理本地状态、如何验证桌面壳和后端联通。

M13 memory 的真实 Postgres/API smoke 见 [M13 Memory 本地 Smoke](./local-m13-memory/)。M15 Chat integrated closeout smoke 见 [Local M15 Chat Smoke](./local-m15-chat-smoke/)。M18 Tool Runtime Catalog 验证见 [Local M18 Tool Runtime Catalog Validation](./local-m18-tool-runtime-catalog/)。

## M2 本地 API + DB 验证

### 1. 准备环境变量

```bash
cp .env.example .env
set -a
source .env
set +a
```

`.env.example` 当前使用本地 PostgreSQL：

```bash
APP_ENV=local
HTTP_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
LOG_LEVEL=info
READINESS_TIMEOUT_SECONDS=5
```

### 2. 启动 PostgreSQL

```bash
docker compose up -d postgres
```

### 3. 应用 schema baseline

如果本机没有 `migrate` CLI，可以使用 Docker image：

```bash
docker run --rm --network host -v "$PWD/migrations:/migrations" migrate/migrate \
  -path=/migrations \
  -database "$DATABASE_URL" \
  up

docker run --rm --network host -v "$PWD/migrations:/migrations" migrate/migrate \
  -path=/migrations \
  -database "$DATABASE_URL" \
  version
```

期望 version 为 `1`，且不是 dirty state。

### 4. 启动 API

```bash
go run ./cmd/loomi-api
```

期望看到包含 `operation_id` 的结构化启动日志。

### 5. 检查 health/readiness

```bash
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
```

期望：

- `/healthz` 返回 HTTP 200 和 `status: "alive"`。
- `/readyz` 在数据库和 schema baseline 可用时返回 HTTP 200 和 `status: "ready"`。

### 6. 验证 not-ready 行为

```bash
docker compose stop postgres
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
docker compose start postgres
```

期望：

- `/healthz` 仍返回 HTTP 200。
- `/readyz` 返回 HTTP 503 和 `status: "not_ready"`。
- 失败原因不包含完整 `DATABASE_URL` 或密码。

### 7. 验证 migration rollback/reapply

```bash
docker run --rm --network host -v "$PWD/migrations:/migrations" migrate/migrate \
  -path=/migrations \
  -database "$DATABASE_URL" \
  down 1

docker run --rm --network host -v "$PWD/migrations:/migrations" migrate/migrate \
  -path=/migrations \
  -database "$DATABASE_URL" \
  up
```

期望 rollback 和 reapply 都能完成，不需要手动清理业务表。

## 常用验证命令

```bash
go test ./...
cd web && bun run build
cd docs-site && bun run build
```

每次变更命令、环境变量、端口、数据库或验证流程时，都应同步更新对应 runbook。

M10 Persona/Skill foundation 的专项验收见 [Local M10 Persona Validation](/runbooks/local-m10-persona/)。
