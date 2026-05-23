---
title: 以后如何开发 Loomi
description: 使用 Spec Kit、Bun 和文档站推进 Loomi 开发的简单教程。
---

这篇文档是 Loomi 后续开发的默认操作指南。目标是让每一步开发都留下可追踪的规格、技术计划、实现结果和文档记录。

## 日常打开文档站

在仓库根目录执行：

```bash
cd /Users/xuean/Repos/personal-projects/Loomi/docs-site
bun run dev
```

默认地址是 `http://localhost:4321/`。如果端口被占用，Astro 会自动换一个端口，终端输出里会显示实际地址。

验证文档站是否能构建：

```bash
cd /Users/xuean/Repos/personal-projects/Loomi/docs-site
bun run build
```

以后文档站相关命令统一使用 Bun，不使用 npm。

## 开发一个新功能的标准流程

非平凡功能默认走 Spec Kit。Loomi 当前安装了 Claude Code 和 Codex CLI 两套 skills，GitHub Copilot 旧集成已被 Claude/Codex 集成替换。默认集成是 Claude Code。

Claude Code 使用 `.claude/skills/`，Codex 使用 `.agents/skills/`。两者的命令名都是横线格式，不是点号格式：

```text
speckit-specify
speckit-clarify
speckit-plan
speckit-tasks
speckit-implement
```

在 Claude Code 里，可以直接用自然语言触发 skill：

```text
使用 speckit-specify：实现 M2 API 与数据库基座，提供 healthz/readyz、配置加载、PostgreSQL 连接、migration up/down，并保持现有 web mock UI 不受影响。
```

在 Codex CLI 里，按 Codex 的技能调用习惯使用项目内 `.agents/skills/speckit-*`，也可以直接说：

```text
按 speckit-specify 流程，为 M2 API 与数据库基座写功能规格。
```

推荐顺序是：

```text
speckit-specify
speckit-clarify
speckit-plan
speckit-tasks
speckit-implement
```

第一步先写清楚要做什么，不急着写代码。如果规格里有歧义，继续执行 `speckit-clarify`。等关键产品问题和边界收敛后，再用 `speckit-plan` 写技术计划，随后用 `speckit-tasks` 拆任务，最后用 `speckit-implement` 实现。

如果功能比较大，优先只实现第一个可运行薄片，不要一次性吞完整个平台能力。

## 每次开发必须同步哪些文档

除非只是修拼写、格式化或无行为变化的小改动，否则实现时必须同步更新 `docs-site/src/content/docs/`。

如果改了模块边界、状态流、事件模型、权限、安全或运行方式，更新 `architecture/`。

如果改了接口、事件 payload、数据结构或兼容性，更新 `api/`。

如果改了启动命令、环境变量、验证命令、migration 或排错方式，更新 `runbooks/`。

如果做了重要技术取舍，例如新增依赖、选择数据库方案、改变运行模式、引入 worker 或 sandbox，新增或更新 `adr/`。

每完成一个阶段性任务，更新 `devlog/`，记录做了什么、怎么验证、还剩什么。

如果功能对应 Spec Kit 规格、计划或任务，更新 `spec-kit/` 中的说明或链接。

## 一次完整开发结束前的检查清单

完成前至少检查这些事情：

```text
代码是否实现了 spec 里的验收标准
是否跑了相关测试、构建或 smoke test
如果文档有变化，是否执行了 bun run build
是否更新了开发日志
是否记录了重要技术取舍
是否说明了已知限制和下一步
```

如果某个验证无法运行，要在最终说明里明确写出原因，不能假装通过。

## 推荐的开发节奏

Loomi 当前最适合按阶段推进。先保持 Web/Electron 壳可运行，再补 M2 API 与 DB 基座，然后做 Thread/Message，接着做 Run/Event/SSE。等这些核心链路可观察、可测试之后，再做 LLM Gateway、工具调用、Worker、Pipeline、桌面运行时和多 Agent 能力。

不要提前把复杂平台能力拉进来。每一步都应该能在文档站里看到：为什么做、怎么设计、怎么验证、还有什么没做。

## 给 AI Agent 的默认要求

在 Loomi 仓库执行非平凡开发时，AI Agent 必须先读 `.specify/memory/constitution.md`，遵循其中的项目原则，并主动同步更新文档站。不需要用户单独提醒“记得写文档”。

当文档被更新后，完成前必须在 `docs-site/` 下运行：

```bash
bun run build
```

这条规则已经写入 `.specify/memory/constitution.md` 和 `.github/copilot-instructions.md`。
