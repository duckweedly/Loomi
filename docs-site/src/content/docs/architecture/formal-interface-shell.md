---
title: Formal Interface Shell
description: UI-01 的浅色桌面壳、窄侧栏、白色聊天画布和固定 composer 边界。
---

UI-01 把 Loomi 的主界面收敛成浅色 desktop window：外层是柔和蓝紫背景，内部是带圆角、轻边框和阴影的主窗口。参考图只用于布局机制和视觉比例，不用于复制品牌、图标、文案或私有表达。

UI-02 在 UI-01 壳上做真实使用态收口：不改 runtime、provider、tool、DB 或后端能力，只让已有状态对真实用户更诚实、更可操作。

## Shell 边界

主窗口由三层组成：

- 左侧窄 sidebar：工作区身份、Settings 入口、Chat/Work 模式入口、当前模式的 thread 列表和轻量 thread actions。
- 右侧主画布：白色聊天 surface、极简顶栏、状态提示、消息内容和 Work projection。
- 浮动/次级面板：Settings、Tools、RunRail 沿用既有结构，只接受新 shell 的外层 containment。

这轮不新增 backend、runtime、tool、provider、memory、database 或 activity recorder 能力。

## Sidebar 规则

Sidebar 保持 native desktop sidebar 的轻量感：宽度收窄，列表滚动在内部完成，thread title 必须截断而不是挤压右侧画布。Chat/Work 模式入口可以改变当前 thread mode，但不能混合两个模式的 thread 列表。

Settings 放在左上工作区区块旁边，作为常驻 gear 入口；它打开既有 Settings/Theme/Update 菜单，不隐藏在底部操作区。

Sidebar 不重复放置底部新建/搜索入口，也不在 thread 列表顶部显示搜索框；新会话入口只在 sidebar 收起时出现在顶栏 chrome，使用当前 Chat/Work mode 创建对应 thread。检索入口后续需要重新放回时必须先有清晰位置和可用行为。Run 状态点保留轻量视觉，但必须带可读状态 label，方便验收和辅助技术读取。

Electron 顶栏图标必须和 macOS traffic light 的中心线对齐，避免左侧 collapse/search 和右侧工具按钮低于三色灯。

## Canvas 规则

聊天画布使用大面积白色 surface。消息、Work plan、provider warning 和 composer 都对齐同一个中心内容列，最大宽度受控，避免在宽屏上铺满。

运行状态仍然优先可见：provider unavailable、active run Stop 和 tool approval notice 不能被固定 composer 遮挡。

Chat 画布不显示旧 runtime/context 调试顶栏。Real API provider 不可用时显示 Provider Settings 入口；mock/real/provider 细节保留在 Settings 和事件视图里，不压在主聊天顶部。

Approval-blocked run 会在 canvas 顶部显示等待用户确认，同时保留 Stop；具体 Approve/Deny 仍由既有 ToolCallCard action 承载。

## Composer 规则

Composer 固定在画布底部居中，保留现有真实入口：

- send / continue
- Stop
- retry / regenerate

UI-02 不再把 Work in Folder、附件、persona/provider selector 或 voice 画成不能用的假入口。Chat placeholder 用发消息语义，Work placeholder 用描述任务语义。

不可新增没有行为支撑的 runtime/tool/provider 按钮。

## Work Plan 规则

WorkPlanView 是用户任务面板，不是 raw event dump。它展示目标、状态、计划/步骤、todos、产物和最近进展。计划/步骤只来自明确的 work metadata；没有 plan/todo/artifact metadata 时，不渲染 WorkPlanView，也不能从普通用户消息或工具事件冒充计划步骤。Recent progress 只能作为已有任务面板的附属信息显示。

## Tool Event 规则

RunRail 和 ToolCallCard 默认先显示人话标签：

- `workspace.read`：Read project files
- `web.fetch`：Visit web page
- `lsp.symbols`：Analyze code
- `artifact.*`：Handle artifact
- `agent.*`：Coordinate subtasks

Tool card 和 Progress rail 只展示安全摘要：搜索词、服务、数量、标题、链接、结果数等用户能理解的字段。`tool_call_id`、`approval_status`、raw tool name、pipeline stage、model delta、persona id、长 run/message id 等 runtime 字段不进入主界面；需要审计时应走后续专门的调试/导出入口，而不是挤在聊天界面里。

Progress rail 只展示当前 thread 当前 run 的关键活动：工具等待确认/执行/完成、run terminal、用户需要处理的错误。pipeline 准备阶段、model streaming delta、worker lease、MCP discovery 等内部流水默认隐藏，避免用户误以为右侧是其它会话的历史。

## 响应式底线

窄屏本轮只要求不明显重叠或溢出。详细移动端交互、抽屉式 sidebar、像素级密度和图标精修留到后续迭代。
