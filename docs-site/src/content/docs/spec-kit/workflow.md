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

## 当前候选：Memory Provider Error UI

当前 Spec Kit 功能目录：

```text
specs/070-memory-provider-error-ui/
```

关键产物：

- `spec.md`：定义 Settings > Memory recent errors 展示 runtime run/event 线索的目标、安全边界和验收标准。
- `plan.md`：确定复用现有 recent errors panel 和前端 error formatter。
- `tasks.md`：按 runtime/docs/validation 拆分。

状态：candidate。该轮只补 runtime error display；不做 run detail navigation、modal redesign 或 raw log viewer。

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

- `spec.md`：定义 `agent.spawn`、`agent.list` 和 `agent.complete` 的 approval-gated、Work-mode-only、coordination-only agent task runtime 和非目标。
- `plan.md`：确定复用 ToolCatalog/RunContext/ToolBroker/worker continuation、Settings/RunRail。
- `research.md`：记录先做 task record coordination、不做 autonomous child runs/cross-thread delegation 的决策。
- `data-model.md`：定义 Agent Task、tool arguments 和 result summary。
- `contracts/`：定义 catalog、arguments、result 和 rejection contract。
- `quickstart.md`：记录 focused/full validation 和 manual smoke。
- `tasks.md`：按 catalog/runtime/spawn/list-complete/safety/UI/docs/validation 拆分。

状态：M29 PG-backed candidate。Agent tools 仅 Work mode 启用，always approval required，经 ToolBroker/worker continuation 创建、列出和完成 bounded coordination task records；真实 API 路径使用 PostgreSQL `agent_tasks`，in-memory 和 PG 均覆盖 spawn/list/complete 与 cross-thread no-leak。Settings > Tools 与 RunRail 显示 agent scope、medium risk、coordination-only、no autonomous execution。M29 不包含 autonomous child model runs、cross-thread delegation、external worker pools、process spawning、filesystem access、network calls、shell execution、long-term multi-agent memory、marketplace packaging 或 background swarm orchestration。

## 近期已完成：M28 Artifact Runtime Foundation

当前 Spec Kit 功能目录：

```text
specs/036-artifact-runtime-foundation/
```

关键产物：

- `spec.md`：定义 `artifact.create_text`、`artifact.read` 和 `artifact.list` 的 approval-gated、Work-mode-only、non-executable text artifact runtime 和非目标。
- `plan.md`：确定复用 ToolCatalog/RunContext/ToolBroker/worker continuation、WorkPlan artifact projection、Settings/RunRail。
- `research.md`：记录先做 text-only storage、不做执行/渲染/下载的决策。
- `data-model.md`：定义 Artifact、tool arguments 和 result summary。
- `contracts/`：定义 catalog、arguments、result 和 rejection contract。
- `quickstart.md`：记录 focused/full validation 和 manual smoke。
- `tasks.md`：按 catalog/runtime/read-list/safety/UI/docs/validation 拆分。

状态：M28 PG-backed candidate。Artifact tools 仅 Work mode 启用，always approval required，经 ToolBroker/worker continuation 创建/读取/列出 bounded UTF-8 text artifacts；真实 API 路径使用 PostgreSQL `artifacts`，in-memory 和 PG 均覆盖 create/read/list 与 cross-thread no-leak。Settings > Tools 与 RunRail 显示 artifact scope、medium risk、non-executable。M28 不包含 binary artifacts、downloads、rendered previews、iframe execution、filesystem export、browser integration、shell integration、artifact version graph、marketplace packaging 或 multi-agent orchestration。

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

状态：M21 candidate。Workspace tools 只在 Work mode RunContext 中启用，并按最新用户意图收窄 enabled tools；日常问候不暴露 workspace/sandbox/agent/artifact/browser/web 工具，文件/目录任务才暴露 workspace read 工具。Chat mode 不扩大 workspace access。工具 root 优先来自本地用户持久化的 workspace root，并同步到 `LOOMI_WORKSPACE_ROOT` 供当前进程执行；未设置时本地桌面/dev 默认用户 Home；桌面端可由用户显式选择目录后切换并持久化运行时 root。路径边界继续拒绝 traversal、absolute escape、symlink escape 和敏感路径。目录分类会先做一次 broad `workspace.glob` 后摘要，后续 continuation 不再重复暴露 `workspace.glob`，只保留 targeted grep/read。浏览器 smoke 需在本轮 closeout 单独记录。M21 不包含 shell、write/edit、sandbox、browser automation、web search/fetch、artifact create 或多工具循环。

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
