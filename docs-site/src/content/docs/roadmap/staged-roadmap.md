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

## M7+：桌面运行时与平台能力

在核心链路稳定后，再推进 Desktop Runtime、Channels、Activity Recorder、Sandbox、插件、多 Agent、长期记忆和更复杂的可观测能力。
