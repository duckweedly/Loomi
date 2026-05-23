---
title: M3.5 前端 Agent Runtime 骨架
description: 解释 Chat Canvas 状态机、runtime adapter 边界、mock run/event 剧本和真实后端能力未接入状态。
---

M3.5 的目标不是提前实现真实 Agent 后端，而是在 M4 run/event/SSE 和 M5 LLM gateway 完成前，让前端拥有可验证的 Agent 交互骨架。

这个骨架有三个边界：

1. `web/src/runtime/`：前端 runtime 语义边界。
2. `web/src/state.ts`：当前选中 thread、messages、run 和 stale event guard 的协调层。
3. `web/src/components/`：Chat Canvas、Run Timeline、Agent 状态徽章的同源渲染层。

## 为什么要单独建 runtime 边界

M3 的 `ApiClient` 负责 durable thread/message 数据。run/event/SSE 属于 M4，LLM/tool 执行属于 M5+。如果现在把未来 runtime 行为直接塞进 M3 API client，mock 和 real 很容易各走一套 UI 逻辑。

因此 M3.5 新增 `ExecutionAdapter`：

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

mock adapter 报告 `available`，并播放 deterministic scripts。real adapter 在 M4/M5 之前报告 `unavailable`，且不会隐藏地 fallback 到 mock 执行。

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

- Chat Canvas 根据 run status 和 events 派生工作区状态。
- RunRail/RunTimeline 渲染同一组 runtime events，并显示 failed/stopped terminal semantics。
- AgentStateMotion 根据 run status/event type 映射到 thinking/speaking/done/error。

这避免三个区域分别维护“装饰状态”，导致一个显示成功、另一个显示失败。

## Stale event guard

异步 runtime event 必须同时满足：

```text
requestedThreadId === currentSelectedThreadId
runId === activeRunId
```

否则事件会被忽略。这个规则保护“用户快速切换 thread”场景，旧 thread 的 mock events 不会污染当前 Chat Canvas、Timeline 或 Agent badge。

## 真实后端边界

当 `VITE_LOOMI_API_BASE_URL` 配置后，thread/message 仍走 real API；runtime execution 则通过 real execution adapter 明确报告 `backend-unavailable`。

这是一条安全和产品诚实边界：真实后端没有 run/event 能力时，前端不能假装执行，也不能静默 fallback 到 mock。

## 当前非目标

M3.5 不实现：

- persisted run/event table
- SSE stream
- LLM gateway
- tool execution
- worker/job queue
- desktop runtime/plugin/channel
- 多 agent 协作

这些能力要等 M4/M5/M6 的后端和执行上下文稳定后再接入同一 adapter/state model。
