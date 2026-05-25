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

## 当前候选完成：M18.5 Local Provider Autodetect Foundation

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

状态：M16 complete candidate。Work mode thread 复用 `Thread.mode = work`，在主区域显示 Work Plan View；progress 来自 messages/current run/run events；artifact 第一版仅 metadata/markdown-like preview；Chat mode 不显示 Work Plan View。M16 不包含 sandbox、shell/filesystem/browser automation、activity recorder、multi-agent、plugin marketplace、real artifact execution/runtime 或 worker queue rewrite。

## 后续路线纠偏：M18+ Tool Runtime 优先

M18 已完成 Tool Runtime + Tool Catalog foundation，后续工具能力应继续复用 broker/catalog，而不是直接在 worker/provider 里增加执行分支。

后续 Spec Kit 应按以下方向拆分：

- `024` 保持当前 M17 Work artifact/evidence closeout，不临时扩大范围。
- `027` 建议用于 M19 Workspace Read Tools：只读 `workspace.glob`、`workspace.grep`、`workspace.read`，绑定 workspace/project scope，禁止文件写入和 host shell。
- `028` 建议用于 M20 Artifact Tools：DB/object-store backed artifact create/read/list，把 Work artifact references 接真实 artifact。
- `029` 建议用于 M21 Sandbox Code/Shell Tools：隔离 sandbox provider、`exec_command`、`python_execute`、continuation/terminate、stdout/stderr streaming、timeout/resource limits。
- `030` 建议用于 M22 Web Search/Fetch Tools：`web.search`、`web.fetch`、citation、URL policy 和 SSRF/private-network guard。
- `031` 建议用于 M23 Browser Automation Tools：sandbox browser navigate/snapshot/screenshot/click/type/console/network。
- `032` 建议用于 M24 Tool Settings + MCP Management：Settings > Tools、MCP server list、本地 stdio config UI、provider health、enable/disable 和 approval policy override。

Desktop Runtime、Channels、Heartbeat、Plugin、Activity Recorder、marketplace 和复杂多 agent 编排应排在上述工具 runtime、安全边界和 sandbox 能力之后。

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
