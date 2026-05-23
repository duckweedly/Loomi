---
title: Spec Kit 工作流
description: Loomi 使用 Spec Kit 管理需求、计划、任务和实现。
---

Loomi 使用 Spec Kit 作为 AI 开发前的对齐层。它的作用不是简单生成代码，而是让每个非平凡功能都有可审查的需求、技术计划和任务拆分。

## 推荐顺序

```text
/speckit.specify
/speckit.clarify
/speckit.plan
/speckit.tasks
/speckit.analyze
/speckit.implement
```

Claude Code 项目内命令使用横线格式：

```text
/speckit-specify
/speckit-clarify
/speckit-plan
/speckit-tasks
/speckit-implement
```

`/speckit-specify` 关注用户目标、功能边界、验收标准和非目标，不应提前写实现细节。

`/speckit-clarify` 用来收敛影响实现的模糊点，避免 AI 自行猜测产品决策。

`/speckit-plan` 把需求翻译成技术方案，并说明依赖、约束、数据模型、接口和验证方式。

`/speckit-tasks` 将计划拆成可独立完成、可验证的任务。

`/speckit-implement` 按任务实现，并在必要时回到 spec 或 plan 修正前提。

## 当前功能：M3.5 Frontend Agent Runtime Skeleton

当前 Spec Kit 功能目录：

```text
specs/004-frontend-agent-runtime/
```

关键产物：

- `spec.md`：定义 M3.5 范围，明确只做前端 Agent runtime 骨架、Chat Canvas 状态机、mock run/event 剧本和 future real adapter 接入点。
- `plan.md`：确定新增 `web/src/runtime/` 边界，保持 M3 thread/message client 和 M4/M5 run/event/LLM 后端语义分离。
- `research.md`：记录 runtime adapter、纯状态派生、deterministic mock scripts、real-mode unavailable 和 Chat-first 范围决策。
- `data-model.md`：定义 Chat Canvas State、Runtime Run、Runtime Event、Runtime Script、Assistant Draft、Execution Adapter、Backend Capability State 和 Stale Event Guard。
- `contracts/execution-adapter.md`：定义 sendMessage/createRun/subscribeRunEvents/appendAssistantDelta/completeRun/failRun/stopRun 契约。
- `contracts/runtime-events.md`：定义 run.created、context.loading、assistant.thinking、assistant.drafting、assistant.message.completed、run.completed、run.failed、run.stopped 的可见语义。
- `contracts/chat-canvas-states.md`：定义 Chat Canvas 状态优先级和跨表面一致性要求。
- `quickstart.md`：记录 mock success/failure/stopped、stale event 和 real-mode unavailable 验证流程。
- `tasks.md`：按 P1/P2 user story 拆分并跟踪实现任务。

## 近期已完成：M4 Run、Event 与 SSE

Spec Kit 功能目录：

```text
specs/003-m4-run-event-sse/
```

关键产物：

- `spec.md`：定义本地 run/event/SSE 用户故事和明确 deferred 的后续平台能力。
- `plan.md`：确定使用 deterministic local simulation、PostgreSQL persistence 和 history-first SSE。
- `data-model.md`：定义 Run、Run Event、Event Stream Cursor、Stop Request、Deterministic Local Simulation、Stream State 和 M4 Schema Revision。
- `contracts/`：记录 HTTP、SSE、migration 与 frontend data-source 契约。
- `tasks.md`：记录 M4 实现任务状态。

2026-05-23 状态：M4 core run/event/SSE slice 已实现，验证结果记录在 M4 devlog。

## 近期已完成：M3 Auth、Thread 与 Message

Spec Kit 功能目录：

```text
specs/002-m3-auth-thread-message/
```

关键产物：

- `spec.md`：定义 M3 范围，明确只做本地 identity、users、threads、messages、real/mock API 切换、M3 readiness、seed 与文档边界。
- `plan.md`：确定复用 M2 Go API/PostgreSQL/migration/diagnostics 基座，并新增 `internal/identity`、`internal/productdata` 与 `/v1` thread/message API。
- `research.md`：记录为什么继续使用 Go stdlib HTTP、pgxpool、显式 migration、固定本地用户、partial unique idempotency、`VITE_LOOMI_API_BASE_URL` 与显式 seed 命令。
- `data-model.md`：定义 Local Identity、User、Thread、Message、Client Message Identifier、API Error、Schema Revision、Seed Data Set 与 Frontend Data Source Mode。
- `contracts/http-m3.openapi.yaml`：定义 `/v1/me`、`/v1/threads` 和 `/v1/threads/{thread_id}/messages` 响应契约。
- `contracts/migration-cli.md`：定义 M3 schema 的 apply/version/rollback/reapply 命令契约。
- `contracts/seed-cli.md`：定义显式本地 seed 命令、固定 demo IDs 和幂等 demo data 行为。
- `contracts/frontend-data-source.md`：定义前端 mock/real API 切换、real API 失败不自动 fallback 和 stale response guard 的规则。
- `quickstart.md`：记录本地 M3 readiness、CRUD、idempotency、seed、前端 mock/real 和 docs 验证流程。
- `tasks.md`：记录 M3 任务拆分与验证门。

## 与文档站的关系

Spec Kit 产物是开发事实来源之一，文档站是面向阅读和检索的长期知识库。开发时应把关键规格、架构设计、接口变化、验证结果和技术取舍同步到 `docs-site/src/content/docs/`。

M2/M3/M3.5/M4 实现都应同步更新 API 文档、架构说明、runbook、Spec Kit 入口和 devlog。后续 LLM Gateway、工具调用、Worker、Pipeline、桌面运行时和多 Agent 能力也应沿用相同模式。
