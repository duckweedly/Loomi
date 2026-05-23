---
title: Loomi Constitution 摘要
description: Loomi Spec Kit 项目宪法的阅读入口。
---

完整宪法位于 `.specify/memory/constitution.md`。这里保留面向阅读的摘要，方便在文档站中检索。

## 核心原则

Loomi 坚持机制对标、表达自研。可以学习公开可观察的产品机制和演进顺序，但不能复制品牌、视觉、文案、图标、私有命名、私有接口或非公开结构。

每个阶段必须交付可运行的纵向薄片。功能不算完成，直到它有可演示、可测试或可观察的结果。

核心链路优先于平台复杂度。默认顺序是项目边界与术语、桌面感 Web 壳、API/DB 基座、Auth/Thread/Message、Run/Event/SSE、Web chat timeline、LLM gateway、工具调用、Worker/Job Queue、Pipeline/Context，再推进复杂桌面运行时和多 Agent 能力。

Agent 执行必须可观察。Run、工具调用、模型增量、状态转换、错误、取消和最终消息都应能通过事件与 timeline/debug 解释。

工具、文件、sandbox、本机活动记录和外部渠道都需要权限边界、审计和失败可见性。

## 文档同步要求

当前宪法要求未来功能工作通过 Spec Kit 产出可审查的规格、计划和任务。文档站进一步规定：非平凡开发必须同步更新文档站，包括架构、开发日志、ADR、API 或 runbook 中受影响的部分。
