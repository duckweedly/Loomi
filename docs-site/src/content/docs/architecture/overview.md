---
title: 架构总览
description: Loomi 的当前模块边界和演进方向。
---

Loomi 当前采用 web-first 的开发方式：先在 `web/` 中建立桌面感产品壳和可解释的 Agent 执行界面，再逐步补齐 Go 后端、数据库、事件流、Worker 和桌面运行时。

## 当前目录意图

`web/` 承载当前可运行的前端与 Electron 壳。它可以暂时使用 mock 数据，但必须保留 API client 边界。

`docs/` 承载学习路线、公开机制拆解和阶段任务说明，是项目早期的源材料。

`docs-site/` 是可浏览文档站，用来把路线、架构、规格、开发日志和技术决策组织成长期可检索知识库。

`.specify/` 和 `specs/` 承载 Spec Kit 的项目宪法、模板、规格、计划和任务。未来非平凡功能应优先通过 Spec Kit 产出可审查的开发材料。

`cmd/`、`internal/`、`services/` 是后续 Go API、内部应用代码和服务边界的预留位置。

## 核心架构原则

Loomi 优先建设从用户输入到 Agent 执行结果的最短闭环。核心链路包括 thread、message、run、event、SSE、LLM gateway、tool call、worker job 和 timeline/debug 可观测界面。

任何高级平台能力都应该依赖这些基础设施，而不是绕过它们单独实现。

## 可观测性要求

Agent 执行必须能够解释。一次 run 至少应能回答：谁触发了它、输入是什么、进入了哪些状态、调用了哪些模型或工具、产生了哪些事件、哪里失败、如何取消、最终消息如何生成。

这些信息应同时体现在数据库事件、API/SSE 输出和 UI timeline/debug 面板里。
