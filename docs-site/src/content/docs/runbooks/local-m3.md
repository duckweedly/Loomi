---
title: M3 本地运行手册
description: M3 migration、readiness、seed、API smoke、前端 real/mock 切换和验证命令。
---

## 环境变量

```bash
APP_ENV=local
HTTP_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
LOG_LEVEL=info
READINESS_TIMEOUT_SECONDS=5
```

前端真实 API 模式：

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080
```

未设置 `VITE_LOOMI_API_BASE_URL` 时，web shell 使用 M1 mock data。

## 启动数据库和 migration

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

M3 需要 version `2` 且 clean。

## Readiness smoke

M2-only baseline 应该 not ready：

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
curl -i http://127.0.0.1:8080/readyz
```

重新 apply M3 后应该 ready：

```bash
migrate -path migrations -database "$DATABASE_URL" up
curl -i http://127.0.0.1:8080/readyz
```

## API smoke

```bash
curl -s http://127.0.0.1:8080/v1/me
curl -s -X POST http://127.0.0.1:8080/v1/threads \
  -H 'Content-Type: application/json' \
  -d '{"title":"M3 smoke thread","mode":"chat"}'
curl -s http://127.0.0.1:8080/v1/threads
curl -s -X POST "http://127.0.0.1:8080/v1/threads/$THREAD_ID/messages" \
  -H 'Content-Type: application/json' \
  -d '{"content":"Persist this local user message.","client_message_id":"smoke-message-001"}'
```

重复最后一个 message 请求应返回同一个 message id。

## Seed

```bash
go run ./cmd/loomi-seed
```

Seed 可重复运行，不应创建重复 demo thread 或 demo message。Demo thread/message 使用固定 ID：`thr_local_demo` 和 `msg_local_demo_001`；如果本地用户改名 demo thread，再次 seed 会复用同一个 thread id 并恢复 seed 标题。Seed 不会删除本地数据。

## Frontend smoke

Mock mode：

```bash
cd web
bun run dev
```

Real API mode：

```bash
cd web
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run dev
```

如果配置了 real API 但 API 不可用，UI 应显示可恢复错误，不自动 fallback 到 mock thread/message。真实 API 模式应检查：加载 thread list、创建 thread、改名、归档、发送 message、刷新后仍能看到持久化数据。API 已支持 loopback origin 的 `/v1/*` CORS preflight，浏览器 JSON write 不需要代理。

M3 的真实 API 只覆盖本地 identity、thread 和 message；run/event/SSE、LLM、tool、worker 和 desktop runtime 仍是延期能力。

## Validation commands

```bash
go test ./...
cd web && bun run build
cd docs-site && bun run build
```

UI 改动完成前还需要在浏览器里分别检查 mock mode 和 real API mode。
