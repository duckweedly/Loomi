---
title: Local M16 Work Mode Smoke
description: 本地验证 Work mode foundation 的测试和浏览器 smoke。
---

## Commands

```sh
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Go backend 没有改动时不需要针对 M16 追加后端 smoke；如果后续新增后端 payload 或 API，再跑：

```sh
go test ./...
```

## Browser Smoke

启动 web：

```sh
bun run --cwd web dev -- --port 5180 --strictPort
```

打开 `http://127.0.0.1:5180`：

1. 选择 Work mode thread。
2. 验证主区域出现 Work Plan View。
3. 验证 goal、steps、status、artifact references、recent progress 可见。
4. 验证 Work mode composer 是 disabled/read-only，不会启动新的 run。
5. 切到 Chat mode thread。
6. 验证 Chat mode 不显示 Work Plan View，消息和 composer 正常。
7. DevTools console 无 error。

## Expected Boundaries

- Artifact cards are safe metadata previews only.
- No shell/filesystem/browser tool controls appear.
- Progress comes from existing run events/messages.
- Chat mode remains unchanged.
