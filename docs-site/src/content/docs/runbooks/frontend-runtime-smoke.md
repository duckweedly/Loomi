---
title: 前端 Runtime Smoke Runbook
description: M3.5 mock success、failure、stopped、stale event 和 real-mode unavailable 的本地验证步骤。
---

这份 runbook 用来验证 M3.5 前端 Agent runtime 骨架。默认使用 mock mode；real API mode 只用于验证 runtime 能力未接入状态。

## 自动验证

从仓库根目录执行：

```bash
bun test ./web/src/**/*.test.ts "web/vite.config.test.ts"
bun run --cwd web build
bun run --cwd docs-site build
```

期望结果：

- 所有 web tests 通过。
- web build 无 TypeScript 错误。
- docs build 通过。

## Mock success smoke

1. 不设置 `VITE_LOOMI_API_BASE_URL`。
2. 启动 web renderer。
3. 进入 Chat mode。
4. 选择或创建一个 Chat thread。
5. 输入一条消息并提交。
6. 验证用户消息立即出现在 Chat Canvas。
7. 验证 Run Timeline 出现这些事件：
   - `run.created`
   - `context.loading`
   - `assistant.thinking`
   - `assistant.drafting`
   - `assistant.message.completed`
   - `run.completed`
8. 验证 Agent 状态徽章进入非 idle，并最终进入 done。
9. 验证 Chat Canvas 只出现一条最终 assistant 回复。

## Mock failure smoke

1. 将 mock runtime script 切到 `failure`。
2. 提交一条 Chat 消息。
3. 验证用户消息立即出现。
4. 验证 Timeline 终态为 `run.failed`。
5. 验证 Chat Canvas 显示“执行失败”。
6. 验证没有追加成功 assistant 回复。

## Stop run smoke

1. 启动一个 mock run。
2. 在完成前点击 Run Rail 中的“停止”。
3. 验证 Timeline 出现 stopped terminal 状态。
4. 验证 Chat Canvas 进入 failed/stopped 语义。
5. 验证 Agent 状态徽章进入 error。
6. 验证后续 script event 不会把 stopped run 推进到 completed。

## Stale event smoke

1. 在 Chat thread A 启动 mock run。
2. run 尚未完成前切换到 Chat thread B。
3. 验证 thread A 的后续 event 不会改变 thread B 的 Chat Canvas、Timeline 或 Agent 状态徽章。

## Real API unavailable smoke

1. 设置 `VITE_LOOMI_API_BASE_URL` 指向本地 API。
2. 启动 web renderer。
3. 选择一个 real API thread。
4. 提交消息触发 runtime execution。
5. 验证 Chat Canvas 在 1 秒内显示“后端能力未接入”。
6. 验证没有 mock run/event 被隐藏执行。

## 常见问题

### 看到后端能力未接入是不是错误？

不是。M3.5 只做前端 runtime 骨架；真实 run/event/SSE 在 M4 以后接入。

### 为什么 mock success 会马上 completed？

测试环境使用 deterministic scripts，目的是验证顺序和状态语义。浏览器里可以通过可控延迟让过程更可见，但不应引入随机事件。

### 为什么 Work mode 不接 runtime？

当前 slice 只验证 Chat execution loop。Work mode 有不同业务语义，等 project/task flow 设计清楚后再复用同一 runtime 模型。
