---
title: M89 Unified Conversation Entry
description: 合并 Chat/Work 可见入口，让目录和计划能力按上下文浮现。
---

M89 将前端入口从 Chat/Work 双模式切换收敛成一个会话入口。Arkloop 的可学习点是分层：外层是统一会话，深层能力由 work folder、run metadata、collaboration/tool 状态决定；Loomi 这轮只做自己的最小前端收口。

完成内容：

- Sidebar 移除 Chat/Work 双按钮，线程列表合并显示。
- 新建入口统一为“新会话 / New thread”，空状态不再写 Chat/Work。
- Composer 始终可显示目录选择和目录状态，不再要求用户先进入 Work mode。
- `deriveWorkPlanProjection` 改为按安全 run metadata 投影计划、todo 和 artifact，不再只接受 `Thread.mode = work`。
- 删除未使用的 mode menu/thread filter helper 和测试。
- 修复删光会话后的空列表回归：`GET /v1/threads` 在没有 active thread 时返回 `threads: []`，不返回 `null`，前端显示正常空状态。

保留边界：

- 后端 `Thread.mode`、工具权限、provider/run 链路保持不变。
- 不新增工具、不改数据库、不引入 Docker/Redis/多 agent。
- 不复制 Arkloop 文案、品牌或视觉表达。

验证：

- `bun test --cwd web App.threadModes.test.ts components/ThreadSidebar.actions.test.ts components/ThreadSidebar.layout.test.ts components/Composer.test.ts workModeProjection.test.ts components/WorkPlanView.test.tsx animalIslandUi.test.ts`
- `go test ./internal/httpapi -run TestThreadListReturnsEmptyArrayAfterAllThreadsArchived -count=1`
