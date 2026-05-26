---
title: M10 Workspace Exec Command 本地验证
description: 本地验证 approval-gated workspace exec command。
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

1. 触发 `workspace.exec_command` 的受控 tool request，例如 `["printf", "hello"]`。
2. 确认 ToolCallCard 显示 approval-required。
3. Deny 一条请求，确认没有执行。
4. Approve 一条请求，确认只执行一次并显示 exit code、stdout/stderr、timeout 和 truncation flags。
5. 刷新页面或重连 SSE，确认历史 replay 顺序不变。
6. 打开浏览器 console，确认没有 error。

## 必须拒绝的输入

```text
cwd: ../
command: ["sh", "-c", "echo no"]
command: ["bash", "-c", "echo no"]
command: ["rm", "-rf", "internal"]
command: ["dd", "if=/dev/zero", "of=file"]
command: ["git", "push"]
command: ["git", "reset", "--hard"]
```

M10 不支持 shell、PTY、持久终端、后台进程管理、MCP、浏览器自动化、长期授权或外部上传。
