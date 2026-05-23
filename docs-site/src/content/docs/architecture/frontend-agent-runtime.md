---
title: M3.5 前端 Agent Runtime 骨架
description: 解释 Chat Canvas 状态机、runtime adapter 边界、mock run/event 剧本和 M4 real API run/event/SSE 接入边界。
---

M3.5 的目标不是提前实现真实 LLM/工具/Worker 后端，而是在 M4 run/event/SSE 和 M5 LLM gateway 分阶段推进时，让前端拥有可验证的 Agent 交互骨架。

这个骨架有三个边界：

1. `web/src/runtime/`：前端 runtime 语义边界。
2. `web/src/state.ts`：当前选中 thread、messages、run、SSE stream state 和 stale event guard 的协调层。
3. `web/src/components/`：Chat Canvas、Run Timeline、Agent 状态徽章的同源渲染层。

## 为什么要单独建 runtime 边界

M3 的 `ApiClient` 负责 durable thread/message 数据。M4 已提供真实 run/event/SSE，M5+ 再接 LLM/tool 执行。如果把 mock runtime、real API 和未来 LLM gateway 都直接塞进组件，mock 和 real 很容易各走一套 UI 逻辑。

因此 M3.5 新增 `ExecutionAdapter` 作为前端 runtime 语义边界：

```text
sendMessage
createRun
subscribeRunEvents
appendAssistantDelta
completeRun
failRun
stopRun
runtimeCapability
```

mock adapter 报告 `available`，并播放 deterministic scripts。real API mode 报告 M4 run/event capability `available`，真实发送、启动 run、SSE 订阅和停止 run 继续由 `realApiClient` 负责，避免把 M4 已接入能力误判为 unavailable，也不隐藏 fallback 到 mock 执行。

## Chat Canvas 状态机

`deriveChatCanvasState` 是纯函数，输入来自 loading/error、selected thread、message count、run、backend capability。优先级固定：

```text
loading
error
backend-unavailable
no-thread
runtime states
empty-thread
history
```

这让 Chat Canvas 可以明确显示：

- 未选择会话
- 新对话
- 加载中
- 加载失败
- 历史消息
- 等待执行
- 执行中
- 已完成
- 执行失败
- 后端能力未接入

产品 UI 只保留短标签；详细解释放在文档和测试里。

## Mock runtime scripts

`runtimeScripts.ts` 定义两个第一阶段剧本：

成功剧本：

```text
run.created
context.loading
assistant.thinking
assistant.drafting
assistant.message.completed
run.completed
```

失败剧本：

```text
run.created
context.loading
assistant.thinking
run.failed
```

每次 run 生成独立 `runId` 和 event id，避免测试和截图出现随机顺序。成功剧本只追加一次 assistant message；失败或停止不会追加伪成功回复。

## 三个前端表面的同源状态

Chat Canvas、Run Timeline 和 Agent 状态徽章都消费同一个 selected run：

- Chat Canvas 根据 run status 和 events 派生工作区状态，并显示 real API/mock、stream state 和 stop 控制。
- RunRail/RunTimeline 渲染同一组 runtime events，并显示 failed/stopped terminal semantics。
- AgentStateMotion 根据 run status/event type 映射到 thinking/speaking/done/error。

这避免三个区域分别维护“装饰状态”，导致一个显示成功、另一个显示失败。

## Stale event guard

异步 runtime event 必须同时满足：

```text
requestedThreadId === currentSelectedThreadId
runId === activeRunId
```

否则事件会被忽略。这个规则保护“用户快速切换 thread”场景，旧 thread 的 mock events 不会污染当前 Chat Canvas、Timeline 或 Agent badge。M4 SSE path 使用 `selectedThreadId + currentRunId` 做同类保护，并按 event id/sequence 去重。

## 真实后端边界

当 `VITE_LOOMI_API_BASE_URL` 配置后，thread/message 和 M4 run/event/SSE 都走 real API：

- `sendMessage` 先创建 message，再通过 `POST /v1/threads/{thread_id}/runs` 启动 local simulated run。
- `subscribeRunEvents` 连接 `GET /v1/runs/{run_id}/events/stream?after_sequence=...`。
- `stopRun` 调用 `POST /v1/runs/{run_id}/stop`。
- RunRail 的 mock scenario selector 只在 mock mode 显示。

这是一条安全和产品诚实边界：真实 M4 能力存在时使用真实 API；未来 LLM/tool/worker 能力不存在时不能假装执行，也不能静默 fallback 到 mock。

## 当前非目标

M3.5/M4 当前仍不实现：

- LLM gateway
- tool execution
- worker/job queue
- desktop runtime/plugin/channel
- 多 agent 协作

这些能力要等 M5/M6+ 的后端和执行上下文稳定后再接入同一 adapter/state model。
