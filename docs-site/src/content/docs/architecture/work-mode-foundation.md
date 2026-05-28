---
title: M16 Work Mode Foundation
description: Work mode 的最小可用计划、进度和 artifact metadata 投影。
---

M16 把 Work mode 从模式壳推进到最小可用薄片：`Thread.mode = work` 的线程会在现有 ChatCanvas 主区域显示 Work Plan View。M89 之后，用户界面不再暴露 Chat/Work 双入口；`Thread.mode` 暂时保留为后端兼容字段，前端能力逐步改为按目录、run metadata 和工具状态浮现。它不新增任务系统、worker queue、执行环境或 sandbox。

M17 把这个薄片从 mock-only seed 推进到可重复 evidence closeout：本地开发可通过 `LOOMI_SEED_SCENARIO=m17-work-artifact go run ./cmd/loomi-seed` 创建或复用 Work thread、seed message、current run 和 work metadata event。前端仍只从现有 thread/message/run/event replay 投影 Work Plan View，不新增生产事件写接口。

## 边界

- 复用现有 thread、message、run、event、Timeline 和右侧面板边界。
- Work Plan View 是只读投影，不是新的任务数据库。
- Progress 从当前 run events、messages 和 safe metadata 推导。
- Artifact 第一版只显示 metadata 和 markdown-like summary。
- 不执行文件，不提供 shell/browser/filesystem controls。
- M17 seed 仅是 local-dev/test evidence path，不是通用生产写接口。

## 投影来源

Work Plan View 优先读取 run event metadata：

- `work_goal`
- `work_steps`
- `work_artifacts`

如果没有 metadata，UI 会用 thread title 和用户消息生成最小 fallback，避免空白。

M17 local seed 会写入 `work.plan.updated` progress event，并在 metadata 中带上 `m17_seed = m17-work-artifact`，用于重复运行时复用同一个 evidence event，避免重复可见证据。

M22 runtime 会在 Work thread 的工具请求和工具成功执行后追加 durable `work.todo.updated` progress event。该事件由后端从已持久化的 tool-call lifecycle events 派生，包含安全的 `todo_items`、`updated_by = runtime` 和 redaction flag。M89 前端投影按安全 metadata 显示计划/todo/artifact，不再要求用户先切换到 Work mode。

## UI 行为

当前主聊天入口保持一套 ChatCanvas 和 Composer。目录选择、目录状态、tool approval controls 和 run timeline 都在同一个会话体验中出现；计划/todo/artifact 只在安全 run metadata 存在时投影，避免空线程出现额外面板。M89 不改变后端工具权限，后续会继续把工具可用性从 legacy mode gate 迁到目录和 run capability gate。

M91 补齐目录浏览的第一类工具选择策略：当用户问“这个目录里有什么”“帮我分类梳理下载目录”这类目录盘点问题时，provider schema 和 Work prompt 都要求优先使用 `workspace.tree_summary` 或 `workspace.list_directory`，再按需要 `workspace.read` 少量代表性文件；`workspace.grep` 只用于内容搜索，不再作为目录 inventory 的默认入口。目录工具共享 workspace root、敏感路径、symlink escape、生成目录跳过和 bounded continuation 边界，RunRail 只显示“读取目录 / 目录概览”这类安全摘要。

Work Plan View 包含：

- goal
- ordered steps
- current status/detail
- artifact references
- recent progress events
- loading/error/empty states

## 安全

Artifact reference 只保留安全字段：id、title、type、source thread/run、summary、created/updated 和 redaction marker。secret-looking string 会被 `[redacted]` 替换；command/path/file/shell/browser/filesystem/execute/url hints 不作为可操作内容显示。Artifact cards 没有 execute/open/run/download 控件。

## 非目标

M16/M17 不做 sandbox、shell/filesystem/browser automation、activity recorder、multi-agent、plugin marketplace、real artifact runtime 或 worker queue rewrite。
