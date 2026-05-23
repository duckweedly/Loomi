---
title: ADR 0001：使用 Starlight 建立 Loomi 文档站
description: 记录 Loomi 选择 Starlight 作为技术文档站的原因。
---

## 状态

Accepted

## 背景

Loomi 需要一个随开发同步更新的技术文档站，用来查看路线、架构、规格、开发日志、技术决策和验证结果。文档需要与代码同仓库，方便 AI Agent 在实现代码时同步修改。

## 决策

使用 Astro Starlight 在 `docs-site/` 下建立独立文档站。文档内容以 Markdown/MDX 为主，构建为静态站点，并使用 Starlight 默认的信息架构和搜索能力。

## 理由

Starlight 是专门面向技术文档的静态站点方案，默认样式、导航和搜索能力已经足够，不需要在 Loomi 早期投入大量文档站定制成本。

它与主项目的 `web/` 解耦，不会影响 React/Vite/Electron 产品代码。文档站可以独立安装依赖、独立构建和独立部署。

## 后果

后续非平凡开发必须同步更新 `docs-site/src/content/docs/`。如果未来需要公开发布，可以直接部署 `docs-site/dist` 到任意静态托管平台。
