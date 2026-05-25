---
title: M16 Work Mode Foundation
description: Work mode 的最小可用计划、进度和 artifact metadata 投影。
---

M16 把 Work mode 从模式壳推进到最小可用薄片：`Thread.mode = work` 的线程会在现有 ChatCanvas 主区域显示 Work Plan View。它不新增任务系统、worker queue、执行环境或 sandbox。

## 边界

- 复用现有 thread、message、run、event、Timeline 和右侧面板边界。
- Work Plan View 是只读投影，不是新的任务数据库。
- Progress 从当前 run events、messages 和 safe metadata 推导。
- Artifact 第一版只显示 metadata 和 markdown-like summary。
- 不执行文件，不提供 shell/browser/filesystem controls。

## 投影来源

Work Plan View 优先读取 run event metadata：

- `work_goal`
- `work_steps`
- `work_artifacts`

如果没有 metadata，UI 会用 thread title 和用户消息生成最小 fallback，避免空白。

## UI 行为

Work mode thread 的主区域先显示 Work Plan View，再显示原有 message history、assistant draft 和 tool approval controls。Chat mode 不挂载 Work Plan View。M16 中 Work mode composer 是只读禁用状态，避免把 plan/progress surface 误解成新的 Work execution runtime；已有 Chat run/runtime 仍只属于 Chat mode 路径。

Work Plan View 包含：

- goal
- ordered steps
- current status/detail
- artifact references
- recent progress events
- loading/error/empty states

## 安全

Artifact reference 只保留安全字段：title、type、source thread/run、summary、created/updated。secret-looking string 会被 `[redacted]` 替换；command/path/browser/filesystem/executable hints 不作为可操作内容显示。

## 非目标

M16 不做 sandbox、shell/filesystem/browser automation、activity recorder、multi-agent、plugin marketplace、real artifact runtime 或 worker queue rewrite。
