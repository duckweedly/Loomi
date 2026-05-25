---
title: Loomi 阶段路线图
description: Loomi 从可视化壳到 Agent 平台核心链路的阶段拆分。
---

Loomi 采用薄片式演进：每个阶段都必须能运行、能测试、能解释，避免在核心链路未稳定前提前堆叠复杂平台能力。

## M1：桌面感 Web 壳

目标是先形成可观察、可演示的产品外壳。当前 `web/` 已包含 React、Vite、TypeScript 和 Electron 相关入口，重点是线程列表、消息画布、Composer、Run Timeline、工具调用卡片和右侧调试面板等前端体验。

右侧 UI 使用统一的浮动卡片 / 右滑卡片语言，详见 [Web Shell Panels](/architecture/web-shell-panels/)。这一阶段允许使用 mock 或本地状态，但必须保留清晰 API client 边界，避免后续接入真实后端时重写 UI。

## M2：API 与数据库基座

目标是建立真实后端边界，包括 healthz/readyz、配置加载、PostgreSQL 连接、migration up/down、基础错误处理和 smoke test。

这一阶段不追求完整 Agent Loop，而是让前端有可以替换 mock 的真实服务边界。

## M3：Auth、Thread、Message

目标是沉淀最小产品数据模型：用户、会话、线程、消息、附件或上下文引用。所有状态变化都应可追踪，并为后续 run/event 模型预留关联关系。

## M4：Run、Event 与 SSE

目标是让一次 Agent 执行变成可持久化、可流式观察的 run。事件应记录状态转换、模型输出、工具调用、错误、取消和最终消息。Web UI 的 timeline/debug 面板应能从真实事件流驱动。

## M5：LLM Gateway 与工具调用

目标是统一模型供应商调用、流式输出、错误处理和工具调用协议。工具执行必须有权限边界和审计记录。

## M6：Worker、Job Queue 与 Pipeline

目标是把 run 从同步请求推进到可恢复、可重试、可取消的后台执行模型。Worker 需要租约、所有权、幂等和失败恢复机制。

## M7-M17：核心执行闭环、Memory 与 Work 壳

M7-M17 的重点是先把 Chat/Work 的可观察执行链闭合：tool approval、worker continuation、RunContext、persona、MCP、memory、Chat real smoke、Work Plan View 和 Work artifact evidence。

这一段允许 Work mode 先形成计划、进度和安全 artifact metadata 的可见面，但不能把 UI 壳误认为真实工作能力。只要缺少工具层，Work mode 就仍然只能展示“将要做什么”，不能真正读项目、生成产物、执行命令或浏览网页。

## M18-M24：工具能力优先

M18 之后的路线改为 Tool Runtime 优先，而不是直接推进 Channels、Desktop Runtime、Plugin 或 Activity Recorder。原因很简单：没有工具，Agent 没有眼睛和手，Work mode 无法可用。

工具路线按风险递增：

- **M18 Tool Runtime + Tool Catalog**：建立一等公民工具目录、risk level、approval policy、schema hash、source、enabled state 和统一 broker/executor。现有 `runtime.get_current_time` 与本地 stdio MCP executor 必须接入同一 runtime。
- **M19 Workspace Read Tools**：先做只读 `workspace.glob`、`workspace.grep`、`workspace.read`，让 Work 能安全理解项目。禁止宿主文件写入和 shell 执行。
- **M20 Artifact Tools**：让 Agent 能创建、读取和引用 DB/object-store backed artifact，例如 markdown/text plan、report、spec；artifact 不是宿主文件写入。
- **M21 Sandbox Code/Shell Tools**：在隔离 provider 中实现 `sandbox.exec_command`、`sandbox.python_execute`、continuation/terminate、stdout/stderr streaming、timeout 和 resource limits。
- **M22 Web Search/Fetch Tools**：实现可审计的 `web.search`、`web.fetch`、citation、URL policy 和 SSRF/private-network guard。
- **M23 Browser Automation Tools**：在 sandbox/browser provider 之后再实现 navigate/snapshot/screenshot/click/type/console/network。
- **M24 Tool Settings + MCP Management**：补完整 Settings > Tools、MCP server list、local stdio config UI、provider health、enable/disable 和 approval policy override。

## M25+：桌面、运营、渠道和插件

Desktop Runtime、SQLite/Bridge、Admin Console、Channels/Heartbeat、Plugin、Activity Recorder 和 Release/CI/Docs 都排在 M25 之后。它们依赖前面的工具管理、安全边界、审计和 sandbox 能力，不应在工具层薄弱时提前推进。
