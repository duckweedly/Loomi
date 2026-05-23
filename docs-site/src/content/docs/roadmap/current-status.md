---
title: 当前状态
description: Loomi 当前已完成内容和建议下一步。
---

## 已有基础

当前仓库已经有 Web/Electron 前端壳、Go API/DB 基座、本地 identity/thread/message、M4 run/event/SSE，以及 M3.5 前端 Agent runtime 骨架。

Spec Kit 已接入仓库，并写入 Loomi constitution。文档站已作为 `docs-site/` 独立子项目创建，用于长期记录技术细节。

已完成的可运行薄片：

- M1：桌面感 Web 壳和基础面板。
- M2：API、配置、PostgreSQL、migration、health/readiness 和 diagnostics 基座。
- M3：本地 identity、thread/message API、seed、mock/real API 切换。
- M3.5：Chat Canvas 状态机、mock runtime scripts、RunRail/Timeline/Agent badge 联动。
- M4：持久化 run/event、history-first SSE、local simulated run、stop/already-terminal 语义。

## 建议下一步

下一步适合推进 M5 LLM Gateway：在现有 thread/message 与 run/event/SSE 之上接入模型调用，把 model delta 映射成同一套 RuntimeEvent/RunEvent 语义，并继续保持工具调用、worker/job queue、desktop runtime 和多 Agent 能力 deferred。

## 开发要求

后续非平凡开发必须同步更新 `docs-site/`。如果实现 M5，应至少更新 API/架构页、runbook、开发日志、验证命令，并在必要时新增 ADR。
