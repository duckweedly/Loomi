---
title: M2 API 与数据库基座
description: Loomi 当前后端基础设施的边界、数据流和非目标。
---

M2 的目标是让 Loomi 从前端 mock 进入真实后端边界。它不追求完整 Agent Loop，而是提供可运行、可测试、可替换 mock 的基础服务。

## 服务边界

M2 新增 `loomi-api` 本地服务。进程入口位于 `cmd/loomi-api/`，内部包位于 `internal/`：

- `internal/config/` 负责读取和校验本地运行配置。
- `internal/diagnostics/` 负责结构化日志、request id 和 operation id。
- `internal/db/` 负责 PostgreSQL pool、ping 和 schema readiness。
- `internal/httpapi/` 负责 `/healthz` 与 `/readyz`。

当前服务只暴露 health/readiness 行为。M1 的 `web/` mock UI 不依赖 M2 服务，后续 M3/M4 会逐步把真实 thread、message、run 和 event 接入现有 UI 边界。

## 配置模型

M2 本地配置来自环境变量：

```bash
APP_ENV=local
HTTP_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
LOG_LEVEL=info
READINESS_TIMEOUT_SECONDS=5
```

缺少或格式错误的配置会让服务启动失败。敏感配置必须在日志中脱敏。

## Health 与 readiness

`/healthz` 只表示进程仍然可响应，不检查 PostgreSQL。即使 PostgreSQL 不可用，服务也应继续启动并保持 liveness 可查。

`/readyz` 表示依赖是否可用。只有以下检查都通过才返回 ready：

1. 配置已经通过启动校验。
2. PostgreSQL ping 成功。
3. schema baseline version 存在且非 dirty。

当 PostgreSQL 停止时，`/readyz` 返回 HTTP 503，并给出不含敏感信息的失败原因。

## 数据库基线

M2 使用本地 PostgreSQL，默认由 `compose.yaml` 提供 `postgres` service。M2 只建立 migration/version baseline，不创建业务表。

```text
migrations/000001_schema_baseline.up.sql
migrations/000001_schema_baseline.down.sql
```

API 服务不会自动运行 migration。开发者必须显式执行 up/down/reapply 验证，这样 schema 变更保持可见、可回滚。

## 可观测性

M2 的最低可观测性是结构化诊断输出：

- HTTP health/readiness 响应包含 `request_id`。
- 启动日志包含 `operation_id`。
- readiness failure reason 不泄露完整连接串或凭证。

完整 metrics、distributed tracing 和 run event timeline 留到后续执行链路成熟后再引入。

## 非目标

M2 不做：

- 认证、用户、thread、message。
- run、run_event、SSE。
- worker、job queue、LLM gateway、tool calling。
- Electron runtime、SQLite、本地 bridge、系统托盘、自动更新。
- 生产部署、发布包和多用户托管环境。
