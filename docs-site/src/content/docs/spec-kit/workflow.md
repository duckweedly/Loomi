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

## 当前功能：M13 MCP Call Tool Bridge

当前 Spec Kit 功能目录：

```text
specs/015-mcp-call-tool-bridge/
```

关键产物：

- `spec.md`：定义 approval-gated `mcp.call_tool` 用户故事、固定 `local.echo` bridge、安全边界和非目标。
- `plan.md`：确定复用 M7-M12 tool lifecycle、catalog、worker execution、ToolCallCard 和 Settings Tools。
- `research.md`：记录 minimal local MCP-style bridge、approval required、no external MCP processes 和 bounded message 决策。
- `data-model.md`：定义 MCP Tool Call Arguments、Local MCP Echo Tool 和 Tool Result。
- `contracts/mcp-call-tool-bridge.md`：定义 tool name、arguments、安全规则和结果形状。
- `quickstart.md`：记录 focused tests、full validation 和 Settings Tools smoke。
- `tasks.md`：按 setup、backend、frontend、documentation 和 validation 拆分任务。

## 近期已完成：M12 Todo Write Planning Tool

当前 Spec Kit 功能目录：

```text
specs/014-todo-write-planning-tool/
```

关键产物：

- `spec.md`：定义 approval-gated `runtime.todo_write` 用户故事、bounded todo items、安全边界和非目标。
- `plan.md`：确定复用 M7-M11 tool lifecycle、catalog、worker execution、ToolCallCard 和 Settings Tools。
- `research.md`：记录 runtime planning tool、approval required、no durable todo table 和 bounded items 决策。
- `data-model.md`：定义 Todo Item、Tool Arguments 和 Tool Result。
- `contracts/todo-write-planning-tool.md`：定义 tool name、arguments、安全规则和结果形状。
- `quickstart.md`：记录 focused tests、full validation 和 Settings Tools smoke。
- `tasks.md`：按 setup、backend、frontend、documentation 和 validation 拆分任务。

## 近期已完成：M11 Tool Catalog Visibility

当前 Spec Kit 功能目录：

```text
specs/013-tool-catalog-visibility/
```

关键产物：

- `spec.md`：定义只读工具目录 API、Settings > Tools 可见性、安全 redaction 和非目标。
- `plan.md`：确定复用 runtime tool registry、HTTP API、real/mock API client 和 Settings surface。
- `research.md`：记录 static catalog、read-only metadata、Settings read-only panel 和 redaction 决策。
- `data-model.md`：定义 Tool Catalog Entry、Tool Capability、Risk Level 和 Side Effect。
- `contracts/tool-catalog.md`：定义 `GET /v1/tools/catalog`、字段形状和安全边界。
- `quickstart.md`：记录 API smoke、Settings Tools smoke 和验证命令。
- `tasks.md`：按 setup、backend、frontend、documentation 和 validation 拆分任务。

## 近期已完成：M10 Safe Workspace Exec Command

当前 Spec Kit 功能目录：

```text
specs/012-safe-workspace-exec-command/
```

关键产物：

- `spec.md`：定义 approval-required `workspace.exec_command` 用户故事、argv-only/no-shell、安全拒绝和 bounded output。
- `plan.md`：确定复用 M7 approval、M8/M9 workspace boundary、worker resume、run events、SSE replay 和现有 frontend tool UI。
- `research.md`：记录 argv-only、shell/destructive command rejection、timeout 和 bounded output 决策。
- `data-model.md`：定义 Exec Command Tool、Arguments 和 Result。
- `contracts/workspace-exec-command.md`：定义 tool name、arguments、安全规则和结果形状。
- `quickstart.md`：记录本地 approve/deny、安全命令、危险命令、timeout、history replay 和验证命令。
- `tasks.md`：按 setup、backend、frontend、documentation 和 validation 拆分任务。

## 近期已完成：M9 Safe Workspace Write Tools

当前 Spec Kit 功能目录：

```text
specs/011-safe-workspace-write-tools/
```

关键产物：

- `spec.md`：定义 approval-required `workspace.write_file` 和 `workspace.edit` 用户故事、安全边界和 no-mutation failure。
- `plan.md`：确定复用 M7 approval、M8 workspace boundary、worker resume、run events、SSE replay 和现有 frontend tool UI。
- `research.md`：记录 write_file 先于 shell、exact replacement、parent directory required 和 text-only bounded content 决策。
- `data-model.md`：定义 Workspace Write Tool、Tool Arguments、Tool Results 和 Sensitive Path Policy。
- `contracts/workspace-write-tools.md`：定义 lifecycle、tool names、安全规则和结果形状。
- `quickstart.md`：记录本地 approve/deny、write、edit、unsafe path、history replay、browser smoke 和验证命令。
- `tasks.md`：按 setup、backend、frontend、documentation 和 validation 拆分任务。

## 近期已完成：M8 Safe Workspace Read Tools

当前 Spec Kit 功能目录：

```text
specs/010-safe-workspace-read-tools/
```

关键产物：

- `spec.md`：定义 approval-required `workspace.glob`、`workspace.grep`、`workspace.read_file` 用户故事和安全边界。
- `plan.md`：确定复用 M7 tool-call lifecycle、worker resume、run events、SSE replay 和现有前端 tool UI。
- `research.md`：记录 approval、workspace root containment、sensitive path denial、bounded summaries 和非目标决策。
- `data-model.md`：定义 Workspace Read Tool、Workspace Root、Tool Arguments、Tool Results 和 Sensitive Path Policy。
- `contracts/workspace-read-tools.md`：定义 lifecycle、tool names、安全规则和结果形状。
- `quickstart.md`：记录本地 approval/deny、unsafe path、history replay、browser smoke 和验证命令。
- `tasks.md`：按 setup、backend、frontend、documentation 和 validation 拆分任务。

## 近期已完成：M7 Tool Call Approval Core

当前 Spec Kit 功能目录：

```text
specs/009-tool-call-approval-core/
```

关键产物：

- `spec.md`：定义 approval-gated tool-call lifecycle、approve/deny、current-time MVP tool 和 UI replay。
- `plan.md`：确定复用 M5 gateway、M6 worker/job pipeline、run/event/SSE 和 ToolCallCard/RunRail/Timeline。
- `research.md`：记录 allowlisted tool registry、`runtime.get_current_time`、tool_calls projection、worker block/resume 和 redaction 决策。
- `data-model.md`：定义 Tool Call、Tool Definition、Approval Decision、Tool Result、Run Event extensions 和 Worker Block/Resume State。
- `contracts/`：定义 HTTP、tool lifecycle、worker resume、frontend tool UI 和 docs update contracts。
- `quickstart.md`：记录 fake/model tool request、approve/deny、execution、cancellation、replay 和 browser smoke。
- `tasks.md`：按 setup、foundation、US1-US4 和 polish 拆分并跟踪实现任务。

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
