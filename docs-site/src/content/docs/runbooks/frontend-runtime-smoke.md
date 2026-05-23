---
title: 前端 Runtime Smoke Runbook
description: M3.5 mock success、failure、stopped、stale event 和 M4 real API run/event/SSE 的本地验证步骤。
---

这份 runbook 用来验证 M3.5 前端 Agent runtime 骨架。默认使用 mock mode 验证 deterministic runtime scripts；设置 `VITE_LOOMI_API_BASE_URL` 后使用 M4 real API 验证 persisted run/event/SSE。

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

1. 启动一个 run。
2. 在 run 仍处于 running 时点击 Run Rail 中的“Stop run”。
3. 验证 Timeline 出现 stopped terminal 状态。
4. 验证 Chat Canvas 进入 failed/stopped 语义。
5. 验证 Agent 状态徽章进入 error。
6. 验证后续 event 不会把 stopped run 推进到 completed。

Mock scripts 在当前测试路径中会同步推进到 terminal；需要可见 stop 流程时，使用 M4 real API local simulated run 或后续引入可控延迟的 mock driver。

## Stale event smoke

1. 在 Chat thread A 启动 run。
2. run 尚未完成前切换到 Chat thread B。
3. 验证 thread A 的后续 event 不会改变 thread B 的 Chat Canvas、Timeline 或 Agent 状态徽章。

Mock scripts 当前同步完成，浏览器里更适合用 M4 real API SSE 或后续可控延迟 mock driver 验证可见切换过程；单元测试覆盖 stale event guard。

## Real API M4 smoke

1. 启动本地 Go API，并确认 `/readyz` ready。
2. 设置 `VITE_LOOMI_API_BASE_URL` 指向本地 API。
3. 启动 web renderer。
4. 选择一个 real API thread。
5. 提交消息触发 runtime execution。
6. 验证 Chat Canvas 显示 `Real API`、`Local simulated` 和 stream state。
7. 验证网络请求命中 `/v1/threads`、`/v1/threads/{thread_id}/runs/current`、`/v1/runs/{run_id}/events/stream`。
8. 验证 Run Timeline 显示 persisted lifecycle/progress/message/final events。
9. 点击 running run 的 stop 控制时，验证 run 进入 stopped 或 already-terminal 语义。
10. 验证没有 mock scenario selector 出现在 real API mode。

## 常见问题

### 看到后端能力未接入是不是错误？

如果配置了 M4 real API 并且 readiness 通过，这表示当前 real API wiring 有问题，应优先检查 `VITE_LOOMI_API_BASE_URL`、API server 日志和 `/readyz`。未配置 real API 时，mock mode 不会显示这个状态。

### 为什么 mock success 会马上 completed？

测试环境使用 deterministic scripts，目的是验证顺序和状态语义。浏览器里可以通过 M4 real API local simulated run 观察更完整的 stream lifecycle；后续也可以引入可控延迟的 mock driver，但不应引入随机事件。

### 为什么 Work mode 不接 runtime？

当前 slice 只验证 Chat execution loop。Work mode 有不同业务语义，等 project/task flow 设计清楚后再复用同一 runtime 模型。
