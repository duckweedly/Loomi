---
title: M8 Workspace Read Tools 本地验证
description: 本地验证 approval-gated workspace read tools。
---

## 自动化验证

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```

## 手动 Smoke

1. 启动本地 API 和 web app。
2. 触发 `workspace.glob`、`workspace.grep` 或 `workspace.read_file` 的受控 tool request。
3. 确认 ToolCallCard 显示 approval-required。
4. Approve 一次，确认只执行一次并显示 succeeded 或 failed。
5. Deny 另一条，确认没有执行。
6. 刷新页面或重连 SSE，确认历史 replay 顺序不变。
7. 打开浏览器 console，确认没有 error。

## 安全用例

这些输入必须失败，且不能泄漏文件内容：

```text
../outside.txt
/Users/example/.ssh/id_ed25519
.env
.ssh/id_ed25519
secrets/token.txt
credentials/prod.json
key.pem
```

M8 不支持写文件、shell、网络、MCP、浏览器自动化、长期授权或外部上传。
