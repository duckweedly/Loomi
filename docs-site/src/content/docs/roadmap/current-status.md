---
title: 当前状态
description: Loomi 当前已完成内容和建议下一步。
---

## 已有基础

当前仓库已经有 Web/Electron 前端壳、Go API/DB 基座、本地 identity/thread/message、M4 run/event/SSE、M3.5 前端 Agent runtime 骨架、M5 LLM Gateway 的 backend/provider/frontend 基础切片、M5.5 Settings Placeholder 的前端设置占位面、M6 Worker Job Pipeline 的 queued background execution MVP、M6.5 Real Testing Console and Background UX，以及 M7 Tool Call Approval Core 的 Phase 2 foundation。

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
- M6.5：Provider Test Console、real mode provider unavailable guidance、read-only Background tasks observer、M6 worker/job Timeline labels、Composer 状态收口和本地真实测试 runbook。
- M7 Phase 2 foundation + US1 observable request：approval-blocked tool-call projection、`runtime.get_current_time` allowlist/schema、tool lifecycle events、safe metadata redaction、worker diagnostics counters、frontend tool-event replay mapping、provider tool-call conversion、scoped tool-call read API 和 approval-required UI placeholder。

## 建议下一步

下一步适合继续完成 M7 的 approve/deny API、approval UI actions 和 approved `runtime.get_current_time` worker execution。Desktop runtime、多 Agent、RAG/memory、持久化 settings、provider 管理、shell/filesystem/MCP/browser automation 等能力仍应通过后续 Spec Kit 功能单独推进。

## 开发要求

后续非平凡开发必须同步更新 `docs-site/`。M5.5 已更新 Settings 架构页、runbook、开发日志和验证命令；M6 已新增 worker/job pipeline API、架构页、runbook、开发日志、Spec Kit 状态和验证命令；M6.5 已新增 Provider Test Console/Background tasks 架构页、本地 provider testing runbook、开发日志和验证命令；M7 Phase 2 已新增 tool-call approval architecture/API/runbook/devlog。后续真实 provider/browser smoke 结果应继续追加到相关 devlog。
