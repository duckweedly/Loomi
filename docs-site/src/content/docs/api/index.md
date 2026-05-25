---
title: API 文档入口
description: Loomi API、事件和数据模型文档入口。
---

M2 已建立 Loomi 的第一个本地 API 服务边界。当前 API 只覆盖进程健康、依赖就绪和本地开发验证，不包含产品数据接口。

## M2 health/readiness

### `GET /healthz`

用于确认 `loomi-api` 进程仍能响应。这个检查不依赖 PostgreSQL，也不表示服务已经可以承接依赖数据库的行为。

成功响应：

```json
{
  "status": "alive",
  "service": "loomi-api",
  "environment": "local",
  "request_id": "req_example"
}
```

### `GET /readyz`

用于确认 M2 必需依赖是否可用。只有配置有效、PostgreSQL 可 ping、schema baseline 已应用且非 dirty 时才返回 ready。

Ready 响应为 HTTP 200：

```json
{
  "status": "ready",
  "service": "loomi-api",
  "environment": "local",
  "request_id": "req_example",
  "checks": [
    { "name": "config", "status": "ok" },
    { "name": "database", "status": "ok" },
    { "name": "schema", "status": "ok" }
  ]
}
```

Not ready 响应为 HTTP 503：

```json
{
  "status": "not_ready",
  "service": "loomi-api",
  "environment": "local",
  "request_id": "req_example",
  "checks": [
    { "name": "config", "status": "ok" },
    { "name": "database", "status": "failed", "reason": "database ping failed" },
    { "name": "schema", "status": "failed", "reason": "schema version unavailable" }
  ]
}
```

Failure reasons must stay non-secret. Full `DATABASE_URL` values must not appear in responses or logs.

## M2 migration contract

M2 uses explicit migration commands. The API process does not auto-apply migrations at startup.

Migration files:

```text
migrations/000001_schema_baseline.up.sql
migrations/000001_schema_baseline.down.sql
```

This baseline intentionally creates no business tables. It only lets `golang-migrate` record schema version state, so M3/M4 can own users, threads, messages, runs, and events.

## Deferred APIs

The following remain out of scope until later milestones:

- Authentication and `/me`
- Threads and messages
- Runs, run events, and SSE
- Worker/job execution
- LLM gateway, tool calls, and tool catalog
- Persona selection and safe persona summaries
- Desktop runtime and local bridge

M18 Tool Runtime Catalog adds `GET /v1/tools/catalog`; see [Tool Runtime Catalog API](./tool-runtime-catalog/).
