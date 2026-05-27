---
title: Local M17 Work Artifact Smoke
description: 本地验证 Work artifact evidence closeout 的 seed、真实 API 和浏览器 smoke。
---

## Commands

```sh
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Seed Evidence

启动本地 API 依赖后运行：

```sh
LOOMI_SEED_SCENARIO=m17-work-artifact go run ./cmd/loomi-seed
```

记录输出里的：

- `thread_id`，预期 `thr_m17_work_artifact`
- `message_id`
- `run_id`
- `event_id`

这个 seed 只用于 local-dev/test evidence，不是生产写接口。

## Browser Smoke

1. 启动 local API。
2. 用 real API mode 启动 web。
3. 打开 Work mode thread `thr_m17_work_artifact`。
4. 验证 Work Plan View 显示 M17 goal、steps、status、artifact references 和 recent progress。
5. 验证 artifact card 显示 redaction marker，但没有 execute/open/run/download 按钮。
6. 验证 command/path/file/shell/browser/filesystem/execute/url 等 unsafe metadata 没有成为 action。
7. 切到 Chat mode thread。
8. 验证 Chat mode 不显示 Work Plan View。
9. 记录 API/web 端口、seed ids、截图路径和 console error 状态。

## Expected Boundaries

- Work 仍复用 `Thread.mode = work`、messages、current run、run events、ChatCanvas 和 RunRail。
- Artifact evidence 是 metadata-only。
- 不做 artifact execution/runtime。
- 不新增 task system 或 worker queue rewrite。
