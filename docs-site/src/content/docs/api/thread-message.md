---
title: M3 Thread 与 Message API
description: M3 本地身份、thread、message 与结构化错误 API 契约。
---

M3 API 使用 M2 的本地 HTTP 服务，新增 `/v1` JSON endpoint。所有 product data 请求都解析为固定本地用户 `user_local_dev`，客户端不能通过 header 或参数选择用户。

## Identity

```http
GET /v1/me
```

返回固定本地用户：

```json
{
  "user": {
    "id": "user_local_dev",
    "display_name": "Local Developer",
    "created_at": "2026-05-23T00:00:00Z",
    "updated_at": "2026-05-23T00:00:00Z"
  },
  "request_id": "req_..."
}
```

## Threads

```http
GET /v1/threads
POST /v1/threads
GET /v1/threads/{thread_id}
PATCH /v1/threads/{thread_id}
POST /v1/threads/{thread_id}/archive
```

创建请求：

```json
{
  "title": "M3 smoke thread",
  "mode": "chat"
}
```

更新请求至少包含一个字段：

```json
{
  "title": "Renamed thread",
  "mode": "work"
}
```

Thread response 中 `lifecycle_status` 只表示 thread 生命周期，不表示 run 状态。

## Messages

```http
GET /v1/threads/{thread_id}/messages
POST /v1/threads/{thread_id}/messages
```

创建请求：

```json
{
  "content": "Persist this local user message.",
  "client_message_id": "web-unique-id"
}
```

`client_message_id` 可选。重复使用同一个 idempotency key 时，API 返回既有 message 并使用 HTTP 200；新建 message 使用 HTTP 201。

M3 message 只支持 `role: "user"`，不会创建 assistant placeholder、run event、streaming delta、tool call 或 LLM request。

Model-gateway assistant messages created by later milestones include safe run linkage in the message projection:

```json
{
  "id": "msg_...",
  "thread_id": "thr_...",
  "role": "assistant",
  "content": "Final answer",
  "metadata": { "run_id": "run_..." },
  "run_id": "run_...",
  "created_at": "2026-05-27T00:00:00Z"
}
```

`run_id` is duplicated from safe metadata so clients can reconcile a terminal run with its persisted assistant message without parsing debug metadata. This field is display/state linkage only; it is not an authorization boundary.

## Frontend real API mode

前端只有在设置 `VITE_LOOMI_API_BASE_URL` 时才调用这些 `/v1` endpoint。未设置时仍使用 mock data；已设置但 API 不可用时，UI 显示可恢复错误，不自动 fallback 到 mock。

本地 API 对 loopback origin 的 `/v1/*` 请求返回 CORS header，并支持 `GET, POST, PATCH, OPTIONS` preflight，所以 Vite dev server 可以从 `127.0.0.1:5173` 调用 `127.0.0.1:8080` 的 JSON write endpoint。非本地 origin 不会获得 `Access-Control-Allow-Origin`。

请求 JSON 会拒绝未知字段，保持后端行为和 OpenAPI 的 `additionalProperties: false` 一致。

M3 API 不提供 run/event/SSE 数据。前端 run timeline 在 M3 仍是 mock、empty 或 deferred surface。

## Structured errors

错误响应统一包含 stable code、human message 和 request id：

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Message content is required.",
    "request_id": "req_..."
  }
}
```

初始错误码：

| Code | HTTP | Meaning |
| --- | --- | --- |
| `invalid_request` | 400 | JSON、title、mode、message、client id 或空 PATCH 无效 |
| `thread_not_found` | 404 | thread 不存在或不属于当前固定本地用户 |
| `method_not_allowed` | 405 | endpoint 不支持该 HTTP method |
| `internal_error` | 500 | 未预期服务端错误，或 product data 在 DB 不可用时未启用 |

错误 message 不包含 database URL、credential 或 secret。
