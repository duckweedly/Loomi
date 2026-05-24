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
7. 验证 pending assistant bubble 出现在 Chat Canvas 中，并随着 `assistant.drafting` 或 `model.delta` 片段增长。
8. 验证 Run Timeline 按组显示事件：
   - Run lifecycle: `run.created`, terminal run events
   - Model stream: `assistant.drafting`, `model.delta`, `model.final`, `model.usage`
   - Worker/job: `job.queued`, `worker.claimed`, `job.retrying` when using richer mock scripts
   - Error: provider/model/stream/backend failures
9. 验证 token usage/provider metadata 只出现在 timeline/debug detail，不进入 assistant message text。
10. 验证 Agent 状态徽章进入 speaking/thinking，并最终进入 done。
11. 验证 Chat Canvas 只出现一条最终 assistant 回复，不重复显示 final draft 和 persisted assistant message。
12. 验证 Regenerate 可见，触发后旧 assistant 回复仍保留，并出现一个关联旧消息的新 pending attempt。
13. 验证 Continue 只在有选中 thread、输入非空且无 active run 时可用；active run 期间 Send/Continue/Retry/Regenerate 都被阻止。

## Mock failure smoke

1. 将 mock runtime script 切到 `failure`。
2. 提交一条 Chat 消息。
3. 验证用户消息立即出现。
4. 验证 Timeline 终态为 `run.failed`。
5. 验证 Chat Canvas 显示“执行失败”，并保留失败前已经生成的 partial assistant draft。
6. 验证没有追加成功 assistant 回复。
7. 验证 Retry 可见，触发后会创建新的 pending attempt，失败上下文在新 attempt 可见前不被清空。

## Stop run smoke

1. 启动一个 run。
2. 在 run 仍处于 running 时点击 Run Rail 中的“Stop run”。
3. 验证 Timeline 出现 stopped terminal 状态。
4. 验证 Chat Canvas 进入 stopped 语义，并保留停止前 partial assistant draft。
5. 验证 Agent 状态徽章进入 error。
6. 验证后续 final event 不会把 stopped run 推进到 completed。

Mock API 的 send path 会先返回 running run，再通过 deterministic subscription path 分步推送事件；浏览器可见 pending bubble、streaming draft、grouped timeline 和 terminal transition。需要更长时间窗口时，使用 M4 real API local simulated run。

## Thread/message state smoke

1. 加载无选中 thread 状态，验证 Chat Canvas 不显示旧 messages。
2. 选择空 Chat thread，验证 empty state 与 loading state 不同。
3. 模拟 thread/message loading，验证 ThreadSidebar 和 Chat Canvas 显示 loading affordance。
4. 模拟加载失败，验证 error 与 Retry 可见，并且 selected thread context 不被清空。
5. 选择带 persisted assistant message 和 run events 的 thread，验证 Chat Canvas 与 Run Timeline 对 latest run outcome 一致。

## Stale event smoke

1. 在 Chat thread A 启动 run。
2. run 尚未完成前切换到 Chat thread B。
3. 验证 thread A 的后续 event 不会改变 thread B 的 Chat Canvas、Timeline 或 Agent 状态徽章。

Mock scripts 通过 deterministic subscription path 分步推进；浏览器里也可以用 M4 real API SSE 验证更完整的可见切换过程。单元测试覆盖 stale event guard 和 out-of-order delta protection。

## Real API M4 smoke

1. 启动本地 Go API，并确认 `/readyz` ready。
2. 设置 `VITE_LOOMI_API_BASE_URL` 指向本地 API。
3. 启动 web renderer。
4. 选择一个 real API thread。
5. 提交消息触发 runtime execution。
6. 验证 Chat Canvas 和 Run Timeline 显示 capability status：mock、Local simulated、Real model、Backend unavailable、Model setup missing、Provider unavailable、Stream disconnected 或 Run recovering 中的当前最高优先级状态。
7. 验证 local simulated 不被描述为 real model output；backend unavailable 不显示 model-thinking copy。
8. 验证网络请求命中 `/v1/threads`、`/v1/threads/{thread_id}/runs/current`、`/v1/runs/{run_id}/events/stream`。
9. 验证 Run Timeline 显示 persisted lifecycle/progress/message/final events。
10. 点击 running run 的 stop 控制时，验证 run 进入 stopped 或 already-terminal 语义。
11. 验证没有 mock scenario selector 出现在 real API mode。

## 常见问题

### 看到后端能力未接入是不是错误？

如果配置了 M4 real API 并且 readiness 通过，这表示当前 real API wiring 有问题，应优先检查 `VITE_LOOMI_API_BASE_URL`、API server 日志和 `/readyz`。未配置 real API 时，mock mode 不会显示这个状态。

### 为什么 mock success 会马上 completed？

测试环境使用 deterministic scripts，目的是验证顺序和状态语义。浏览器里可以通过 M4 real API local simulated run 观察更完整的 stream lifecycle；后续也可以引入可控延迟的 mock driver，但不应引入随机事件。

### 为什么 Work mode 不接 runtime？

当前 slice 只验证 Chat execution loop。Work mode 有不同业务语义，等 project/task flow 设计清楚后再复用同一 runtime 模型。
