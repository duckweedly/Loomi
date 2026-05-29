---
title: Spec Kit 工作流
description: Loomi 使用 Spec Kit 管理需求、计划、任务和实现。
---

Loomi 使用 Spec Kit 作为 AI 开发前的对齐层。它的作用不是简单生成代码，而是让每个非平凡功能都有可审查的需求、技术计划和任务拆分。

## 当前候选完成：M31 Child Agent Run Handoff

M31 补齐 Arkloop/Craft 对标中的下一块 agent 能力差距：已有 agent task 可以通过 approval-gated `agent.delegate` 创建独立 child thread 和 queued child model-gateway run。

关键产物：

- Data：`agent_tasks` 增加 `child_thread_id`、`child_run_id` 和 `delegated_at`。
- Runtime：`agent.delegate` 走现有 ToolBroker/approval/worker continuation，不绕过审批。
- Productdata：child run 复用现有 `StartRun` 背景 job pipeline，不新增外部队列。
- Projection：HTTP/CLI agent task list 返回 optional child ids。
- Safety：终态 task、跨线程 task、重复 delegate 均拒绝；父 run continuation 只得到 safe child ids，不得到 child run raw messages 或工具日志。

状态：candidate。该轮不是 worker pool、swarm scheduler、远端 guest agent、Docker/Firecracker sandbox、OS process 或 Redis queue rewrite。

## 当前候选完成：M32 Context Source Registry

M32 补齐 Arkloop/Craft 对标中的 context/source 边界底座：先把“这个线程有哪些可用来源”持久化登记下来，再在后续 slice 接入 connector 执行、同步或 RunContext source selection。

关键产物：

- Data：新增 `context_sources`，记录 thread-scoped source id、kind、safe title、normalized locator、summary、status 和 redacted metadata。
- API：新增 `POST /v1/threads/{thread_id}/sources` 和 `GET /v1/threads/{thread_id}/sources`。
- Safety：URL 去 query/fragment，拒绝 localhost/private host/URL credentials；workspace path 只允许相对路径并拒绝 `.env*`、`.git`、private key、`secrets` 和 `credentials`。
- Boundary：不抓取、不同步、不调用 GitHub、不做 OAuth、不做 marketplace、不改 UI。

状态：candidate。该轮只是安全来源注册层，不是 connector runtime。

## 当前候选完成：M89 Unified Conversation Entry

M89 学习 Arkloop 的分层方向，但保留 Loomi 自己的产品表达：用户只看到一个会话入口，深层能力按目录、run metadata 和工具状态浮现。

关键产物：

- UI：Sidebar 不再显示 Chat/Work 双模式切换；线程列表合并显示，创建入口统一为“新会话”。
- Composer：目录选择和目录状态不再只挂在 Work mode 上，用户可以从同一个输入框进入目录相关任务。
- Projection：计划/todo/artifact 投影由安全 run metadata 决定，不再只由 legacy `Thread.mode = work` 决定。
- Tests：更新 sidebar、composer、projection 和 shell style 断言，防止模式切换入口回流。

状态：candidate。该轮只合并前端入口和投影条件；保留后端 `Thread.mode`、现有工具权限和 provider/run 链路，不新增工具、不改数据库、不引入多 agent。

## 近期候选完成：M88 Chat Response Polish Followups

M88 继续沿着 Arkloop 的可感知生成体验做最小 follow-up，但不复制 Arkloop 文案、架构或工具面。

关键产物：

- UI：Run Rail 在 assistant draft 为空时显示 run-level thinking line；文案使用 shimmer 和短暂渐进更新，避免点状 loading 和卡片套卡片。
- Safety：完成后的 thought summary 只来自 allowlisted metadata；frontend event 会剥离 raw/hidden thinking 字段。
- Runtime：provider continuation 前会压缩超大 redacted tool result 字符串，保留 path/status/error 等信号和 compaction marker，小结果不变。
- Persona：默认 Loomi prompt 明确 result-first、brief final answer 和不暴露 hidden chain-of-thought。

状态：candidate。该轮只做真实使用 polish；不新增工具，不改 provider 路由，不引入外部 runtime，不做 UI 内容截断。

## 近期候选完成：M87 Chat Response Polish

当前实现对齐 Arkloop 的方向，但保留 Loomi 自己的命名、文案、视觉和安全边界。

关键产物：

- UI：Chat Canvas 在 assistant 内容为空时显示 run-scoped short thinking hint；同一 `run_id` 在浏览器本地保持同一句，SSR 使用稳定 hash。
- Markdown：streaming draft 不渲染半截 Markdown，最终 heading 渲染不显示可见 `#`。
- Persona：Gateway prompt policy 增加精炼输出规则，约束最终回复先给结论、无开场白、不复述用户请求、代码改动只报变更和验证。
- Docs：runbook/devlog/spec-kit 记录本轮 UX 行为和验证命令。

状态：candidate。该轮只做真实使用体验 polish；不新增工具，不改 provider 路由，不做 UI 截断，不复制 Arkloop 文案。

## 当前候选完成：M79 Agent Harness Smoke

当前实现沿用 M76 continuation reliability、M78 sandbox process foundation 和 M6 worker/job path，不新增 Docker、Redis 或外部服务。

关键产物：

- CLI：`loomi smoke agent` 执行真实 API/provider/run stream，打印 `thread_id`、`run_id`、final stage、provider `check_stage` 和最后事件摘要。
- Diagnostics：`loomi doctor --provider <id>` 和 smoke 对 401/403、429、503 输出可行动原因。
- Runtime：provider completion check 使用语义 `check_code`，不会把 upstream body 或 token 写入输出。
- Docs：runbook/devlog/spec-kit 记录真实 smoke 的 env、blocked 边界和验证命令。

状态：candidate。该轮只做 harness smoke closeout；不新增 Docker/Redis，不改 web UI，不复制 Arkloop 命名或文案。

## 当前候选：M81 Sandbox Process Lifecycle Recovery

当前实现继续沿用 M78 sandbox process foundation，只补 Arkloop 对标里的下一块最小机制差距，不新增 Docker/Firecracker、guest agent 或 shell service。

关键产物：

- Runtime：process stdout buffer 改为 bounded latest-tail，同时 `next_cursor` 使用绝对捕获字节 cursor，长输出继续读取不会重复、丢失新尾部或无限增长。
- Lifecycle：exited/terminated process 的 `continue_process` 只返回终态摘要和安全状态，不再尝试写 stdin 或执行新动作。
- Safety：run-scoped process ownership、argv-only、approval required、workspace root/path/secret redaction 继续保留。
- Tests：覆盖长输出 cursor、exit 后 terminal summary、terminate 后 state-only continue、跨 run 拒绝和 output redaction。

状态：candidate。该轮只补进程输出/生命周期/恢复边界；不做 Firecracker/Docker、guest agent、PTY/shell/resize、Redis、artifact sync、provider/gateway continuation 改动或新 sandbox service。

## 近期候选完成：M80 Durable Run Resume

当前实现沿用 M76 tool continuation reliability 和 M6 worker/job durable event path，不新增字段、队列或外部状态层。

关键产物：

- Runtime：worker retry 遇到已 `succeeded` 的 approved tool call 时，按 durable run events 判断是否需要补发 provider continuation。
- Guard：只有该 tool 的 `tool_call_succeeded` 之后没有 continuation start、后续 tool request 或 terminal event，才恢复 continuation。
- Tests：覆盖 pending tool call 不进入 continuation、worker restart 后 approval/executing/succeeded 顺序保留、M76 多工具 final assistant 单写，以及 terminal run late-write guard。
- Docs：architecture/API/devlog/spec-kit 状态记录 Arkloop 对标点和 Loomi 实现边界。

状态：candidate。该轮只修 durable resume 语义；不做 interrupted status 字段、batch API、Redis、外部 queue、terminal shell runtime 或多 agent 编排。

## 当前候选：M78 Sandbox Process Foundation

当前实现沿用 `specs/032-sandbox-exec-command-tools/` 的 sandbox exec 基础，不新增 Docker/Firecracker 或独立 sandbox service。

关键产物：

- Runtime：`sandbox.start_process` / `sandbox.continue_process` / `sandbox.terminate_process` 走 Work-mode approval-gated ToolBroker path，使用本地内存 run-scoped process registry。
- Safety：argv-only、workspace cwd、allowlist、timeout/output bounds、secret/path redaction、terminal run/denied/unapproved/cross-run guard。
- UI：RunRail / ToolCallCard 显示 process lifecycle 安全摘要。
- Docs：API/architecture/runbook/devlog 记录 Arkloop 对标和明确非目标。

状态：candidate。该轮只做本地受控进程 foundation；不做 Arkloop 的 Firecracker/Docker 隔离、guest agent、sandbox template、shell/PTY/resize、Redis 或 artifact sync。

## 近期候选完成：M77 Long Run Recovery

当前实现沿用 M4 run/event/SSE、M6 worker event persistence、M71 CLI control-plane 和 real API frontend state，不改 provider、tool execution 或 workspace mutation 语义。

关键产物：

- CLI：`runs attach` 先读 run projection，再 replay persisted events，再按最后 replay sequence 继续 live SSE；terminal run replay 后退出。
- CLI：`runs follow` 默认只从当前 last sequence 之后 tail，不回放历史。
- API：events history 和 SSE 都保持 `sequence > after_sequence` exclusive cursor，不重复边界事件。
- UI：real API replay + live merge 按 id/sequence 去重，避免 duplicate assistant delta 和 tool lifecycle event。
- Docs：architecture/runbook/devlog 记录 Arkloop 对标后的 Loomi M77 边界。

状态：candidate。该轮只补长 run 断线恢复和 attach/follow cursor；不新增 worker ownership 语义，不改 provider/tool/workspace mutation，不引入 Redis 或多实例 stream fanout。

## 近期候选完成：M76 Tool Continuation Reliability

当前实现沿用 M75 Work-mode 工具闭环，不新增工具、batch API 或多 agent。

关键产物：

- HTTP smoke：模拟 provider 连续请求 `workspace.grep`、`workspace.read`、`workspace.patch_preview`、`workspace.patch_apply`、`sandbox.exec_command`、`workspace.read`，并验证 final assistant message。
- Runtime：Gateway continuation request 按 run events 重建已成功 tool result 的有序前缀，覆盖 5-8 次工具调用里的顺序、去重和终止边界。
- Docs：architecture/runbook/devlog 记录 Arkloop 对标观察和 Loomi 本 slice 取舍。

状态：candidate。该轮只做 tool continuation reliability；不新增工具，不引入 batch API，不引入多 agent，不改 provider HTTP 实现。

## 近期候选：M75 Code-Agent Daily Loop

当前实现沿用 Work-mode 工具闭环，不新增工具或 provider 路由。

关键产物：

- HTTP smoke：模拟 provider 连续请求 `workspace.grep`、`workspace.read`、`workspace.patch_preview`、`workspace.patch_apply`、`sandbox.exec_command`，并验证 approval、事件、continuation 和最终 assistant message。
- UI smoke：RunRail / ToolCallCard 可读地展示 patch preview/apply 和 sandbox validation 步骤，同时隐藏 raw path/content/diff payload。
- Docs：新增 architecture/runbook/devlog，说明本地复现和 done 证据。

状态：candidate。该轮只打通 dogfoodable daily loop；不处理 provider 503、CLI thread title、新工具、多 agent、Docker 或 Firecracker。

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

## 当前候选：Memory Provider Error UI

## 近期已完成：M82 Real Local Usability Closeout

状态：M82 closeout。CLI 默认 API host 对齐本地 Loomi API 的 `127.0.0.1:18080`，避免误连本机其他 8080 服务导致 `401 missing bearer token`。CLI config/env 支持 `LOOMI_API_TOKEN` / `api_token` bearer token，但输出只显示 `api_token_set`。`loomi doctor --provider local_codex` 现在区分 detected-but-not-enabled blocked reason 和 enabled ready state。真实本机 smoke 已通过 `loomi smoke agent --auto-approve --prompt "Read AGENTS.md with workspace.read, then reply M82 smoke complete."`，产出 `thread_id=thr_1779861294575417000_71c96fe2b8eb`、`run_id=run_1779861294596954000_78e89c7fc75e`、`final_stage=run_completed`，并记录 `workspace.read AGENTS.md` request/approved/executing/succeeded 事件。

本轮没有新增工具、Docker/Redis/Firecracker、多 agent 架构，也没有把 token 写入日志或文档。

当前 Spec Kit 功能目录：

```text
specs/070-memory-provider-error-ui/
```

关键产物：

- `spec.md`：定义 Settings > Memory recent errors 展示 runtime run/event 线索的目标、安全边界和验收标准。
- `plan.md`：确定复用现有 recent errors panel 和前端 error formatter。
- `tasks.md`：按 runtime/docs/validation 拆分。

状态：candidate。该轮只补 runtime error display；不做 run detail navigation、modal redesign 或 raw log viewer。

## 追加候选完成：M72 Provider 503 Diagnosis

当前实现目录仍沿用本轮工作上下文，不新增 provider router/fallback spec。

关键产物：

- runtime：新增 completion smoke 诊断状态 `configured`、`reachable`、`completion-ok`、`completion-failed`，并归类 `completion-failed-503`。
- API：`GET /v1/model-providers` 不外呼 provider；`POST /v1/model-providers/check` 才执行安全 completion smoke。
- CLI/UI：doctor 和 Settings > Providers 都显示 `completion-failed-503`，不把配置完整误报成 completion 可用。

状态：candidate。该轮只诊断 provider 503；不修 M71 thread title bug，不新增 M81 provider router/fallback。

## 近期候选完成：Memory Provider Runtime Errors

当前 Spec Kit 功能目录：

```text
specs/069-memory-provider-runtime-errors/
```

关键产物：

- `spec.md`：定义 Settings > Memory recent errors 纳入 runtime provider failure 的目标、安全边界和验收标准。
- `plan.md`：确定复用 Gateway error event、provider error read model 和 `/v1/memory/errors`。
- `tasks.md`：按 runtime/docs/validation 拆分。

状态：candidate。该轮只补 safe runtime provider errors；不做 retry scheduler、provider process restart、raw provider log viewer 或 Settings UI redesign。

## 近期候选完成：Memory Nowledge Prompt Snapshot Regression

当前 Spec Kit 功能目录：

```text
specs/068-memory-nowledge-prompt-snapshot/
```

关键产物：

- `spec.md`：定义 Nowledge prompt recall 与 OpenViking 同等可验证的目标、安全边界和验收标准。
- `plan.md`：确定复用 Gateway prompt enrichment、Nowledge adapter 和 safe run event metadata。
- `tasks.md`：按 runtime/docs/validation 拆分。

状态：candidate。该轮只补 Nowledge-specific regression；不做 UI redesign、provider endpoint redesign、background snapshot cache 或 automatic provider install。

## 近期候选完成：Memory External Snapshot Event

当前 Spec Kit 功能目录：

```text
specs/067-memory-external-snapshot-event/
```

关键产物：

- `spec.md`：定义 external provider prompt recall 的 timeline 可观测性目标、安全边界和验收标准。
- `plan.md`：确定复用 Gateway prompt enrichment 和 safe run event metadata。
- `tasks.md`：按 runtime/docs/validation 拆分。

状态：candidate。该轮只补 `memory_external_snapshot_loaded` progress event；不做 UI redesign、memory audit history、background snapshot cache 或 raw provider payload event。

## 近期候选完成：Memory External Prompt Snapshot

当前 Spec Kit 功能目录：

```text
specs/066-memory-external-prompt-snapshot/
```

关键产物：

- `spec.md`：定义外部 provider 在初始模型请求前贡献 safe memory prompt 的目标、安全边界和验收标准。
- `plan.md`：确定复用 Gateway、`MemoryToolExecutor.externalMemorySearch` 和现有 `<memory>` prompt block。
- `tasks.md`：按 runtime/docs/validation 拆分。

状态：candidate。该轮只补 external provider recall before initial provider request；不做 background snapshot cache、LLM distill、recursive OpenViking tree snapshot 或 raw provider payload prompt injection。

## 近期候选完成：Memory OpenViking Connections

当前 Spec Kit 功能目录：

```text
specs/065-memory-openviking-connections/
```

关键产物：

- `spec.md`：定义 OpenViking `memory.connections` 的用户目标、安全边界和验收标准。
- `plan.md`：确定复用 `MemoryToolExecutor.connections`、OpenViking `/api/v1/fs/ls` 和 safe tool result projection。
- `tasks.md`：按 runtime/docs/validation 拆分。

状态：candidate。该轮只补 `viking://...` direct child connections；不做递归树遍历、raw provider payload、OpenViking install/start/restart bridge 或 provider write/delete 语义变更。

## 近期候选完成：Memory OpenViking Detect

当前 Spec Kit 功能目录：

```text
specs/064-memory-openviking-detect/
```

关键产物：

- `spec.md`：定义 OpenViking 本地实例检测的用户目标、安全边界和验收标准。
- `plan.md`：确定复用 Nowledge detect 模式、Memory provider modal、real/mock API client 和 docs-site。
- `tasks.md`：按 httpapi/frontend/docs/validation 拆分。

状态：candidate。该轮只补 OpenViking localhost detect convenience；不做 install/start/restart bridge、key discovery、remote scan、provider auto-enable 或外部 provider adapter 语义变更。

## 近期候选完成：Memory Proposal Edit

当前 Spec Kit 功能目录：

```text
specs/046-memory-proposal-edit/
```

关键产物：

- `spec.md`：定义 pending proposal 审批前编辑标题/摘要的闭环。
- `plan.md`：确定复用 `memory_write_proposals`、PATCH API、workspace state 和 MemoryPanel。
- `tasks.md`：按 productdata/httpapi/frontend/docs/validation 拆分。

状态：candidate。该轮只补待审批记忆的编辑闭环；不做 bulk edit、已审批内容编辑、LLM distillation、embedding/vector search 或外部 semantic provider 执行。

## 近期候选完成：Memory Proposal Review

当前 Spec Kit 功能目录：

```text
specs/045-memory-proposal-review/
```

关键产物：

- `spec.md`：定义 pending proposal 安全列表和 Settings > Memory 审批/拒绝闭环。
- `plan.md`：确定复用 `memory_write_proposals`、现有 approve/deny API、workspace state 和 MemoryPanel。
- `tasks.md`：按 productdata/httpapi/frontend/docs/validation 拆分。

状态：candidate。该轮只补待审批记忆的可见和决策闭环；不做 bulk approve、LLM distillation、embedding/vector search 或外部 semantic provider 执行。

## 近期候选完成：Memory Post-run Proposals

当前 Spec Kit 功能目录：

```text
specs/044-memory-post-run-proposals/
```

关键产物：

- `spec.md`：定义 `commit_after_run` 打开后 completed run 生成待审批记忆提案，不自动批准。
- `plan.md`：确定复用 runtime closeout、productdata provider status、assistant message、write-proposal/audit。
- `tasks.md`：按 runtime helper、runner/gateway wiring、Settings 文案、docs、validation 拆分。

状态：candidate。该轮只把每轮后自动整理接成 pending proposal；不做 LLM distillation、embedding/vector search、外部 semantic provider 执行或 background memory worker。

## 近期候选完成：Memory Agent Tools

当前 Spec Kit 功能目录：

```text
specs/043-memory-agent-tools/
```

关键产物：

- `spec.md`：定义 `memory.search`、`memory.read`、`memory.write`、`memory.forget`、`memory.status` 的 Work-mode、approval-gated、safe-summary-only 行为。
- `plan.md`：确定复用 ToolCatalog、provider schema、ToolBroker、worker continuation、productdata memory service、Settings > Tools。
- `tasks.md`：按 productdata/runtime/provider/worker/web/docs/validation 拆分。

状态：candidate。该轮补 agent 可调用的 memory tools；不做自动 distill、embedding/vector search、外部 semantic provider 执行或 background memory worker。

## 近期候选：Web Search Providers

当前 Spec Kit 功能目录：

```text
specs/041-web-search-providers/
```

关键产物：

- `spec.md`：定义 Chat/Work 可用的 read-only auto-approved `web.search`，支持 Tavily 和 Brave Search provider。
- `plan.md`：确定复用 M26 WebToolExecutor、ToolBroker、worker continuation、RunRail；网页搜索配置放在 Settings > Web Search，Settings > Tools 保持只读 catalog，不新增 crawler/browser/social 搜索。
- `tasks.md`：按 productdata/runtime/provider/worker/web/docs 验证拆分。

状态：candidate。该轮只补真实 web search provider 能力；`web.search` 与 `web.fetch` 都可在 Chat/Work 中按 persona allowlist 作为 read-only auto-approved public web 工具使用，搜索结果只返回 bounded title/url/snippet safe summary。

## M71 CLI Real Run Closeout

当前关联：CLI control-plane closeout，支撑真实 Chat/Work run dogfood。

状态：complete candidate。`loomi run` 无 `--thread` 时会先创建带 mode 和短标题的真实 thread，再写入 user message 并启动 run；`loomi doctor` 对默认 `local_codex` 未注册给出明确 operator 提示但不写配置文件。Provider 503 根因保留为 M72 blocker。

## 近期候选完成：UI-02 Real Usage Readiness

当前 Spec Kit 功能目录：

```text
specs/040-real-usage-readiness/
```

关键产物：

- `spec.md`：定义 Mock/Real 状态前置、Work folder 诚实受限态、WorkPlanView 任务面板、RunRail/ToolCallCard 人话工具标签、approval blocked 前置、Sidebar 精简入口和 Composer 模式文案。
- `plan.md`：确定只复用现有 React/CSS 和已存在运行状态，不改 backend/runtime/provider/tool execution/database，不推进 M38。
- `tasks.md`：按状态可见性、Work mode 真实态、工具事件可读性、Sidebar/Composer 基础使用和验证拆分。

状态：UI-02 candidate。该轮只做真实使用态收口，不新增工具能力，不改变执行链路，不做像素级美化；RunRail/ToolCallCard 已统一使用安全预览，Chat mode 不显示 Work folder 限制态。

## 近期候选完成：UI-01 Formal Interface Shell Redesign

当前 Spec Kit 功能目录：

```text
specs/039-formal-interface-shell-redesign/
```

关键产物：

- `spec.md`：定义浅色桌面壳、窄 sidebar、白色聊天画布、固定 composer、Chat/Work 保持可用，以及非目标。
- `plan.md`：确定复用现有 React/CSS 结构，不改 backend/runtime/tool/provider/memory/database。
- `tasks.md`：按整体壳、视觉 token、左侧栏、主内容布局、composer、Settings/Tools/RunRail compatibility 和 browser smoke 拆分。

状态：第一轮 UI shell redesign candidate。该轮只让界面可正式打开测试，不包含像素级精修、Settings redesign、Tools redesign、RunRail redesign 或 M38/activity recorder。

## 近期候选完成：M29 Multi-agent Runtime Foundation

当前 Spec Kit 功能目录：

```text
specs/037-multi-agent-runtime-foundation/
```

关键产物：

- `spec.md`：定义 `agent.spawn`、`agent.list`、`agent.start`、`agent.complete` 和 `agent.fail` 的 approval-gated、Work-mode-only、coordination-only agent task runtime 和非目标。
- `plan.md`：确定复用 ToolCatalog/RunContext/ToolBroker/worker continuation、Settings/RunRail。
- `research.md`：记录先做 task record coordination、不做 autonomous child runs/cross-thread delegation 的决策。
- `data-model.md`：定义 Agent Task lifecycle、tool arguments 和 result summary。
- `contracts/`：定义 catalog、arguments、result 和 rejection contract。
- `quickstart.md`：记录 focused/full validation 和 manual smoke。
- `tasks.md`：按 catalog/runtime/spawn/list-start-complete-fail/safety/UI/docs/validation 拆分。

状态：M29/M30 PG-backed candidate。Agent tools 仅 Work mode 启用，always approval required，经 ToolBroker/worker continuation 创建、列出、启动、完成和失败 bounded coordination task records；真实 API 路径使用 PostgreSQL `agent_tasks`，in-memory 和 PG 均覆盖 spawned/in_progress/completed/failed lifecycle 与 cross-thread no-leak。Settings > Tools 与 RunRail 显示 agent scope、medium risk、coordination-only、no autonomous execution。当前 multi-agent runtime 不包含 autonomous child model runs、cross-thread delegation、external worker pools、process spawning、filesystem access、network calls、shell execution、long-term multi-agent memory、marketplace packaging 或 background swarm orchestration。

## 近期已完成：M28 Artifact Runtime Foundation

当前 Spec Kit 功能目录：

```text
specs/036-artifact-runtime-foundation/
```

关键产物：

- `spec.md`：定义 `artifact.create_text`、`artifact.create_visual`、`artifact.read` 和 `artifact.list` 的 approval-gated、Work-mode-only artifact runtime 和非目标。
- `plan.md`：确定复用 ToolCatalog/RunContext/ToolBroker/worker continuation、WorkPlan artifact projection、Settings/RunRail。
- `research.md`：记录先做 bounded artifact storage、不做下载/文件系统导出/浏览器集成的决策。
- `data-model.md`：定义 Artifact、tool arguments 和 result summary。
- `contracts/`：定义 catalog、arguments、result 和 rejection contract。
- `quickstart.md`：记录 focused/full validation 和 manual smoke。
- `tasks.md`：按 catalog/runtime/read-list/safety/UI/docs/validation 拆分。

状态：M28 PG-backed candidate。Artifact tools 仅 Work mode 启用，always approval required，经 ToolBroker/worker continuation 创建/读取/列出 bounded UTF-8 text artifacts，并支持 `artifact.create_visual` 创建 bounded SVG/HTML visual artifacts。真实 API 路径使用 PostgreSQL `artifacts`，in-memory 和 PG 均覆盖 create/read/list 与 cross-thread no-leak。Settings > Tools 与 RunRail 显示 artifact scope 和 medium risk；文本 artifact 仍 non-executable，视觉 artifact 仅在 sandboxed Preview frame 中渲染。M28 不包含 binary artifacts、downloads、filesystem export、browser integration、shell integration、artifact version graph、marketplace packaging 或 multi-agent orchestration。

## 近期已完成：M27 Browser Automation Foundation

当前 Spec Kit 功能目录：

```text
specs/035-browser-automation-foundation/
```

关键产物：

- `spec.md`：定义 `browser.open`、`browser.snapshot` 和 `browser.click_link` 的 public HTTP(S)-only、approval-gated、Work-mode-only、run-scoped browser session 和非目标。
- `plan.md`：确定复用 ToolCatalog/RunContext/ToolBroker/worker continuation、Settings/RunRail 和 Go stdlib HTTP。
- `research.md`：记录先做 HTTP-backed browser session、拒绝 private/local network、只持久化 safe summaries 的决策。
- `data-model.md`：定义 Browser Tool、Browser Session、page snapshot 和 link target。
- `contracts/`：定义 catalog、arguments、result 和 rejection contract。
- `quickstart.md`：记录 focused/full validation 和 manual smoke。
- `tasks.md`：按 catalog/runtime/click_link/safety/UI/docs/validation 拆分。

状态：M27 candidate。Browser tools 仅 Work mode 启用，always approval required，经 ToolBroker/worker continuation 执行 public HTTP(S)-only 的 run-scoped browser session；生产默认拒绝 credentialed URL、非 HTTP(S)、localhost、loopback、private/link-local/multicast/unspecified host、blocked redirect 和 unsafe link target；Settings > Tools 与 RunRail 显示 browser scope、medium risk、public HTTP only。浏览器 smoke 需在本轮 closeout 单独记录。M27 不包含 Chrome profile、cookies、JavaScript rendering、forms、screenshots、downloads、authenticated browsing、crawler、artifact runtime、activity recorder 或多 agent 编排。

## 近期已完成：M26 Web Fetch Tool Foundation

当前 Spec Kit 功能目录：

```text
specs/034-web-fetch-tool-foundation/
```

关键产物：

- `spec.md`：定义 `web.fetch` 的 public HTTP(S)-only、Chat/Work persona-gated、auto-approved bounded network read 和非目标。
- `plan.md`：确定复用 ToolCatalog/RunContext/ToolBroker/worker continuation、Settings/RunRail 和 Go stdlib HTTP。
- `research.md`：记录先做 fetch-only、拒绝 private/local network、只持久化 summaries 的决策。
- `data-model.md`：定义 Web Fetch Tool、request summary 和 result summary。
- `contracts/`：定义 catalog、arguments、result 和 rejection contract。
- `quickstart.md`：记录 focused/full validation 和 manual smoke。
- `tasks.md`：按 catalog/runtime/safety/UI/docs/validation 拆分。

状态：M26 candidate。`web.fetch` 可在 Chat/Work 中按 persona allowlist 启用，作为 read-only auto-approved public web 工具，经 ToolBroker/worker continuation 执行一次 bounded public HTTP(S) 读取；生产默认拒绝 credentialed URL、非 HTTP(S)、localhost、loopback、private/link-local/multicast/unspecified host 和 blocked redirect；Settings > Tools 与 RunRail 显示 web scope、medium risk、public HTTP only。浏览器 smoke 需在本轮 closeout 单独记录。M26 不包含 browser automation、JavaScript rendering、cookies、authenticated fetch、crawler、artifact runtime、activity recorder 或多 agent 编排。

## 近期已完成：M25 MCP Management + LSP Read-only Foundation

当前 Spec Kit 功能目录：

```text
specs/033-mcp-management-lsp-readonly/
```

关键产物：

- `spec.md`：定义 Settings MCP read-only status、LSP read-only tools、approval-gated execution、safety visibility 和非目标。
- `plan.md`：确定复用 MCP config/discovery、ToolCatalog/RunContext/ToolBroker/worker continuation 和 Settings/RunRail。
- `tasks.md`：按 MCP status、LSP runtime、安全可见性、docs/validation 拆分。

状态：M25 candidate。`GET /v1/mcp/servers` 和 Settings > MCP 提供 local stdio MCP server safe status；`lsp.diagnostics`、`lsp.symbols`、`lsp.references` 仅 Work mode 启用，approval-gated、workspace-scoped、bounded、read-only，并经 ToolBroker/worker continuation 执行。浏览器 smoke 需在本轮 closeout 单独记录。M25 不包含 MCP config write UI、remote MCP/OAuth、marketplace、真实 language server lifecycle、browser/web/artifact runtime 或多 agent 编排。

## 近期已完成：M21 Workspace Read Tools

当前 Spec Kit 功能目录：

```text
specs/029-workspace-read-tools/
```

关键产物：

- `spec.md`：定义 `workspace.glob`、`workspace.grep`、`workspace.read` 的 bounded read-only scope、边界保护、UI/timeline 可见性和非目标。
- `plan.md`：确定复用 M18 ToolCatalog/ToolBroker/RunContext/worker approval path，并使用 Go stdlib 文件扫描。
- `research.md`：记录单 workspace root、deny-before-read、bounded results 和 mechanism-only reference 决策。
- `data-model.md`：定义 Workspace Tool Definition、Workspace Scope、Tool Arguments、Tool Result 和 Tool Call Event。
- `contracts/`：定义 catalog、read-only auto-approval、failure contract。
- `quickstart.md`：记录 backend smoke、UI smoke 和 required validation。
- `tasks.md`：按 backend foundation、approved execution、安全边界、UI/docs/validation 拆分。

状态：M21 candidate。Workspace tools 只在 Work mode RunContext 中启用，并按最新用户意图收窄 enabled tools；日常问候不暴露 workspace/sandbox/agent/artifact/browser/web 工具，文件/目录任务才暴露 workspace read 工具。Chat mode 不扩大 workspace access。工具 root 在 run 创建时从本地用户持久化 workspace root 快照进 background job、RunContext 和 tool invocation；`/v1/workspace/root` 不再写进程级 `LOOMI_WORKSPACE_ROOT`，该 env 只作为本地测试 fallback 在 run 创建时读取。未设置目录时不再默认用户 Home，workspace 工具会在读取前失败，桌面端必须显式选择目录或测试环境显式设置 `LOOMI_WORKSPACE_ROOT`。路径边界继续拒绝 traversal、absolute escape、symlink escape 和敏感路径。目录分类优先使用 `workspace.tree_summary` 或 `workspace.list_directory`，`workspace.glob` 仅用于文件名匹配或窄范围补充，`workspace.grep` 仅用于内容搜索。浏览器 smoke 需在本轮 closeout 单独记录。M21 不包含 shell、write/edit、sandbox、browser automation、web search/fetch、artifact create 或多工具循环。

## 近期已完成：M20 Local Codex Execution Bridge

当前 Spec Kit 功能目录：

```text
specs/028-local-codex-execution-bridge/
```

关键产物：

- `spec.md`：定义显式启用后的 Local Codex 可执行 provider、Chat 发送、run event 可观察性和 redaction 边界。
- `plan.md`：确定 auth.json direct bridge、复用 runtime.Provider/Gateway/worker，不新增聊天链路。
- `research.md`：记录 auth.json direct 与 CLI 的 tradeoff，拒绝 CLI 交互/泄漏风险。
- `data-model.md`：定义 LocalCodexCredentialSnapshot、LocalCodexProvider、LocalProviderEnablement 和 ProviderCapability extension。
- `contracts/`：定义 enable/model provider/run event/UI contract。
- `quickstart.md`：记录自动验证、手动 smoke 和 fixture proof。
- `tasks.md`：按 backend/web/docs/validation 拆分。

状态：M20 complete candidate。Local Codex 只有显式 detect + enable 后才会注册为 session-local Gateway provider；`GET /v1/model-providers` 返回 `local_codex` 为 available/supported；Chat Composer 对 supported Local Codex 不再显示 provider unavailable warning。发送后走现有 model_gateway run、worker、Gateway、run events、SSE、RunTimeline/RunRail。M20 不调用 CLI、不刷新 OAuth、不读 keychain、不安装 CLI、不新增 sandbox/browser/filesystem/shell/workspace tools；自动化证据使用 temp `CODEX_HOME` fixture 和本地 OpenAI-compatible server。真实本机 Chat 需要在目标机器手动 smoke 验证，失败时记录 provider failure，不伪造回复。

## 近期已完成：M19 Local Provider Opt-in Bridge

当前 Spec Kit 功能目录：

```text
specs/027-local-provider-opt-in-bridge/
```

关键产物：

- `spec.md`：定义 Local Codex 显式本会话启用、安全 provider capability、unsupported execution 和非目标。
- `plan.md`：确定 enable action 可触发安全 detection，model provider list 不触发 detection，状态仅保存在进程内。
- `research.md`：记录 session-local enablement、no auto detection、unsupported execution 等决策。
- `data-model.md`：定义 LocalProviderEnablement、ProviderCapability extension 和 redaction rules。
- `contracts/`：定义 enable/disable API 和 Settings provider UI 契约。
- `quickstart.md`：记录 focused/full validation。
- `tasks.md`：按 backend tests/API、web tests/UI、docs/validation 拆分。

状态：M19 complete candidate。`POST /v1/local-provider-detections/local_codex/enable` 在显式动作后把 Local Codex 加入 session-local configured provider list；`DELETE` 禁用；`GET /v1/model-providers` 只返回已显式启用的 local provider safe capability，不做 detection。Local Codex 当前 `execution_state=unsupported`，所以 Chat 仍阻止发送。M19 不执行 CLI、不刷新 OAuth、不读 keychain、不调用外网、不保存 token，也不新增 sandbox/browser/filesystem/shell/workspace tools。

## 近期已完成：M18.5 Local Provider Autodetect Foundation

当前 Spec Kit 功能目录：

```text
specs/026-local-provider-autodetect-foundation/
```

关键产物：

- `spec.md`：定义 Claude Code/Codex 本地 provider 检测、safe capability fields、explicit opt-in、Settings 状态和非目标。
- `plan.md`：确定 detector 在 backend runtime/provider 边界，API 使用 read-only endpoint，Settings 只显示 detection evidence，不接入 model gateway。
- `research.md`：记录 detection-only、fixture roots/env map、dedicated endpoint、safe status/model labels、helper/keychain/refresh unsupported 等决策。
- `data-model.md`：定义 LocalProviderDetectionInput、LocalProviderCapability、LocalProviderDetectionResponse 和状态转换。
- `contracts/`：定义 `GET /v1/local-provider-detections` 和 Settings provider UI 契约。
- `quickstart.md`：记录 detector/httpapi/web/full validation。
- `tasks.md`：按 detector、API、Settings UI、docs/validation 拆分。

状态：M18.5 complete candidate。当前实现可识别 fixture Claude Code `.claude.json` primaryApiKey、`.claude/settings.json` env、`.claude/.credentials.json` OAuth shape，以及 Codex env/API key/auth-file/OAuth token presence；通过 `GET /v1/local-provider-detections` 和 Settings > Providers 显示 Local Claude Code / Local Codex 状态。M18.5 不执行 CLI/helper，不读 keychain，不刷新 OAuth，不调用外网，不保存 token，不自动启用或切换 provider，不做 Tool Runtime、workspace read/grep、sandbox、browser、web search 或 plugin marketplace。

## 当前候选完成：M18 Tool Runtime + Tool Catalog Foundation

当前 Spec Kit 功能目录：

```text
specs/025-tool-runtime-catalog-foundation/
```

关键产物：

- `spec.md`：定义统一 Tool catalog、broker/executor、RunContext/persona/discovery policy、approval/event/redaction、read-only API/UI 和明确非目标。
- `plan.md`：确定复用 M7/M12 approval projection、M9 RunContext、M10 persona、M11/M12 MCP discovery/execution、现有 Settings/docs-site，不新增强工具。
- `research.md`：记录 computed catalog、single broker entrypoint、reuse lifecycle events、local stdio MCP only、Settings read-only 决策。
- `data-model.md`：定义 Tool Catalog Entry、Tool Invocation、Tool Result 和 RunContext Tool Runtime Summary。
- `contracts/`：定义 Tools catalog API、broker checks、event/redaction、Settings UI safe display。
- `quickstart.md`：记录验证命令和 builtin/MCP approval-to-broker smoke。
- `tasks.md`：按 Spec Kit、tests、catalog/broker、API/smoke、web UI、docs/validation 拆分。

状态：M18 complete candidate。`runtime.get_current_time` 与本地 stdio MCP approved execution 都从 worker resume 进入 `ToolBroker`；RunContext enabled tools 由 catalog + persona allowlist + MCP discovery schema hash 交集生成；`GET /v1/tools/catalog` 和 Settings > Tools 仅展示 safe catalog。M18 不包含 workspace/shell/sandbox/browser/web/artifact tools、plugin marketplace、remote MCP/OAuth、Local Provider autodetect、多工具循环、多 agent 或 worker queue rewrite。

## 近期已完成：M17 Work Artifact Evidence Closeout

当前 Spec Kit 功能目录：

```text
specs/024-work-artifact-evidence-closeout/
```

关键产物：

- `spec.md`：定义 Work artifact evidence closeout，覆盖 local evidence seed、safe artifact metadata、redaction marker、browser smoke 和 Chat/Work isolation。
- `plan.md`：确定复用现有 thread/message/run/event/ChatCanvas/RunRail 边界，通过 local-dev/test seed 而不是生产事件写 API 产出 evidence。
- `research.md`：记录 `loomi-seed` local evidence path、metadata-only artifact evidence 和 projection redaction marker 决策。
- `data-model.md`：定义 Work Evidence Seed、Work Event Metadata、Artifact Evidence Reference 和 browser smoke evidence。
- `contracts/`：定义 local seed contract、event metadata shape 和 safe display contract。
- `quickstart.md`：记录验证命令、seed 命令和 browser smoke 步骤。
- `tasks.md`：按 setup、seed evidence、projection、safe artifact、mode isolation、docs/validation 拆分。

状态：M17 complete candidate。`LOOMI_SEED_SCENARIO=m17-work-artifact go run ./cmd/loomi-seed` 可创建或复用 `thr_m17_work_artifact`，写入 message/current run/`work.plan.updated` event metadata；Work Plan View 通过 real API replay 显示 goal、steps、status、artifact references、redaction marker 和 recent progress；Chat mode 不显示 Work Plan View。M17 不包含 artifact execution/runtime、sandbox、shell/filesystem/browser automation、activity recorder、multi-agent、plugin marketplace、新 task system、worker queue rewrite 或生产 event-write API。

## 近期已完成：M16 Work Mode Foundation

当前 Spec Kit 功能目录：

```text
specs/023-work-mode-foundation/
```

关键产物：

- `spec.md`：定义 Work mode 最小可用薄片，覆盖 goal、steps、status/progress、artifact references、recent events、Chat/Work isolation、安全 metadata 和明确非目标。
- `plan.md`：确定复用现有 thread/message/run/event/ChatCanvas/Timeline 边界，不新增 backend API、task system、worker queue 或 sandbox。
- `research.md`：记录 frontend projection、safe artifact metadata、主区域 Work Plan View 和 no-new-API 决策。
- `data-model.md`：定义 Work Thread、Work Plan Projection、Work Step、Artifact Reference 和 Recent Progress Event。
- `contracts/`：定义现有 run event metadata 上的 optional safe payload shape 和 redaction contract。
- `quickstart.md`：记录 web/docs validation 和 browser smoke。
- `tasks.md`：按 setup、projection、US1 Work view、US2 safe artifacts、US3 mode isolation、docs/validation 拆分。

状态：M16 candidate。Work mode thread 复用 `Thread.mode = work`，在主区域显示 Work Plan View；progress 来自 messages/current run/run events；artifact 第一版仅 metadata/markdown-like preview；Chat mode 不显示 Work Plan View；Work mode Composer 在无 active run 且 backend/provider 可用时可以发起真实 model_gateway run。M16 不包含 sandbox、shell/filesystem/browser automation、activity recorder、multi-agent、plugin marketplace、real artifact execution/runtime 或 worker queue rewrite。

## 后续路线纠偏：M18+ Tool Runtime 优先

M18 已完成 Tool Runtime + Tool Catalog foundation，后续工具能力应继续复用 broker/catalog，而不是直接在 worker/provider 里增加执行分支。

后续 Spec Kit 应按以下方向拆分：

- `024` 保持当前 M17 Work artifact/evidence closeout，不临时扩大范围。
- `029` 已用于 M21 Workspace Read Tools：只读 `workspace.glob`、`workspace.grep`、`workspace.read`，绑定 workspace/project scope，禁止文件写入和 host shell。
- `030` 已用于 M22 Bounded Agent Loop：允许 bounded multi-tool approval loop 和 todo projection，不引入外部 agent。
- `031` 用于 M23 Workspace Mutation Tools：approval-gated `workspace.write_file` / `workspace.edit`、workspace scope guard、event-safe mutation previews、RunRail high-risk/write-capable visibility。
- `032` 用于 M24 Bounded Read-only Command：保留 `sandbox.exec_command` 名称，但第一版只允许 read-only command allowlist，不宣传隔离 sandbox。
- `033` 已用于 M25 MCP Management + LSP Read-only Foundation：Settings > MCP safe status 和 read-only LSP tools。
- `034` 已用于 M26 Web Fetch Tool Foundation：public HTTP(S)-only bounded fetch。
- `035` 已用于 M27 Browser Automation Foundation：public HTTP(S)-only run-scoped browser session。
- `036` 已用于 M28 Artifact Runtime Foundation：PG-backed non-executable text artifact tools。
- `037` 已用于 M29 Multi-agent Runtime Foundation：PG-backed coordination-only agent task tools。
- `038` 保持 Activity Recorder draft/future，不纳入当前完成项。

Desktop Runtime、Channels、Heartbeat、Plugin、Activity Recorder、marketplace 和复杂多 agent 编排应排在上述工具 runtime、安全边界和真实隔离方案之后。

## 近期已完成：M15 Chat Real Integrated Smoke Closeout

当前 Spec Kit 功能目录：

```text
specs/022-chat-real-integrated-smoke-closeout/
```

关键产物：

- `spec.md`：定义 M15 closeout/evidence slice，覆盖 real API path、deterministic provider、approved memory snapshot、MCP approval/execution、continuation、final assistant message、history replay 和 redaction。
- `plan.md`：确定复用现有 Go API/service/worker/runtime/productdata 边界，不引入新平台能力。
- `research.md`：记录 gated Go smoke、deterministic provider、复用 M7/M9/M11/M12/M13、敏感 canary redaction 断言等决策。
- `data-model.md`：定义 M15 Smoke Scenario、Approved Memory Snapshot、MCP Candidate Fixture、Tool Approval Projection 和 Replay Evidence Set。
- `contracts/`：定义 smoke command、evidence milestone 和 redaction contract。
- `quickstart.md`：记录 gated smoke 与完整验证命令。
- `tasks.md`：按 setup、foundation、US1 approval boundary、US2 execution/continuation、US3 replay/redaction、docs/validation 拆分。

状态：M15 complete candidate。`TestM15ChatRealIntegratedSmoke` 通过真实 HTTP handler 创建 thread/message/run，worker 准备 RunContext 并加载 approved memory，provider fixture 请求 discovered + persona-allowed MCP tool，run 进入 approval required，HTTP approve 后 worker 执行一次 local stdio MCP `tools/call`，记录 redacted result，provider continuation 写入 final assistant message，replay API 验证完整事件链。M15 不包含 sandbox、filesystem/shell/browser automation tools、activity recorder、OpenViking/vector/RAG/distill、marketplace/plugin install、multi-agent 或 worker queue rewrite。

## 近期已完成：M14 Memory Management Audit UX

当前 Spec Kit 功能目录：

```text
specs/021-memory-management-audit-ux/
```

关键产物：

- `spec.md`：定义 Settings > Memory 管理面、safe audit/history、delete confirmation、状态覆盖、seeded browser smoke 和非目标。
- `plan.md`：确定复用 M13 productdata memory、HTTP API、realApiClient、SettingsView 和 docs-site，不引入 distill/RAG/OpenViking/MCP/worker/sandbox。
- `research.md`：记录复用 M13 audit/event、最小 scoped audit read、grounded filters、tombstone confirmation、safe projection 和 blocker findings 决策。
- `data-model.md`：定义 Memory Management Item、Detail、Filter、Audit Item 和状态规则。
- `contracts/`：定义 list/search/detail/delete/audit endpoint、payload、forbidden fields 和 no-existence-leak 行为。
- `quickstart.md`：记录 M14 prep blockers、验证命令和 full UI seeded browser smoke 标准。
- `tasks.md`：把 prep blockers 与后续完整 UX 实现任务分开。

状态：full UX complete candidate。已完成 thread-scoped read/delete authorization、thread list/search missing `scope_id` invalid request、terminal-run durable memory audit、redaction hardening、search/list/audit filter shape 收口、Settings > Memory list/search/filter/detail/delete confirmation、真实 `/v1/memory/audit` history、backend unavailable/error/empty 不造假，以及 seeded browser smoke。M14.1 已修复 audit filter shape 与 latest-request guard，避免 M15 evidence 混入其它 thread history 或被旧响应覆盖。M14 仍不包含 distill、OpenViking、vector/embedding/RAG、activity recorder、MCP、worker queue、sandbox 或多 agent 重写。

## 当前进行：M42 Memory Provider Foundation

当前 Spec Kit 功能目录：

```text
specs/042-memory-provider-foundation/
```

关键产物：

- `spec.md`：定义 memory provider foundation 第一薄片，覆盖 backend-owned provider config/status、semantic readiness diagnostics、安全 run readiness、Settings > Memory provider 状态和 M13/M14 行为不回归。
- `plan.md`：确定复用 productdata/httpapi/runtime/web/docs-site 边界，不引入 full provider adapter、memory tools、distillation、embedding/vector store 或外部语义服务。
- `tasks.md`：按 backend provider foundation、Settings UI、docs/validation 拆分。

状态：M42 candidate。`GET/PUT /v1/memory/provider` 提供 enabled/provider/commit-after-run/status/diagnostic safe projection；默认 local provider 保留现有 approved memory store；semantic provider 仅作为 readiness-capable future mode，缺配置时返回 `unconfigured`，未知 provider 降级到 local。RunContext safe summary 增加 memory readiness metadata；Settings > Memory 增加 backend-derived Memory Service 面板。M42 不包含 agent memory tools、automatic distillation、external semantic read/write、embedding/vector search、activity recorder 或 multi-agent long-term memory。

## 当前已完成：M13 Memory Foundation + M13.5 Closeout

当前 Spec Kit 功能目录：

```text
specs/019-memory-foundation/
specs/020-memory-real-pg-smoke-closeout/
```

关键产物：

- `spec.md`：定义 PG v1 memory、RunContext safe memory snapshot、approval-gated memory write、用户查看/删除、隐私/安全/删除/审计/redaction 边界。
- `plan.md`：确定 v1 只实现 PG provider，复用 RunContext/Pipeline、productdata、HTTP API、frontend shell/docs-site 边界，不引入向量库/embedding/RAG/OpenViking。
- `research.md`：记录 PG-first、approval-gated writes、tombstone deletion、redact-before-exposure、MemoryProvider PG-only 和 distill deferred 决策。
- `data-model.md`：定义 Memory Entry、Memory Search、Memory Write Proposal、Approval Decision、Snapshot、Tombstone、Audit Event 和 MemoryProvider。
- `contracts/`：定义 memory API、memory events/audit、MemoryProvider PG v1 契约。
- `quickstart.md`：记录 backend/web/docs validation 和 smoke expectations。
- `tasks.md`：按 setup、foundation、US1 safe snapshot、US2 approval-gated writes、US3 user control、US4 planned-only provider/distill、docs/validation 拆分实现任务。
- `020` closeout：记录真实 Postgres migration + HTTP API smoke evidence，不新增 memory platform 功能。

状态：第一实现薄片和 M13.5 evidence closeout 完成。已实现 PG-backed memory entries/search/write proposal、RunContext safe memory snapshot、approval-gated writes、用户 list/search/delete、redaction/tombstone/audit 边界；`TestM13MemoryRealPGHTTPAPISmoke` 覆盖真实 Postgres/httpapi migrated path。未实现向量库/embedding/RAG、OpenViking、自动 distill、marketplace/plugin、sandbox/browser/activity recorder、多 agent 长期自动化、worker/job queue rewrite 或 MCP rewrite。

## 近期已完成：M12.5 Real MCP Smoke Closeout

当前 Spec Kit 功能目录：

```text
specs/018-m12-real-mcp-smoke-closeout/
```

关键产物：

- `spec.md`：定义 M12.5 只补真实本地 smoke/evidence closeout，不扩展 remote MCP、marketplace、plugin install、sandbox、automation 或多工具循环。
- `plan.md`：确定用现有 M12 本地 stdio MCP、M7 approval、M6 worker、M10 persona、M11 discovery 和 provider continuation 边界补证据。
- `research.md`：记录 Go test subprocess fixture、HTTP approve、worker real executor、browser smoke limitation 和 closeout-only scope 决策。
- `data-model.md`：定义 M12.5 Smoke Run、Local MCP Fixture 和 Closeout Evidence。
- `contracts/`：定义从 discovery 到 approval、execution、redacted result、continuation 和 final 的证据链。
- `quickstart.md`：记录 targeted smoke、完整验证命令和 browser smoke 条件。
- `tasks.md`：按 setup、fixture、US1 smoke、US2 docs 和 validation 拆分任务。

## 近期已完成：M12 MCP Approval-Gated Execution

当前 Spec Kit 功能目录：

```text
specs/017-mcp-approval-gated-execution/
```

关键产物：

- `spec.md`：定义已发现本地 stdio MCP tool 如何进入 M7 approval/tool-call/audit/worker/run-event 边界并执行一个最小安全闭环。
- `plan.md`：确定复用 M7 approval、M6 worker/job、M9 RunContext/pipeline、M10 persona allowed-tools、M11 discovery/candidate mapping 和现有 Timeline/debug。
- `research.md`：记录 approval-only entry、discovered+persona gate、at-most-once execution、stdio lifecycle redaction、single continuation 和 deferred scope 决策。
- `data-model.md`：定义 MCP Tool Execution Request、Scoped Tool-Call Projection、Execution Attempt、Stdio Invocation、Result Summary、Continuation Context 和 Audit Event。
- `contracts/`：定义 approval gate、worker execution、continuation、redaction/events 契约。
- `quickstart.md`：记录 backend/web/docs validation 和 local smoke expectations。
- `tasks.md`：按 setup、foundation、US1-US3 和 docs/validation 拆分实现任务。

## 近期已完成：M11 MCP Stdio Foundation

当前 Spec Kit 功能目录：

```text
specs/016-mcp-stdio-foundation/
```

关键产物：

- `spec.md`：定义本地 stdio MCP 配置、discover/list-tools、read-only ToolSpec candidate、persona 非执行引用、RunContext availability summary 和安全边界。
- `plan.md`：确定复用 M7 approval、M9 RunContext/pipeline、M10 persona allowed-tools、M6 worker/job 和现有 Timeline/debug 边界。
- `research.md`：记录 local explicit config、sensitive redaction、discovery-only、namespacing、persona non-executable reference 和 future approval execution 决策。
- `data-model.md`：定义 MCP Server Config、Discovery Session、Tool Candidate、Availability Summary、Safety Error 和 Execution Boundary。
- `contracts/`：定义 config、discovery/mapping、RunContext observability 和 future execution boundary 契约。
- `quickstart.md`：记录 backend/web/docs validation 和 browser/debug smoke。
- `tasks.md`：按 setup、foundation、US1-US3 和 docs/validation 拆分实现任务。

## 近期已完成：M10 Persona Skill Foundation

当前 Spec Kit 功能目录：

```text
specs/015-persona-skill-foundation/
```

关键产物：

- `spec.md`：定义 persona 数据模型、内置 persona 同步、thread/run 选择或继承、RunContext snapshot/version 和安全 Timeline/debug summary。
- `plan.md`：确定复用 productdata、M9 RunContext pipeline、M7 MVP tool allowlist、现有 run/event/SSE 和 frontend Composer/RunRail 边界。
- `research.md`：记录 built-in persona sync、immutable run snapshot、安全 summary、最小 selector 和非目标决策。
- `data-model.md`：定义 Persona、Persona Version、Built-In Persona Config、Persona Selection、Persona Snapshot、Persona Safe Summary 和 Skill Stub。
- `contracts/`：定义 persona sync、persona resolution、frontend safe summary 契约。
- `quickstart.md`：记录 backend/web/docs validation 和 browser smoke。
- `tasks.md`：按 setup、foundation、US1-US3 和 docs/validation 拆分实现任务。

## 近期已完成：M9 RunContext Pipeline Foundation

Spec Kit 功能目录：

```text
specs/014-run-context-pipeline-foundation/
```

关键产物：

- `spec.md`：定义 durable RunContext loader、worker 不依赖 API request memory、线性 pipeline trace 和非目标。
- `plan.md`：确定复用 M6/M8 worker/job queue、M7 continuation、现有 run/event/SSE 和 frontend Timeline/RunRail 边界。
- `research.md`：记录 durable context、窄 RunContext 字段、线性 stage list、safe stage metadata 和 M7 continuation preservation 决策。
- `data-model.md`：定义 RunContext、ContextSource、Pipeline Stage、Pipeline State、Pipeline Trace Event 和 Stage Failure。
- `contracts/`：定义 RunContext loader、pipeline stage events、frontend debug trace 契约。
- `quickstart.md`：记录 backend/web/docs validation 和 browser smoke。
- `tasks.md`：按 setup、foundation、US1-US3 和 docs/validation 拆分实现任务。

## 近期已完成：M6 Worker Job Pipeline

当前 Spec Kit 功能目录：

```text
specs/008-worker-job-pipeline/
```

关键产物：

- `spec.md`：定义 background worker、durable job queue、recovery、cancellation 和 diagnostics user stories。
- `plan.md`：确定复用 M4/M5 run/event/SSE/message 基座，新增 database-backed job queue 和 local in-process worker。
- `research.md`：记录 durable queue、worker lease、terminal idempotency、safe cancellation、minimal pipeline 和 diagnostics 决策。
- `data-model.md`：定义 Background Job、Worker Lease、Pipeline Step、Queue Diagnostics 和扩展 Run 状态。
- `contracts/http-m6.openapi.yaml`：定义 queued run creation、stop、event history/SSE 和 diagnostics 契约。
- `contracts/worker-queue.md`：定义 claim、lease、retry、cancellation、completion 和 diagnostics 语义。
- `contracts/pipeline-events.md`：定义 `run_queued`、`job_claimed`、pipeline、recovery、retry、stop 和 terminal events。
- `quickstart.md`：记录 queued ack、reconnect、recovery、cancellation、diagnostics、frontend 和 rollback validation。
- `tasks.md`：按 setup、foundation、US1-US4 和 polish 拆分并跟踪实现任务；当前已完成 US1-US4。

## 近期已完成：M5.5 Settings Placeholder

Spec Kit 功能目录：

```text
specs/007-settings-placeholder/
```

关键产物：

- `spec.md`：定义临时 Settings 占位界面、当前可用设置、read-only 状态和 mock 安全边界。
- `plan.md`：确定在现有 web shell 内实现两列 Settings surface，不引入新依赖或持久化设置。
- `research.md`：记录 in-app Settings、session-local controls、read-only provider state、placeholder panels 和 docs/smoke 决策。
- `data-model.md`：定义 Settings Category、Setting Section、Setting Row、Local Settings State、Runtime Capability Summary 和 Placeholder Setting。
- `contracts/settings-ui.md`：定义入口、布局、分类、working rows、placeholder safety 和视觉契约。
- `quickstart.md`：记录 mock desktop、General working settings、placeholder navigation、real API visibility 和验证命令。
- `tasks.md`：按 setup、foundation、US1、US2、US3 和 polish 拆分并跟踪实现任务。

## 近期已完成：M5 LLM Gateway

Spec Kit 功能目录：

```text
specs/005-llm-gateway/
```

关键产物：

- `spec.md`：定义模型网关用户故事，明确成功响应、失败可见性和工具边界。
- `plan.md`：确定复用 M3 thread/message 与 M4 run/event/SSE 基座，使用后端本地 provider 配置和 Go stdlib HTTP。
- `research.md`：记录 server-side gateway、provider stream normalization、redacted errors、current-thread context 和非执行 tool boundary 决策。
- `data-model.md`：扩展 Message/Run/Run Event/Provider Capability/Gateway Request Context 等实体。
- `contracts/http-m5.openapi.yaml`：定义 provider capability、model_gateway run creation 和 run event/SSE 契约。
- `contracts/provider-event-mapping.md`：定义 Anthropic、OpenAI、Gemini 和 OpenAI-compatible provider 事件到 Loomi run events 的映射。
- `contracts/frontend-runtime.md`：定义前端真实 API 模式下的 assistant draft、provider failure 和 tool-boundary 行为。
- `quickstart.md`：记录 provider 配置、API/SSE、failure、tool-boundary、frontend 和验证命令。
- `tasks.md`：按 foundation、US1、US2、US3 和 polish 拆分并跟踪实现任务。

## 近期已完成：006 Streaming Chat Runtime

Spec Kit 功能目录：

```text
specs/006-streaming-chat-runtime/
```

US1-US5 are implemented: streaming Chat Canvas draft bubbles, grouped Timeline/debug events, backend capability status, composer stop/retry/regenerate/continue controls, and synchronized thread/message loading/error/history states. Final validation is tracked in tasks T060-T063.

## 近期已完成：M3.5 Frontend Agent Runtime Skeleton

Spec Kit 功能目录：

```text
specs/004-frontend-agent-runtime/
```

关键产物：

- `spec.md`：定义 M3.5 范围，明确只做前端 Agent runtime 骨架、Chat Canvas 状态机、mock run/event 剧本和 future real adapter 接入点。
- `plan.md`：确定新增 `web/src/runtime/` 边界，保持 M3 thread/message client 和 M4/M5 run/event/LLM 后端语义分离。
- `research.md`：记录 runtime adapter、纯状态派生、deterministic mock scripts、real-mode unavailable 和 Chat-first 范围决策。
- `data-model.md`：定义 Chat Canvas State、Runtime Run、Runtime Event、Runtime Script、Assistant Draft、Execution Adapter、Backend Capability State 和 Stale Event Guard。
