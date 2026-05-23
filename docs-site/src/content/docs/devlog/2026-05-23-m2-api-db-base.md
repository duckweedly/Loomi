---
title: 2026-05-23 M2 API 与数据库基座
description: 为 Loomi 增加本地 Go API、health/readiness、PostgreSQL schema baseline 和验证流程。
---

## 关联阶段

M2 API 与数据库基座。

## 完成内容

新增 `loomi-api` 本地服务，提供 `/healthz` 和 `/readyz`。服务在 PostgreSQL 不可用时仍可启动，`/healthz` 继续报告 alive，`/readyz` 报告 not ready 并给出不含敏感信息的失败原因。

新增运行配置加载、结构化诊断、request id、operation id、PostgreSQL pool 和 schema readiness 检查。

新增本地 PostgreSQL `compose.yaml`、`.env.example`、M2 schema baseline migration，以及 `services/api/README.md`。

更新 Spec Kit 产物和文档站，记录 M2 的 API 契约、migration 契约、数据模型、quickstart、runbook 和非目标。

## 验证结果

已执行：

```bash
go test ./internal/config
go test ./internal/diagnostics
go test ./internal/db
go test ./internal/httpapi
go test ./...
```

结果：通过。

已执行本地 PostgreSQL smoke：

```bash
docker compose up -d postgres
docker run --rm --network host -v "$PWD/migrations:/migrations" migrate/migrate -path=/migrations -database "postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable" up
docker run --rm --network host -v "$PWD/migrations:/migrations" migrate/migrate -path=/migrations -database "postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable" version
docker run --rm --network host -v "$PWD/migrations:/migrations" migrate/migrate -path=/migrations -database "postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable" down 1
docker run --rm --network host -v "$PWD/migrations:/migrations" migrate/migrate -path=/migrations -database "postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable" up
```

结果：up/version/down/reapply 通过，version 为 `1`。

已执行 API smoke：

```bash
go run ./cmd/loomi-api
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
docker compose stop postgres
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
docker compose start postgres
```

结果：数据库可用时 `/readyz` 返回 HTTP 200；数据库停止时 `/healthz` 仍返回 HTTP 200，`/readyz` 返回 HTTP 503。

## 已知限制

M2 不包含认证、用户、thread、message、run、event、SSE、worker、LLM gateway、工具调用、桌面运行时或生产部署。

本地 `migrate` CLI 不是必须项；当前 smoke 使用 Docker image `migrate/migrate`。

本地 PostgreSQL 端口使用 `55433`，用于避开已有开发容器占用的 `55432`。

## 下一步

继续 M3：认证、thread 和 message 数据模型。M3 应在 M2 service/readiness 基础上增加真实业务表、访问边界和前端 API client 切换路径。
