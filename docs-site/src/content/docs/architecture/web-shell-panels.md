---
title: Web Shell Panels
description: Loomi M1 桌面壳的右侧浮动卡片、右滑卡片和顶栏入口规则。
---

Loomi 的 M1 Web 壳把右侧区域拆成两类表面：运行概览的浮动卡片，以及从右侧滑出的详情卡片。两者都属于同一套桌面面板语言，不能再引入一套全宽 drawer。

## 左侧导航原则

左侧栏只承载当前模式的主要工作对象。Chat 模式显示创建对话和 chat thread 列表；Work 模式显示 Projects、Scheduled 和 work thread 列表。两个模式的最近会话列表按 `thread.mode` 分开过滤，不能把 Work 任务混进 Chat 的最近会话，也不能把 Chat 对话混进 Work 的项目/任务流。Search 已经属于顶部 chrome，因此不在左侧重复出现。

Context、Files、Memory、Run、Preview、Terminal、Diff 和其他工具详情属于右侧辅助栏。它们依附于当前 thread 或 run，不是左侧的主导航对象。

## 右侧表面分工

### Run rail 浮动卡片

Run rail 是默认可见的运行概览层，用来展示当前 run 的 progress、相关文件和 context connector。它保持浮动白卡形态，视觉上压在主画布之上，而不是嵌入成普通右栏。

Progress 卡片顶部有一个紧凑的 agent state motion 徽章。它不是完整 artifact runtime，也不是独立聊天空状态，而是把当前 run/event 的状态投影成一眼可读的运动信号：`idle`、`thinking`、`tool`、`speaking`、`confirm`、`done`、`error`。当前 M3 还没有真实 run/event/SSE，所以前端从 mock run 和 event type 推导状态；后续 M4 引入持久 run events 后，只需要替换状态来源，右侧表面语言不变。

这张卡片内部不提供额外关闭按钮。用户如果要收起运行概览，使用顶栏的 run details 按钮；这样入口和退出都保持在同一层级，避免在卡片底部出现孤立的控制点。

### Right tool 详情卡片

当用户从 run rail 或顶栏右侧菜单打开某个具体内容，例如 Preview、Files 或 Terminal，界面使用窄的 right tool card 从右侧滑出。它和 run rail 共享圆角、浅色面、细分隔线和右滑动效，所以用户会把它理解成同一套右侧工作区。

Preview 承接原先的 artifact placeholder，但不再打开旧的全宽 artifact drawer。这个限制很重要：宽 drawer 会像另一套产品语言，遮住主画布太多内容，也和 Run rail 的浮动卡片关系不清。

## 左下 Settings popover

Settings 入口是左侧栏底部的轻量命令菜单，不是完整设置页。M1 只保留三行：`Settings` 作为未来设置面板入口，`Theme` 使用 Light/Dark segmented toggle 表示当前主题，`Update` 显示当前更新状态并触发 refresh。帮助、退出登录、用户资料卡和 run 状态摘要都不放在这里，避免这个底部入口变成臃肿账户菜单。

## 交互规则

- 顶栏 run details 按钮负责打开或收起 Run rail。
- 顶栏 right tools 按钮在没有详情卡片时打开 panel menu；如果详情卡片已经打开，它负责收起详情卡片。
- Run rail 内部内容点击只选择对应的 right tool card，例如 context connector 打开 Preview。
- 面板内部不放冗余关闭按钮；关闭动作保留在顶栏入口，避免用户在多个层级寻找同一种控制。

## 为什么不用全宽 drawer

全宽 drawer 适合独立的 artifact/runtime 工作区，但 M1 还没有真实 browser、terminal 或 artifact runtime。当前阶段的目标是展示桌面壳语言和未来入口，而不是提前占用完整右侧大画布。

因此 M1 只保留窄卡片式 preview placeholder。后续如果真实 artifact runtime 需要更大空间，应先在 spec 中定义它和主画布、right tool card 的关系，再引入新的 surface。
