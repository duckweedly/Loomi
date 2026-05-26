---
title: 当前状态
description: Loomi 当前已完成内容和建议下一步。
---

## 已有基础

当前仓库已经有 Web/Electron 前端壳、Go API/DB 基座、本地 identity/thread/message、M4 run/event/SSE、M3.5 前端 Agent runtime 骨架、M5 LLM Gateway 的 backend/provider/frontend 基础切片、M5.5 Settings Placeholder 的前端设置占位面、M6 Worker Job Pipeline 的 queued background execution MVP、M7 Tool Call Approval Core 的 approval-gated current-time execution slice、M8 Safe Workspace Read Tools 的 approval-gated `glob` / `grep` / `read_file` 基础切片、M9 Safe Workspace Write Tools 的 approval-gated `write_file` / `edit` 文本 mutation 基础切片、M10 Safe Workspace Exec Command 的 approval-gated argv command 基础切片、M11 Tool Catalog Visibility 的只读工具目录切片、M12 Todo Write Planning Tool 的结构化计划工具切片，以及 M13 MCP Call Tool Bridge 的最小 MCP-style bridge 切片。

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
- M7：approval-blocked tool-call projection、`runtime.get_current_time` allowlist/schema、tool lifecycle events、safe metadata redaction、blocked/resumable worker diagnostics counters、frontend tool-event replay mapping、provider tool-call conversion、scoped tool-call read/approve/deny API、approval UI controls、approval 后 worker resume、approved current-time execution、result/error/cancel terminal projection、ToolCallCard 终态展示、pending/approved/executing cancellation precedence、worker duplicate-terminal guard、RunRail/Timeline 混合模型/工具 replay polish。
- M8：approval-required `workspace.glob`、`workspace.grep`、`workspace.read_file` allowlist；workspace root containment；path traversal/sensitive path/binary file rejection；bounded relative-path result summaries；worker execution reuse；ToolCallCard workspace grep result rendering。
- M9：approval-required `workspace.write_file`、`workspace.edit` allowlist；mutation root containment；symlink escape/sensitive path/missing parent rejection；bounded UTF-8 writes；exact single replacement edit；failed edit no-mutation；worker workspace root injection；ToolCallCard write/edit result rendering。
- M10：approval-required `workspace.exec_command` allowlist；argv-only/no-shell execution；workspace-contained cwd；timeout bounds；stdout/stderr bounded summaries；shell/destructive command rejection；ToolCallCard exec result rendering。
- M11：`GET /v1/tools/catalog` 只读工具目录；七个当前 allowlisted tools 的 deterministic metadata；Settings > Tools 真实只读面板；approval/risk/side-effect/safety-class 可见性；mock/real API catalog mapping。
- M12：approval-required `runtime.todo_write` allowlist；bounded todo items；`pending` / `in_progress` / `completed` 状态校验；worker result_summary 计数；ToolCallCard todo plan rendering；Settings Tools catalog planning tool visibility。
- M13：approval-required `mcp.call_tool` allowlist；固定 `local.echo` MCP-style bridge；bounded message validation；secret-looking message rejection；worker result_summary echo；ToolCallCard nested MCP summary rendering；Settings Tools catalog MCP bridge visibility。

## 建议下一步

下一步适合继续补 workspace/project context、project files context assembly，或扩展 MCP/browser automation 的最小只读切片。多 Agent / spawn-agent、RAG/memory、Activity Recorder 等能力仍应通过后续 Spec Kit 功能单独推进。

## 开发要求

后续非平凡开发必须同步更新 `docs-site/`。M5.5 已更新 Settings 架构页、runbook、开发日志和验证命令；M6 已新增 worker/job pipeline API、架构页、runbook、开发日志、Spec Kit 状态和验证命令；M7 已新增 tool-call approval architecture/API/runbook/devlog，并记录 worker execution、cancellation、mixed replay 和 browser smoke 验证；M8 已新增 workspace read tools architecture/API/runbook/devlog，并记录 read-only、approval-required、bounded output 和 sensitive path 边界；M9 已新增 workspace write tools architecture/API/runbook/devlog，并记录 approval-required、bounded text mutation、exact edit 和 no-mutation failure 边界；M10 已新增 workspace exec command architecture/API/runbook/devlog，并记录 approval-required、argv-only、timeout、bounded output 和危险命令拒绝边界；M11 已新增 tool catalog architecture/API/runbook/devlog，并记录只读 metadata、redaction 和 Settings visibility 边界；M12 已新增 todo write planning architecture/API/runbook/devlog，并记录 bounded planning metadata、approval 和 timeline visibility；M13 已新增 MCP call tool bridge architecture/API/runbook/devlog，并记录 fixed local MCP-style allowlist、approval、catalog 和 timeline visibility。后续真实 provider/browser smoke 结果应继续追加到相关 devlog。
