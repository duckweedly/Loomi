---
title: M3 Auth、Thread 与 Message 架构
description: M3 的本地身份、thread/message 数据层、运行边界和后续延期能力。
---

M3 在 M2 API/DB 基座上引入第一层真实产品数据：固定本地开发身份、用户记录、thread 容器和完整用户消息。它不是生产认证系统，也不是 agent run 系统。

## 分层边界

```text
cmd/loomi-api
└── internal/httpapi
    └── internal/productdata
        ├── internal/identity
        └── internal/db / pgxpool
```

`internal/identity` 只解析固定本地用户 `user_local_dev`。这个边界存在的原因是：M3 需要所有持久数据都有 owner，但还不能提前引入 session、组织、多用户切换或生产 auth。

`internal/productdata` 是 thread/message 的用例层。它负责：

- 确保本地用户存在。
- 创建、列出、读取、更新、归档 thread。
- 创建和列出用户消息。
- 用 `client_message_id` 防止重复消息。
- 保证 message 创建和 thread `updated_at` 更新在同一个可见操作里完成。

## Thread lifecycle 不等于 run status

M3 的 thread 只有 `active` 和 `archived` 生命周期。它不表达 agent 是否 running、completed 或 stopped。

前端仍然可以显示 mock run timeline，但真实 API 返回的 thread/message 不代表任何真实执行状态。这个分离能避免 M3 数据模型污染 M4 的 run/event/SSE 设计。

## Message 语义

M3 只持久化完整用户文本消息：

- `role` 固定为 `user`。
- `content` 是最终文本，不是 streaming delta。
- `metadata` 保留为空对象，不承载 run/tool/model 语义。
- 不创建 assistant placeholder。
- 不创建 run event、tool call、worker job 或 LLM request。

## Idempotency

`client_message_id` 是可选的。提供时，数据库通过 `(thread_id, user_id, client_message_id)` 的 partial unique index 保证同一 thread 和本地用户下重复请求返回原消息。

重复请求不会创建新消息，也不会推进 thread `updated_at`。Web real API client 会为每次发送生成 `web-<timestamp>-<random>` 形式的 idempotency key，避免同一毫秒内的用户发送被误判为重复请求。

## Seed boundary

Seed command 使用 `productdata.SeedService` 的固定 ID upsert 边界创建本地 demo 数据：`thr_local_demo` 和 `msg_local_demo_001`。普通产品 API 仍通过随机文本 ID 创建 thread/message，避免把 demo data determinism 暴露给常规用户操作。

## Deferred beyond M3

以下能力明确延期：

- run/event/SSE execution timeline
- LLM gateway 和 assistant message generation
- tool calling
- worker/job queue
- desktop runtime、SQLite adapter、Electron bridge、tray、auto-update
- attachments/file upload
- RAG/context ingestion
- catalog、marketplace、plugin runtime
- production auth、organization、hosted deployment
