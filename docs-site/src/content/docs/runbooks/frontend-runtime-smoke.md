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

## P0-2 real desktop regression projection

前端 render regression 覆盖真实 smoke 的 UI 投影，不再只依赖截图：

- completed real API run 如果没有任何 assistant final 内容，Chat Canvas 会显示 `Final assistant message missing` / `最终回复缺失`，不能静默展示成完成。
- failed run 的 RunRail 仍保留工具历史，包括成功和失败的 workspace tool rows。
- persisted final assistant message 的 Markdown table、inline code、fenced code block 在当前轮就渲染，不依赖刷新。

对应测试：

```bash
bun test --cwd web src/components/ChatCanvas.states.test.ts src/components/RunRail.runtime.test.ts --test-name-pattern "flags a completed real API run that has no final assistant content|keeps failed run tool history visible|renders real smoke final markdown"
```

## Mock success smoke

1. 不设置 `VITE_LOOMI_API_BASE_URL`。
2. 启动 web renderer。
3. 进入默认会话视图；侧边栏不显示 Chat/Work 双模式切换。
4. 选择或创建一个会话。
5. 输入一条消息并提交。
6. 验证用户消息立即出现在 Chat Canvas。
7. 验证 pending assistant bubble 出现在 Chat Canvas 中；assistant 内容为空时只显示短的 run-scoped thinking hint，例如“组织回复”或“梳理线索”，不会显示“模型正在生成回复”这类长句。
8. 验证 Run Rail 在 draft 仍为空时也显示同类 thinking line，并随着 elapsed seconds 轻量更新；文案使用 text shimmer，不出现点状 loading。
9. 验证 pending draft 是内联透明状态，不出现卡片套卡片；如果已有 streaming markdown fragments，Chat Canvas 等最终内容完成后再渲染，避免半截 Markdown 标题露出 `#`。
10. 验证 Run Timeline 按组显示事件：
   - Run lifecycle: `run.created`, terminal run events
   - Model stream: `assistant.drafting`, `model.delta`, `model.final`, `model.usage`
   - Worker/job: `job.queued`, `worker.claimed`, `job.retrying` when using richer mock scripts
   - Error: provider/model/stream/backend failures
11. 验证 token usage/provider metadata 只出现在 timeline/debug detail，不进入 assistant message text。
12. 验证 safe thought summary 只显示折叠式一行摘要；raw/hidden thinking 不出现在页面、event detail 或 message text。
13. 验证 Agent 状态徽章进入 speaking/thinking，并最终进入 done。
14. 验证 Chat Canvas 只出现一条最终 assistant 回复，不重复显示 final draft 和 persisted assistant message。
15. 验证 Regenerate 可见，触发后旧 assistant 回复仍保留，并出现一个关联旧消息的新 pending attempt。
16. 验证 Continue 只在有选中 thread、输入非空且无 active run 时可用；active run 期间 Send/Continue/Retry/Regenerate 都被阻止。

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
2. 选择空会话，验证 empty state 与 loading state 不同。
3. 模拟 thread/message loading，验证 ThreadSidebar 和 Chat Canvas 显示 loading affordance。
4. 模拟加载失败，验证 error 与 Retry 可见，并且 selected thread context 不被清空。
5. 选择带 persisted assistant message 和 run events 的 thread，验证 Chat Canvas 与 Run Timeline 对 latest run outcome 一致。

## Stale event smoke

1. 在会话 A 启动 run。
2. run 尚未完成前切换到会话 B。
3. 验证会话 A 的后续 event 不会改变会话 B 的 Chat Canvas、Timeline 或 Agent 状态徽章。

Mock scripts 通过 deterministic subscription path 分步推进；浏览器里也可以用 M4 real API SSE 验证更完整的可见切换过程。单元测试覆盖 stale event guard 和 out-of-order delta protection。

## M92 real task smoke

用真实后端和 Work mode 依次跑这些提示，每个 run 都要有工具事件、最终 assistant message，且 final 内容不是 `[redacted]`：

1. `帮我看当前目录有什么，分类梳理`
2. `读取 README，告诉我这个项目是什么`
3. `搜索 package.json 或 go.mod 并总结技术栈`
4. `对一个测试 fixture 做 patch_preview，不自动 apply`
5. `访问一个公开网页并总结 title/excerpt`

验收点：

- 目录类任务先出现 workspace listing/glob 类工具事件。
- README/技术栈任务出现 grep/read 工具事件。
- 修改类任务先 read，再 patch preview；没有用户批准前不 apply。
- web 任务出现 `web.fetch` 成功事件。
- 失败时 RunRail/Chat 能显示 provider、validation、permission、workspace、timeout/bounded limit 中的具体原因。
- run terminal 后，迟到的 model/tool event 不改变已经完成的消息或下一轮状态。
- persisted assistant message 的 Markdown 在当前轮就正常渲染，不依赖下一轮刷新。

CLI closeout should use the same real API/provider/workspace path documented in `local-m79-agent-harness-smoke`. For desktop/web smoke, run the renderer with `VITE_LOOMI_API_BASE_URL` pointing at the local API, select the same workspace folder in the Composer, then compare the browser RunRail with `loomi smoke agent --workspace ...` output:

- `tool_chain` summarizes the same workspace tools the browser shows.
- `final_message` matches the final assistant answer and is not `[redacted]`.
- A completed run without a persisted assistant message is a blocker, even if SSE emitted `run_completed`.

## P0-3 workspace reference smoke

真实 workspace 指代任务要验证当前选择目录不会被历史 thread 污染：

1. 先在一个 Work thread 选择 `Arkloop`，发起任何 workspace run。
2. 新建另一个 Work thread，选择 `Downloads`。
3. 发送 `帮我看刚选目录有什么` 或 `帮我看当前目录有什么`。
4. RunRail / ToolCallCard 只能显示安全 workspace label，例如 `Downloads` / `正在读取：Downloads`，不能显示绝对路径。
5. 工具参数使用相对 path；目录类任务第一步应是 `workspace.tree_summary` 或 `workspace.list_directory`，不能从 broad `workspace.glob` / `workspace.grep` 开始。
6. 发送 `下载目录有什么` 时，如果当前 workspace label 不是 `Downloads`，agent 必须提示选择 Downloads，不能默认读 Loomi、Arkloop、进程 cwd 或历史目录。
7. URL-only prompt 不应出现 workspace tool；普通聊天不应出现 workspace tool。

对应自动覆盖：

```bash
bun test --cwd web src/components/ToolCallCard.test.tsx src/components/Composer.test.ts src/components/RunRail.runtime.test.ts src/workModeProjection.test.ts src/components/WorkPlanView.test.tsx
```

## M91 directory classification smoke

真实目录分类任务：

```text
请看一下当前选择目录都有哪些东西，按源码/文档/配置/构建产物/临时文件分类列出。
```

期望工具链：

1. `workspace.list_directory` 或 `workspace.tree_summary`
2. 必要时 `workspace.read` 少量代表性文件
3. final answer 用自然语言列出分类

验收点：

- 不能只跑 `workspace.grep`。
- RunRail 显示“Read directory / Summarize directory”安全摘要。
- ToolCallCard 不展示 host absolute path、secret-looking filename、raw JSON dump 或未渲染协议文本。
- `node_modules`、`dist`、`build`、`.next`、`.vite`、`.venv`、`.cache`、`.git`、`target`、`vendor` 等生成/缓存目录被跳过或计入 skipped count。

## Real API M4 smoke

1. 启动本地 Go API，并确认 `/readyz` ready。
2. 设置 `VITE_LOOMI_API_BASE_URL` 指向本地 API；如果本地 API 前面有需要 bearer 的代理，同时设置 `VITE_LOOMI_API_TOKEN`，或在浏览器开发者工具中写入 `localStorage.setItem('loomi.api_token', '<token>')`。
3. 启动 web renderer。
4. 选择一个 real API thread。
5. 提交消息触发 runtime execution。
6. 验证 Chat Canvas 和 Run Timeline 显示 capability status：mock、Local simulated、Real model、Backend unavailable、Model setup missing、Provider unavailable、Stream disconnected 或 Run recovering 中的当前最高优先级状态。
7. 如果 `/v1/model-providers` 没有 available provider，验证 Chat 在发送前显示 provider unavailable 引导，并且 Open Settings / 打开设置会进入 Settings > Providers。
8. 验证 local simulated 不被描述为 real model output；backend unavailable/provider unavailable 不显示 model-thinking copy。
9. 验证网络请求命中 `/v1/threads`、`/v1/threads/{thread_id}/runs/current`、`/v1/runs/{run_id}/events/stream`。事件流使用 fetch-based SSE，因此与普通 JSON 请求一样能携带 bearer token。
9. 验证 Run Timeline 显示 persisted lifecycle/progress/message/final events。
10. 点击 running run 的 stop 控制时，验证 run 进入 stopped 或 already-terminal 语义。
11. 验证没有 mock scenario selector 出现在 real API mode。

## Desktop readiness smoke

真实桌面端启动链路按这个顺序排查：

```text
Postgres/schema -> loomi-api /readyz -> provider list/check -> local provider detection/enable -> tool catalog -> workspace root -> renderer
```

前端不应只显示浏览器原始 `Failed to fetch`。在 real API mode 下，Chat Canvas 顶部 readiness 面板要显示当前最高优先级阻断项和下一步动作：

- `Loomi API 未连接`: Retry。
- `DB/schema 未 ready`: Retry，并回到 `/readyz` 和后端日志。
- `provider 未配置`: Open Settings，必要时 Detect Local Provider。
- `Local Codex detected 但未启用`: Detect Local Provider / Enable Local Codex。
- `tool catalog 不可用`: Retry，并检查 `/v1/tools/catalog`。
- `workspace 未选择`: Choose folder。

CLI closeout 可先跑：

```bash
go run ./cmd/loomi doctor --desktop --provider "$LOOMI_PROVIDER"
```

`doctor --desktop` 需要同时报告 API/DB、provider、tool catalog 和 workspace；如果 workspace 未选择，应给出选择目录或设置 `LOOMI_WORKSPACE_ROOT` 的修复提示。

P0 real desktop closeout 需要再跑一次真实 Work-mode dogfood，不能只看 readiness：

1. 使用安全临时 workspace，不要指向项目根目录或用户真实资料目录。
2. 确认 `local_codex` 或指定 provider 是 `available`，并且 tool catalog 已返回 workspace/sandbox 工具。
3. 发送目录盘点任务；第一步必须是 `workspace.tree_summary` 或 `workspace.list_directory`，不能从 broad `workspace.grep` 起步。
4. 继续要求读取一个具体文件，先 `workspace.patch_preview`，批准后 `workspace.patch_apply`。
5. 再批准并执行 bounded validation command，例如 `go test ./...`、`bun test` 或同等安全命令。
6. 结束后记录 `thread_id`、`run_id`、provider status、workspace label、工具顺序、pending approval 状态，以及最终 assistant Markdown 是否刷新后仍显示。

如果这条链路出现 CORS、provider unavailable、workspace false-ready、工具卡片显示、RunRail 或 final Markdown 落库问题，先补一个失败测试，再修复。

## 常见问题

### 看到后端能力未接入是不是错误？

如果配置了 M4 real API 并且 readiness 通过，这表示当前 real API wiring 有问题，应优先检查 `VITE_LOOMI_API_BASE_URL`、API server 日志和 `/readyz`。未配置 real API 时，mock mode 不会显示这个状态。

### 看到 Loomi API 未连接怎么办？

前端会把浏览器原始 `Failed to fetch` 转成 `Loomi API 未连接`，并显示当前 API base。先确认 `go run ./cmd/loomi-api` 正在运行、`HTTP_ADDR` 与 `VITE_LOOMI_API_BASE_URL` 一致；local/development API 会放行 `127.0.0.1` / `localhost` 的本地 dev 端口，并允许 `Authorization` header。如果返回 401 missing bearer token，设置 `VITE_LOOMI_API_TOKEN` 或 `localStorage['loomi.api_token']` 后刷新页面。

### 为什么 mock success 会马上 completed？

测试环境使用 deterministic scripts，目的是验证顺序和状态语义。浏览器里可以通过 M4 real API local simulated run 观察更完整的 stream lifecycle；后续也可以引入可控延迟的 mock driver，但不应引入随机事件。

### 为什么 Work mode 不接 runtime？

当前 slice 只验证 Chat execution loop。Work mode 有不同业务语义，等 project/task flow 设计清楚后再复用同一 runtime 模型。
