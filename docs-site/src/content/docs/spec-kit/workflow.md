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

## 与文档站的关系

Spec Kit 产物是开发事实来源之一，文档站是面向阅读和检索的长期知识库。开发时应把关键规格、架构设计、接口变化、验证结果和技术取舍同步到 `docs-site/src/content/docs/`。

M2 实现同步更新了 API 文档、架构说明、runbook、Spec Kit 入口和 devlog。后续 M3/M4 应继续沿用相同模式，把 thread/message、run/event/SSE 的规格、契约和验证结果写入 Spec Kit 与文档站。
