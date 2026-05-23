---
title: 2026-05-23 M3 Auth、Thread 与 Message
description: M3 本地身份、thread/message 数据层、seed、readiness 和前端 real/mock 切换的实现记录。
---

## 完成范围

- 新增固定本地开发身份 `user_local_dev`。
- 新增 M3 `users`、`threads`、`messages` migration。
- 新增 thread/message API handler、`/v1/*` CORS preflight 和结构化错误 envelope。
- 新增 message `client_message_id` 幂等行为。
- 新增 M3 schema readiness：version 必须至少为 `2` 且 clean。
- 新增显式 seed command：`go run ./cmd/loomi-seed`，使用固定 `thr_local_demo` / `msg_local_demo_001`。
- 新增前端 API seam：未配置 backend 时走 mock，配置 `VITE_LOOMI_API_BASE_URL` 时走真实 API。
- 修复 mock `New Chat` 的运行状态边界：新 thread 会同步创建可读取的 idle run，首条消息会持久化 run event，避免 `Run not found` 把画布清空。
- 新增 Run rail 顶部的 agent state motion 徽章，把 `run.status` 和最新 event type 映射成紧凑运动状态，作为未来 M4 run/event/SSE 的 UI 接口占位。

## 验证记录

- `go test ./...`：通过（2026-05-23 15:49，本地执行）。
- `bun test ./web/src/*.test.ts ./web/src/components/*.test.ts`：通过（2026-05-23 15:49，33 tests / 58 assertions）。
- `bun run --cwd web build`：通过（2026-05-23 15:49；当时仍提示 chunk size warning）。16:07 拆分 React/UI/icons/motion vendor chunks 后再次通过，最大 JS chunk 为 `ui-vendor` 约 376KB，入口 app chunk 约 27KB，chunk-size warning 消失。
- migration rollback/reapply smoke：未执行；本机未检测到 `migrate` CLI。使用 Docker `psql` 手动执行 `000002_m3_auth_thread_message.up.sql` 后，thread/message API 和 seed smoke 通过；readiness 仍正确依赖 `schema_migrations` version 2。
- API smoke：M2-only schema 下 `/readyz` 返回 503；手动建 M3 表后 `/v1/me`、thread create/list、message create、重复 `client_message_id` 返回同一 message id。Review 修复补充了 loopback-only `/v1/*` CORS preflight、405 method errors、unknown JSON field 拒绝、empty PATCH 拒绝和 DB pool unavailable 的结构化错误。
- seed smoke：`go run ./cmd/loomi-seed` 连续运行两次后，`seed-m3-local-demo-message` 只有 1 条；review 修复后 seed 使用固定 `thr_local_demo` / `msg_local_demo_001`，不再按 title 扫描。
- browser mock/real smoke：Playwright MCP 仍被现有 browser profile 锁定，无法交互点击；CORS 修复后已用 loopback `OPTIONS /v1/threads` 验证 preflight 204，并用 18080 临时 API 端口跑通 create/rename/message duplicate/archive 的 real API fallback smoke。
- `bun run --cwd docs-site build`：通过（2026-05-23 15:49，20 pages built）。

## UI 修复记录

截图里的重复 `New thread`、archive 图标挤位和 `Run not found` 属于同一类前端 mock 状态问题：创建 thread 只插入 thread list，没有同步插入 run；thread row 只有两列却渲染了状态点、标题和 archive 三个元素；创建标题也始终使用固定字符串。现在 mock thread 使用递增标题，thread row 有独立 archive 列，Run rail 可以从新 thread 的 idle run 开始展示状态。

Agent state motion 徽章必须复用 `loomi-agent-states.html` 里的真实 Loomi 刺猬资产，而不是用 CSS 临时画一个近似图形。左侧栏保留 `New Chat` 作为唯一创建入口，Threads 标题旁不再显示重复的 `+` 按钮。Chat 和 Work 的最近会话按 mode 分开显示，避免 Chat 视图混入 Work 任务。

## 已知限制

M3 仍不包含 run/event/SSE、LLM gateway、assistant message generation、tool calling、worker/job queue、desktop runtime、attachments、RAG、catalog/plugin runtime 或 production auth。

## 下一步

完成最终验证后，M4 可以在 thread/message 数据层之上引入 run/event/SSE 和可观测执行流。
