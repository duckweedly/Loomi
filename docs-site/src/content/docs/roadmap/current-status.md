---
title: 当前状态
description: Loomi 当前已完成内容和建议下一步。
---

## 已有基础

当前仓库已经有 Web/Electron 前端壳，包含线程侧栏、聊天画布、Composer、Run Timeline、工具调用卡片、顶部栏、右侧面板数据和相关测试文件。

Spec Kit 已接入仓库，并写入 Loomi constitution。文档站已作为 `docs-site/` 独立子项目创建，用于长期记录技术细节。

## 建议下一步

建议第一个正式走 Spec Kit 的功能是 M2 API 与数据库基座。这个阶段能把 Loomi 从可展示 UI 推进到真实服务边界，为后续 Thread/Message、Run/Event/SSE、LLM Gateway 和 Worker 打基础。

## 开发要求

后续非平凡开发必须同步更新 `docs-site/`。如果实现 M2，应至少更新 API/DB 架构页、开发日志、验证命令，并在必要时新增 ADR。
