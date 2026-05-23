# Loomi Docs

Loomi 的 Starlight 技术文档站，用来沉淀产品路线、架构设计、Spec Kit 规格、开发日志、ADR、API 和运行手册。

## 使用 Bun

所有文档站命令统一使用 Bun。

```bash
bun install
bun run dev
bun run build
bun run preview
```

## 本地开发

```bash
cd /Users/xuean/Repos/personal-projects/Loomi/docs-site
bun run dev
```

默认地址是 `http://localhost:4321/`。如果端口被占用，Astro 会自动换端口，以终端输出为准。

## 内容目录

文档页面位于：

```text
src/content/docs/
```

主要分区：

```text
roadmap/       阶段路线和当前状态
architecture/  架构边界、状态流、事件模型和安全边界
spec-kit/      Spec Kit 工作流、constitution 摘要和功能规格状态
devlog/        开发日志和验证结果
adr/           技术决策记录
api/           API、事件 payload、数据模型和兼容性说明
runbooks/      启动、验证、排错和运维手册
workflow/      开发流程和文档同步规则
```

## 开发规则

Loomi 的非平凡开发必须同步更新文档站。完整规则见：

```text
src/content/docs/workflow/how-to-develop.md
src/content/docs/workflow/documentation-sync.md
```

完成文档变更后至少运行：

```bash
bun run build
```
