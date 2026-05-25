---
title: 当前状态
description: Loomi 当前已完成内容和建议下一步。
---

## 已有基础

当前仓库已经有 Web/Electron 前端壳、Go API/DB 基座、本地 identity/thread/message、M4 run/event/SSE、M3.5 前端 Agent runtime 骨架、M5 LLM Gateway 的 backend/provider/frontend 基础切片、M5.5 Settings Placeholder 的前端设置占位面、M6 Worker Job Pipeline 的 queued background execution MVP、M6.5 Real Testing Console and Background UX、M7 Tool Call Approval Execution + Tool Result Continuation 的最小闭环、M9 RunContext + Pipeline foundation 的最小薄片、M10 Persona/Skill foundation 的 persona 选择与版本快照薄片、M11 MCP stdio foundation 的本地 discovery/read-only candidate 薄片、M12/M12.5 MCP approval-gated execution 的单工具本地 stdio 执行薄片和真实本地 smoke closeout 证据，以及 M13 Memory Foundation 的 PG-backed memory 第一实现薄片。

M13 Memory Foundation 的 Spec Kit 目录为 `specs/019-memory-foundation/`，M13.5 closeout 目录为 `specs/020-memory-real-pg-smoke-closeout/`。当前已实现第一薄片：PG v1 memory entries/search/write proposal、RunContext safe memory snapshot、approval-gated agent memory writes、用户查看/检索/删除记忆的最小 API/UI、tombstone deletion、redaction 和 source-run audit events。M13.5 已补真实 Postgres/httpapi smoke，覆盖 migrated `memory_entries`/`memory_write_proposals`、proposal/approve/list/search/RunContext/delete/idempotency/out-of-scope/redaction。它明确不包含向量库/embedding/RAG、OpenViking provider、marketplace/plugin、sandbox/browser/activity recorder、多 agent 长期记忆自动化、worker/job queue rewrite 或 MCP rewrite。

M14 Memory Management Audit UX 的 Spec Kit 目录为 `specs/021-memory-management-audit-ux/`。当前已推进到 full UX complete candidate：thread-scoped memory detail/delete 需要匹配 thread/source scope，thread list/search 缺少 `scope_id` 会返回参数错误，terminal run 后 memory proposal/approve/deny/delete audit 写入 durable `memory_audit_events`，redaction 覆盖 `/home`、Windows path、stdout/stderr、tool output、provider trace、key/env markers，list/search API 与前端 client 统一到 `query/q`、`limit`、`scope_type`、`scope_id`、`source_thread_id`、`source_run_id`、`source_type`、`include_tombstoned`。Settings > Memory 已覆盖真实 approved/tombstoned list、search、grounded filters、safe detail panel、delete confirmation、loading/empty/error/deleted states、thread-scoped detail/delete context guard、真实 `/v1/memory/audit` history、backend unavailable/error/empty 不造假，以及 seeded browser smoke。M14 仍不包含 distill 自动总结、OpenViking、vector/embedding/RAG、browser/activity recorder ingestion、MCP、worker queue、sandbox 或多 agent 重写。

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

## 建议下一步

下一步适合做 M14 code review / closeout，或进入后续 memory distill 之前的规划评审；不要把 M14 扩展成 distill、worker/job queue、MCP、sandbox/browser/activity recorder、marketplace/plugin 或多 agent 长期自动化重做。

## 开发要求

后续非平凡开发必须同步更新 `docs-site/`。M5.5 已更新 Settings 架构页、runbook、开发日志和验证命令；M6 已新增 worker/job pipeline API、架构页、runbook、开发日志、Spec Kit 状态和验证命令；013 已记录 M8 closeout passed 和 retry/backoff 补丁；M6.5 已新增 Provider Test Console/Background tasks 架构页、本地 provider testing runbook、开发日志和验证命令；M7 已新增并更新 tool-call approval architecture/API/runbook/devlog，并新增 tool-result continuation architecture/devlog；M9 foundation 已更新 worker-job pipeline 架构/API、本地 M9 runbook、devlog 和 Spec Kit 状态；M10 foundation 已新增 persona architecture/API/runbook/devlog，并更新当前状态与 Spec Kit 状态；M11 foundation 已新增 MCP stdio architecture/API/runbook/devlog，并更新当前状态与 Spec Kit 状态；M12/M12.5 已新增并补齐 MCP approval execution architecture/API/runbook/devlog、真实本地 smoke closeout 证据，并更新当前状态与 Spec Kit 状态；M13/M13.5 Memory Foundation 已新增 memory architecture/API/runbook/devlog、真实 PG/httpapi smoke closeout，并更新当前状态与 Spec Kit 状态；M14 已补 memory management/audit UX contract、完整 Settings > Memory UX 和 seeded browser smoke 证据。
