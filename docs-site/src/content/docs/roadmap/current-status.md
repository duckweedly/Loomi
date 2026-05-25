---
title: 当前状态
description: Loomi 当前已完成内容和建议下一步。
---

## 已有基础

当前仓库已经有 Web/Electron 前端壳、Go API/DB 基座、本地 identity/thread/message、M4 run/event/SSE、M3.5 前端 Agent runtime 骨架、M5 LLM Gateway 的 backend/provider/frontend 基础切片、M5.5 Settings Placeholder 的前端设置占位面、M6 Worker Job Pipeline 的 queued background execution MVP、M6.5 Real Testing Console and Background UX、M7 Tool Call Approval Execution + Tool Result Continuation 的最小闭环、M9 RunContext + Pipeline foundation 的最小薄片、M10 Persona/Skill foundation 的 persona 选择与版本快照薄片、M11 MCP stdio foundation 的本地 discovery/read-only candidate 薄片、M12/M12.5 MCP approval-gated execution 的单工具本地 stdio 执行薄片和真实本地 smoke closeout 证据、M13 Memory Foundation 的 PG-backed memory 第一实现薄片、M15 Chat real integrated smoke closeout 的 deterministic 后端证据链、M16 Work mode foundation 的最小计划/进度/artifact metadata UI 薄片、M17 Work artifact evidence closeout 的可重复 local seed + real API/browser evidence path、M18 Tool Runtime + Tool Catalog foundation 的统一 catalog/broker/API/Settings 只读薄片，M18.5 Local Provider Autodetect 的检测-only Settings/API 证据薄片，M19 Local Provider Opt-in Bridge 的 Local Codex 显式本会话启用薄片，以及 M20 Local Codex Execution Bridge 的 fixture-backed Gateway 执行薄片。

M13 Memory Foundation 的 Spec Kit 目录为 `specs/019-memory-foundation/`，M13.5 closeout 目录为 `specs/020-memory-real-pg-smoke-closeout/`。当前已实现第一薄片：PG v1 memory entries/search/write proposal、RunContext safe memory snapshot、approval-gated agent memory writes、用户查看/检索/删除记忆的最小 API/UI、tombstone deletion、redaction 和 source-run audit events。M13.5 已补真实 Postgres/httpapi smoke，覆盖 migrated `memory_entries`/`memory_write_proposals`、proposal/approve/list/search/RunContext/delete/idempotency/out-of-scope/redaction。它明确不包含向量库/embedding/RAG、OpenViking provider、marketplace/plugin、sandbox/browser/activity recorder、多 agent 长期记忆自动化、worker/job queue rewrite 或 MCP rewrite。

M14 Memory Management Audit UX 的 Spec Kit 目录为 `specs/021-memory-management-audit-ux/`。当前已推进到 full UX complete candidate：thread-scoped memory detail/delete 需要匹配 thread/source scope，thread list/search 缺少 `scope_id` 会返回参数错误，terminal run 后 memory proposal/approve/deny/delete audit 写入 durable `memory_audit_events`，redaction 覆盖 `/home`、Windows path、stdout/stderr、tool output、provider trace、key/env markers，list/search/audit API 与前端 client 统一到 `query/q`、`limit`、`scope_type`、`scope_id`、`source_thread_id`、`source_run_id`、`source_type`、`include_tombstoned`。Settings > Memory 已覆盖真实 approved/tombstoned list、search、grounded filters、safe detail panel、delete confirmation、loading/empty/error/deleted states、thread-scoped detail/delete context guard、真实 `/v1/memory/audit` history、backend unavailable/error/empty 不造假，以及 seeded browser smoke。M14.1 review fix 补齐 audit thread filter contract、memory list/audit/detail latest-request guard 和 delete confirmation/error polish。M14 仍不包含 distill 自动总结、OpenViking、vector/embedding/RAG、browser/activity recorder ingestion、MCP、worker queue、sandbox 或多 agent 重写。

M15 Chat Real Integrated Smoke Closeout 的 Spec Kit 目录为 `specs/022-chat-real-integrated-smoke-closeout/`。当前候选证据是 gated backend smoke `TestM15ChatRealIntegratedSmoke`：通过 HTTP 创建 thread/message/model-gateway run，加载 approved memory snapshot，经 deterministic provider 请求 discovered + persona-allowed local stdio MCP tool，进入 approval required，通过 HTTP approve，worker 执行一次 MCP `tools/call`，记录 redacted result，provider continuation 写入 final assistant message，并从 replay API 验证 memory/MCP/approval/execution/continuation/completed 事件链。M15 是 closeout/evidence slice，不新增 sandbox、automation tools、activity recorder、OpenViking/vector/RAG/distill、marketplace/plugin、多 agent 或 worker queue rewrite。

M16 Work Mode Foundation 的 Spec Kit 目录为 `specs/023-work-mode-foundation/`。当前候选实现是在现有 `Thread.mode = work` 上增加 Work Plan View：从 messages、当前 run 和 run events safe metadata 投影 goal、steps、status、artifact references 和 recent progress。Artifact 第一版仅显示 title/type/source thread/run/summary/created/updated，secret-looking 内容 redacted，command/path/shell/browser/filesystem/executable metadata 不作为 action 暴露。M16 没有新增 backend API、task system、sandbox、filesystem/shell/browser tools、activity recorder、multi-agent 或 worker queue rewrite。

M17 Work Artifact Evidence Closeout 的 Spec Kit 目录为 `specs/024-work-artifact-evidence-closeout/`。当前候选实现新增 `LOOMI_SEED_SCENARIO=m17-work-artifact` local-dev/test seed：复用现有 productdata thread/message/run/event 服务创建 Work thread、message、current run 和 `work.plan.updated` metadata event，前端通过 real API replay 投影 goal、steps、status、artifact references 和 recent progress。Artifact cards 增加 redaction marker，但仍无 executable controls。M17 没有新增生产事件写 API、artifact execution/runtime、sandbox、shell/filesystem/browser automation、activity recorder、multi-agent、plugin marketplace、新 task system 或 worker queue rewrite。

M18 Tool Runtime + Tool Catalog Foundation 的 Spec Kit 目录为 `specs/025-tool-runtime-catalog-foundation/`。当前候选实现新增 safe tool catalog、read-only `GET /v1/tools/catalog`、Settings > Tools 只读 catalog、统一 `ToolBroker`/`ToolInvocation`/`ToolResult` envelope，并让 `runtime.get_current_time` 与本地 stdio MCP 执行都从 worker approved-tool resume 进入同一 broker。RunContext enabled tools 从 catalog、persona allowlist 和 MCP discovery schema hash 交集生成；未发现、未 allowlist、disabled/non-executable 或 schema mismatch 的工具会在执行前失败。M18 没有新增 workspace/shell/sandbox/browser/web/artifact tools、plugin marketplace、remote MCP/OAuth、Local Provider autodetect、多工具循环、多 agent 或 worker queue rewrite。

M18.5 Local Provider Autodetect 的 Spec Kit 目录为 `specs/026-local-provider-autodetect-foundation/`。当前候选实现新增 Claude Code/Codex 本地 provider 检测：支持 fixture HOME/CODEX_HOME/CLAUDE_CONFIG_DIR 和 env map，返回 safe capability fields，经 `GET /v1/local-provider-detections` 暴露，并在 Settings > Providers 显示 Local Claude Code / Local Codex detected/not detected、explicit opt-in 和 no secrets copy。M18.5 不执行 CLI/helper，不刷新 OAuth，不读 keychain，不调用外网，不保存 token，不自动启用或切换 provider。

M19 Local Provider Opt-in Bridge 的 Spec Kit 目录为 `specs/027-local-provider-opt-in-bridge/`。当前候选实现新增 `POST/DELETE /v1/local-provider-detections/{provider_id}/enable`，支持 Local Codex 检测为 available 后显式本会话启用；`GET /v1/model-providers` 只返回已显式启用的 local provider safe capability，不触发本机 auth detection。Local Codex 会标记 `local_provider`、`session_local`、`credential_reference=redacted` 和 `execution_state=unsupported`，Chat 仍阻止发送。M19 不能实际通过 Local Codex Chat；剩余 blocker 是 Local Codex execution bridge。

M20 Local Codex Execution Bridge 的 Spec Kit 目录为 `specs/028-local-codex-execution-bridge/`。当前候选实现选择 auth.json direct bridge，不调用 CLI、不刷新 OAuth、不读 keychain；Local Codex 仅在显式 detect + enable 后注册为现有 Gateway provider，`GET /v1/model-providers` 返回 `status=available`、`execution_state=supported`。Chat Composer 不再因 available/supported Local Codex 阻止发送，发送后走原有 model_gateway run、worker、Gateway、run events、SSE、RunTimeline/RunRail 路径。自动化证据是 temp `CODEX_HOME` + 本地 OpenAI-compatible fixture；真实本机 OAuth 是否可用取决于本机 token 对目标 compatible endpoint 是否有效，失败时记录 provider failure，不伪造回复。

原路线 M8 Worker + Job Queue 已由 M6 Worker Job Pipeline 覆盖并通过 013 closeout：`background_jobs` 表、API 同事务 queued run + job、worker claim、lease renew、retry/backoff、failed terminal、lost-lock ownership guard、202 queued acknowledgement、worker crash recovery、旧 worker 失锁后不能写 terminal 均已有证据或已由 closeout 补齐。M8 closeout passed；下一步不要重做 worker/job queue。

Spec Kit 已接入仓库，并写入 Loomi constitution。文档站已作为 `docs-site/` 独立子项目创建，用于长期记录技术细节。

已完成的可运行薄片：

- M1：桌面感 Web 壳和基础面板。
- M2：API、配置、PostgreSQL、migration、health/readiness 和 diagnostics 基座。
- M3：本地 identity、thread/message API、seed、mock/real API 切换。
- M3.5：Chat Canvas 状态机、mock runtime scripts、RunRail/Timeline/Agent badge 联动。
- M4：持久化 run/event、history-first SSE、local simulated run、stop/already-terminal 语义。
- M5：后端 LLM gateway、provider capability、provider-normalized run events、streaming assistant draft、redacted failure states 和 non-executed tool boundary。
- M5.5：Settings Placeholder、General session-local controls、read-only runtime/provider status，以及安全的 mock/preview future settings categories。
- M6：queued run acknowledgement、durable background job、local in-process worker、pipeline events、frontend queued/worker timeline replay、lease recovery、cancellation 和 worker diagnostics。
- M8 closeout：原 M8 Worker + Job Queue 已由 M6 覆盖；013 仅补齐 stale lease recovery 的 retry backoff scheduling，并记录审计证据。
- M6.5：Provider Test Console、real mode provider unavailable guidance、read-only Background tasks observer、M6 worker/job Timeline labels、Composer 状态收口和本地真实测试 runbook。
- M7 approval execution closure：approval-blocked tool-call projection、`runtime.get_current_time` allowlist/schema、tool lifecycle events、safe metadata redaction、worker diagnostics counters、frontend tool-event replay mapping、provider tool-call conversion、scoped tool-call read API、idempotent approve/deny API、enabled approval UI actions、approved current-time worker execution、result/error/denied SSE replay。
- M7 tool-result continuation closure：approved `runtime.get_current_time` 执行成功后自动续调 provider、provider-neutral continuation context、OpenAI-compatible tool-result serialization、`model_phase = continuation` replay、one final assistant message、denied/tool-failed no-continuation、continuation provider failure redacted terminal，以及 continuation 再请求工具时的 `unsupported_tool_loop` 安全失败。
- M9 RunContext/Pipeline foundation：worker 执行前从 durable run/thread/messages/job/provider route/tool summary 准备 RunContext，按 `prepare_context`、`resolve_tools`、`invoke_runtime`、`finalize` 记录线性 pipeline trace，Timeline/debug/Background tasks 可从 live SSE 与 history replay 看到安全摘要。
- M10 Persona/Skill foundation：内置 persona 配置同步到 DB，thread/run 可选择或继承 persona，run 创建时记录 persona snapshot/version，RunContext 应用 persona model route 和 allowed tool names，Timeline/debug 只显示安全 persona summary，前端提供最小 persona selector。
- M11 MCP stdio foundation：显式本地 stdio MCP 配置校验、bounded discovery/list-tools parser、namespaced read-only ToolSpec candidate、persona allowed-tools 非执行引用、RunContext MCP availability 安全摘要、Timeline/debug discovery 成功/失败/禁用标签；MCP 工具执行仍未开放。
- M12 MCP approval-gated execution：已发现且 persona allowed-tools 引用的本地 stdio MCP candidate 可以进入 M7 approval projection；approve 后由 worker 执行一个 bounded stdio `tools/call`，记录 redacted result/error，并进行一次 provider continuation；retry/recovery 不重复执行已 started/terminal 的 MCP call。
- M12.5 real MCP smoke closeout：新增真实本地 stdio MCP fixture 证据，覆盖 `Content-Length` discovery `tools/list`、`candidate_schema_hashes`、persona allowed tool、provider 请求 MCP tool、HTTP approve、worker 通过 `LOOMI_MCP_SERVERS_JSON` 使用 `StdioMCPToolExecutor` 执行一次 `tools/call`、redacted result、continuation 和 final assistant message。
- M13 Memory Foundation + M13.5 closeout：新增 `memory_entries`/`memory_write_proposals`、approved safe memory search、RunContext memory snapshot、`memory_snapshot_loaded`、approval-gated write proposal/approve/deny、source-run memory audit events、用户 list/search/delete API、Settings > Memory 最小 UI，并补真实 Postgres/httpapi smoke closeout 证据。
- M14 Memory Management Audit UX：Settings > Memory 已接真实 memory list/search/filter/detail/delete confirmation 和 `/v1/memory/audit` safe history；seeded browser smoke 覆盖 list、search/filter、detail、delete confirmation、post-delete refresh、`memory_deleted`/proposed/approved/denied/snapshot audit history，console 无 error。
- M14.1 Review Fix：memory audit filter shape 与 list/search 对齐，Settings memory list/audit/detail 加 latest-only guard，避免跨 thread history 混入和旧响应覆盖新状态。
- M15 Chat Real Integrated Smoke Closeout：gated backend smoke 覆盖 real API path、approved memory RunContext snapshot、MCP discovery candidate hash、persona allowed provider tool request、approval required、HTTP approve、worker MCP `tools/call`、redacted result、provider continuation、final assistant message 和 history replay。
- M16 Work Mode Foundation：Work mode thread 在主区域显示 Work Plan View，复用 existing thread/message/run/event 投影 goal、steps、status、artifact references 和 recent progress；Chat mode 不显示该 surface；artifact references 仅 metadata preview 且做 safe redaction。
- M17 Work Artifact Evidence Closeout：新增可重复 local-dev/test seed evidence path，使用现有 thread/message/run/event 数据链写入 Work metadata event；real API/browser smoke 可验证 Work Plan View、artifact redaction marker、无 executable artifact controls 和 Chat mode isolation。
- M18 Tool Runtime + Tool Catalog Foundation：新增 safe catalog、read-only Tools API、Settings > Tools catalog、统一 broker/executor envelope，并把 builtin current-time 与 local stdio MCP approved execution 接入同一执行入口。
- M18.5 Local Provider Autodetect：新增检测-only Claude Code/Codex local provider capability，read-only API 和 Settings > Providers 状态；检测结果不进入 model gateway provider list，不自动启用，不保存或显示 secrets。
- M19 Local Provider Opt-in Bridge：新增 Local Codex 显式 session-local enable/disable API、configured provider safe capability、Settings enable/disable UI 和 Chat unsupported 阻止；不执行真实 Local Codex 模型调用。
- M20 Local Codex Execution Bridge：新增 auth.json direct Local Codex provider，显式 enable 后返回 available/supported，并通过现有 worker/Gateway/run event/SSE 路径完成 fixture-backed Chat assistant reply；unsupported/unavailable local provider 仍阻止发送且不泄漏 token/path。

## 建议下一步

M18 已作为 Tool Runtime + Tool Catalog foundation complete candidate，保持其统一 catalog/broker/read-only UI 范围，不在 M18 临时塞入 workspace/shell/sandbox/browser/web/artifact runtime。

M18/M18.5 开始应转向工具能力和本地能力边界，而不是继续堆 Work 外壳或提前推进 Channels/Desktop/Plugin。当前 Loomi 已有 approval-gated tool-call projection、MCP stdio discovery/execution、RunContext、worker continuation、Work Plan View 和 local provider detection evidence，但实际可用工具仍很少：`runtime.get_current_time` 是演示级 builtin，本地 stdio MCP 已有执行通道但缺少产品化工具管理，Memory 是上下文能力不是执行工具，Work artifact metadata 只是展示面。因此 Work mode 要真正可用，后续必须先补“眼睛和手”。

推荐后续顺序：

1. **M21 Workspace Read Tools**：只读 `workspace.glob`、`workspace.grep`、`workspace.read`，绑定 workspace/project scope，限制敏感文件、路径穿越、文件大小和分页，并通过 M18 broker/catalog 注册。
2. **M22 Artifact Tools**：DB/object-store backed artifact create/read/list，让 Work mode 能生成和引用真实 report/plan/spec，而不是宿主文件写入。
3. **M23 Sandbox Code/Shell Tools**：在隔离 sandbox provider 中做 `exec_command`、`python_execute`、process continuation、stdout/stderr streaming、timeout、resource limits 和 approval。
4. **M24 Web Search/Fetch Tools**：做 `web.search`、`web.fetch`、citation、URL policy 和 SSRF/private-network guard。
5. **M25 Browser Automation Tools**：等 sandbox/browser provider 存在后再做 navigate/snapshot/screenshot/click/type/console/network。

Channels、Heartbeat、Desktop Runtime、Plugin、Activity Recorder、多 agent 编排和复杂 marketplace 都应排在工具 runtime、安全边界和 sandbox 之后。

## 开发要求

后续非平凡开发必须同步更新 `docs-site/`。M5.5 已更新 Settings 架构页、runbook、开发日志和验证命令；M6 已新增 worker/job pipeline API、架构页、runbook、开发日志、Spec Kit 状态和验证命令；013 已记录 M8 closeout passed 和 retry/backoff 补丁；M6.5 已新增 Provider Test Console/Background tasks 架构页、本地 provider testing runbook、开发日志和验证命令；M7 已新增并更新 tool-call approval architecture/API/runbook/devlog，并新增 tool-result continuation architecture/devlog；M9 foundation 已更新 worker-job pipeline 架构/API、本地 M9 runbook、devlog 和 Spec Kit 状态；M10 foundation 已新增 persona architecture/API/runbook/devlog，并更新当前状态与 Spec Kit 状态；M11 foundation 已新增 MCP stdio architecture/API/runbook/devlog，并更新当前状态 与 Spec Kit 状态；M12/M12.5 已新增并补齐 MCP approval execution architecture/API/runbook/devlog、真实本地 smoke closeout 证据，并更新当前状态与 Spec Kit 状态；M13/M13.5 Memory Foundation 已新增 memory architecture/API/runbook/devlog、真实 PG/httpapi smoke closeout，并更新当前状态与 Spec Kit 状态；M14/M14.1 已补 memory management/audit UX contract、完整 Settings > Memory UX、seeded browser smoke 和 review fix 证据；M15 已补 Chat integrated smoke runbook/devlog 和 Spec Kit 状态；M16 已补 Work mode foundation architecture/API/runbook/devlog、roadmap 和 Spec Kit 状态；M17 已补 Work artifact evidence closeout runbook/devlog、architecture/API、roadmap 和 Spec Kit 状态；M18 已补 tool runtime catalog architecture/API/runbook/devlog、roadmap 和 Spec Kit 状态；M18.5 已补 local provider autodetect architecture/API/runbook/devlog、roadmap 和 Spec Kit 状态；M19 已补 local provider opt-in architecture/API/runbook/devlog、roadmap 和 Spec Kit 状态；M20 已补 local codex execution architecture/API/runbook/devlog、roadmap 和 Spec Kit 状态。
