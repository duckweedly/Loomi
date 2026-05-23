---
title: M3.5 Frontend Agent Runtime Skeleton
description: 记录 M3.5 前端 runtime 状态机、mock scripts、adapter 边界和验证结果。
---

## 完成内容

M3.5 增加了前端 Agent runtime 骨架，让 Loomi 在真实 LLM/工具/Worker 后端完成前也能验证一次 Chat execution loop；合入 M4 后，real API mode 使用 M4 run/event/SSE。

主要变化：

- 新增 `web/src/runtime/`，包含 execution adapter、mock adapter、real adapter、runtime scripts 和 Chat Canvas state derivation。
- 新增 runtime domain types：RuntimeStatus、RuntimeEvent、RuntimeRun、AssistantDraft、RuntimeScript、BackendCapabilityState、StaleEventGuard。
- Chat Canvas 改为显式状态渲染，覆盖 no-thread、empty-thread、loading、error、history、waiting-run、running、completed、failed、backend-unavailable。
- Mock adapter 提供 deterministic success/failure scripts，并保证每次 run/event ID 独立。
- Real API mode 通过 M4 `realApiClient` 启动 run、订阅 SSE 和停止 run；mock scenario selector 只在 mock mode 暴露。
- RunRail/RunTimeline/AgentStateMotion 从同一个 selected run 派生状态。
- `tasks.md` 已按实现进度勾选 M3.5 任务。

## 验证结果

当前验证：

```bash
bun test ./web/src/*.test.ts ./web/src/components/*.test.ts ./web/src/runtime/*.test.ts
# 79 pass, 0 fail

bun run --cwd web build
# passed

bun run --cwd docs-site build
# passed, 29 pages built
```

浏览器 smoke：

- Mock success：提交 Chat 消息后，用户消息立即出现；Timeline 显示 6 个里程碑；Agent 状态进入 Done；Chat Canvas 显示“已完成”。
- Mock failure：在 Run Rail 选择 Fail 后提交消息，Timeline 只显示一个 `run.failed` terminal event；Chat Canvas 显示“执行失败”；没有追加成功 assistant 回复。
- Stopped/failed 语义：默认 stopped thread 显示 Agent Error 和失败类 Chat Canvas 状态。
- Real API M4：M4 smoke 通过；页面命中 real API `/v1/threads`、`/runs/current`、`/events`，timeline 显示 local simulated events，并支持 stop/already-terminal 语义。
- Clean dev server console：重启 Vite 后重新加载页面，无浏览器控制台 error。

## 已知限制

- 当前 mock script 在测试路径中同步推进，浏览器中的可见延迟仍可继续打磨。
- Failure script 的选择入口是 mock mode 的 smoke 控件，不是正式产品控件。
- Work mode 暂不接 runtime execution，只保留 mode-specific thread 列表和现有行为。
- LLM/tool/worker 仍未接入；M4 real run/event/SSE 只是 local simulated execution。

## 下一步

- M5：接入 LLM gateway，并把 model delta 映射到同一 RuntimeEvent 语义。
- 后续：为 Work mode 设计 project/task execution flow，再复用 M3.5 runtime adapter/state model。
